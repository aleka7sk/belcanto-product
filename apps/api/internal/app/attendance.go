package app

import (
	"context"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.4 per-participant attendance (domain/lesson.md; Figma
// TCH-JOURNAL-01/02). A late mark carries minutes, an absence carries a
// mandatory note, and changing a recorded mark carries a reason.

type MarkAttendanceInput struct {
	OccurrenceID   string
	StudentID      string
	Status         string
	LateMinutes    int
	Note           string
	ChangeReason   string
	IdempotencyKey string
}

func (s *Service) MarkAttendance(ctx context.Context, principal core.Principal, input MarkAttendanceInput) ([]core.AttendanceRecord, error) {
	occurrenceID, err := security.ValidateIdentifier("occurrenceId", input.OccurrenceID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	switch input.Status {
	case core.AttendancePresent, core.AttendanceLate, core.AttendanceAbsent:
	default:
		return nil, core.E(core.CodeInvalidInput, "status must be present, late or absent", nil)
	}
	if input.Status == core.AttendanceLate {
		if input.LateMinutes < 1 || input.LateMinutes > 240 {
			return nil, core.E(core.CodeInvalidInput, "a late mark carries minutes between 1 and 240", nil)
		}
	} else if input.LateMinutes != 0 {
		return nil, core.E(core.CodeInvalidInput, "minutes belong to a late mark only", nil)
	}
	note := ""
	if input.Note != "" {
		note, err = security.ValidateText("note", input.Note, 1, 500)
		if err != nil {
			return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.Status == core.AttendanceAbsent && note == "" {
		return nil, core.E(core.CodeInvalidInput, "an absence requires a note", nil)
	}
	changeReason := ""
	if input.ChangeReason != "" {
		changeReason, err = security.ValidateText("reason", input.ChangeReason, 1, 500)
		if err != nil {
			return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		OccurrenceID, StudentID, Status, Note, Reason string
		LateMinutes                                   int
	}{occurrenceID, studentID, input.Status, note, changeReason, input.LateMinutes})
	if err != nil {
		return nil, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	records, err := s.store.MarkAttendance(ctx, core.MarkAttendanceCommand{
		Principal: principal, OccurrenceID: occurrenceID, StudentID: studentID,
		Status: input.Status, LateMinutes: input.LateMinutes, Note: note,
		ChangeReason: changeReason, IdempotencyKey: idempotencyKey,
		PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return nil, normalizeStoreError("mark attendance", err)
	}
	return records, nil
}

func (s *Service) ListLessonAttendance(ctx context.Context, principal core.Principal, occurrenceID string) ([]core.AttendanceRecord, error) {
	normalizedID, err := security.ValidateIdentifier("occurrenceId", occurrenceID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	records, err := s.store.ListLessonAttendance(ctx, principal, normalizedID)
	if err != nil {
		return nil, normalizeStoreError("list attendance", err)
	}
	return records, nil
}
