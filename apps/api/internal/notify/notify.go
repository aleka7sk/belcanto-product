// Package notify is the outbox worker: it drains pending outbox events,
// routes a curated subset to per-recipient in-app activity entries and
// records delivery status with a retry policy — an outbox row without a
// worker, retry and delivery status is not a notification flow.
package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// Store is the slice of the application store the worker needs.
type Store interface {
	PendingOutboxEvents(context.Context, int, time.Time) ([]core.OutboxEvent, error)
	DeliverOutboxEvent(context.Context, core.DeliverOutboxCommand) error
	FailOutboxEvent(context.Context, core.FailOutboxCommand) error
	AccountIDForStudent(context.Context, string, string) (string, error)
	HomeworkTeacherAccountID(context.Context, string, string) (string, error)
	AdministratorAccountIDs(context.Context, string) ([]string, error)
}

// Resolver turns event payload references into recipient accounts.
type Resolver interface {
	AccountIDForStudent(ctx context.Context, tenantID, studentID string) (string, error)
	HomeworkTeacherAccountID(ctx context.Context, tenantID, homeworkID string) (string, error)
	AdministratorAccountIDs(ctx context.Context, tenantID string) ([]string, error)
}

type route struct {
	category  string
	recipient func(ctx context.Context, resolver Resolver, tenantID string, payload map[string]string) ([]string, error)
}

func studentRecipient(ctx context.Context, resolver Resolver, tenantID string, payload map[string]string) ([]string, error) {
	studentID := payload["studentId"]
	if studentID == "" {
		return nil, fmt.Errorf("payload has no studentId")
	}
	accountID, err := resolver.AccountIDForStudent(ctx, tenantID, studentID)
	if err != nil {
		return nil, err
	}
	if accountID == "" {
		// The student has not activated an account yet — nobody to tell.
		return nil, nil
	}
	return []string{accountID}, nil
}

func homeworkTeacherRecipient(ctx context.Context, resolver Resolver, tenantID string, payload map[string]string) ([]string, error) {
	homeworkID := payload["homeworkId"]
	if homeworkID == "" {
		return nil, fmt.Errorf("payload has no homeworkId")
	}
	accountID, err := resolver.HomeworkTeacherAccountID(ctx, tenantID, homeworkID)
	if err != nil {
		return nil, err
	}
	return []string{accountID}, nil
}

func administratorRecipients(ctx context.Context, resolver Resolver, tenantID string, _ map[string]string) ([]string, error) {
	return resolver.AdministratorAccountIDs(ctx, tenantID)
}

// Routes: the curated event → recipients/category table (ACT-02
// categories). Events outside the table deliver with no recipients —
// not every domain event is a notification.
var routes = map[string]route{
	"JournalPublished":          {category: "learning", recipient: studentRecipient},
	"JournalCorrected":          {category: "learning", recipient: studentRecipient},
	"HomeworkAssigned":          {category: "learning", recipient: studentRecipient},
	"HomeworkStarted":           {category: "learning", recipient: homeworkTeacherRecipient},
	"HomeworkSubmitted":         {category: "learning", recipient: homeworkTeacherRecipient},
	"HomeworkReviewed":          {category: "learning", recipient: studentRecipient},
	"HomeworkCompleted":         {category: "learning", recipient: studentRecipient},
	"HomeworkCancelled":         {category: "learning", recipient: studentRecipient},
	"GoalCompleted":             {category: "learning", recipient: studentRecipient},
	"AchievementAwarded":        {category: "learning", recipient: studentRecipient},
	"SongStageChanged":          {category: "learning", recipient: studentRecipient},
	"AttendanceAbsenceRecorded": {category: "important", recipient: administratorRecipients},
}

// RouteEvent resolves the activity entries an event produces.
func RouteEvent(ctx context.Context, resolver Resolver, event core.OutboxEvent) ([]core.ActivityInsert, error) {
	table, known := routes[event.EventType]
	if !known {
		return nil, nil
	}
	payload := map[string]string{}
	raw := map[string]any{}
	if len(event.Payload) > 0 {
		if err := json.Unmarshal(event.Payload, &raw); err != nil {
			return nil, fmt.Errorf("decode payload: %w", err)
		}
	}
	for key, value := range raw {
		if text, ok := value.(string); ok {
			payload[key] = text
		}
	}
	recipients, err := table.recipient(ctx, resolver, event.TenantID, payload)
	if err != nil {
		return nil, err
	}
	entries := make([]core.ActivityInsert, 0, len(recipients))
	for _, recipient := range recipients {
		entryID, err := security.NewID("act")
		if err != nil {
			return nil, fmt.Errorf("create activity id: %w", err)
		}
		entries = append(entries, core.ActivityInsert{
			EntryID:            entryID,
			RecipientAccountID: recipient,
			Category:           table.category,
			Kind:               event.EventType,
			TargetType:         event.AggregateType,
			TargetID:           event.AggregateID,
			Payload:            event.Payload,
		})
	}
	return entries, nil
}

// Options tune the worker; zero values take deploy-side defaults.
type Options struct {
	PollInterval time.Duration
	BatchSize    int
	MaxAttempts  int
	BackoffBase  time.Duration
	Logger       *slog.Logger
	Clock        func() time.Time
}

type Worker struct {
	store       Store
	interval    time.Duration
	batchSize   int
	maxAttempts int
	backoffBase time.Duration
	logger      *slog.Logger
	clock       func() time.Time
}

func NewWorker(store Store, options Options) *Worker {
	interval := options.PollInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = 50
	}
	maxAttempts := options.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 8
	}
	backoffBase := options.BackoffBase
	if backoffBase <= 0 {
		backoffBase = 30 * time.Second
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	clock := options.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Worker{
		store: store, interval: interval, batchSize: batchSize,
		maxAttempts: maxAttempts, backoffBase: backoffBase,
		logger: logger, clock: clock,
	}
}

// Backoff doubles per attempt from the base, capped at one hour.
func (w *Worker) Backoff(attempt int) time.Duration {
	delay := w.backoffBase
	for count := 0; count < attempt && delay < time.Hour; count++ {
		delay *= 2
	}
	if delay > time.Hour {
		delay = time.Hour
	}
	return delay
}

// DrainOnce processes one batch and reports processed/failed counts.
func (w *Worker) DrainOnce(ctx context.Context) (processed, failed int, err error) {
	now := w.clock().UTC()
	events, err := w.store.PendingOutboxEvents(ctx, w.batchSize, now)
	if err != nil {
		return 0, 0, err
	}
	for _, event := range events {
		entries, routeErr := RouteEvent(ctx, w.store, event)
		if routeErr == nil {
			routeErr = w.store.DeliverOutboxEvent(ctx, core.DeliverOutboxCommand{
				EventID: event.ID, Tenant: event.TenantID,
				Entries: entries, Now: w.clock().UTC(),
			})
		}
		if routeErr == nil {
			processed++
			continue
		}
		failed++
		attempt := event.AttemptCount + 1
		dead := attempt >= w.maxAttempts
		var nextAttempt *time.Time
		if !dead {
			at := w.clock().UTC().Add(w.Backoff(event.AttemptCount))
			nextAttempt = &at
		}
		message := routeErr.Error()
		if len(message) > 500 {
			message = message[:500]
		}
		if failErr := w.store.FailOutboxEvent(ctx, core.FailOutboxCommand{
			EventID: event.ID, ErrorMessage: message,
			NextAttemptAt: nextAttempt, DeadLetter: dead,
			Now: w.clock().UTC(),
		}); failErr != nil {
			w.logger.Error("outbox fail-mark failed", "eventId", event.ID, "error", failErr)
		}
		if dead {
			w.logger.Error("outbox event dead-lettered", "eventId", event.ID,
				"eventType", event.EventType, "error", message)
		} else {
			w.logger.Warn("outbox event delivery failed", "eventId", event.ID,
				"eventType", event.EventType, "attempt", attempt, "error", message)
		}
	}
	return processed, failed, nil
}

// Run polls until the context ends.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processed, failed, err := w.DrainOnce(ctx)
			if err != nil {
				w.logger.Error("outbox drain failed", "error", err)
				continue
			}
			if processed > 0 || failed > 0 {
				w.logger.Info("outbox drained", "delivered", processed, "failed", failed)
			}
		}
	}
}
