package memory

import (
	"context"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.5 outbox delivery, activity feed and notification preferences —
// parity with PostgreSQL.

type activityEntry struct {
	ID           string
	TenantID     string
	Recipient    string
	SourceOutbox int64
	Category     string
	Kind         string
	TargetType   string
	TargetID     string
	Payload      []byte
	OccurredAt   time.Time
	ReadAt       *time.Time
}

func (s *Store) PendingOutboxEvents(_ context.Context, limit int, now time.Time) ([]core.OutboxEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.OutboxEvent{}
	for index := range s.outbox {
		record := &s.outbox[index]
		if record.Status != "pending" {
			continue
		}
		if record.NextAttemptAt != nil && record.NextAttemptAt.After(now) {
			continue
		}
		result = append(result, core.OutboxEvent{
			ID: record.ID, TenantID: record.TenantID, EventType: record.EventType,
			AggregateType: record.AggregateType, AggregateID: record.AggregateID,
			Payload: record.Payload, RecordedAt: record.RecordedAt,
			AttemptCount: record.AttemptCount,
		})
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (s *Store) DeliverOutboxEvent(_ context.Context, command core.DeliverOutboxCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var record *OutboxRecord
	for index := range s.outbox {
		if s.outbox[index].ID == command.EventID {
			record = &s.outbox[index]
			break
		}
	}
	if record == nil || record.Status != "pending" {
		return core.E(core.CodeConflict, "the outbox event is no longer pending", nil)
	}
	for _, entry := range command.Entries {
		duplicate := false
		for _, existing := range s.activity {
			if existing.TenantID == command.Tenant &&
				existing.SourceOutbox == command.EventID &&
				existing.Recipient == entry.RecipientAccountID {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		s.activity = append(s.activity, &activityEntry{
			ID: entry.EntryID, TenantID: command.Tenant,
			Recipient: entry.RecipientAccountID, SourceOutbox: command.EventID,
			Category: entry.Category, Kind: entry.Kind,
			TargetType: entry.TargetType, TargetID: entry.TargetID,
			Payload: entry.Payload, OccurredAt: command.Now,
		})
	}
	record.Status = "delivered"
	record.LastError = ""
	record.NextAttemptAt = nil
	return nil
}

func (s *Store) FailOutboxEvent(_ context.Context, command core.FailOutboxCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.outbox {
		record := &s.outbox[index]
		if record.ID != command.EventID || record.Status != "pending" {
			continue
		}
		record.AttemptCount++
		record.LastError = command.ErrorMessage
		record.NextAttemptAt = command.NextAttemptAt
		if command.DeadLetter {
			record.Status = "dead_letter"
		}
		return nil
	}
	return nil
}

func (s *Store) AccountIDForStudent(_ context.Context, tenantID, studentID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.students[studentID]
	if stored == nil || stored.TenantID != tenantID {
		return "", core.E(core.CodeNotFound, "Student not found", nil)
	}
	return stored.AccountID, nil
}

func (s *Store) HomeworkTeacherAccountID(_ context.Context, tenantID, homeworkID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.homework[homeworkID]
	if stored == nil || stored.TenantID != tenantID {
		return "", core.E(core.CodeNotFound, "homework not found", nil)
	}
	return stored.TeacherAccountID, nil
}

func (s *Store) AdministratorAccountIDs(_ context.Context, tenantID string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []string{}
	for _, account := range s.accounts {
		if account.TenantID == tenantID && account.Status == "active" &&
			account.Roles[core.RoleAdministrator] != "" {
			result = append(result, account.ID)
		}
	}
	sort.Strings(result)
	return result, nil
}

func (s *Store) ActivityFeed(_ context.Context, principal core.Principal, limit int) (core.ActivityFeed, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	feed := core.ActivityFeed{Entries: []core.ActivityEntry{}}
	entries := []*activityEntry{}
	for _, entry := range s.activity {
		if entry.TenantID != principal.TenantID || entry.Recipient != principal.AccountID {
			continue
		}
		if entry.ReadAt == nil {
			feed.UnreadCount++
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(left, right int) bool {
		if !entries[left].OccurredAt.Equal(entries[right].OccurredAt) {
			return entries[left].OccurredAt.After(entries[right].OccurredAt)
		}
		return entries[left].ID < entries[right].ID
	})
	for index, entry := range entries {
		if index >= limit {
			break
		}
		view := core.ActivityEntry{
			ID: entry.ID, Category: entry.Category, Kind: entry.Kind,
			TargetType: entry.TargetType, TargetID: entry.TargetID,
			Payload: entry.Payload, OccurredAt: entry.OccurredAt,
		}
		if entry.ReadAt != nil {
			read := *entry.ReadAt
			view.ReadAt = &read
		}
		feed.Entries = append(feed.Entries, view)
	}
	return feed, nil
}

func (s *Store) MarkActivityRead(_ context.Context, principal core.Principal, upTo, now time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	marked := 0
	for _, entry := range s.activity {
		if entry.TenantID != principal.TenantID || entry.Recipient != principal.AccountID {
			continue
		}
		if entry.ReadAt == nil && !entry.OccurredAt.After(upTo) {
			readAt := now
			entry.ReadAt = &readAt
			marked++
		}
	}
	return marked, nil
}

func (s *Store) NotificationPreferences(_ context.Context, principal core.Principal) ([]core.NotificationPreference, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return nil, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	result := make([]core.NotificationPreference, 0, len(core.NotificationCategories))
	for _, category := range core.NotificationCategories {
		enabled, ok := s.notificationPrefs[principal.AccountID+"\x00"+category]
		if !ok {
			enabled = true
		}
		result = append(result, core.NotificationPreference{Category: category, PushEnabled: enabled})
	}
	return result, nil
}

func (s *Store) UpdateNotificationPreference(ctx context.Context, command core.UpdateNotificationPreferenceCommand) ([]core.NotificationPreference, error) {
	s.mu.Lock()
	principal := command.Principal
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		s.mu.Unlock()
		return nil, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	s.notificationPrefs[principal.AccountID+"\x00"+command.Category] = command.PushEnabled
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "NotificationPreferenceUpdated",
		"notification_preference", command.Category, "allow", "", command.Now, nil)
	s.mu.Unlock()
	return s.NotificationPreferences(ctx, principal)
}

// AppendBrokenOutboxEvent is a test hook: an event whose payload cannot
// be routed, to exercise the worker's retry and dead-letter path.
func (s *Store) AppendBrokenOutboxEvent(tenantID, eventType string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appendOutboxPayload(tenantID, eventType, "homework", "hw_broken", map[string]any{}, at)
}
