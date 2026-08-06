package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 homework and practice (domain/homework.md Approved 1.0.0).
// Homework belongs to a Lesson and one Student; the author is the
// Lesson's Teacher. Completed is final, nothing is ever deleted, and
// submissions/feedback are immutable history (DB triggers enforce it).
// Expiry is lazy: a due assignment flips to expired on the next read or
// mutation — the deadline has no other consequence (DEC-102 adjacent).

func expireDueHomework(ctx context.Context, tx pgx.Tx, tenantID, homeworkID string, now time.Time) error {
	_, err := tx.Exec(ctx, `
		UPDATE homework_assignments
		SET status = 'expired', updated_at = $3, version = version + 1
		WHERE tenant_id = $1 AND id = $2
		  AND status IN ('assigned', 'in_progress')
		  AND due_at IS NOT NULL AND due_at < $3
	`, tenantID, homeworkID, now)
	if err != nil {
		return fmt.Errorf("expire due homework: %w", err)
	}
	return nil
}

func (s *Store) expireStudentHomework(ctx context.Context, tenantID, studentID string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE homework_assignments
		SET status = 'expired', updated_at = $3, version = version + 1
		WHERE tenant_id = $1 AND student_id = $2
		  AND status IN ('assigned', 'in_progress')
		  AND due_at IS NOT NULL AND due_at < $3
	`, tenantID, studentID, now)
	if err != nil {
		return fmt.Errorf("expire student homework: %w", err)
	}
	return nil
}

type homeworkRow struct {
	occurrenceID string
	studentID    string
	teacherID    string
	status       string
	version      int
}

func lockHomework(ctx context.Context, tx pgx.Tx, tenantID, homeworkID string) (homeworkRow, error) {
	var row homeworkRow
	err := tx.QueryRow(ctx, `
		SELECT occurrence_id, student_id, teacher_account_id, status, version
		FROM homework_assignments
		WHERE tenant_id = $1 AND id = $2
		FOR UPDATE
	`, tenantID, homeworkID).Scan(&row.occurrenceID, &row.studentID, &row.teacherID, &row.status, &row.version)
	if errors.Is(err, pgx.ErrNoRows) {
		return homeworkRow{}, core.E(core.CodeNotFound, "homework not found", nil)
	}
	if err != nil {
		return homeworkRow{}, fmt.Errorf("lock homework: %w", err)
	}
	return row, nil
}

func studentSelfAuthority(ctx context.Context, tx pgx.Tx, tenantID, accountID, studentID string) error {
	ownStudentID, err := studentIDForAccount(ctx, tx, tenantID, accountID)
	if err != nil {
		return err
	}
	if ownStudentID == "" || ownStudentID != studentID {
		return core.E(core.CodeForbidden, "only the assigned Student performs this action", nil)
	}
	return nil
}

func validateReadyMedia(ctx context.Context, tx pgx.Tx, tenantID, ownerAccountID string, mediaIDs []string) error {
	for _, mediaID := range mediaIDs {
		var owner, status string
		err := tx.QueryRow(ctx, `
			SELECT owner_account_id, status FROM media_objects
			WHERE tenant_id = $1 AND id = $2
		`, tenantID, mediaID).Scan(&owner, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidInput, "attached media was not found", nil)
		}
		if err != nil {
			return fmt.Errorf("read attached media: %w", err)
		}
		if owner != ownerAccountID {
			return core.E(core.CodeForbidden, "attached media must belong to the actor", nil)
		}
		if status != core.MediaStatusReady {
			return core.E(core.CodeInvalidState, "attached media must finish uploading first", nil)
		}
	}
	return nil
}

func (s *Store) CreateHomework(ctx context.Context, command core.CreateHomeworkCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	var homework core.HomeworkAssignment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := journalTeacherAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.OccurrenceID, command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_homework", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			homework, err = decodeReplay[core.HomeworkAssignment](claim)
			return err
		}
		if err := validateReadyMedia(ctx, tx, principal.TenantID, principal.AccountID, command.AttachmentMediaIDs); err != nil {
			return err
		}
		status := core.HomeworkStatusDraft
		if command.Assign {
			status = core.HomeworkStatusAssigned
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO homework_assignments (
				id, tenant_id, occurrence_id, student_id, teacher_account_id,
				status, goal, readiness_criteria, due_at, version, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, 1, $10, $10)
		`, command.HomeworkID, principal.TenantID, command.OccurrenceID, command.StudentID,
			principal.AccountID, status, command.Goal, command.ReadinessCriteria,
			command.DueAt, command.Now); err != nil {
			return mapWriteError(err, "homework conflicts with existing data")
		}
		for index, task := range command.Tasks {
			if _, err := tx.Exec(ctx, `
				INSERT INTO homework_tasks (
					tenant_id, homework_id, id, position, title, description,
					recommended_minutes, skill_area, song_title, status
				) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, 0), NULLIF($8, ''), NULLIF($9, ''), 'pending')
			`, principal.TenantID, command.HomeworkID, command.TaskIDs[index], index+1,
				task.Title, task.Description, task.RecommendedMinutes,
				task.SkillArea, task.SongTitle); err != nil {
				return mapWriteError(err, "homework task conflicts with existing data")
			}
		}
		for index, mediaID := range command.AttachmentMediaIDs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO homework_attachments (tenant_id, homework_id, media_id, position)
				VALUES ($1, $2, $3, $4)
			`, principal.TenantID, command.HomeworkID, mediaID, index+1); err != nil {
				return mapWriteError(err, "homework attachment conflicts with existing data")
			}
		}
		homework, err = readHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_homework", command.IdempotencyKey, homework, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "HomeworkCreated", targetType: "homework", targetID: command.HomeworkID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"studentId": command.StudentID, "tasks": len(command.Tasks), "assigned": command.Assign},
			at:       command.Now,
		}); err != nil {
			return err
		}
		if !command.Assign {
			return nil
		}
		return appendOutbox(ctx, tx, principal.TenantID, "HomeworkAssigned", "homework", command.HomeworkID,
			map[string]any{"homeworkId": command.HomeworkID, "studentId": command.StudentID}, command.Now)
	})
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	return homework, nil
}

func (s *Store) homeworkTransition(ctx context.Context, command core.HomeworkTransitionCommand, operation string,
	transition func(ctx context.Context, tx pgx.Tx, row homeworkRow) (string, error)) (core.HomeworkAssignment, error) {
	principal := command.Principal
	var homework core.HomeworkAssignment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, operation, command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			homework, err = decodeReplay[core.HomeworkAssignment](claim)
			return err
		}
		if err := expireDueHomework(ctx, tx, principal.TenantID, command.HomeworkID, command.Now); err != nil {
			return err
		}
		row, err := lockHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		action, err := transition(ctx, tx, row)
		if err != nil {
			return err
		}
		homework, err = readHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, operation, command.IdempotencyKey, homework, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "homework", targetID: command.HomeworkID,
			decision: "allow", idempotencyKey: command.IdempotencyKey, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, action, "homework", command.HomeworkID,
			map[string]any{"homeworkId": command.HomeworkID, "studentId": row.studentID}, command.Now)
	})
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	return homework, nil
}

func (s *Store) AssignHomework(ctx context.Context, command core.HomeworkTransitionCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	return s.homeworkTransition(ctx, command, "assign_homework",
		func(ctx context.Context, tx pgx.Tx, row homeworkRow) (string, error) {
			if row.teacherID != principal.AccountID {
				return "", core.E(core.CodeForbidden, "only the homework's Teacher assigns it", nil)
			}
			if row.status != core.HomeworkStatusDraft {
				return "", core.E(core.CodeInvalidState, "only a draft homework can be assigned", nil)
			}
			if _, err := tx.Exec(ctx, `
				UPDATE homework_assignments
				SET status = 'assigned', updated_at = $3, version = version + 1
				WHERE tenant_id = $1 AND id = $2
			`, principal.TenantID, command.HomeworkID, command.Now); err != nil {
				return "", mapWriteError(err, "homework transition conflicts with existing data")
			}
			return "HomeworkAssigned", nil
		})
}

func (s *Store) StartHomework(ctx context.Context, command core.HomeworkTransitionCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	return s.homeworkTransition(ctx, command, "start_homework",
		func(ctx context.Context, tx pgx.Tx, row homeworkRow) (string, error) {
			if err := studentSelfAuthority(ctx, tx, principal.TenantID, principal.AccountID, row.studentID); err != nil {
				return "", err
			}
			if row.status != core.HomeworkStatusAssigned {
				return "", core.E(core.CodeInvalidState, "the homework is not awaiting a start", nil)
			}
			if _, err := tx.Exec(ctx, `
				UPDATE homework_assignments
				SET status = 'in_progress', updated_at = $3, version = version + 1
				WHERE tenant_id = $1 AND id = $2
			`, principal.TenantID, command.HomeworkID, command.Now); err != nil {
				return "", mapWriteError(err, "homework transition conflicts with existing data")
			}
			return "HomeworkStarted", nil
		})
}

func (s *Store) CancelHomework(ctx context.Context, command core.HomeworkTransitionCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	return s.homeworkTransition(ctx, command, "cancel_homework",
		func(ctx context.Context, tx pgx.Tx, row homeworkRow) (string, error) {
			if row.teacherID != principal.AccountID {
				return "", core.E(core.CodeForbidden, "only the homework's Teacher cancels it", nil)
			}
			switch row.status {
			case core.HomeworkStatusDraft, core.HomeworkStatusAssigned,
				core.HomeworkStatusInProgress, core.HomeworkStatusSubmitted,
				core.HomeworkStatusReviewed:
			default:
				return "", core.E(core.CodeInvalidState, "the homework is already closed", nil)
			}
			if _, err := tx.Exec(ctx, `
				UPDATE homework_assignments
				SET status = 'cancelled', cancel_reason = $3, updated_at = $4, version = version + 1
				WHERE tenant_id = $1 AND id = $2
			`, principal.TenantID, command.HomeworkID, command.Reason, command.Now); err != nil {
				return "", mapWriteError(err, "homework transition conflicts with existing data")
			}
			return "HomeworkCancelled", nil
		})
}

func (s *Store) MarkHomeworkTask(ctx context.Context, command core.MarkHomeworkTaskCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	var homework core.HomeworkAssignment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := expireDueHomework(ctx, tx, principal.TenantID, command.HomeworkID, command.Now); err != nil {
			return err
		}
		row, err := lockHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		if err := studentSelfAuthority(ctx, tx, principal.TenantID, principal.AccountID, row.studentID); err != nil {
			return err
		}
		if row.status != core.HomeworkStatusInProgress {
			return core.E(core.CodeInvalidState, "tasks are marked while the homework is in progress", nil)
		}
		status := "pending"
		if command.Done {
			status = "done"
		}
		tag, err := tx.Exec(ctx, `
			UPDATE homework_tasks
			SET status = $4
			WHERE tenant_id = $1 AND homework_id = $2 AND id = $3
		`, principal.TenantID, command.HomeworkID, command.TaskID, status)
		if err != nil {
			return mapWriteError(err, "homework task conflicts with existing data")
		}
		if tag.RowsAffected() == 0 {
			return core.E(core.CodeNotFound, "homework task not found", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE homework_assignments SET updated_at = $3
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.HomeworkID, command.Now); err != nil {
			return fmt.Errorf("touch homework: %w", err)
		}
		homework, err = readHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		action := "TaskReopened"
		if command.Done {
			action = "TaskCompleted"
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "homework_task", targetID: command.TaskID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	return homework, nil
}

func (s *Store) SubmitHomework(ctx context.Context, command core.SubmitHomeworkCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	var homework core.HomeworkAssignment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "submit_homework", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			homework, err = decodeReplay[core.HomeworkAssignment](claim)
			return err
		}
		if err := expireDueHomework(ctx, tx, principal.TenantID, command.HomeworkID, command.Now); err != nil {
			return err
		}
		row, err := lockHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		if err := studentSelfAuthority(ctx, tx, principal.TenantID, principal.AccountID, row.studentID); err != nil {
			return err
		}
		if row.status != core.HomeworkStatusInProgress && row.status != core.HomeworkStatusReviewed {
			return core.E(core.CodeInvalidState, "the homework is not open for a submission", nil)
		}
		if command.ExpectedVersion != row.version {
			return core.E(core.CodeConflict, "the homework changed; reload and retry", nil)
		}
		if err := validateReadyMedia(ctx, tx, principal.TenantID, principal.AccountID, command.MediaIDs); err != nil {
			return err
		}
		var attempt int
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(MAX(attempt), 0) + 1 FROM practice_submissions
			WHERE tenant_id = $1 AND homework_id = $2
		`, principal.TenantID, command.HomeworkID).Scan(&attempt); err != nil {
			return fmt.Errorf("next submission attempt: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice_submissions (
				id, tenant_id, homework_id, student_id, attempt, note, submitted_at
			) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7)
		`, command.SubmissionID, principal.TenantID, command.HomeworkID, row.studentID,
			attempt, command.Note, command.Now); err != nil {
			return mapWriteError(err, "practice submission conflicts with existing data")
		}
		for index, mediaID := range command.MediaIDs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO practice_submission_media (tenant_id, submission_id, media_id, position)
				VALUES ($1, $2, $3, $4)
			`, principal.TenantID, command.SubmissionID, mediaID, index+1); err != nil {
				return mapWriteError(err, "submission media conflicts with existing data")
			}
		}
		if _, err := tx.Exec(ctx, `
			UPDATE homework_assignments
			SET status = 'submitted', updated_at = $3, version = version + 1
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.HomeworkID, command.Now); err != nil {
			return mapWriteError(err, "homework transition conflicts with existing data")
		}
		homework, err = readHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "submit_homework", command.IdempotencyKey, homework, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "HomeworkSubmitted", targetType: "homework", targetID: command.HomeworkID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"attempt": attempt, "mediaCount": len(command.MediaIDs)},
			at:       command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "HomeworkSubmitted", "homework", command.HomeworkID,
			map[string]any{"homeworkId": command.HomeworkID, "studentId": row.studentID, "attempt": attempt}, command.Now)
	})
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	return homework, nil
}

func (s *Store) ReviewHomework(ctx context.Context, command core.ReviewHomeworkCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	var homework core.HomeworkAssignment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "review_homework", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			homework, err = decodeReplay[core.HomeworkAssignment](claim)
			return err
		}
		row, err := lockHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		if row.teacherID != principal.AccountID {
			return core.E(core.CodeForbidden, "only the homework's Teacher reviews it", nil)
		}
		if row.status != core.HomeworkStatusSubmitted {
			return core.E(core.CodeInvalidState, "there is no submission awaiting review", nil)
		}
		if command.ExpectedVersion != row.version {
			return core.E(core.CodeConflict, "the homework changed; reload and retry", nil)
		}
		var submissionID string
		if err := tx.QueryRow(ctx, `
			SELECT id FROM practice_submissions
			WHERE tenant_id = $1 AND homework_id = $2
			ORDER BY attempt DESC
			LIMIT 1
		`, principal.TenantID, command.HomeworkID).Scan(&submissionID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return core.E(core.CodeInvalidState, "there is no submission awaiting review", nil)
			}
			return fmt.Errorf("find submission for review: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice_feedback (
				id, tenant_id, homework_id, submission_id, teacher_account_id,
				decision, body, next_step, evidence_area, evidence_note, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11)
		`, command.FeedbackID, principal.TenantID, command.HomeworkID, submissionID,
			principal.AccountID, command.Decision, command.Body, command.NextStep,
			command.EvidenceArea, command.EvidenceNote, command.Now); err != nil {
			return mapWriteError(err, "practice feedback conflicts with existing data")
		}
		nextStatus := core.HomeworkStatusReviewed
		action := "HomeworkReviewed"
		if command.Decision == core.FeedbackDecisionAccepted {
			nextStatus = core.HomeworkStatusCompleted
			action = "HomeworkCompleted"
			if command.EvidenceArea != "" {
				if _, err := tx.Exec(ctx, `
					INSERT INTO progress_evidence (
						id, tenant_id, student_id, source_kind, source_id, area, note,
						recorded_by_account_id, recorded_at
					) VALUES ($1, $2, $3, 'practice', $4, $5, $6, $7, $8)
				`, command.EvidenceID, principal.TenantID, row.studentID, submissionID,
					command.EvidenceArea, command.EvidenceNote, principal.AccountID, command.Now); err != nil {
					return mapWriteError(err, "progress evidence conflicts with existing data")
				}
			}
		}
		if _, err := tx.Exec(ctx, `
			UPDATE homework_assignments
			SET status = $3, updated_at = $4, version = version + 1
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.HomeworkID, nextStatus, command.Now); err != nil {
			return mapWriteError(err, "homework transition conflicts with existing data")
		}
		homework, err = readHomework(ctx, tx, principal.TenantID, command.HomeworkID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "review_homework", command.IdempotencyKey, homework, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "homework", targetID: command.HomeworkID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"decision": command.Decision, "submissionId": submissionID},
			at:       command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, action, "homework", command.HomeworkID,
			map[string]any{"homeworkId": command.HomeworkID, "studentId": row.studentID, "decision": command.Decision}, command.Now)
	})
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	return homework, nil
}

func readHomework(ctx context.Context, reader lessonReader, tenantID, homeworkID string) (core.HomeworkAssignment, error) {
	var homework core.HomeworkAssignment
	var readiness, cancelReason *string
	err := reader.QueryRow(ctx, `
		SELECT h.id, h.occurrence_id, h.student_id,
		       h.teacher_account_id, teacher_person.full_name,
		       h.status, h.goal, h.readiness_criteria, h.due_at, h.cancel_reason,
		       h.version, h.created_at, h.updated_at
		FROM homework_assignments h
		JOIN accounts teacher_account
		  ON teacher_account.tenant_id = h.tenant_id AND teacher_account.id = h.teacher_account_id
		JOIN people teacher_person
		  ON teacher_person.tenant_id = teacher_account.tenant_id
		 AND teacher_person.id = teacher_account.person_id
		WHERE h.tenant_id = $1 AND h.id = $2
	`, tenantID, homeworkID).Scan(
		&homework.ID, &homework.OccurrenceID, &homework.StudentID,
		&homework.Teacher.AccountID, &homework.Teacher.FullName,
		&homework.Status, &homework.Goal, &readiness, &homework.DueAt, &cancelReason,
		&homework.Version, &homework.CreatedAt, &homework.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework not found", nil)
	}
	if err != nil {
		return core.HomeworkAssignment{}, fmt.Errorf("read homework: %w", err)
	}
	if readiness != nil {
		homework.ReadinessCriteria = *readiness
	}
	if cancelReason != nil {
		homework.CancelReason = *cancelReason
	}
	homework.CreatedAt = homework.CreatedAt.UTC()
	homework.UpdatedAt = homework.UpdatedAt.UTC()
	if homework.DueAt != nil {
		utc := homework.DueAt.UTC()
		homework.DueAt = &utc
	}

	homework.Tasks = make([]core.HomeworkTask, 0)
	rows, err := reader.Query(ctx, `
		SELECT id, position, title, COALESCE(description, ''),
		       COALESCE(recommended_minutes, 0), COALESCE(skill_area, ''),
		       COALESCE(song_title, ''), status
		FROM homework_tasks
		WHERE tenant_id = $1 AND homework_id = $2
		ORDER BY position
	`, tenantID, homeworkID)
	if err != nil {
		return core.HomeworkAssignment{}, fmt.Errorf("read homework tasks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var task core.HomeworkTask
		if err := rows.Scan(&task.ID, &task.Position, &task.Title, &task.Description,
			&task.RecommendedMinutes, &task.SkillArea, &task.SongTitle, &task.Status); err != nil {
			return core.HomeworkAssignment{}, fmt.Errorf("scan homework task: %w", err)
		}
		homework.Tasks = append(homework.Tasks, task)
	}
	if err := rows.Err(); err != nil {
		return core.HomeworkAssignment{}, fmt.Errorf("iterate homework tasks: %w", err)
	}
	rows.Close()

	homework.Attachments, err = readMediaList(ctx, reader, tenantID, `
		SELECT m.id, m.kind, m.content_type, m.byte_size, m.uploaded_bytes, m.status,
		       m.created_at, m.updated_at
		FROM homework_attachments a
		JOIN media_objects m ON m.tenant_id = a.tenant_id AND m.id = a.media_id
		WHERE a.tenant_id = $1 AND a.homework_id = $2
		ORDER BY a.position
	`, homeworkID)
	if err != nil {
		return core.HomeworkAssignment{}, err
	}

	homework.Submissions = make([]core.PracticeSubmission, 0)
	submissionRows, err := reader.Query(ctx, `
		SELECT id, attempt, COALESCE(note, ''), submitted_at
		FROM practice_submissions
		WHERE tenant_id = $1 AND homework_id = $2
		ORDER BY attempt DESC
	`, tenantID, homeworkID)
	if err != nil {
		return core.HomeworkAssignment{}, fmt.Errorf("read submissions: %w", err)
	}
	defer submissionRows.Close()
	for submissionRows.Next() {
		var submission core.PracticeSubmission
		if err := submissionRows.Scan(&submission.ID, &submission.Attempt,
			&submission.Note, &submission.SubmittedAt); err != nil {
			return core.HomeworkAssignment{}, fmt.Errorf("scan submission: %w", err)
		}
		submission.SubmittedAt = submission.SubmittedAt.UTC()
		homework.Submissions = append(homework.Submissions, submission)
	}
	if err := submissionRows.Err(); err != nil {
		return core.HomeworkAssignment{}, fmt.Errorf("iterate submissions: %w", err)
	}
	submissionRows.Close()
	for index := range homework.Submissions {
		media, err := readMediaList(ctx, reader, tenantID, `
			SELECT m.id, m.kind, m.content_type, m.byte_size, m.uploaded_bytes, m.status,
			       m.created_at, m.updated_at
			FROM practice_submission_media sm
			JOIN media_objects m ON m.tenant_id = sm.tenant_id AND m.id = sm.media_id
			WHERE sm.tenant_id = $1 AND sm.submission_id = $2
			ORDER BY sm.position
		`, homework.Submissions[index].ID)
		if err != nil {
			return core.HomeworkAssignment{}, err
		}
		homework.Submissions[index].Media = media
	}

	homework.Feedback = make([]core.PracticeFeedback, 0)
	feedbackRows, err := reader.Query(ctx, `
		SELECT f.id, f.submission_id, f.teacher_account_id, teacher_person.full_name,
		       f.decision, f.body, COALESCE(f.next_step, ''),
		       COALESCE(f.evidence_area, ''), COALESCE(f.evidence_note, ''), f.created_at
		FROM practice_feedback f
		JOIN accounts teacher_account
		  ON teacher_account.tenant_id = f.tenant_id AND teacher_account.id = f.teacher_account_id
		JOIN people teacher_person
		  ON teacher_person.tenant_id = teacher_account.tenant_id
		 AND teacher_person.id = teacher_account.person_id
		WHERE f.tenant_id = $1 AND f.homework_id = $2
		ORDER BY f.created_at DESC, f.id
	`, tenantID, homeworkID)
	if err != nil {
		return core.HomeworkAssignment{}, fmt.Errorf("read feedback: %w", err)
	}
	defer feedbackRows.Close()
	for feedbackRows.Next() {
		var feedback core.PracticeFeedback
		if err := feedbackRows.Scan(&feedback.ID, &feedback.SubmissionID,
			&feedback.Teacher.AccountID, &feedback.Teacher.FullName,
			&feedback.Decision, &feedback.Body, &feedback.NextStep,
			&feedback.EvidenceArea, &feedback.EvidenceNote, &feedback.CreatedAt); err != nil {
			return core.HomeworkAssignment{}, fmt.Errorf("scan feedback: %w", err)
		}
		feedback.CreatedAt = feedback.CreatedAt.UTC()
		homework.Feedback = append(homework.Feedback, feedback)
	}
	if err := feedbackRows.Err(); err != nil {
		return core.HomeworkAssignment{}, fmt.Errorf("iterate feedback: %w", err)
	}
	return homework, nil
}

func readMediaList(ctx context.Context, reader lessonReader, tenantID, query, scopeID string) ([]core.MediaObject, error) {
	result := make([]core.MediaObject, 0)
	rows, err := reader.Query(ctx, query, tenantID, scopeID)
	if err != nil {
		return nil, fmt.Errorf("read media list: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var object core.MediaObject
		if err := rows.Scan(&object.ID, &object.Kind, &object.ContentType, &object.ByteSize,
			&object.UploadedBytes, &object.Status, &object.CreatedAt, &object.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan media object: %w", err)
		}
		object.CreatedAt = object.CreatedAt.UTC()
		object.UpdatedAt = object.UpdatedAt.UTC()
		result = append(result, object)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media list: %w", err)
	}
	return result, nil
}

// homeworkViewerScope: the homework's teacher, managers, and the student
// see an assignment; drafts stay teacher-only.
func (s *Store) homeworkViewer(homework core.HomeworkAssignment, principal core.Principal, manager, isSelf bool) error {
	if homework.Teacher.AccountID == principal.AccountID || manager {
		return nil
	}
	if isSelf && homework.Status != core.HomeworkStatusDraft {
		return nil
	}
	return core.E(core.CodeNotFound, "homework not found", nil)
}

func (s *Store) GetHomework(ctx context.Context, principal core.Principal, homeworkID string, now time.Time) (core.HomeworkAssignment, error) {
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return expireDueHomework(ctx, tx, principal.TenantID, homeworkID, now)
	}); err != nil {
		return core.HomeworkAssignment{}, err
	}
	homework, err := readHomework(ctx, s.pool, principal.TenantID, homeworkID)
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	manager, isSelf, err := s.journalViewerScope(ctx, principal, homework.StudentID)
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	if err := s.homeworkViewer(homework, principal, manager, isSelf); err != nil {
		return core.HomeworkAssignment{}, err
	}
	return homework, nil
}

func (s *Store) ListStudentHomework(ctx context.Context, principal core.Principal, studentID string, now time.Time) ([]core.HomeworkAssignment, error) {
	manager, isSelf, err := s.journalViewerScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	if err := s.expireStudentHomework(ctx, principal.TenantID, studentID, now); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM homework_assignments
		WHERE tenant_id = $1 AND student_id = $2
		  AND ($3::boolean OR teacher_account_id = $4 OR ($5::boolean AND status <> 'draft'))
		ORDER BY updated_at DESC
		LIMIT 100
	`, principal.TenantID, studentID, manager, principal.AccountID, isSelf)
	if err != nil {
		return nil, fmt.Errorf("list student homework: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan homework id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate homework ids: %w", err)
	}
	rows.Close()
	if !manager && !isSelf {
		// A teacher who is neither manager nor the student sees only
		// their own assignments; the filter above already ensured it,
		// but an empty scope should read as forbidden, not empty.
		var teaches bool
		if err := s.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM homework_assignments
				WHERE tenant_id = $1 AND student_id = $2 AND teacher_account_id = $3
			) OR EXISTS (
				SELECT 1 FROM teacher_assignments
				WHERE tenant_id = $1 AND student_id = $2 AND teacher_account_id = $3
				  AND status = 'active'
			)
		`, principal.TenantID, studentID, principal.AccountID).Scan(&teaches); err != nil {
			return nil, fmt.Errorf("check homework teacher scope: %w", err)
		}
		if !teaches {
			return nil, core.E(core.CodeForbidden, "homework is visible to the Student and assigned staff", nil)
		}
	}
	result := make([]core.HomeworkAssignment, 0, len(ids))
	for _, id := range ids {
		homework, err := readHomework(ctx, s.pool, principal.TenantID, id)
		if err != nil {
			return nil, err
		}
		result = append(result, homework)
	}
	return result, nil
}
