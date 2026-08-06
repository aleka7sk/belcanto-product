package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.2 events and RSVP. Every seat mutation serializes on a per-occurrence
// advisory lock, first expires overdue spot offers and cascades freed
// seats to the waitlist head, then applies the caller's change. The DB
// triggers stay as the second line of defence for the capacity invariant.

func (s *Store) eventManagerAuthority(ctx context.Context, tx pgx.Tx, tenantID, accountID string) error {
	manager, err := lessonManagementAuthority(ctx, tx, tenantID, accountID)
	if err != nil {
		return err
	}
	if !manager {
		return core.E(core.CodeForbidden, "event management permission is required", nil)
	}
	return nil
}

func (s *Store) CreateEventCategory(ctx context.Context, command core.CreateEventCategoryCommand) (core.EventCategory, error) {
	principal := command.Principal
	var category core.EventCategory
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.eventManagerAuthority(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_categories (id, tenant_id, name, status, created_at)
			VALUES ($1, $2, $3, 'active', $4)
		`, command.CategoryID, principal.TenantID, command.Name, command.Now); err != nil {
			return mapWriteError(err, "event category name is already in use")
		}
		category = core.EventCategory{ID: command.CategoryID, Name: command.Name, Status: "active"}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "EventCategoryCreated", targetType: "event_category", targetID: command.CategoryID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.EventCategory{}, err
	}
	return category, nil
}

func (s *Store) ListEventCategories(ctx context.Context, principal core.Principal) ([]core.EventCategory, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, status FROM event_categories
		WHERE tenant_id = $1 AND status = 'active'
		ORDER BY name, id
	`, principal.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list event categories: %w", err)
	}
	defer rows.Close()
	categories := []core.EventCategory{}
	for rows.Next() {
		var category core.EventCategory
		if err := rows.Scan(&category.ID, &category.Name, &category.Status); err != nil {
			return nil, fmt.Errorf("scan event category: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event categories: %w", err)
	}
	return categories, nil
}

func (s *Store) CreateEventSeries(ctx context.Context, command core.CreateEventSeriesCommand) (core.EventSeries, error) {
	var series core.EventSeries
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.eventManagerAuthority(ctx, tx, command.TenantID, command.ActorAccountID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_event_series", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			series, err = decodeReplay[core.EventSeries](claim)
			return err
		}
		if err := checkEventReferences(ctx, tx, command.TenantID, command.CategoryID, command.HostAccountID, command.RoomID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_series (
				id, tenant_id, category_id, title, description, host_account_id, room_id,
				capacity, weekday, start_minutes, duration_minutes,
				effective_from, effective_until, status, version,
				created_by_account_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, NULLIF($7, ''),
			          $8, $9, $10, $11, $12::date, NULLIF($13, '')::date, 'active', 0,
			          $14, $15, $15)
		`, command.SeriesID, command.TenantID, command.CategoryID, command.Title,
			command.Description, command.HostAccountID, command.RoomID,
			command.Capacity, command.Weekday, command.StartMinutes, command.DurationMinutes,
			command.EffectiveFrom, command.EffectiveUntil,
			command.ActorAccountID, command.Now); err != nil {
			return mapWriteError(err, "event series conflicts with existing data")
		}
		series, err = readEventSeries(ctx, tx, command.TenantID, command.SeriesID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_event_series", command.IdempotencyKey, series, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "EventSeriesCreated", targetType: "event_series", targetID: command.SeriesID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"hostAccountId": command.HostAccountID}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "EventSeriesCreated", "event_series", command.SeriesID,
			map[string]any{"seriesId": command.SeriesID}, command.Now)
	})
	if err != nil {
		return core.EventSeries{}, err
	}
	return series, nil
}

func checkEventReferences(ctx context.Context, tx pgx.Tx, tenantID, categoryID, hostAccountID, roomID string) error {
	var categoryActive bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM event_categories WHERE tenant_id = $1 AND id = $2 AND status = 'active')
	`, tenantID, categoryID).Scan(&categoryActive); err != nil {
		return fmt.Errorf("check event category: %w", err)
	}
	if !categoryActive {
		return core.E(core.CodeInvalidInput, "event category is not active in this school", nil)
	}
	var hostActive bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM accounts WHERE tenant_id = $1 AND id = $2 AND status = 'active')
	`, tenantID, hostAccountID).Scan(&hostActive); err != nil {
		return fmt.Errorf("check event host: %w", err)
	}
	if !hostActive {
		return core.E(core.CodeInvalidInput, "host is not active in this school", nil)
	}
	if roomID != "" {
		var roomActive bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (SELECT 1 FROM rooms WHERE tenant_id = $1 AND id = $2 AND status = 'active')
		`, tenantID, roomID).Scan(&roomActive); err != nil {
			return fmt.Errorf("check event room: %w", err)
		}
		if !roomActive {
			return core.E(core.CodeInvalidInput, "room is not active in this school", nil)
		}
	}
	return nil
}

func readEventSeries(ctx context.Context, reader lessonReader, tenantID, seriesID string) (core.EventSeries, error) {
	var series core.EventSeries
	var description, roomID, effectiveUntil *string
	err := reader.QueryRow(ctx, `
		SELECT es.id, es.category_id, es.title, es.description,
		       es.host_account_id, host_person.full_name, es.room_id, es.capacity,
		       es.weekday, es.start_minutes, es.duration_minutes,
		       to_char(es.effective_from, 'YYYY-MM-DD'),
		       to_char(es.effective_until, 'YYYY-MM-DD'),
		       es.status, es.version
		FROM event_series es
		JOIN accounts host_account
		  ON host_account.tenant_id = es.tenant_id AND host_account.id = es.host_account_id
		JOIN people host_person
		  ON host_person.tenant_id = host_account.tenant_id AND host_person.id = host_account.person_id
		WHERE es.tenant_id = $1 AND es.id = $2
	`, tenantID, seriesID).Scan(
		&series.ID, &series.CategoryID, &series.Title, &description,
		&series.Host.AccountID, &series.Host.FullName, &roomID, &series.Capacity,
		&series.Weekday, &series.StartMinutes, &series.DurationMinutes,
		&series.EffectiveFrom, &effectiveUntil, &series.Status, &series.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.EventSeries{}, core.E(core.CodeNotFound, "event series not found", nil)
	}
	if err != nil {
		return core.EventSeries{}, fmt.Errorf("read event series: %w", err)
	}
	if description != nil {
		series.Description = *description
	}
	if roomID != nil {
		series.RoomID = *roomID
	}
	if effectiveUntil != nil {
		series.EffectiveUntil = *effectiveUntil
	}
	return series, nil
}

func (s *Store) GenerateEventOccurrences(ctx context.Context, command core.GenerateEventOccurrencesCommand) (core.SeriesOccurrenceGenerationResult, error) {
	var result core.SeriesOccurrenceGenerationResult
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.eventManagerAuthority(ctx, tx, command.TenantID, command.ActorAccountID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "generate_event_occurrences", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.SeriesOccurrenceGenerationResult](claim)
			return err
		}
		var status string
		err = tx.QueryRow(ctx, `
			SELECT status FROM event_series WHERE tenant_id = $1 AND id = $2 FOR UPDATE
		`, command.TenantID, command.SeriesID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "event series not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock event series: %w", err)
		}
		if status != "active" {
			return core.E(core.CodeInvalidState, "only an active series can generate occurrences", nil)
		}
		created := []string{}
		for _, planned := range command.Occurrences {
			var exists bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM event_occurrences
					WHERE tenant_id = $1 AND series_id = $2 AND starts_at = $3
				)
			`, command.TenantID, command.SeriesID, planned.StartsAt).Scan(&exists); err != nil {
				return fmt.Errorf("check generated event occurrence: %w", err)
			}
			if exists {
				continue
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO event_occurrences (
					id, tenant_id, series_id, category_id, title, description,
					starts_at, duration_minutes, host_account_id, room_id, capacity,
					status, version, created_by_account_id, created_at, updated_at
				)
				SELECT $3, es.tenant_id, es.id, es.category_id, es.title, es.description,
				       $4, es.duration_minutes, es.host_account_id, es.room_id, es.capacity,
				       'scheduled', 0, $5, $6, $6
				FROM event_series es
				WHERE es.tenant_id = $1 AND es.id = $2
			`, command.TenantID, command.SeriesID, planned.OccurrenceID, planned.StartsAt,
				command.ActorAccountID, command.Now); err != nil {
				return mapWriteError(err, "event occurrence conflicts with existing data")
			}
			created = append(created, planned.OccurrenceID)
		}
		result = core.SeriesOccurrenceGenerationResult{SeriesID: command.SeriesID, CreatedCount: len(created), OccurrenceIDs: created}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "generate_event_occurrences", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "EventOccurrencesGenerated", targetType: "event_series", targetID: command.SeriesID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"createdCount": len(created)}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "EventOccurrencesGenerated", "event_series", command.SeriesID,
			map[string]any{"createdCount": len(created)}, command.Now)
	})
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, err
	}
	return result, nil
}

func (s *Store) CreateEventOccurrence(ctx context.Context, command core.CreateEventOccurrenceCommand) (core.EventOccurrence, error) {
	var occurrence core.EventOccurrence
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.eventManagerAuthority(ctx, tx, command.TenantID, command.ActorAccountID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_event_occurrence", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			occurrence, err = decodeReplay[core.EventOccurrence](claim)
			return err
		}
		if !command.StartsAt.After(command.Now) {
			return core.E(core.CodeInvalidState, "event must start in the future", nil)
		}
		if err := checkEventReferences(ctx, tx, command.TenantID, command.CategoryID, command.HostAccountID, command.RoomID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_occurrences (
				id, tenant_id, category_id, title, description,
				starts_at, duration_minutes, host_account_id, room_id, capacity,
				status, version, created_by_account_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, NULLIF($9, ''), $10,
			          'scheduled', 0, $11, $12, $12)
		`, command.OccurrenceID, command.TenantID, command.CategoryID, command.Title,
			command.Description, command.StartsAt, command.DurationMinutes,
			command.HostAccountID, command.RoomID, command.Capacity,
			command.ActorAccountID, command.Now); err != nil {
			return mapWriteError(err, "event conflicts with existing data")
		}
		occurrence, err = readEventOccurrence(ctx, tx, command.TenantID, command.OccurrenceID, core.Principal{TenantID: command.TenantID})
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_event_occurrence", command.IdempotencyKey, occurrence, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "EventScheduled", targetType: "event_occurrence", targetID: command.OccurrenceID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"hostAccountId": command.HostAccountID}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "EventScheduled", "event_occurrence", command.OccurrenceID,
			map[string]any{"occurrenceId": command.OccurrenceID}, command.Now)
	})
	if err != nil {
		return core.EventOccurrence{}, err
	}
	return occurrence, nil
}

func readEventOccurrence(ctx context.Context, reader lessonReader, tenantID, occurrenceID string, principal core.Principal) (core.EventOccurrence, error) {
	var view core.EventOccurrence
	var seriesID, description, roomID *string
	err := reader.QueryRow(ctx, `
		SELECT eo.id, eo.series_id, eo.category_id, category.name, eo.title, eo.description,
		       eo.starts_at, eo.duration_minutes,
		       eo.host_account_id, host_person.full_name, eo.room_id, eo.capacity,
		       (SELECT count(*) FROM event_rsvps r
		        WHERE r.tenant_id = eo.tenant_id AND r.occurrence_id = eo.id AND r.status = 'confirmed'),
		       eo.status, eo.version
		FROM event_occurrences eo
		JOIN event_categories category
		  ON category.tenant_id = eo.tenant_id AND category.id = eo.category_id
		JOIN accounts host_account
		  ON host_account.tenant_id = eo.tenant_id AND host_account.id = eo.host_account_id
		JOIN people host_person
		  ON host_person.tenant_id = host_account.tenant_id AND host_person.id = host_account.person_id
		WHERE eo.tenant_id = $1 AND eo.id = $2
	`, tenantID, occurrenceID).Scan(
		&view.ID, &seriesID, &view.CategoryID, &view.CategoryName, &view.Title, &description,
		&view.StartsAt, &view.DurationMinutes,
		&view.Host.AccountID, &view.Host.FullName, &roomID, &view.Capacity,
		&view.ConfirmedCount, &view.Status, &view.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.EventOccurrence{}, core.E(core.CodeNotFound, "event not found", nil)
	}
	if err != nil {
		return core.EventOccurrence{}, fmt.Errorf("read event occurrence: %w", err)
	}
	view.StartsAt = view.StartsAt.UTC()
	if seriesID != nil {
		view.SeriesID = *seriesID
	}
	if description != nil {
		view.Description = *description
	}
	if roomID != nil {
		view.RoomID = *roomID
	}
	studentID, err := studentIDForAccount(ctx, reader, tenantID, principal.AccountID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	if studentID == "" {
		return view, nil
	}
	var rsvpStatus *string
	if err := reader.QueryRow(ctx, `
		SELECT status FROM event_rsvps
		WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
	`, tenantID, occurrenceID, studentID).Scan(&rsvpStatus); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return core.EventOccurrence{}, fmt.Errorf("read caller RSVP: %w", err)
	}
	if rsvpStatus != nil {
		view.MyRsvp = *rsvpStatus
	}
	var position *int
	if err := reader.QueryRow(ctx, `
		SELECT position FROM event_waitlist_entries
		WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
	`, tenantID, occurrenceID, studentID).Scan(&position); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return core.EventOccurrence{}, fmt.Errorf("read caller waitlist position: %w", err)
	}
	if position != nil {
		view.MyWaitlistPosition = *position
	}
	var offer core.SpotOffer
	err = reader.QueryRow(ctx, `
		SELECT id, occurrence_id, status, offered_at, expires_at
		FROM event_spot_offers
		WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3 AND status = 'pending'
	`, tenantID, occurrenceID, studentID).Scan(
		&offer.ID, &offer.OccurrenceID, &offer.Status, &offer.OfferedAt, &offer.ExpiresAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return core.EventOccurrence{}, fmt.Errorf("read caller spot offer: %w", err)
	}
	if err == nil {
		offer.OfferedAt = offer.OfferedAt.UTC()
		offer.ExpiresAt = offer.ExpiresAt.UTC()
		view.MyOffer = &offer
	}
	return view, nil
}

func studentIDForAccount(ctx context.Context, reader lessonReader, tenantID, accountID string) (string, error) {
	if accountID == "" {
		return "", nil
	}
	var studentID string
	err := reader.QueryRow(ctx, `
		SELECT id FROM students WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(&studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("resolve caller student: %w", err)
	}
	return studentID, nil
}

func (s *Store) ListEventOccurrences(ctx context.Context, principal core.Principal, query core.EventListQuery) ([]core.EventOccurrence, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM event_occurrences
		WHERE tenant_id = $1 AND starts_at >= $2 AND starts_at < $3
		ORDER BY starts_at, id
	`, principal.TenantID, query.From, query.To)
	if err != nil {
		return nil, fmt.Errorf("list event occurrences: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan event occurrence id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event occurrence ids: %w", err)
	}
	rows.Close()
	result := make([]core.EventOccurrence, 0, len(ids))
	for _, id := range ids {
		view, err := readEventOccurrence(ctx, s.pool, principal.TenantID, id, principal)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	return result, nil
}

func lockEventOccurrence(ctx context.Context, tx pgx.Tx, tenantID, occurrenceID string) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
		advisoryLockKey("event-occurrence", tenantID, occurrenceID)); err != nil {
		return fmt.Errorf("lock event occurrence: %w", err)
	}
	var status string
	err := tx.QueryRow(ctx, `
		SELECT status FROM event_occurrences WHERE tenant_id = $1 AND id = $2 FOR UPDATE
	`, tenantID, occurrenceID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.E(core.CodeNotFound, "event not found", nil)
	}
	if err != nil {
		return fmt.Errorf("read event occurrence status: %w", err)
	}
	if status != "scheduled" {
		return core.E(core.CodeInvalidState, "event is not open for changes", nil)
	}
	return nil
}

func heldEventSeats(ctx context.Context, tx pgx.Tx, tenantID, occurrenceID string) (held, capacity int, err error) {
	err = tx.QueryRow(ctx, `
		SELECT eo.capacity,
		       (SELECT count(*) FROM event_rsvps r
		        WHERE r.tenant_id = eo.tenant_id AND r.occurrence_id = eo.id AND r.status = 'confirmed')
		     + (SELECT count(*) FROM event_spot_offers o
		        WHERE o.tenant_id = eo.tenant_id AND o.occurrence_id = eo.id AND o.status = 'pending')
		FROM event_occurrences eo
		WHERE eo.tenant_id = $1 AND eo.id = $2
	`, tenantID, occurrenceID).Scan(&capacity, &held)
	if err != nil {
		return 0, 0, fmt.Errorf("count held event seats: %w", err)
	}
	return held, capacity, nil
}

// expireAndCascadeOffers retires the overdue pending offer (if any) and
// hands a free seat to the waitlist head as one fresh pending offer.
func expireAndCascadeOffers(ctx context.Context, tx pgx.Tx, tenantID, occurrenceID string, now time.Time, ttl time.Duration) error {
	if _, err := tx.Exec(ctx, `
		UPDATE event_spot_offers
		SET status = 'expired', resolved_at = $3
		WHERE tenant_id = $1 AND occurrence_id = $2 AND status = 'pending' AND expires_at <= $3
	`, tenantID, occurrenceID, now); err != nil {
		return fmt.Errorf("expire overdue spot offers: %w", err)
	}
	held, capacity, err := heldEventSeats(ctx, tx, tenantID, occurrenceID)
	if err != nil {
		return err
	}
	if held >= capacity {
		return nil
	}
	var pendingExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM event_spot_offers
			WHERE tenant_id = $1 AND occurrence_id = $2 AND status = 'pending'
		)
	`, tenantID, occurrenceID).Scan(&pendingExists); err != nil {
		return fmt.Errorf("check pending spot offer: %w", err)
	}
	if pendingExists {
		return nil
	}
	var headStudent string
	err = tx.QueryRow(ctx, `
		SELECT student_id FROM event_waitlist_entries
		WHERE tenant_id = $1 AND occurrence_id = $2
		ORDER BY position
		LIMIT 1
	`, tenantID, occurrenceID).Scan(&headStudent)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read waitlist head: %w", err)
	}
	if err := removeWaitlistEntry(ctx, tx, tenantID, occurrenceID, headStudent); err != nil {
		return err
	}
	var headAccountID string
	if err := tx.QueryRow(ctx, `
		SELECT account_id FROM students WHERE tenant_id = $1 AND id = $2
	`, tenantID, headStudent).Scan(&headAccountID); err != nil {
		return fmt.Errorf("resolve waitlist head account: %w", err)
	}
	offerID, err := security.NewID("offer")
	if err != nil {
		return core.E(core.CodeInternal, "could not create the spot offer", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO event_spot_offers (id, tenant_id, occurrence_id, student_id, status, offered_at, expires_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, $6)
	`, offerID, tenantID, occurrenceID, headStudent, now, now.Add(ttl)); err != nil {
		return mapWriteError(err, "spot offer conflicts with existing data")
	}
	if err := appendAudit(ctx, tx, auditInput{
		tenantID: tenantID, actorID: headAccountID,
		action: "SpotOfferCreated", targetType: "spot_offer", targetID: offerID,
		decision: "allow", at: now,
	}); err != nil {
		return err
	}
	return appendOutbox(ctx, tx, tenantID, "SpotOfferCreated", "spot_offer", offerID,
		map[string]any{"offerId": offerID, "occurrenceId": occurrenceID}, now)
}

func removeWaitlistEntry(ctx context.Context, tx pgx.Tx, tenantID, occurrenceID, studentID string) error {
	var position int
	err := tx.QueryRow(ctx, `
		DELETE FROM event_waitlist_entries
		WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
		RETURNING position
	`, tenantID, occurrenceID, studentID).Scan(&position)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.E(core.CodeInvalidState, "not on the waitlist", nil)
	}
	if err != nil {
		return fmt.Errorf("leave waitlist: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE event_waitlist_entries
		SET position = position - 1
		WHERE tenant_id = $1 AND occurrence_id = $2 AND position > $3
	`, tenantID, occurrenceID, position); err != nil {
		return fmt.Errorf("shift waitlist positions: %w", err)
	}
	return nil
}

func requireEventStudent(ctx context.Context, tx pgx.Tx, principal core.Principal) (string, error) {
	isStudent, err := hasActiveRole(ctx, tx, principal.TenantID, principal.AccountID, core.RoleStudent)
	if err != nil {
		return "", err
	}
	if !isStudent {
		return "", core.E(core.CodeForbidden, "only a Student can manage own RSVP", nil)
	}
	studentID, err := studentIDForAccount(ctx, tx, principal.TenantID, principal.AccountID)
	if err != nil {
		return "", err
	}
	if studentID == "" {
		return "", core.E(core.CodeForbidden, "only a Student can manage own RSVP", nil)
	}
	return studentID, nil
}

func (s *Store) RsvpToEvent(ctx context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	principal := command.Principal
	var view core.EventOccurrence
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		studentID, err := requireEventStudent(ctx, tx, principal)
		if err != nil {
			return err
		}
		if err := lockEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID); err != nil {
			return err
		}
		if err := expireAndCascadeOffers(ctx, tx, principal.TenantID, command.OccurrenceID, command.Now, command.OfferTTL); err != nil {
			return err
		}
		var existingStatus *string
		if err := tx.QueryRow(ctx, `
			SELECT status FROM event_rsvps
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
		`, principal.TenantID, command.OccurrenceID, studentID).Scan(&existingStatus); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("read existing RSVP: %w", err)
		}
		if existingStatus != nil && *existingStatus == "confirmed" {
			view, err = readEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID, principal)
			return err
		}
		var myOfferID *string
		if err := tx.QueryRow(ctx, `
			SELECT id FROM event_spot_offers
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3 AND status = 'pending'
		`, principal.TenantID, command.OccurrenceID, studentID).Scan(&myOfferID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("read caller spot offer: %w", err)
		}
		if myOfferID != nil {
			if _, err := tx.Exec(ctx, `
				UPDATE event_spot_offers SET status = 'confirmed', resolved_at = $2 WHERE id = $1
			`, *myOfferID, command.Now); err != nil {
				return fmt.Errorf("confirm spot offer: %w", err)
			}
		} else {
			held, capacity, err := heldEventSeats(ctx, tx, principal.TenantID, command.OccurrenceID)
			if err != nil {
				return err
			}
			if held >= capacity {
				return core.E(core.CodeConflict, "event is full — join the waitlist", nil)
			}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_rsvps (tenant_id, occurrence_id, student_id, status, confirmed_at, updated_at)
			VALUES ($1, $2, $3, 'confirmed', $4, $4)
			ON CONFLICT (tenant_id, occurrence_id, student_id)
			DO UPDATE SET status = 'confirmed', confirmed_at = EXCLUDED.confirmed_at,
			              cancelled_at = NULL, updated_at = EXCLUDED.updated_at
		`, principal.TenantID, command.OccurrenceID, studentID, command.Now); err != nil {
			return mapWriteError(err, "RSVP conflicts with existing data")
		}
		if _, err := tx.Exec(ctx, `
			DELETE FROM event_waitlist_entries
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
		`, principal.TenantID, command.OccurrenceID, studentID); err != nil {
			return fmt.Errorf("clear waitlist after RSVP: %w", err)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "EventRsvpConfirmed", targetType: "event_occurrence", targetID: command.OccurrenceID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		if err := appendOutbox(ctx, tx, principal.TenantID, "EventRsvpConfirmed", "event_occurrence", command.OccurrenceID,
			map[string]any{"occurrenceId": command.OccurrenceID}, command.Now); err != nil {
			return err
		}
		view, err = readEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID, principal)
		return err
	})
	if err != nil {
		return core.EventOccurrence{}, err
	}
	return view, nil
}

func (s *Store) CancelEventRsvp(ctx context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	principal := command.Principal
	var view core.EventOccurrence
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		studentID, err := requireEventStudent(ctx, tx, principal)
		if err != nil {
			return err
		}
		if err := lockEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID); err != nil {
			return err
		}
		updated, err := tx.Exec(ctx, `
			UPDATE event_rsvps
			SET status = 'cancelled', cancelled_at = $4, updated_at = $4
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3 AND status = 'confirmed'
		`, principal.TenantID, command.OccurrenceID, studentID, command.Now)
		if err != nil {
			return fmt.Errorf("cancel RSVP: %w", err)
		}
		if updated.RowsAffected() == 0 {
			return core.E(core.CodeInvalidState, "no confirmed RSVP to cancel", nil)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "EventRsvpCancelled", targetType: "event_occurrence", targetID: command.OccurrenceID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		if err := appendOutbox(ctx, tx, principal.TenantID, "EventRsvpCancelled", "event_occurrence", command.OccurrenceID,
			map[string]any{"occurrenceId": command.OccurrenceID}, command.Now); err != nil {
			return err
		}
		if err := expireAndCascadeOffers(ctx, tx, principal.TenantID, command.OccurrenceID, command.Now, command.OfferTTL); err != nil {
			return err
		}
		view, err = readEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID, principal)
		return err
	})
	if err != nil {
		return core.EventOccurrence{}, err
	}
	return view, nil
}

func (s *Store) JoinEventWaitlist(ctx context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	principal := command.Principal
	var view core.EventOccurrence
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		studentID, err := requireEventStudent(ctx, tx, principal)
		if err != nil {
			return err
		}
		if err := lockEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID); err != nil {
			return err
		}
		if err := expireAndCascadeOffers(ctx, tx, principal.TenantID, command.OccurrenceID, command.Now, command.OfferTTL); err != nil {
			return err
		}
		var rsvpStatus *string
		if err := tx.QueryRow(ctx, `
			SELECT status FROM event_rsvps
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
		`, principal.TenantID, command.OccurrenceID, studentID).Scan(&rsvpStatus); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("read existing RSVP: %w", err)
		}
		if rsvpStatus != nil && *rsvpStatus == "confirmed" {
			return core.E(core.CodeInvalidState, "RSVP is already confirmed", nil)
		}
		var myPendingOffer bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM event_spot_offers
				WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3 AND status = 'pending'
			)
		`, principal.TenantID, command.OccurrenceID, studentID).Scan(&myPendingOffer); err != nil {
			return fmt.Errorf("check caller spot offer: %w", err)
		}
		if myPendingOffer {
			return core.E(core.CodeInvalidState, "a spot offer is already waiting for your decision", nil)
		}
		var alreadyWaitlisted bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM event_waitlist_entries
				WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
			)
		`, principal.TenantID, command.OccurrenceID, studentID).Scan(&alreadyWaitlisted); err != nil {
			return fmt.Errorf("check waitlist membership: %w", err)
		}
		if alreadyWaitlisted {
			view, err = readEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID, principal)
			return err
		}
		held, capacity, err := heldEventSeats(ctx, tx, principal.TenantID, command.OccurrenceID)
		if err != nil {
			return err
		}
		if held < capacity {
			return core.E(core.CodeInvalidState, "seats are available — RSVP directly", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_waitlist_entries (tenant_id, occurrence_id, student_id, position, joined_at)
			VALUES ($1, $2, $3,
			        COALESCE((SELECT max(position) FROM event_waitlist_entries
			                  WHERE tenant_id = $1 AND occurrence_id = $2), 0) + 1,
			        $4)
		`, principal.TenantID, command.OccurrenceID, studentID, command.Now); err != nil {
			return mapWriteError(err, "waitlist entry conflicts with existing data")
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "EventWaitlistJoined", targetType: "event_occurrence", targetID: command.OccurrenceID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		if err := appendOutbox(ctx, tx, principal.TenantID, "EventWaitlistJoined", "event_occurrence", command.OccurrenceID,
			map[string]any{"occurrenceId": command.OccurrenceID}, command.Now); err != nil {
			return err
		}
		view, err = readEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID, principal)
		return err
	})
	if err != nil {
		return core.EventOccurrence{}, err
	}
	return view, nil
}

func (s *Store) LeaveEventWaitlist(ctx context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	principal := command.Principal
	var view core.EventOccurrence
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		studentID, err := requireEventStudent(ctx, tx, principal)
		if err != nil {
			return err
		}
		if err := lockEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID); err != nil {
			return err
		}
		if err := removeWaitlistEntry(ctx, tx, principal.TenantID, command.OccurrenceID, studentID); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "EventWaitlistLeft", targetType: "event_occurrence", targetID: command.OccurrenceID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		view, err = readEventOccurrence(ctx, tx, principal.TenantID, command.OccurrenceID, principal)
		return err
	})
	if err != nil {
		return core.EventOccurrence{}, err
	}
	return view, nil
}

func (s *Store) ConfirmSpotOffer(ctx context.Context, command core.SpotOfferDecisionCommand) (core.EventOccurrence, error) {
	principal := command.Principal
	var view core.EventOccurrence
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		studentID, err := requireEventStudent(ctx, tx, principal)
		if err != nil {
			return err
		}
		var occurrenceID, status string
		var expiresAt time.Time
		err = tx.QueryRow(ctx, `
			SELECT occurrence_id, status, expires_at FROM event_spot_offers
			WHERE tenant_id = $1 AND id = $2 AND student_id = $3
		`, principal.TenantID, command.OfferID, studentID).Scan(&occurrenceID, &status, &expiresAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "spot offer not found", nil)
		}
		if err != nil {
			return fmt.Errorf("read spot offer: %w", err)
		}
		if err := lockEventOccurrence(ctx, tx, principal.TenantID, occurrenceID); err != nil {
			return err
		}
		if status != "pending" || !expiresAt.After(command.Now) {
			if err := expireAndCascadeOffers(ctx, tx, principal.TenantID, occurrenceID, command.Now, command.OfferTTL); err != nil {
				return err
			}
			return core.E(core.CodeInvalidState, "spot offer has expired", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE event_spot_offers SET status = 'confirmed', resolved_at = $2 WHERE id = $1
		`, command.OfferID, command.Now); err != nil {
			return fmt.Errorf("confirm spot offer: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_rsvps (tenant_id, occurrence_id, student_id, status, confirmed_at, updated_at)
			VALUES ($1, $2, $3, 'confirmed', $4, $4)
			ON CONFLICT (tenant_id, occurrence_id, student_id)
			DO UPDATE SET status = 'confirmed', confirmed_at = EXCLUDED.confirmed_at,
			              cancelled_at = NULL, updated_at = EXCLUDED.updated_at
		`, principal.TenantID, occurrenceID, studentID, command.Now); err != nil {
			return mapWriteError(err, "RSVP conflicts with existing data")
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "SpotOfferConfirmed", targetType: "spot_offer", targetID: command.OfferID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		if err := appendOutbox(ctx, tx, principal.TenantID, "SpotOfferConfirmed", "spot_offer", command.OfferID,
			map[string]any{"offerId": command.OfferID}, command.Now); err != nil {
			return err
		}
		view, err = readEventOccurrence(ctx, tx, principal.TenantID, occurrenceID, principal)
		return err
	})
	if err != nil {
		return core.EventOccurrence{}, err
	}
	return view, nil
}

func (s *Store) DeclineSpotOffer(ctx context.Context, command core.SpotOfferDecisionCommand) (core.EventOccurrence, error) {
	principal := command.Principal
	var view core.EventOccurrence
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		studentID, err := requireEventStudent(ctx, tx, principal)
		if err != nil {
			return err
		}
		var occurrenceID, status string
		err = tx.QueryRow(ctx, `
			SELECT occurrence_id, status FROM event_spot_offers
			WHERE tenant_id = $1 AND id = $2 AND student_id = $3
		`, principal.TenantID, command.OfferID, studentID).Scan(&occurrenceID, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "spot offer not found", nil)
		}
		if err != nil {
			return fmt.Errorf("read spot offer: %w", err)
		}
		if err := lockEventOccurrence(ctx, tx, principal.TenantID, occurrenceID); err != nil {
			return err
		}
		if status != "pending" {
			return core.E(core.CodeInvalidState, "spot offer is already resolved", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE event_spot_offers SET status = 'declined', resolved_at = $2 WHERE id = $1
		`, command.OfferID, command.Now); err != nil {
			return fmt.Errorf("decline spot offer: %w", err)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "SpotOfferDeclined", targetType: "spot_offer", targetID: command.OfferID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		if err := expireAndCascadeOffers(ctx, tx, principal.TenantID, occurrenceID, command.Now, command.OfferTTL); err != nil {
			return err
		}
		view, err = readEventOccurrence(ctx, tx, principal.TenantID, occurrenceID, principal)
		return err
	})
	if err != nil {
		return core.EventOccurrence{}, err
	}
	return view, nil
}

func (s *Store) GetEventSeries(ctx context.Context, principal core.Principal, seriesID string) (core.EventSeries, error) {
	return readEventSeries(ctx, s.pool, principal.TenantID, seriesID)
}

func (s *Store) GetEventOccurrence(ctx context.Context, principal core.Principal, occurrenceID string) (core.EventOccurrence, error) {
	return readEventOccurrence(ctx, s.pool, principal.TenantID, occurrenceID, principal)
}
