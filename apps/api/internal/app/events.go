package app

import (
	"context"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.2 events and RSVP (DEC-001/003/101). Seat mutations carry the spot
// offer TTL from configuration so an overdue offer can expire and pass
// to the next waitlisted student inside the same transaction.

func (s *Service) CreateEventCategory(ctx context.Context, principal core.Principal, name string) (core.EventCategory, error) {
	normalized, err := security.ValidateText("name", name, 1, 100)
	if err != nil {
		return core.EventCategory{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	categoryID, err := security.NewID("evcat")
	if err != nil {
		return core.EventCategory{}, core.E(core.CodeInternal, "could not create the category", err)
	}
	category, err := s.store.CreateEventCategory(ctx, core.CreateEventCategoryCommand{
		Principal: principal, CategoryID: categoryID, Name: normalized, Now: s.clock.Now(),
	})
	if err != nil {
		return core.EventCategory{}, normalizeStoreError("create event category", err)
	}
	return category, nil
}

func (s *Service) ListEventCategories(ctx context.Context, principal core.Principal) ([]core.EventCategory, error) {
	categories, err := s.store.ListEventCategories(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list event categories", err)
	}
	return categories, nil
}

type CreateEventSeriesInput struct {
	CategoryID      string
	Title           string
	Description     string
	HostAccountID   string
	RoomID          string
	Capacity        int
	Weekday         int
	StartMinutes    int
	DurationMinutes int
	EffectiveFrom   string
	EffectiveUntil  string
	IdempotencyKey  string
}

func (s *Service) CreateEventSeries(ctx context.Context, principal core.Principal, input CreateEventSeriesInput) (core.EventSeries, error) {
	categoryID, err := security.ValidateIdentifier("categoryId", input.CategoryID, 128)
	if err != nil {
		return core.EventSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	title, err := security.ValidateText("title", input.Title, 1, 200)
	if err != nil {
		return core.EventSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	description := ""
	if input.Description != "" {
		description, err = security.ValidateText("description", input.Description, 1, 2000)
		if err != nil {
			return core.EventSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	hostID, err := security.ValidateIdentifier("hostAccountId", input.HostAccountID, 128)
	if err != nil {
		return core.EventSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	roomID := ""
	if input.RoomID != "" {
		roomID, err = security.ValidateIdentifier("roomId", input.RoomID, 128)
		if err != nil {
			return core.EventSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.Capacity < 1 || input.Capacity > 500 {
		return core.EventSeries{}, core.E(core.CodeInvalidInput, "capacity must be between 1 and 500", nil)
	}
	if err := validateWeeklyAnchor(input.Weekday, input.StartMinutes, input.DurationMinutes, input.EffectiveFrom, input.EffectiveUntil); err != nil {
		return core.EventSeries{}, err
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.EventSeries{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	seriesID, err := security.NewID("evser")
	if err != nil {
		return core.EventSeries{}, core.E(core.CodeInternal, "could not create the series", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		Category, Title, Description, Host, Room string
		Capacity, Weekday, Start, Duration       int
		EffectiveFrom, EffectiveUntil            string
	}{categoryID, title, description, hostID, roomID, input.Capacity,
		input.Weekday, input.StartMinutes, input.DurationMinutes,
		input.EffectiveFrom, input.EffectiveUntil})
	if err != nil {
		return core.EventSeries{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	series, err := s.store.CreateEventSeries(ctx, core.CreateEventSeriesCommand{
		TenantID: principal.TenantID, ActorAccountID: principal.AccountID,
		SeriesID: seriesID, CategoryID: categoryID, Title: title, Description: description,
		HostAccountID: hostID, RoomID: roomID, Capacity: input.Capacity,
		Weekday: input.Weekday, StartMinutes: input.StartMinutes,
		DurationMinutes: input.DurationMinutes,
		EffectiveFrom:   input.EffectiveFrom, EffectiveUntil: input.EffectiveUntil,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.EventSeries{}, normalizeStoreError("create event series", err)
	}
	return series, nil
}

func validateWeeklyAnchor(weekday, startMinutes, durationMinutes int, effectiveFrom, effectiveUntil string) error {
	if weekday < 0 || weekday > 6 {
		return core.E(core.CodeInvalidInput, "weekday must be between 0 (Monday) and 6 (Sunday)", nil)
	}
	if startMinutes < 0 || startMinutes > 1439 {
		return core.E(core.CodeInvalidInput, "startMinutes must be between 0 and 1439", nil)
	}
	if durationMinutes < 1 || durationMinutes > 1440 {
		return core.E(core.CodeInvalidInput, "durationMinutes must be between 1 and 1440", nil)
	}
	zone, err := time.LoadLocation(schoolTimeZone)
	if err != nil {
		return core.E(core.CodeInternal, "school time zone is unavailable", err)
	}
	from, err := time.ParseInLocation("2006-01-02", effectiveFrom, zone)
	if err != nil {
		return core.E(core.CodeInvalidInput, "effectiveFrom must be a YYYY-MM-DD date", nil)
	}
	if effectiveUntil != "" {
		until, parseErr := time.ParseInLocation("2006-01-02", effectiveUntil, zone)
		if parseErr != nil {
			return core.E(core.CodeInvalidInput, "effectiveUntil must be a YYYY-MM-DD date", nil)
		}
		if until.Before(from) {
			return core.E(core.CodeInvalidInput, "effectiveUntil must not precede effectiveFrom", nil)
		}
	}
	return nil
}

func (s *Service) GenerateEventSeriesOccurrences(ctx context.Context, principal core.Principal, seriesID string, weeks int, idempotencyKey string) (core.SeriesOccurrenceGenerationResult, error) {
	normalizedID, err := security.ValidateIdentifier("seriesId", seriesID, 128)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if weeks < 1 || weeks > maxGenerationWeeks {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidInput, "weeks must be between 1 and 12", nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	series, err := s.store.GetEventSeries(ctx, principal, normalizedID)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, normalizeStoreError("read event series", err)
	}
	now := s.clock.Now()
	planned, err := planWeeklyOccurrences(core.CoreLessonSeries{
		Weekday: series.Weekday, StartMinutes: series.StartMinutes,
		EffectiveFrom: series.EffectiveFrom, EffectiveUntil: series.EffectiveUntil,
	}, now, weeks)
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, err
	}
	for index := range planned {
		occurrenceID, idErr := security.NewID("evocc")
		if idErr != nil {
			return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInternal, "could not create occurrence ids", idErr)
		}
		planned[index].OccurrenceID = occurrenceID
	}
	fingerprint, err := security.Fingerprint(struct {
		SeriesID string
		Weeks    int
	}{normalizedID, weeks})
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	result, err := s.store.GenerateEventOccurrences(ctx, core.GenerateEventOccurrencesCommand{
		TenantID: principal.TenantID, ActorAccountID: principal.AccountID,
		SeriesID: normalizedID, Occurrences: planned,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint, Now: now,
	})
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, normalizeStoreError("generate event occurrences", err)
	}
	return result, nil
}

type CreateEventInput struct {
	CategoryID      string
	Title           string
	Description     string
	StartsAt        time.Time
	DurationMinutes int
	HostAccountID   string
	RoomID          string
	Capacity        int
	IdempotencyKey  string
}

func (s *Service) CreateEvent(ctx context.Context, principal core.Principal, input CreateEventInput) (core.EventOccurrence, error) {
	categoryID, err := security.ValidateIdentifier("categoryId", input.CategoryID, 128)
	if err != nil {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	title, err := security.ValidateText("title", input.Title, 1, 200)
	if err != nil {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	description := ""
	if input.Description != "" {
		description, err = security.ValidateText("description", input.Description, 1, 2000)
		if err != nil {
			return core.EventOccurrence{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	hostID, err := security.ValidateIdentifier("hostAccountId", input.HostAccountID, 128)
	if err != nil {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	roomID := ""
	if input.RoomID != "" {
		roomID, err = security.ValidateIdentifier("roomId", input.RoomID, 128)
		if err != nil {
			return core.EventOccurrence{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.Capacity < 1 || input.Capacity > 500 {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, "capacity must be between 1 and 500", nil)
	}
	if input.DurationMinutes < 1 || input.DurationMinutes > 1440 {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, "durationMinutes must be between 1 and 1440", nil)
	}
	if input.StartsAt.IsZero() {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, "startsAt is required", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	occurrenceID, err := security.NewID("evocc")
	if err != nil {
		return core.EventOccurrence{}, core.E(core.CodeInternal, "could not create the event", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		Category, Title, Description, Host, Room string
		StartsAt                                 time.Time
		Duration, Capacity                       int
	}{categoryID, title, description, hostID, roomID, input.StartsAt.UTC(), input.DurationMinutes, input.Capacity})
	if err != nil {
		return core.EventOccurrence{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	occurrence, err := s.store.CreateEventOccurrence(ctx, core.CreateEventOccurrenceCommand{
		TenantID: principal.TenantID, ActorAccountID: principal.AccountID,
		OccurrenceID: occurrenceID, CategoryID: categoryID,
		Title: title, Description: description,
		StartsAt: input.StartsAt.UTC(), DurationMinutes: input.DurationMinutes,
		HostAccountID: hostID, RoomID: roomID, Capacity: input.Capacity,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("create event", err)
	}
	return occurrence, nil
}

func (s *Service) ListEvents(ctx context.Context, principal core.Principal, from, to time.Time) ([]core.EventOccurrence, error) {
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, core.E(core.CodeInvalidInput, "a valid from/to window is required", nil)
	}
	if to.Sub(from) > 200*24*time.Hour {
		return nil, core.E(core.CodeInvalidInput, "window must not exceed 200 days", nil)
	}
	events, err := s.store.ListEventOccurrences(ctx, principal, core.EventListQuery{From: from.UTC(), To: to.UTC()})
	if err != nil {
		return nil, normalizeStoreError("list events", err)
	}
	return events, nil
}

func (s *Service) eventSeatCommand(principal core.Principal, occurrenceID string) (core.EventSeatCommand, error) {
	normalizedID, err := security.ValidateIdentifier("occurrenceId", occurrenceID, 128)
	if err != nil {
		return core.EventSeatCommand{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	return core.EventSeatCommand{
		Principal: principal, OccurrenceID: normalizedID,
		OfferTTL: s.spotOfferTTL, Now: s.clock.Now(),
	}, nil
}

func (s *Service) RsvpToEvent(ctx context.Context, principal core.Principal, occurrenceID string) (core.EventOccurrence, error) {
	command, err := s.eventSeatCommand(principal, occurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	view, err := s.store.RsvpToEvent(ctx, command)
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("confirm RSVP", err)
	}
	return view, nil
}

func (s *Service) CancelEventRsvp(ctx context.Context, principal core.Principal, occurrenceID string) (core.EventOccurrence, error) {
	command, err := s.eventSeatCommand(principal, occurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	view, err := s.store.CancelEventRsvp(ctx, command)
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("cancel RSVP", err)
	}
	return view, nil
}

func (s *Service) JoinEventWaitlist(ctx context.Context, principal core.Principal, occurrenceID string) (core.EventOccurrence, error) {
	command, err := s.eventSeatCommand(principal, occurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	view, err := s.store.JoinEventWaitlist(ctx, command)
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("join waitlist", err)
	}
	return view, nil
}

func (s *Service) LeaveEventWaitlist(ctx context.Context, principal core.Principal, occurrenceID string) (core.EventOccurrence, error) {
	command, err := s.eventSeatCommand(principal, occurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	view, err := s.store.LeaveEventWaitlist(ctx, command)
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("leave waitlist", err)
	}
	return view, nil
}

func (s *Service) spotOfferCommand(principal core.Principal, offerID string) (core.SpotOfferDecisionCommand, error) {
	normalizedID, err := security.ValidateIdentifier("offerId", offerID, 128)
	if err != nil {
		return core.SpotOfferDecisionCommand{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	return core.SpotOfferDecisionCommand{
		Principal: principal, OfferID: normalizedID,
		OfferTTL: s.spotOfferTTL, Now: s.clock.Now(),
	}, nil
}

func (s *Service) ConfirmSpotOffer(ctx context.Context, principal core.Principal, offerID string) (core.EventOccurrence, error) {
	command, err := s.spotOfferCommand(principal, offerID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	view, err := s.store.ConfirmSpotOffer(ctx, command)
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("confirm spot offer", err)
	}
	return view, nil
}

func (s *Service) DeclineSpotOffer(ctx context.Context, principal core.Principal, offerID string) (core.EventOccurrence, error) {
	command, err := s.spotOfferCommand(principal, offerID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	view, err := s.store.DeclineSpotOffer(ctx, command)
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("decline spot offer", err)
	}
	return view, nil
}

func (s *Service) GetEvent(ctx context.Context, principal core.Principal, occurrenceID string) (core.EventOccurrence, error) {
	normalizedID, err := security.ValidateIdentifier("occurrenceId", occurrenceID, 128)
	if err != nil {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	view, err := s.store.GetEventOccurrence(ctx, principal, normalizedID)
	if err != nil {
		return core.EventOccurrence{}, normalizeStoreError("read event", err)
	}
	return view, nil
}
