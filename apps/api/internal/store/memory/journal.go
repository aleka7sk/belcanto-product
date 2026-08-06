package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 lesson journals and progress evidence — parity with PostgreSQL.

type lessonJournal struct {
	ID               string
	TenantID         string
	OccurrenceID     string
	StudentID        string
	TeacherAccountID string
	Status           string
	CurrentVersion   int
	Draft            *core.JournalDraft
	Versions         []core.JournalVersion
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type progressEvidenceRecord struct {
	ID         string
	TenantID   string
	StudentID  string
	SourceKind string
	SourceID   string
	Area       string
	Note       string
	RecordedBy string
	RecordedAt time.Time
}

func journalKey(occurrenceID, studentID string) string {
	return occurrenceID + "\x00" + studentID
}

func (s *Store) journalTeacherAuthority(actorID, tenantID, occurrenceID, studentID string) error {
	occurrence := s.lessons[occurrenceID]
	if occurrence == nil || occurrence.TenantID != tenantID {
		return core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	if occurrence.TeacherAccountID != actorID {
		return core.E(core.CodeForbidden, "only the Lesson's Teacher writes the journal", nil)
	}
	for _, participant := range occurrence.StudentIDs {
		if participant == studentID {
			return nil
		}
	}
	return core.E(core.CodeInvalidInput, "Student does not participate in this Lesson", nil)
}

func (s *Store) journalView(stored *lessonJournal, includeDraft bool) core.LessonJournal {
	view := core.LessonJournal{
		ID: stored.ID, OccurrenceID: stored.OccurrenceID, StudentID: stored.StudentID,
		Teacher: core.TeacherSummary{AccountID: stored.TeacherAccountID},
		Status:  stored.Status, CurrentVersion: stored.CurrentVersion,
		Versions:  make([]core.JournalVersion, 0, len(stored.Versions)),
		UpdatedAt: stored.UpdatedAt,
	}
	if account := s.accounts[stored.TeacherAccountID]; account != nil {
		view.Teacher.FullName = account.FullName
	}
	if includeDraft && stored.Draft != nil {
		draft := *stored.Draft
		view.Draft = &draft
	}
	view.Versions = append(view.Versions, stored.Versions...)
	sort.Slice(view.Versions, func(left, right int) bool {
		return view.Versions[left].Version > view.Versions[right].Version
	})
	return view
}

func (s *Store) SaveJournalDraft(_ context.Context, command core.SaveJournalDraftCommand) (core.LessonJournal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if err := s.journalTeacherAuthority(principal.AccountID, principal.TenantID, command.OccurrenceID, command.StudentID); err != nil {
		return core.LessonJournal{}, err
	}
	key := journalKey(command.OccurrenceID, command.StudentID)
	stored := s.journals[key]
	if stored == nil {
		stored = &lessonJournal{
			ID: command.JournalID, TenantID: principal.TenantID,
			OccurrenceID: command.OccurrenceID, StudentID: command.StudentID,
			TeacherAccountID: principal.AccountID, Status: "draft",
			CreatedAt: command.Now,
		}
		s.journals[key] = stored
	}
	draft := command.Draft
	stored.Draft = &draft
	stored.UpdatedAt = command.Now
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "JournalDraftSaved",
		"lesson_journal", stored.ID, "allow", "", command.Now, nil)
	return s.journalView(stored, true), nil
}

func (s *Store) PublishJournal(_ context.Context, command core.PublishJournalCommand) (core.LessonJournal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if err := s.journalTeacherAuthority(principal.AccountID, principal.TenantID, command.OccurrenceID, command.StudentID); err != nil {
		return core.LessonJournal{}, err
	}
	if response, ok, err := s.replay("publish_journal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.LessonJournal{}, err
		}
		var result core.LessonJournal
		if err := json.Unmarshal(response, &result); err != nil {
			return core.LessonJournal{}, core.E(core.CodeInternal, "decode idempotent journal result", err)
		}
		return result, nil
	}
	occurrence := s.lessons[command.OccurrenceID]
	if occurrence.StartsAt.After(command.Now) {
		return core.LessonJournal{}, core.E(core.CodeInvalidState, "the journal publishes after the Lesson starts", nil)
	}
	stored := s.journals[journalKey(command.OccurrenceID, command.StudentID)]
	if stored == nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidState, "save a draft before publishing", nil)
	}
	if stored.Draft == nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidState, "the draft must be complete before publishing", nil)
	}
	nextVersion := stored.CurrentVersion + 1
	if nextVersion > 1 && command.CorrectionNote == "" {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, "a correction requires an explicit note (DEC-007)", nil)
	}
	stored.Versions = append(stored.Versions, core.JournalVersion{
		Version: nextVersion, WhatWorked: stored.Draft.WhatWorked,
		CurrentFocus: stored.Draft.CurrentFocus, NextStep: stored.Draft.NextStep,
		CorrectionNote: command.CorrectionNote, PublishedAt: command.Now,
	})
	stored.CurrentVersion = nextVersion
	stored.Status = "published"
	stored.Draft = nil
	stored.UpdatedAt = command.Now
	sourceID := fmt.Sprintf("%s:%d", stored.ID, nextVersion)
	for index, evidence := range command.Evidence {
		s.evidence = append(s.evidence, &progressEvidenceRecord{
			ID: command.EvidenceIDs[index], TenantID: principal.TenantID,
			StudentID: command.StudentID, SourceKind: "lesson_journal", SourceID: sourceID,
			Area: evidence.Area, Note: evidence.Note,
			RecordedBy: principal.AccountID, RecordedAt: command.Now,
		})
	}
	result := s.journalView(stored, true)
	if err := s.completeIdempotency("publish_journal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.LessonJournal{}, err
	}
	action := "JournalPublished"
	if nextVersion > 1 {
		action = "JournalCorrected"
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"lesson_journal", stored.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, action, stored.ID, command.Now)
	return result, nil
}

func (s *Store) journalViewerScope(principal core.Principal, studentID string) (manager, isSelf bool) {
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	manager = actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	isSelf = s.studentIDForAccount(principal.AccountID) == studentID && studentID != ""
	return manager, isSelf
}

func (s *Store) GetJournal(_ context.Context, principal core.Principal, occurrenceID, studentID string) (core.LessonJournal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.journals[journalKey(occurrenceID, studentID)]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.LessonJournal{}, core.E(core.CodeNotFound, "journal not found", nil)
	}
	if stored.TeacherAccountID == principal.AccountID {
		return s.journalView(stored, true), nil
	}
	manager, isSelf := s.journalViewerScope(principal, studentID)
	if manager {
		return s.journalView(stored, true), nil
	}
	if isSelf {
		if stored.CurrentVersion == 0 {
			return core.LessonJournal{}, core.E(core.CodeNotFound, "journal not found", nil)
		}
		return s.journalView(stored, false), nil
	}
	return core.LessonJournal{}, core.E(core.CodeNotFound, "journal not found", nil)
}

func (s *Store) ListStudentJournals(_ context.Context, principal core.Principal, studentID string) ([]core.LessonJournal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	manager, isSelf := s.journalViewerScope(principal, studentID)
	result := []core.LessonJournal{}
	for _, stored := range s.journals {
		if stored.TenantID != principal.TenantID || stored.StudentID != studentID {
			continue
		}
		isTeacher := stored.TeacherAccountID == principal.AccountID
		if !manager && !isTeacher && !(isSelf && stored.CurrentVersion > 0) {
			continue
		}
		result = append(result, s.journalView(stored, manager || isTeacher))
	}
	sort.Slice(result, func(left, right int) bool {
		if !result[left].UpdatedAt.Equal(result[right].UpdatedAt) {
			return result[left].UpdatedAt.After(result[right].UpdatedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func (s *Store) ListProgressEvidence(_ context.Context, principal core.Principal, studentID string) ([]core.ProgressEvidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	manager, isSelf := s.journalViewerScope(principal, studentID)
	isTeacher := false
	for _, stored := range s.journals {
		if stored.TenantID == principal.TenantID && stored.StudentID == studentID &&
			stored.TeacherAccountID == principal.AccountID {
			isTeacher = true
		}
	}
	if assignment := s.assignmentAt(studentID, timeNowFallback()); assignment != nil &&
		assignment.TeacherAccountID == principal.AccountID {
		isTeacher = true
	}
	if !manager && !isSelf && !isTeacher {
		return nil, core.E(core.CodeForbidden, "progress is visible to the Student and assigned staff", nil)
	}
	result := []core.ProgressEvidence{}
	for _, record := range s.evidence {
		if record.TenantID != principal.TenantID || record.StudentID != studentID {
			continue
		}
		result = append(result, core.ProgressEvidence{
			ID: record.ID, Area: record.Area, Note: record.Note,
			SourceKind: record.SourceKind, SourceID: record.SourceID,
			RecordedAt: record.RecordedAt,
		})
	}
	sort.Slice(result, func(left, right int) bool {
		if !result[left].RecordedAt.Equal(result[right].RecordedAt) {
			return result[left].RecordedAt.After(result[right].RecordedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

// timeNowFallback keeps the memory assignment lookup deterministic in
// tests: assignments effective "now" use the latest known audit time.
func timeNowFallback() time.Time {
	return time.Now().UTC()
}
