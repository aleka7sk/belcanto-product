package memory

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.4 per-participant attendance — parity with PostgreSQL. Corrections
// require a reason; rows are never removed; an empty group seat has no
// row.

type attendanceRecord struct {
	TenantID     string
	OccurrenceID string
	StudentID    string
	Status       string
	LateMinutes  int
	Note         string
	RecordedBy   string
	RecordedAt   time.Time
	UpdatedAt    time.Time
}

func (s *Store) attendanceMarkerAuthority(actorID, tenantID, occurrenceID, studentID string) error {
	occurrence := s.lessons[occurrenceID]
	if occurrence == nil || occurrence.TenantID != tenantID {
		return core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	if occurrence.TeacherAccountID != actorID {
		actor := s.activeAccount(actorID, tenantID)
		if actor == nil || actor.Roles[core.RoleAdministrator] == "" {
			return core.E(core.CodeForbidden, "attendance is marked by the Lesson's Teacher or an Administrator", nil)
		}
	}
	for _, participant := range occurrence.StudentIDs {
		if participant == studentID {
			return nil
		}
	}
	return core.E(core.CodeInvalidInput, "Student does not participate in this Lesson", nil)
}

func (s *Store) lessonAttendanceView(tenantID, occurrenceID string) []core.AttendanceRecord {
	result := []core.AttendanceRecord{}
	for _, record := range s.attendance {
		if record.TenantID != tenantID || record.OccurrenceID != occurrenceID {
			continue
		}
		view := core.AttendanceRecord{
			StudentID: record.StudentID, Status: record.Status,
			LateMinutes: record.LateMinutes, Note: record.Note,
			RecordedAt: record.RecordedAt, UpdatedAt: record.UpdatedAt,
		}
		if stored := s.students[record.StudentID]; stored != nil {
			view.StudentName = stored.FullName
		}
		result = append(result, view)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].StudentName != result[right].StudentName {
			return result[left].StudentName < result[right].StudentName
		}
		return result[left].StudentID < result[right].StudentID
	})
	return result
}

func (s *Store) MarkAttendance(_ context.Context, command core.MarkAttendanceCommand) ([]core.AttendanceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if err := s.attendanceMarkerAuthority(principal.AccountID, principal.TenantID, command.OccurrenceID, command.StudentID); err != nil {
		return nil, err
	}
	if response, ok, err := s.replay("mark_attendance", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return nil, err
		}
		var result []core.AttendanceRecord
		if err := json.Unmarshal(response, &result); err != nil {
			return nil, core.E(core.CodeInternal, "decode idempotent attendance result", err)
		}
		return result, nil
	}
	key := command.OccurrenceID + "\x00" + command.StudentID
	existing := s.attendance[key]
	if existing != nil && command.ChangeReason == "" {
		return nil, core.E(core.CodeInvalidInput, "changing a recorded attendance requires a reason", nil)
	}
	action := "AttendanceMarked"
	if existing != nil {
		action = "AttendanceCorrected"
	}
	recordedAt := command.Now
	if existing != nil {
		recordedAt = existing.RecordedAt
	}
	s.attendance[key] = &attendanceRecord{
		TenantID: principal.TenantID, OccurrenceID: command.OccurrenceID,
		StudentID: command.StudentID, Status: command.Status,
		LateMinutes: command.LateMinutes, Note: command.Note,
		RecordedBy: principal.AccountID, RecordedAt: recordedAt, UpdatedAt: command.Now,
	}
	result := s.lessonAttendanceView(principal.TenantID, command.OccurrenceID)
	if err := s.completeIdempotency("mark_attendance", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return nil, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"lesson_attendance", command.OccurrenceID+":"+command.StudentID,
		"allow", command.ChangeReason, command.Now, nil)
	if command.Status == core.AttendanceAbsent {
		s.appendOutboxPayload(principal.TenantID, "AttendanceAbsenceRecorded", "lesson_attendance",
			command.OccurrenceID+":"+command.StudentID,
			map[string]any{"occurrenceId": command.OccurrenceID, "studentId": command.StudentID}, command.Now)
	}
	return result, nil
}

func (s *Store) ListLessonAttendance(_ context.Context, principal core.Principal, occurrenceID string) ([]core.AttendanceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	occurrence := s.lessons[occurrenceID]
	if occurrence == nil || occurrence.TenantID != principal.TenantID {
		return nil, core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	records := s.lessonAttendanceView(principal.TenantID, occurrenceID)
	if occurrence.TeacherAccountID == principal.AccountID {
		return records, nil
	}
	if actor := s.activeAccount(principal.AccountID, principal.TenantID); actor != nil &&
		(actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "") {
		return records, nil
	}
	if ownStudentID := s.studentIDForAccount(principal.AccountID); ownStudentID != "" {
		own := []core.AttendanceRecord{}
		for _, record := range records {
			if record.StudentID == ownStudentID {
				record.Note = ""
				own = append(own, record)
			}
		}
		return own, nil
	}
	return nil, core.E(core.CodeForbidden, "attendance is visible to the Lesson's staff and its Students", nil)
}
