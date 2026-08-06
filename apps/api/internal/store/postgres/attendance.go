package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.4 per-participant attendance. The Lesson's Teacher or an
// Administrator marks it (domain/lesson.md: Owner is read-only);
// corrections require an explicit reason and the audit keeps the
// previous and new values. An empty group seat has no row.

func attendanceMarkerAuthority(ctx context.Context, tx pgx.Tx, tenantID, actorID, occurrenceID, studentID string) error {
	var teacherAccountID string
	err := tx.QueryRow(ctx, `
		SELECT teacher_account_id FROM core_lesson_occurrences
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, occurrenceID).Scan(&teacherAccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	if err != nil {
		return fmt.Errorf("read occurrence teacher: %w", err)
	}
	if teacherAccountID != actorID {
		admin, adminErr := hasActiveRole(ctx, tx, tenantID, actorID, core.RoleAdministrator)
		if adminErr != nil {
			return adminErr
		}
		if !admin {
			return core.E(core.CodeForbidden, "attendance is marked by the Lesson's Teacher or an Administrator", nil)
		}
	}
	var participates bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM core_lesson_occurrence_participants
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
		)
	`, tenantID, occurrenceID, studentID).Scan(&participates); err != nil {
		return fmt.Errorf("check attendance participant: %w", err)
	}
	if !participates {
		return core.E(core.CodeInvalidInput, "Student does not participate in this Lesson", nil)
	}
	return nil
}

func readLessonAttendance(ctx context.Context, reader lessonReader, tenantID, occurrenceID string) ([]core.AttendanceRecord, error) {
	rows, err := reader.Query(ctx, `
		SELECT a.student_id, person.full_name, a.status,
		       COALESCE(a.late_minutes, 0), COALESCE(a.note, ''),
		       a.recorded_at, a.updated_at
		FROM core_lesson_attendance a
		JOIN students s ON s.tenant_id = a.tenant_id AND s.id = a.student_id
		JOIN accounts account ON account.tenant_id = s.tenant_id AND account.id = s.account_id
		JOIN people person ON person.tenant_id = account.tenant_id AND person.id = account.person_id
		WHERE a.tenant_id = $1 AND a.occurrence_id = $2
		ORDER BY person.full_name, a.student_id
	`, tenantID, occurrenceID)
	if err != nil {
		return nil, fmt.Errorf("read lesson attendance: %w", err)
	}
	defer rows.Close()
	result := []core.AttendanceRecord{}
	for rows.Next() {
		var record core.AttendanceRecord
		if err := rows.Scan(&record.StudentID, &record.StudentName, &record.Status,
			&record.LateMinutes, &record.Note, &record.RecordedAt, &record.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan attendance record: %w", err)
		}
		record.RecordedAt = record.RecordedAt.UTC()
		record.UpdatedAt = record.UpdatedAt.UTC()
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attendance records: %w", err)
	}
	return result, nil
}

func (s *Store) MarkAttendance(ctx context.Context, command core.MarkAttendanceCommand) ([]core.AttendanceRecord, error) {
	principal := command.Principal
	var records []core.AttendanceRecord
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := attendanceMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.OccurrenceID, command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "mark_attendance", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			records, err = decodeReplay[[]core.AttendanceRecord](claim)
			return err
		}
		var previousStatus string
		var previousMinutes *int
		err = tx.QueryRow(ctx, `
			SELECT status, late_minutes FROM core_lesson_attendance
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
			FOR UPDATE
		`, principal.TenantID, command.OccurrenceID, command.StudentID).Scan(&previousStatus, &previousMinutes)
		exists := err == nil
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("lock attendance row: %w", err)
		}
		if exists && command.ChangeReason == "" {
			return core.E(core.CodeInvalidInput, "changing a recorded attendance requires a reason", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO core_lesson_attendance (
				tenant_id, occurrence_id, student_id, status, late_minutes, note,
				recorded_by_account_id, recorded_at, updated_at
			) VALUES ($1, $2, $3, $4, NULLIF($5, 0), NULLIF($6, ''), $7, $8, $8)
			ON CONFLICT (tenant_id, occurrence_id, student_id)
			DO UPDATE SET status = EXCLUDED.status,
			              late_minutes = EXCLUDED.late_minutes,
			              note = EXCLUDED.note,
			              recorded_by_account_id = EXCLUDED.recorded_by_account_id,
			              updated_at = EXCLUDED.updated_at
		`, principal.TenantID, command.OccurrenceID, command.StudentID, command.Status,
			command.LateMinutes, command.Note, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "attendance conflicts with existing data")
		}
		records, err = readLessonAttendance(ctx, tx, principal.TenantID, command.OccurrenceID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "mark_attendance", command.IdempotencyKey, records, command.Now); err != nil {
			return err
		}
		action := "AttendanceMarked"
		if exists {
			action = "AttendanceCorrected"
		}
		metadata := map[string]any{
			"studentId": command.StudentID, "status": command.Status,
		}
		if exists {
			metadata["previousStatus"] = previousStatus
			metadata["reason"] = command.ChangeReason
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "lesson_attendance",
			targetID: command.OccurrenceID + ":" + command.StudentID,
			decision: "allow", reason: command.ChangeReason,
			idempotencyKey: command.IdempotencyKey, metadata: metadata,
			at: command.Now,
		}); err != nil {
			return err
		}
		if command.Status != core.AttendanceAbsent {
			return nil
		}
		return appendOutbox(ctx, tx, principal.TenantID, "AttendanceAbsenceRecorded", "lesson_attendance",
			command.OccurrenceID+":"+command.StudentID,
			map[string]any{"occurrenceId": command.OccurrenceID, "studentId": command.StudentID}, command.Now)
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (s *Store) ListLessonAttendance(ctx context.Context, principal core.Principal, occurrenceID string) ([]core.AttendanceRecord, error) {
	var teacherAccountID string
	err := s.pool.QueryRow(ctx, `
		SELECT teacher_account_id FROM core_lesson_occurrences
		WHERE tenant_id = $1 AND id = $2
	`, principal.TenantID, occurrenceID).Scan(&teacherAccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("read occurrence for attendance: %w", err)
	}
	records, err := readLessonAttendance(ctx, s.pool, principal.TenantID, occurrenceID)
	if err != nil {
		return nil, err
	}
	if teacherAccountID == principal.AccountID {
		return records, nil
	}
	manager := false
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		authority, authorityErr := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if authorityErr != nil {
			return authorityErr
		}
		manager = authority
		return nil
	}); err != nil {
		return nil, err
	}
	if manager {
		return records, nil
	}
	ownStudentID, err := studentIDForAccount(ctx, s.pool, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, err
	}
	if ownStudentID != "" {
		// The Student sees their own mark; the teacher note stays private.
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
