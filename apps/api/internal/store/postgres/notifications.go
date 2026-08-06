package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.5 outbox delivery, activity feed and notification preferences.
// Delivery is idempotent per (outbox event, recipient); the in-app
// channel is always on and preferences gate the future push channel.

func (s *Store) PendingOutboxEvents(ctx context.Context, limit int, now time.Time) ([]core.OutboxEvent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, event_type, aggregate_type, aggregate_id,
		       payload, recorded_at, attempt_count
		FROM outbox_events
		WHERE status = 'pending'
		  AND (next_attempt_at IS NULL OR next_attempt_at <= $2)
		ORDER BY id
		LIMIT $1
	`, limit, now)
	if err != nil {
		return nil, fmt.Errorf("list pending outbox events: %w", err)
	}
	defer rows.Close()
	result := []core.OutboxEvent{}
	for rows.Next() {
		var event core.OutboxEvent
		if err := rows.Scan(&event.ID, &event.TenantID, &event.EventType,
			&event.AggregateType, &event.AggregateID, &event.Payload,
			&event.RecordedAt, &event.AttemptCount); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		event.RecordedAt = event.RecordedAt.UTC()
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox events: %w", err)
	}
	return result, nil
}

func (s *Store) DeliverOutboxEvent(ctx context.Context, command core.DeliverOutboxCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, entry := range command.Entries {
			if _, err := tx.Exec(ctx, `
				INSERT INTO activity_entries (
					id, tenant_id, recipient_account_id, source_outbox_id,
					category, kind, target_type, target_id, payload, occurred_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10)
				ON CONFLICT (tenant_id, source_outbox_id, recipient_account_id) DO NOTHING
			`, entry.EntryID, command.Tenant, entry.RecipientAccountID, command.EventID,
				entry.Category, entry.Kind, entry.TargetType, entry.TargetID,
				entry.Payload, command.Now); err != nil {
				return mapWriteError(err, "activity entry conflicts with existing data")
			}
		}
		tag, err := tx.Exec(ctx, `
			UPDATE outbox_events
			SET status = 'delivered', published_at = $2, last_error = NULL, next_attempt_at = NULL
			WHERE id = $1 AND status = 'pending'
		`, command.EventID, command.Now)
		if err != nil {
			return fmt.Errorf("mark outbox delivered: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return core.E(core.CodeConflict, "the outbox event is no longer pending", nil)
		}
		return nil
	})
}

func (s *Store) FailOutboxEvent(ctx context.Context, command core.FailOutboxCommand) error {
	status := "pending"
	if command.DeadLetter {
		status = "dead_letter"
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE outbox_events
		SET attempt_count = attempt_count + 1,
		    last_error = $2,
		    next_attempt_at = $3,
		    status = $4
		WHERE id = $1 AND status = 'pending'
	`, command.EventID, command.ErrorMessage, command.NextAttemptAt, status)
	if err != nil {
		return fmt.Errorf("mark outbox failed: %w", err)
	}
	return nil
}

func (s *Store) AccountIDForStudent(ctx context.Context, tenantID, studentID string) (string, error) {
	var accountID *string
	err := s.pool.QueryRow(ctx, `
		SELECT account_id FROM students
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, studentID).Scan(&accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", core.E(core.CodeNotFound, "Student not found", nil)
	}
	if err != nil {
		return "", fmt.Errorf("read student account: %w", err)
	}
	if accountID == nil {
		return "", nil
	}
	return *accountID, nil
}

func (s *Store) HomeworkTeacherAccountID(ctx context.Context, tenantID, homeworkID string) (string, error) {
	var teacherAccountID string
	err := s.pool.QueryRow(ctx, `
		SELECT teacher_account_id FROM homework_assignments
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, homeworkID).Scan(&teacherAccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", core.E(core.CodeNotFound, "homework not found", nil)
	}
	if err != nil {
		return "", fmt.Errorf("read homework teacher: %w", err)
	}
	return teacherAccountID, nil
}

func (s *Store) AdministratorAccountIDs(ctx context.Context, tenantID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT account_id FROM role_grants
		WHERE tenant_id = $1 AND role_type = 'Administrator' AND status = 'active'
		ORDER BY account_id
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list administrators: %w", err)
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var accountID string
		if err := rows.Scan(&accountID); err != nil {
			return nil, fmt.Errorf("scan administrator: %w", err)
		}
		result = append(result, accountID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate administrators: %w", err)
	}
	return result, nil
}

func (s *Store) ActivityFeed(ctx context.Context, principal core.Principal, limit int) (core.ActivityFeed, error) {
	feed := core.ActivityFeed{Entries: []core.ActivityEntry{}}
	if err := s.pool.QueryRow(ctx, `
		SELECT count(*) FROM activity_entries
		WHERE tenant_id = $1 AND recipient_account_id = $2 AND read_at IS NULL
	`, principal.TenantID, principal.AccountID).Scan(&feed.UnreadCount); err != nil {
		return core.ActivityFeed{}, fmt.Errorf("count unread activity: %w", err)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, category, kind, target_type, target_id, payload, occurred_at, read_at
		FROM activity_entries
		WHERE tenant_id = $1 AND recipient_account_id = $2
		ORDER BY occurred_at DESC, id
		LIMIT $3
	`, principal.TenantID, principal.AccountID, limit)
	if err != nil {
		return core.ActivityFeed{}, fmt.Errorf("list activity: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var entry core.ActivityEntry
		if err := rows.Scan(&entry.ID, &entry.Category, &entry.Kind,
			&entry.TargetType, &entry.TargetID, &entry.Payload,
			&entry.OccurredAt, &entry.ReadAt); err != nil {
			return core.ActivityFeed{}, fmt.Errorf("scan activity entry: %w", err)
		}
		entry.OccurredAt = entry.OccurredAt.UTC()
		if entry.ReadAt != nil {
			utc := entry.ReadAt.UTC()
			entry.ReadAt = &utc
		}
		feed.Entries = append(feed.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		return core.ActivityFeed{}, fmt.Errorf("iterate activity: %w", err)
	}
	return feed, nil
}

func (s *Store) MarkActivityRead(ctx context.Context, principal core.Principal, upTo, now time.Time) (int, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE activity_entries
		SET read_at = $4
		WHERE tenant_id = $1 AND recipient_account_id = $2
		  AND read_at IS NULL AND occurred_at <= $3
	`, principal.TenantID, principal.AccountID, upTo, now)
	if err != nil {
		return 0, fmt.Errorf("mark activity read: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func notificationPreferenceDefaults() []core.NotificationPreference {
	result := make([]core.NotificationPreference, 0, len(core.NotificationCategories))
	for _, category := range core.NotificationCategories {
		result = append(result, core.NotificationPreference{Category: category, PushEnabled: true})
	}
	return result
}

func (s *Store) NotificationPreferences(ctx context.Context, principal core.Principal) ([]core.NotificationPreference, error) {
	if err := activeAccountExistsPool(ctx, s.pool, principal.TenantID, principal.AccountID); err != nil {
		return nil, err
	}
	stored := map[string]bool{}
	rows, err := s.pool.Query(ctx, `
		SELECT category, push_enabled FROM notification_preferences
		WHERE tenant_id = $1 AND account_id = $2
	`, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, fmt.Errorf("list notification preferences: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var category string
		var enabled bool
		if err := rows.Scan(&category, &enabled); err != nil {
			return nil, fmt.Errorf("scan notification preference: %w", err)
		}
		stored[category] = enabled
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification preferences: %w", err)
	}
	result := notificationPreferenceDefaults()
	for index := range result {
		if enabled, ok := stored[result[index].Category]; ok {
			result[index].PushEnabled = enabled
		}
	}
	return result, nil
}

func (s *Store) UpdateNotificationPreference(ctx context.Context, command core.UpdateNotificationPreferenceCommand) ([]core.NotificationPreference, error) {
	principal := command.Principal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := activeAccountExists(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO notification_preferences (
				tenant_id, account_id, category, push_enabled, updated_at
			) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tenant_id, account_id, category)
			DO UPDATE SET push_enabled = EXCLUDED.push_enabled,
			              updated_at = EXCLUDED.updated_at
		`, principal.TenantID, principal.AccountID, command.Category,
			command.PushEnabled, command.Now); err != nil {
			return mapWriteError(err, "notification preference conflicts with existing data")
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "NotificationPreferenceUpdated", targetType: "notification_preference",
			targetID: command.Category, decision: "allow",
			metadata: map[string]any{"pushEnabled": command.PushEnabled},
			at:       command.Now,
		})
	})
	if err != nil {
		return nil, err
	}
	return s.NotificationPreferences(ctx, principal)
}
