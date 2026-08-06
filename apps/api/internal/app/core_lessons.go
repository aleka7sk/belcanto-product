package app

import (
	"context"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.2 rooms and core lesson series (DEC-002/DEC-004). Recurrence is
// weekly and anchored in the school's civil time zone: storage stays
// UTC, the weekday and start minutes are Asia/Almaty wall-clock values.

const schoolTimeZone = "Asia/Almaty"

const maxGenerationWeeks = 12

type CreateRoomInput struct {
	Name     string
	Capacity *int
}

func (s *Service) CreateRoom(ctx context.Context, principal core.Principal, input CreateRoomInput) (core.Room, error) {
	name, err := security.ValidateText("name", input.Name, 1, 200)
	if err != nil {
		return core.Room{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.Capacity != nil && (*input.Capacity < 1 || *input.Capacity > 500) {
		return core.Room{}, core.E(core.CodeInvalidInput, "capacity must be between 1 and 500", nil)
	}
	roomID, err := security.NewID("room")
	if err != nil {
		return core.Room{}, core.E(core.CodeInternal, "could not create the room", err)
	}
	room, err := s.store.CreateRoom(ctx, core.CreateRoomCommand{
		Principal: principal, RoomID: roomID, Name: name,
		Capacity: input.Capacity, Now: s.clock.Now(),
	})
	if err != nil {
		return core.Room{}, normalizeStoreError("create room", err)
	}
	return room, nil
}

func (s *Service) ListRooms(ctx context.Context, principal core.Principal) ([]core.Room, error) {
	rooms, err := s.store.ListRooms(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list rooms", err)
	}
	return rooms, nil
}

type CreateCoreLessonSeriesInput struct {
	Format           string
	Title            string
	TeacherAccountID string
	RoomID           string
	Weekday          int
	StartMinutes     int
	DurationMinutes  int
	EffectiveFrom    string
	EffectiveUntil   string
	StudentIDs       []string
	IdempotencyKey   string
}

func (s *Service) CreateCoreLessonSeries(ctx context.Context, principal core.Principal, input CreateCoreLessonSeriesInput) (core.CoreLessonSeries, error) {
	if input.Format != "individual" && input.Format != "group" {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "format must be individual or group", nil)
	}
	title, err := security.ValidateText("title", input.Title, 1, 200)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	teacherID, err := security.ValidateIdentifier("teacherAccountId", input.TeacherAccountID, 128)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	roomID := ""
	if input.RoomID != "" {
		roomID, err = security.ValidateIdentifier("roomId", input.RoomID, 128)
		if err != nil {
			return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.Weekday < 0 || input.Weekday > 6 {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "weekday must be between 0 (Monday) and 6 (Sunday)", nil)
	}
	if input.StartMinutes < 0 || input.StartMinutes > 1439 {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "startMinutes must be between 0 and 1439", nil)
	}
	if input.DurationMinutes < 1 || input.DurationMinutes > 1440 {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "durationMinutes must be between 1 and 1440", nil)
	}
	zone, err := time.LoadLocation(schoolTimeZone)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInternal, "school time zone is unavailable", err)
	}
	effectiveFrom, err := time.ParseInLocation("2006-01-02", input.EffectiveFrom, zone)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "effectiveFrom must be a YYYY-MM-DD date", nil)
	}
	if input.EffectiveUntil != "" {
		effectiveUntil, parseErr := time.ParseInLocation("2006-01-02", input.EffectiveUntil, zone)
		if parseErr != nil {
			return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "effectiveUntil must be a YYYY-MM-DD date", nil)
		}
		if effectiveUntil.Before(effectiveFrom) {
			return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "effectiveUntil must not precede effectiveFrom", nil)
		}
	}
	studentIDs, err := validateUniqueIdentifiers("student id", input.StudentIDs)
	if err != nil {
		return core.CoreLessonSeries{}, err
	}
	if input.Format == "individual" && len(studentIDs) != 1 {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "an individual series holds exactly one Student (DEC-002)", nil)
	}
	if input.Format == "group" && (len(studentIDs) < 1 || len(studentIDs) > 3) {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "a group series holds one to three Students (DEC-002)", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	seriesID, err := security.NewID("clser")
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInternal, "could not create the series", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		Format, Title, Teacher, Room    string
		Weekday, StartMinutes, Duration int
		EffectiveFrom, EffectiveUntil   string
		StudentIDs                      []string
	}{input.Format, title, teacherID, roomID, input.Weekday, input.StartMinutes,
		input.DurationMinutes, input.EffectiveFrom, input.EffectiveUntil, studentIDs})
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	series, err := s.store.CreateCoreLessonSeries(ctx, core.CreateCoreLessonSeriesCommand{
		TenantID: principal.TenantID, ActorAccountID: principal.AccountID,
		SeriesID: seriesID, Format: input.Format, Title: title,
		TeacherAccountID: teacherID, RoomID: roomID,
		Weekday: input.Weekday, StartMinutes: input.StartMinutes,
		DurationMinutes: input.DurationMinutes,
		EffectiveFrom:   input.EffectiveFrom, EffectiveUntil: input.EffectiveUntil,
		StudentIDs: studentIDs, IdempotencyKey: idempotencyKey,
		PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.CoreLessonSeries{}, normalizeStoreError("create lesson series", err)
	}
	return series, nil
}

func (s *Service) ListCoreLessonSeries(ctx context.Context, principal core.Principal) ([]core.CoreLessonSeries, error) {
	series, err := s.store.ListCoreLessonSeries(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list lesson series", err)
	}
	return series, nil
}

func (s *Service) GetCoreLessonSeries(ctx context.Context, principal core.Principal, seriesID string) (core.CoreLessonSeries, error) {
	normalizedID, err := security.ValidateIdentifier("seriesId", seriesID, 128)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	series, err := s.store.GetCoreLessonSeries(ctx, principal, normalizedID)
	if err != nil {
		return core.CoreLessonSeries{}, normalizeStoreError("read lesson series", err)
	}
	return series, nil
}

// GenerateSeriesOccurrences materializes the weekly recurrence into
// concrete occurrences up to the requested horizon. Generation is
// idempotent: already-materialized start times are skipped, and the same
// idempotency key replays the same result.
func (s *Service) GenerateSeriesOccurrences(ctx context.Context, principal core.Principal, seriesID string, weeks int, idempotencyKey string) (core.SeriesOccurrenceGenerationResult, error) {
	normalizedID, err := security.ValidateIdentifier("seriesId", seriesID, 128)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if weeks < 1 || weeks > maxGenerationWeeks {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidInput, "weeks must be between 1 and 12", nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	series, err := s.store.GetCoreLessonSeries(ctx, principal, normalizedID)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, normalizeStoreError("read lesson series", err)
	}
	now := s.clock.Now()
	planned, err := planWeeklyOccurrences(series, now, weeks)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, err
	}
	fingerprint, err := security.Fingerprint(struct {
		SeriesID string
		Weeks    int
	}{normalizedID, weeks})
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	result, err := s.store.GenerateSeriesOccurrences(ctx, core.GenerateSeriesOccurrencesCommand{
		TenantID: principal.TenantID, ActorAccountID: principal.AccountID,
		SeriesID: normalizedID, Occurrences: planned,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint, Now: now,
	})
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, normalizeStoreError("generate occurrences", err)
	}
	return result, nil
}

type ChangeSeriesStatusInput struct {
	SeriesID        string
	Status          string
	ExpectedVersion int64
	IdempotencyKey  string
}

// ChangeCoreLessonSeriesStatus pauses, resumes or ends a weekly series.
// The change gates future occurrence generation only; already-scheduled
// Lessons stay and are changed through the explicit Lesson operations.
func (s *Service) ChangeCoreLessonSeriesStatus(ctx context.Context, principal core.Principal, input ChangeSeriesStatusInput) (core.CoreLessonSeries, error) {
	normalizedID, err := security.ValidateIdentifier("seriesId", input.SeriesID, 128)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	switch input.Status {
	case "active", "paused", "ended":
	default:
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "status must be active, paused or ended", nil)
	}
	if input.ExpectedVersion < 0 {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, "expectedVersion must be at least 0", nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		SeriesID, Status string
		ExpectedVersion  int64
	}{normalizedID, input.Status, input.ExpectedVersion})
	if err != nil {
		return core.CoreLessonSeries{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	series, err := s.store.ChangeCoreLessonSeriesStatus(ctx, core.ChangeCoreLessonSeriesStatusCommand{
		TenantID: principal.TenantID, ActorAccountID: principal.AccountID,
		SeriesID: normalizedID, Status: input.Status, ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.CoreLessonSeries{}, normalizeStoreError("change series status", err)
	}
	return series, nil
}

func planWeeklyOccurrences(series core.CoreLessonSeries, now time.Time, weeks int) ([]core.PlannedOccurrence, error) {
	zone, err := time.LoadLocation(schoolTimeZone)
	if err != nil {
		return nil, core.E(core.CodeInternal, "school time zone is unavailable", err)
	}
	effectiveFrom, err := time.ParseInLocation("2006-01-02", series.EffectiveFrom, zone)
	if err != nil {
		return nil, core.E(core.CodeInternal, "series effectiveFrom is invalid", err)
	}
	var effectiveUntil *time.Time
	if series.EffectiveUntil != "" {
		parsed, parseErr := time.ParseInLocation("2006-01-02", series.EffectiveUntil, zone)
		if parseErr != nil {
			return nil, core.E(core.CodeInternal, "series effectiveUntil is invalid", parseErr)
		}
		endOfDay := parsed.AddDate(0, 0, 1)
		effectiveUntil = &endOfDay
	}
	localNow := now.In(zone)
	cursor := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, zone)
	if effectiveFrom.After(cursor) {
		cursor = effectiveFrom
	}
	// Weekday contract: 0 = Monday … 6 = Sunday.
	for isoWeekday(cursor.Weekday()) != series.Weekday {
		cursor = cursor.AddDate(0, 0, 1)
	}
	horizon := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, zone).
		AddDate(0, 0, weeks*7)
	planned := []core.PlannedOccurrence{}
	for !cursor.After(horizon) {
		if effectiveUntil != nil && !cursor.Before(*effectiveUntil) {
			break
		}
		startsAt := cursor.Add(time.Duration(series.StartMinutes) * time.Minute)
		if startsAt.After(now) {
			occurrenceID, idErr := security.NewID("cocc")
			if idErr != nil {
				return nil, core.E(core.CodeInternal, "could not create occurrence ids", idErr)
			}
			planned = append(planned, core.PlannedOccurrence{
				OccurrenceID: occurrenceID,
				StartsAt:     startsAt.UTC(),
			})
		}
		cursor = cursor.AddDate(0, 0, 7)
	}
	return planned, nil
}

func isoWeekday(weekday time.Weekday) int {
	return (int(weekday) + 6) % 7
}
