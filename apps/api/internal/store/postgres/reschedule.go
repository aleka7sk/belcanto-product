package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 reschedule and cancellation requests (flows J/K/L). A participant
// student or the lesson's teacher asks; Owner/Administrator decides.
// Approval applies the change to the occurrence in the same transaction.
// DEC-102 stays open: nothing here computes a late-cancellation
// consequence.

func (s *Store) CreateRescheduleRequest(ctx context.Context, command core.CreateRescheduleRequestCommand) (core.RescheduleRequest, error) {
	principal := command.Principal
	var request core.RescheduleRequest
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_reschedule_request", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			request, err = decodeReplay[core.RescheduleRequest](claim)
			return err
		}
		var status, teacherAccountID string
		var startsAt time.Time
		err = tx.QueryRow(ctx, `
			SELECT status, teacher_account_id, starts_at
			FROM core_lesson_occurrences
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.OccurrenceID).Scan(&status, &teacherAccountID, &startsAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "Lesson not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock occurrence for request: %w", err)
		}
		if status != "scheduled" || !startsAt.After(command.Now) {
			return core.E(core.CodeInvalidState, "only a scheduled future Lesson can be changed", nil)
		}
		isTeacher := teacherAccountID == principal.AccountID
		if !isTeacher {
			var participates bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1
					FROM core_lesson_occurrence_participants participant
					JOIN students s
					  ON s.tenant_id = participant.tenant_id AND s.id = participant.student_id
					WHERE participant.tenant_id = $1 AND participant.occurrence_id = $2
					  AND s.account_id = $3
				)
			`, principal.TenantID, command.OccurrenceID, principal.AccountID).Scan(&participates); err != nil {
				return fmt.Errorf("check request authority: %w", err)
			}
			if !participates {
				return core.E(core.CodeForbidden, "only a participant or the Lesson's Teacher can ask for a change", nil)
			}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO lesson_reschedule_requests (
				id, tenant_id, occurrence_id, requested_by_account_id,
				kind, proposed_starts_at, reason, status, version, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', 0, $8, $8)
		`, command.RequestID, principal.TenantID, command.OccurrenceID, principal.AccountID,
			command.Kind, command.ProposedStartsAt, command.Reason, command.Now); err != nil {
			return mapWriteError(err, "an open request already exists for this Lesson")
		}
		request, err = readRescheduleRequest(ctx, tx, principal.TenantID, command.RequestID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_reschedule_request", command.IdempotencyKey, request, command.Now); err != nil {
			return err
		}
		action := "LessonRescheduleRequested"
		if command.Kind == "cancellation" {
			action = "LessonCancellationRequested"
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "lesson_occurrence", targetID: command.OccurrenceID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"requestId": command.RequestID}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, action, "reschedule_request", command.RequestID,
			map[string]any{"requestId": command.RequestID, "occurrenceId": command.OccurrenceID}, command.Now)
	})
	if err != nil {
		return core.RescheduleRequest{}, err
	}
	return request, nil
}

func readRescheduleRequest(ctx context.Context, reader lessonReader, tenantID, requestID string) (core.RescheduleRequest, error) {
	var request core.RescheduleRequest
	var proposed, decidedAt *time.Time
	var decisionNote *string
	err := reader.QueryRow(ctx, `
		SELECT r.id, r.occurrence_id, r.kind, r.proposed_starts_at, r.reason, r.status,
		       r.requested_by_account_id, requester_person.full_name,
		       r.decision_note, r.decided_at, r.created_at, r.version
		FROM lesson_reschedule_requests r
		JOIN accounts requester_account
		  ON requester_account.tenant_id = r.tenant_id AND requester_account.id = r.requested_by_account_id
		JOIN people requester_person
		  ON requester_person.tenant_id = requester_account.tenant_id
		 AND requester_person.id = requester_account.person_id
		WHERE r.tenant_id = $1 AND r.id = $2
	`, tenantID, requestID).Scan(
		&request.ID, &request.OccurrenceID, &request.Kind, &proposed, &request.Reason, &request.Status,
		&request.RequestedBy.AccountID, &request.RequestedBy.FullName,
		&decisionNote, &decidedAt, &request.CreatedAt, &request.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.RescheduleRequest{}, core.E(core.CodeNotFound, "request not found", nil)
	}
	if err != nil {
		return core.RescheduleRequest{}, fmt.Errorf("read reschedule request: %w", err)
	}
	request.CreatedAt = request.CreatedAt.UTC()
	if proposed != nil {
		utc := proposed.UTC()
		request.ProposedStartsAt = &utc
	}
	if decisionNote != nil {
		request.DecisionNote = *decisionNote
	}
	if decidedAt != nil {
		utc := decidedAt.UTC()
		request.DecidedAt = &utc
	}
	return request, nil
}

func (s *Store) ListRescheduleRequests(ctx context.Context, principal core.Principal) ([]core.RescheduleRequest, error) {
	var manager bool
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		authority, err := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if err != nil {
			return err
		}
		manager = authority
		return nil
	}); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM lesson_reschedule_requests
		WHERE tenant_id = $1 AND ($2::boolean OR requested_by_account_id = $3)
		ORDER BY (status = 'pending') DESC, created_at DESC, id
		LIMIT 100
	`, principal.TenantID, manager, principal.AccountID)
	if err != nil {
		return nil, fmt.Errorf("list reschedule requests: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan reschedule request id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reschedule request ids: %w", err)
	}
	rows.Close()
	result := make([]core.RescheduleRequest, 0, len(ids))
	for _, id := range ids {
		request, err := readRescheduleRequest(ctx, s.pool, principal.TenantID, id)
		if err != nil {
			return nil, err
		}
		result = append(result, request)
	}
	return result, nil
}

func (s *Store) DecideRescheduleRequest(ctx context.Context, command core.DecideRescheduleRequestCommand) (core.RescheduleRequest, error) {
	principal := command.Principal
	var request core.RescheduleRequest
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, err := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if err != nil {
			return err
		}
		if !manager {
			return core.E(core.CodeForbidden, "schedule decision permission is required", nil)
		}
		var occurrenceID, kind, status, requesterAccountID string
		var proposed *time.Time
		var version int64
		err = tx.QueryRow(ctx, `
			SELECT occurrence_id, kind, status, requested_by_account_id, proposed_starts_at, version
			FROM lesson_reschedule_requests
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.RequestID).Scan(
			&occurrenceID, &kind, &status, &requesterAccountID, &proposed, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "request not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock reschedule request: %w", err)
		}
		if status != "pending" {
			return core.E(core.CodeInvalidState, "request is already resolved", nil)
		}
		if version != command.ExpectedVersion {
			return core.E(core.CodeConflict, "request version is stale", nil)
		}
		if command.Approve {
			var occurrenceStatus, teacherAccountID string
			var durationMinutes int
			err = tx.QueryRow(ctx, `
				SELECT status, teacher_account_id, duration_minutes
				FROM core_lesson_occurrences
				WHERE tenant_id = $1 AND id = $2
				FOR UPDATE
			`, principal.TenantID, occurrenceID).Scan(&occurrenceStatus, &teacherAccountID, &durationMinutes)
			if err != nil {
				return fmt.Errorf("lock occurrence for decision: %w", err)
			}
			if occurrenceStatus != "scheduled" {
				return core.E(core.CodeInvalidState, "Lesson is no longer scheduled", nil)
			}
			if kind == "reschedule" {
				if proposed == nil || !proposed.After(command.Now) {
					return core.E(core.CodeInvalidState, "proposed time is already in the past", nil)
				}
				studentRows, err := tx.Query(ctx, `
					SELECT student_id FROM core_lesson_occurrence_participants
					WHERE tenant_id = $1 AND occurrence_id = $2
					ORDER BY student_id
				`, principal.TenantID, occurrenceID)
				if err != nil {
					return fmt.Errorf("read occurrence participants: %w", err)
				}
				studentIDs := []string{}
				for studentRows.Next() {
					var studentID string
					if err := studentRows.Scan(&studentID); err != nil {
						studentRows.Close()
						return fmt.Errorf("scan occurrence participant: %w", err)
					}
					studentIDs = append(studentIDs, studentID)
				}
				studentRows.Close()
				if err := studentRows.Err(); err != nil {
					return fmt.Errorf("iterate occurrence participants: %w", err)
				}
				conflict, err := lessonScheduleConflict(ctx, tx, principal.TenantID, *proposed, durationMinutes, teacherAccountID, studentIDs, []string{occurrenceID})
				if err != nil {
					return err
				}
				if conflict {
					return core.E(core.CodeConflict, "proposed time overlaps an existing Lesson", nil)
				}
				if _, err := tx.Exec(ctx, `
					UPDATE core_lesson_occurrences
					SET starts_at = $3, version = version + 1, updated_at = $4
					WHERE tenant_id = $1 AND id = $2
				`, principal.TenantID, occurrenceID, *proposed, command.Now); err != nil {
					return fmt.Errorf("apply reschedule: %w", err)
				}
			} else {
				// DEC-102 guard: cancellation records who initiated and
				// applies the status; no consequence is computed anywhere.
				cancelledStatus := "cancelled_student"
				if requesterAccountID == teacherAccountID {
					cancelledStatus = "cancelled_school"
				}
				if _, err := tx.Exec(ctx, `
					UPDATE core_lesson_occurrences
					SET status = $3, version = version + 1, updated_at = $4
					WHERE tenant_id = $1 AND id = $2
				`, principal.TenantID, occurrenceID, cancelledStatus, command.Now); err != nil {
					return fmt.Errorf("apply cancellation: %w", err)
				}
			}
		}
		decidedStatus := "declined"
		if command.Approve {
			decidedStatus = "approved"
		}
		if _, err := tx.Exec(ctx, `
			UPDATE lesson_reschedule_requests
			SET status = $3, decided_by_account_id = $4, decision_note = NULLIF($5, ''),
			    decided_at = $6, version = version + 1, updated_at = $6
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.RequestID, decidedStatus, principal.AccountID,
			command.DecisionNote, command.Now); err != nil {
			return fmt.Errorf("record decision: %w", err)
		}
		request, err = readRescheduleRequest(ctx, tx, principal.TenantID, command.RequestID)
		if err != nil {
			return err
		}
		action := "LessonRescheduleDeclined"
		if command.Approve && kind == "reschedule" {
			action = "LessonRescheduled"
		} else if command.Approve {
			action = "LessonCancelled"
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "lesson_occurrence", targetID: occurrenceID,
			decision: "allow",
			metadata: map[string]any{"requestId": command.RequestID}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, action, "lesson_occurrence", occurrenceID,
			map[string]any{"requestId": command.RequestID, "occurrenceId": occurrenceID}, command.Now)
	})
	if err != nil {
		return core.RescheduleRequest{}, err
	}
	return request, nil
}

func (s *Store) WithdrawRescheduleRequest(ctx context.Context, command core.WithdrawRescheduleRequestCommand) (core.RescheduleRequest, error) {
	principal := command.Principal
	var request core.RescheduleRequest
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		updated, err := tx.Exec(ctx, `
			UPDATE lesson_reschedule_requests
			SET status = 'withdrawn', version = version + 1, updated_at = $4
			WHERE tenant_id = $1 AND id = $2 AND requested_by_account_id = $3 AND status = 'pending'
		`, principal.TenantID, command.RequestID, principal.AccountID, command.Now)
		if err != nil {
			return fmt.Errorf("withdraw request: %w", err)
		}
		if updated.RowsAffected() == 0 {
			return core.E(core.CodeInvalidState, "no pending request of yours to withdraw", nil)
		}
		request, err = readRescheduleRequest(ctx, tx, principal.TenantID, command.RequestID)
		if err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "LessonRescheduleWithdrawn", targetType: "reschedule_request", targetID: command.RequestID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.RescheduleRequest{}, err
	}
	return request, nil
}
