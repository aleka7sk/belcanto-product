package memory

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 rooms and core lesson series — parity with PostgreSQL.

type room struct {
	ID       string
	TenantID string
	Name     string
	Capacity *int
	Status   string
	Version  int64
}

type coreLessonSeries struct {
	ID               string
	TenantID         string
	Format           string
	Title            string
	TeacherAccountID string
	RoomID           string
	Weekday          int
	StartMinutes     int
	DurationMinutes  int
	EffectiveFrom    string
	EffectiveUntil   string
	Status           string
	Version          int64
	StudentIDs       []string
	CreatedAt        time.Time
}

func (s *Store) CreateRoom(_ context.Context, command core.CreateRoomCommand) (core.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	if !manager {
		return core.Room{}, core.E(core.CodeForbidden, "room management permission is required", nil)
	}
	for _, existing := range s.rooms {
		if existing.TenantID == principal.TenantID && existing.Name == command.Name {
			return core.Room{}, core.E(core.CodeConflict, "room name is already in use", nil)
		}
	}
	stored := &room{
		ID: command.RoomID, TenantID: principal.TenantID, Name: command.Name,
		Capacity: command.Capacity, Status: "active", Version: 0,
	}
	s.rooms[stored.ID] = stored
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "RoomCreated",
		"room", stored.ID, "allow", "", command.Now, nil)
	return roomView(stored), nil
}

func roomView(stored *room) core.Room {
	view := core.Room{ID: stored.ID, Name: stored.Name, Status: stored.Status, Version: stored.Version}
	if stored.Capacity != nil {
		capacity := *stored.Capacity
		view.Capacity = &capacity
	}
	return view
}

func (s *Store) ListRooms(_ context.Context, principal core.Principal) ([]core.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.Room{}
	for _, stored := range s.rooms {
		if stored.TenantID == principal.TenantID && stored.Status == "active" {
			result = append(result, roomView(stored))
		}
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Name != result[right].Name {
			return result[left].Name < result[right].Name
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func (s *Store) CreateCoreLessonSeries(_ context.Context, command core.CreateCoreLessonSeriesCommand) (core.CoreLessonSeries, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(command.ActorAccountID, command.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	teacher := actor != nil && actor.Roles[core.RoleTeacher] != ""
	if !manager && !teacher {
		return core.CoreLessonSeries{}, core.E(core.CodeForbidden, "Lesson scheduling permission is required", nil)
	}
	if teacher && !manager && command.TeacherAccountID != command.ActorAccountID {
		return core.CoreLessonSeries{}, core.E(core.CodeForbidden, "Teacher can only create series for self", nil)
	}
	if response, ok, err := s.replay("create_lesson_series", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.CoreLessonSeries{}, err
		}
		var result core.CoreLessonSeries
		if err := json.Unmarshal(response, &result); err != nil {
			return core.CoreLessonSeries{}, core.E(core.CodeInternal, "decode idempotent series result", err)
		}
		return result, nil
	}
	teacherRecord := s.activeAccount(command.TeacherAccountID, command.TenantID)
	if teacherRecord == nil || teacherRecord.Roles[core.RoleTeacher] == "" {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "Teacher is not active in this school", nil)
	}
	if command.RoomID != "" {
		stored := s.rooms[command.RoomID]
		if stored == nil || stored.TenantID != command.TenantID || stored.Status != "active" {
			return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "room is not active in this school", nil)
		}
	}
	studentIDs := append([]string(nil), command.StudentIDs...)
	sort.Strings(studentIDs)
	for _, studentID := range studentIDs {
		studentRecord := s.students[studentID]
		if studentRecord == nil || studentRecord.TenantID != command.TenantID || !s.activeStudent(studentRecord) {
			return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "Student is not active in this school", nil)
		}
	}
	stored := &coreLessonSeries{
		ID: command.SeriesID, TenantID: command.TenantID, Format: command.Format,
		Title: command.Title, TeacherAccountID: command.TeacherAccountID, RoomID: command.RoomID,
		Weekday: command.Weekday, StartMinutes: command.StartMinutes,
		DurationMinutes: command.DurationMinutes,
		EffectiveFrom:   command.EffectiveFrom, EffectiveUntil: command.EffectiveUntil,
		Status: "active", Version: 0, StudentIDs: studentIDs, CreatedAt: command.Now,
	}
	s.lessonSeries[stored.ID] = stored
	result := s.seriesView(stored)
	if err := s.completeIdempotency("create_lesson_series", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.CoreLessonSeries{}, err
	}
	s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "LessonSeriesCreated", stored.ID, "allow", "", command.Now, map[string]any{
		"teacherAccountId": stored.TeacherAccountID, "format": stored.Format,
	})
	s.appendOutbox(command.TenantID, "LessonSeriesCreated", stored.ID, command.Now)
	return result, nil
}

func (s *Store) seriesView(stored *coreLessonSeries) core.CoreLessonSeries {
	view := core.CoreLessonSeries{
		ID: stored.ID, Format: stored.Format, Title: stored.Title,
		Teacher: core.TeacherSummary{AccountID: stored.TeacherAccountID},
		RoomID:  stored.RoomID, Weekday: stored.Weekday, StartMinutes: stored.StartMinutes,
		DurationMinutes: stored.DurationMinutes, EffectiveFrom: stored.EffectiveFrom,
		EffectiveUntil: stored.EffectiveUntil, Status: stored.Status, Version: stored.Version,
		Students: make([]core.LessonStudent, 0, len(stored.StudentIDs)),
	}
	if teacherAccount := s.accounts[stored.TeacherAccountID]; teacherAccount != nil {
		view.Teacher.FullName = teacherAccount.FullName
	}
	for _, studentID := range stored.StudentIDs {
		student := core.LessonStudent{StudentID: studentID}
		if record := s.students[studentID]; record != nil {
			student.FullName = record.FullName
		}
		view.Students = append(view.Students, student)
	}
	sort.Slice(view.Students, func(left, right int) bool {
		if view.Students[left].FullName != view.Students[right].FullName {
			return view.Students[left].FullName < view.Students[right].FullName
		}
		return view.Students[left].StudentID < view.Students[right].StudentID
	})
	return view
}

func (s *Store) seriesVisible(principal core.Principal, stored *coreLessonSeries) bool {
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	if actor == nil {
		return false
	}
	if actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "" {
		return true
	}
	if actor.Roles[core.RoleTeacher] != "" && stored.TeacherAccountID == principal.AccountID {
		return true
	}
	if studentID := s.studentIDForAccount(principal.AccountID); studentID != "" {
		for _, enrolled := range stored.StudentIDs {
			if enrolled == studentID {
				return true
			}
		}
	}
	return false
}

func (s *Store) studentIDForAccount(accountID string) string {
	for studentID, record := range s.students {
		if record.AccountID == accountID {
			return studentID
		}
	}
	return ""
}

func (s *Store) ListCoreLessonSeries(_ context.Context, principal core.Principal) ([]core.CoreLessonSeries, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.CoreLessonSeries{}
	for _, stored := range s.lessonSeries {
		if stored.TenantID != principal.TenantID || !s.seriesVisible(principal, stored) {
			continue
		}
		result = append(result, s.seriesView(stored))
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Weekday != result[right].Weekday {
			return result[left].Weekday < result[right].Weekday
		}
		if result[left].StartMinutes != result[right].StartMinutes {
			return result[left].StartMinutes < result[right].StartMinutes
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func (s *Store) GetCoreLessonSeries(_ context.Context, principal core.Principal, seriesID string) (core.CoreLessonSeries, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.lessonSeries[seriesID]
	if stored == nil || stored.TenantID != principal.TenantID || !s.seriesVisible(principal, stored) {
		return core.CoreLessonSeries{}, core.E(core.CodeNotFound, "lesson series not found", nil)
	}
	return s.seriesView(stored), nil
}

func (s *Store) GenerateSeriesOccurrences(_ context.Context, command core.GenerateSeriesOccurrencesCommand) (core.SeriesOccurrenceGenerationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(command.ActorAccountID, command.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	teacher := actor != nil && actor.Roles[core.RoleTeacher] != ""
	if !manager && !teacher {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeForbidden, "Lesson scheduling permission is required", nil)
	}
	if response, ok, err := s.replay("generate_series_occurrences", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.SeriesOccurrenceGenerationResult{}, err
		}
		var result core.SeriesOccurrenceGenerationResult
		if err := json.Unmarshal(response, &result); err != nil {
			return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInternal, "decode idempotent generation result", err)
		}
		return result, nil
	}
	stored := s.lessonSeries[command.SeriesID]
	if stored == nil || stored.TenantID != command.TenantID {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeNotFound, "lesson series not found", nil)
	}
	if stored.Status != "active" {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidState, "only an active series can generate occurrences", nil)
	}
	if teacher && !manager && stored.TeacherAccountID != command.ActorAccountID {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeForbidden, "Teacher can only generate for own series", nil)
	}
	created := []string{}
	for _, planned := range command.Occurrences {
		duplicate := false
		for _, existing := range s.lessons {
			if existing.TenantID == command.TenantID && existing.SeriesID == command.SeriesID &&
				existing.StartsAt.Equal(planned.StartsAt) {
				duplicate = true
			}
		}
		if duplicate {
			continue
		}
		if s.lessonScheduleConflict(command.TenantID, planned.StartsAt, stored.DurationMinutes, stored.TeacherAccountID, stored.StudentIDs, nil) {
			return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeConflict, "generated occurrence overlaps an existing Lesson", nil)
		}
		s.lessons[planned.OccurrenceID] = &lesson{
			ID: planned.OccurrenceID, TenantID: command.TenantID, SeriesID: command.SeriesID,
			Title: stored.Title, StartsAt: planned.StartsAt, DurationMinutes: stored.DurationMinutes,
			TeacherAccountID: stored.TeacherAccountID,
			StudentIDs:       append([]string(nil), stored.StudentIDs...),
			Status:           core.LessonScheduled, Version: 0,
			CreatedBy: command.ActorAccountID, CreatedAt: command.Now, UpdatedAt: command.Now,
		}
		created = append(created, planned.OccurrenceID)
	}
	result := core.SeriesOccurrenceGenerationResult{
		SeriesID: command.SeriesID, CreatedCount: len(created), OccurrenceIDs: created,
	}
	if err := s.completeIdempotency("generate_series_occurrences", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.SeriesOccurrenceGenerationResult{}, err
	}
	s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "SeriesOccurrencesGenerated", command.SeriesID, "allow", "", command.Now, map[string]any{
		"createdCount": len(created),
	})
	s.appendOutbox(command.TenantID, "SeriesOccurrencesGenerated", command.SeriesID, command.Now)
	return result, nil
}
