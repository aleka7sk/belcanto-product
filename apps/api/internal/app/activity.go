package app

import (
	"context"
	"slices"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.5 activity feed and notification preferences (Page 31). The in-app
// channel is always on; preferences gate the future push channel.

const activityFeedLimit = 100

func (s *Service) ActivityFeed(ctx context.Context, principal core.Principal) (core.ActivityFeed, error) {
	feed, err := s.store.ActivityFeed(ctx, principal, activityFeedLimit)
	if err != nil {
		return core.ActivityFeed{}, normalizeStoreError("read activity", err)
	}
	return feed, nil
}

func (s *Service) MarkActivityRead(ctx context.Context, principal core.Principal, upTo time.Time) (int, error) {
	if upTo.IsZero() {
		return 0, core.E(core.CodeInvalidInput, "upTo must be a timestamp", nil)
	}
	marked, err := s.store.MarkActivityRead(ctx, principal, upTo, s.clock.Now())
	if err != nil {
		return 0, normalizeStoreError("mark activity read", err)
	}
	return marked, nil
}

func (s *Service) NotificationPreferences(ctx context.Context, principal core.Principal) ([]core.NotificationPreference, error) {
	preferences, err := s.store.NotificationPreferences(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("read notification preferences", err)
	}
	return preferences, nil
}

func (s *Service) UpdateNotificationPreference(ctx context.Context, principal core.Principal, category string, pushEnabled bool) ([]core.NotificationPreference, error) {
	if !slices.Contains(core.NotificationCategories, category) {
		return nil, core.E(core.CodeInvalidInput, "category must be important, learning, messages or community", nil)
	}
	preferences, err := s.store.UpdateNotificationPreference(ctx, core.UpdateNotificationPreferenceCommand{
		Principal: principal, Category: category, PushEnabled: pushEnabled,
		Now: s.clock.Now(),
	})
	if err != nil {
		return nil, normalizeStoreError("update notification preference", err)
	}
	return preferences, nil
}
