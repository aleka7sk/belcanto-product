package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 events and RSVP — parity with PostgreSQL. A seat is held by a
// confirmed RSVP or one pending spot offer; every seat mutation first
// expires overdue offers and cascades the freed seat to the next
// waitlisted student (DEC-101: the TTL always arrives with the command).

type eventCategory struct {
	ID       string
	TenantID string
	Name     string
	Status   string
}

type eventSeries struct {
	ID              string
	TenantID        string
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
	Status          string
	Version         int64
}

type eventOccurrence struct {
	ID              string
	TenantID        string
	SeriesID        string
	CategoryID      string
	Title           string
	Description     string
	StartsAt        time.Time
	DurationMinutes int
	HostAccountID   string
	RoomID          string
	Capacity        int
	Status          string
	Version         int64
}

type eventRsvp struct {
	Status      string
	ConfirmedAt *time.Time
	CancelledAt *time.Time
}

type waitlistEntry struct {
	StudentID string
	JoinedAt  time.Time
}

type spotOffer struct {
	ID           string
	TenantID     string
	OccurrenceID string
	StudentID    string
	Status       string
	OfferedAt    time.Time
	ExpiresAt    time.Time
	ResolvedAt   *time.Time
}

func (s *Store) eventManager(tenantID, accountID string) bool {
	actor := s.activeAccount(accountID, tenantID)
	return actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
}

func (s *Store) CreateEventCategory(_ context.Context, command core.CreateEventCategoryCommand) (core.EventCategory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if !s.eventManager(principal.TenantID, principal.AccountID) {
		return core.EventCategory{}, core.E(core.CodeForbidden, "event management permission is required", nil)
	}
	for _, existing := range s.eventCategories {
		if existing.TenantID == principal.TenantID && existing.Name == command.Name {
			return core.EventCategory{}, core.E(core.CodeConflict, "event category name is already in use", nil)
		}
	}
	stored := &eventCategory{ID: command.CategoryID, TenantID: principal.TenantID, Name: command.Name, Status: "active"}
	s.eventCategories[stored.ID] = stored
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "EventCategoryCreated",
		"event_category", stored.ID, "allow", "", command.Now, nil)
	return core.EventCategory{ID: stored.ID, Name: stored.Name, Status: stored.Status}, nil
}

func (s *Store) ListEventCategories(_ context.Context, principal core.Principal) ([]core.EventCategory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.EventCategory{}
	for _, stored := range s.eventCategories {
		if stored.TenantID == principal.TenantID && stored.Status == "active" {
			result = append(result, core.EventCategory{ID: stored.ID, Name: stored.Name, Status: stored.Status})
		}
	}
	sort.Slice(result, func(left, right int) bool { return result[left].Name < result[right].Name })
	return result, nil
}

func (s *Store) CreateEventSeries(_ context.Context, command core.CreateEventSeriesCommand) (core.EventSeries, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.eventManager(command.TenantID, command.ActorAccountID) {
		return core.EventSeries{}, core.E(core.CodeForbidden, "event management permission is required", nil)
	}
	if response, ok, err := s.replay("create_event_series", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.EventSeries{}, err
		}
		var result core.EventSeries
		if err := json.Unmarshal(response, &result); err != nil {
			return core.EventSeries{}, core.E(core.CodeInternal, "decode idempotent event series result", err)
		}
		return result, nil
	}
	category := s.eventCategories[command.CategoryID]
	if category == nil || category.TenantID != command.TenantID || category.Status != "active" {
		return core.EventSeries{}, core.E(core.CodeInvalidInput, "event category is not active in this school", nil)
	}
	host := s.activeAccount(command.HostAccountID, command.TenantID)
	if host == nil {
		return core.EventSeries{}, core.E(core.CodeInvalidInput, "host is not active in this school", nil)
	}
	if command.RoomID != "" {
		stored := s.rooms[command.RoomID]
		if stored == nil || stored.TenantID != command.TenantID || stored.Status != "active" {
			return core.EventSeries{}, core.E(core.CodeInvalidInput, "room is not active in this school", nil)
		}
	}
	stored := &eventSeries{
		ID: command.SeriesID, TenantID: command.TenantID, CategoryID: command.CategoryID,
		Title: command.Title, Description: command.Description,
		HostAccountID: command.HostAccountID, RoomID: command.RoomID, Capacity: command.Capacity,
		Weekday: command.Weekday, StartMinutes: command.StartMinutes, DurationMinutes: command.DurationMinutes,
		EffectiveFrom: command.EffectiveFrom, EffectiveUntil: command.EffectiveUntil,
		Status: "active", Version: 0,
	}
	s.eventSeriesMap[stored.ID] = stored
	result := s.eventSeriesView(stored)
	if err := s.completeIdempotency("create_event_series", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.EventSeries{}, err
	}
	s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "EventSeriesCreated", stored.ID, "allow", "", command.Now, map[string]any{
		"hostAccountId": stored.HostAccountID,
	})
	s.appendOutbox(command.TenantID, "EventSeriesCreated", stored.ID, command.Now)
	return result, nil
}

func (s *Store) eventSeriesView(stored *eventSeries) core.EventSeries {
	view := core.EventSeries{
		ID: stored.ID, CategoryID: stored.CategoryID, Title: stored.Title,
		Description: stored.Description,
		Host:        core.TeacherSummary{AccountID: stored.HostAccountID},
		RoomID:      stored.RoomID, Capacity: stored.Capacity,
		Weekday: stored.Weekday, StartMinutes: stored.StartMinutes,
		DurationMinutes: stored.DurationMinutes,
		EffectiveFrom:   stored.EffectiveFrom, EffectiveUntil: stored.EffectiveUntil,
		Status: stored.Status, Version: stored.Version,
	}
	if account := s.accounts[stored.HostAccountID]; account != nil {
		view.Host.FullName = account.FullName
	}
	return view
}

func (s *Store) GenerateEventOccurrences(_ context.Context, command core.GenerateEventOccurrencesCommand) (core.SeriesOccurrenceGenerationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.eventManager(command.TenantID, command.ActorAccountID) {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeForbidden, "event management permission is required", nil)
	}
	if response, ok, err := s.replay("generate_event_occurrences", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.SeriesOccurrenceGenerationResult{}, err
		}
		var result core.SeriesOccurrenceGenerationResult
		if err := json.Unmarshal(response, &result); err != nil {
			return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInternal, "decode idempotent event generation result", err)
		}
		return result, nil
	}
	stored := s.eventSeriesMap[command.SeriesID]
	if stored == nil || stored.TenantID != command.TenantID {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeNotFound, "event series not found", nil)
	}
	if stored.Status != "active" {
		return core.SeriesOccurrenceGenerationResult{}, core.E(core.CodeInvalidState, "only an active series can generate occurrences", nil)
	}
	created := []string{}
	for _, planned := range command.Occurrences {
		duplicate := false
		for _, existing := range s.eventOccurrences {
			if existing.TenantID == command.TenantID && existing.SeriesID == command.SeriesID &&
				existing.StartsAt.Equal(planned.StartsAt) {
				duplicate = true
			}
		}
		if duplicate {
			continue
		}
		s.eventOccurrences[planned.OccurrenceID] = &eventOccurrence{
			ID: planned.OccurrenceID, TenantID: command.TenantID, SeriesID: command.SeriesID,
			CategoryID: stored.CategoryID, Title: stored.Title, Description: stored.Description,
			StartsAt: planned.StartsAt, DurationMinutes: stored.DurationMinutes,
			HostAccountID: stored.HostAccountID, RoomID: stored.RoomID, Capacity: stored.Capacity,
			Status: "scheduled", Version: 0,
		}
		created = append(created, planned.OccurrenceID)
	}
	result := core.SeriesOccurrenceGenerationResult{SeriesID: command.SeriesID, CreatedCount: len(created), OccurrenceIDs: created}
	if err := s.completeIdempotency("generate_event_occurrences", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.SeriesOccurrenceGenerationResult{}, err
	}
	s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "EventOccurrencesGenerated", command.SeriesID, "allow", "", command.Now, map[string]any{
		"createdCount": len(created),
	})
	s.appendOutbox(command.TenantID, "EventOccurrencesGenerated", command.SeriesID, command.Now)
	return result, nil
}

func (s *Store) CreateEventOccurrence(_ context.Context, command core.CreateEventOccurrenceCommand) (core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.eventManager(command.TenantID, command.ActorAccountID) {
		return core.EventOccurrence{}, core.E(core.CodeForbidden, "event management permission is required", nil)
	}
	if response, ok, err := s.replay("create_event_occurrence", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.EventOccurrence{}, err
		}
		var result core.EventOccurrence
		if err := json.Unmarshal(response, &result); err != nil {
			return core.EventOccurrence{}, core.E(core.CodeInternal, "decode idempotent event result", err)
		}
		return result, nil
	}
	category := s.eventCategories[command.CategoryID]
	if category == nil || category.TenantID != command.TenantID || category.Status != "active" {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, "event category is not active in this school", nil)
	}
	host := s.activeAccount(command.HostAccountID, command.TenantID)
	if host == nil {
		return core.EventOccurrence{}, core.E(core.CodeInvalidInput, "host is not active in this school", nil)
	}
	if command.RoomID != "" {
		stored := s.rooms[command.RoomID]
		if stored == nil || stored.TenantID != command.TenantID || stored.Status != "active" {
			return core.EventOccurrence{}, core.E(core.CodeInvalidInput, "room is not active in this school", nil)
		}
	}
	if !command.StartsAt.After(command.Now) {
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "event must start in the future", nil)
	}
	stored := &eventOccurrence{
		ID: command.OccurrenceID, TenantID: command.TenantID,
		CategoryID: command.CategoryID, Title: command.Title, Description: command.Description,
		StartsAt: command.StartsAt, DurationMinutes: command.DurationMinutes,
		HostAccountID: command.HostAccountID, RoomID: command.RoomID, Capacity: command.Capacity,
		Status: "scheduled", Version: 0,
	}
	s.eventOccurrences[stored.ID] = stored
	result := s.eventOccurrenceView(stored, core.Principal{TenantID: command.TenantID})
	if err := s.completeIdempotency("create_event_occurrence", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.EventOccurrence{}, err
	}
	s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "EventScheduled", stored.ID, "allow", "", command.Now, map[string]any{
		"hostAccountId": stored.HostAccountID,
	})
	s.appendOutbox(command.TenantID, "EventScheduled", stored.ID, command.Now)
	return result, nil
}

func (s *Store) confirmedEventCount(occurrenceID string) int {
	count := 0
	for _, rsvp := range s.eventRsvps[occurrenceID] {
		if rsvp.Status == "confirmed" {
			count++
		}
	}
	return count
}

func (s *Store) pendingOfferFor(occurrenceID string) *spotOffer {
	for _, offer := range s.spotOffers {
		if offer.OccurrenceID == occurrenceID && offer.Status == "pending" {
			return offer
		}
	}
	return nil
}

func (s *Store) heldEventSeats(occurrenceID string) int {
	held := s.confirmedEventCount(occurrenceID)
	if s.pendingOfferFor(occurrenceID) != nil {
		held++
	}
	return held
}

// expireAndCascade retires overdue pending offers and hands every free
// seat to the head of the waitlist, one pending offer at a time.
func (s *Store) expireAndCascade(tenantID, occurrenceID string, now time.Time, ttl time.Duration) {
	if offer := s.pendingOfferFor(occurrenceID); offer != nil && !offer.ExpiresAt.After(now) {
		resolved := now
		offer.Status = "expired"
		offer.ResolvedAt = &resolved
		s.appendSecurityAudit(tenantID, "", "SpotOfferExpired", "spot_offer", offer.ID, "allow", "", now, nil)
	}
	occurrence := s.eventOccurrences[occurrenceID]
	if occurrence == nil {
		return
	}
	for s.heldEventSeats(occurrenceID) < occurrence.Capacity {
		queue := s.eventWaitlists[occurrenceID]
		if len(queue) == 0 {
			return
		}
		head := queue[0]
		s.eventWaitlists[occurrenceID] = queue[1:]
		offerID, err := newMemoryID("offer")
		if err != nil {
			return
		}
		s.spotOffers[offerID] = &spotOffer{
			ID: offerID, TenantID: tenantID, OccurrenceID: occurrenceID,
			StudentID: head.StudentID, Status: "pending",
			OfferedAt: now, ExpiresAt: now.Add(ttl),
		}
		s.appendSecurityAudit(tenantID, "", "SpotOfferCreated", "spot_offer", offerID, "allow", "", now, nil)
		s.appendOutbox(tenantID, "SpotOfferCreated", offerID, now)
		// Single-pending invariant: exactly one open offer per occurrence.
		return
	}
}

func (s *Store) eventStudent(principal core.Principal) (string, error) {
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	if actor == nil || actor.Roles[core.RoleStudent] == "" {
		return "", core.E(core.CodeForbidden, "only a Student can manage own RSVP", nil)
	}
	studentID := s.studentIDForAccount(principal.AccountID)
	if studentID == "" {
		return "", core.E(core.CodeForbidden, "only a Student can manage own RSVP", nil)
	}
	return studentID, nil
}

func (s *Store) eventOccurrenceView(stored *eventOccurrence, principal core.Principal) core.EventOccurrence {
	view := core.EventOccurrence{
		ID: stored.ID, SeriesID: stored.SeriesID, CategoryID: stored.CategoryID,
		Title: stored.Title, Description: stored.Description,
		StartsAt: stored.StartsAt, DurationMinutes: stored.DurationMinutes,
		Host:   core.TeacherSummary{AccountID: stored.HostAccountID},
		RoomID: stored.RoomID, Capacity: stored.Capacity,
		ConfirmedCount: s.confirmedEventCount(stored.ID),
		Status:         stored.Status, Version: stored.Version,
	}
	if category := s.eventCategories[stored.CategoryID]; category != nil {
		view.CategoryName = category.Name
	}
	if account := s.accounts[stored.HostAccountID]; account != nil {
		view.Host.FullName = account.FullName
	}
	if studentID := s.studentIDForAccount(principal.AccountID); studentID != "" {
		if rsvp := s.eventRsvps[stored.ID][studentID]; rsvp != nil {
			view.MyRsvp = rsvp.Status
		}
		for index, entry := range s.eventWaitlists[stored.ID] {
			if entry.StudentID == studentID {
				view.MyWaitlistPosition = index + 1
			}
		}
		if offer := s.pendingOfferFor(stored.ID); offer != nil && offer.StudentID == studentID {
			view.MyOffer = &core.SpotOffer{
				ID: offer.ID, OccurrenceID: offer.OccurrenceID, Status: offer.Status,
				OfferedAt: offer.OfferedAt, ExpiresAt: offer.ExpiresAt,
			}
		}
	}
	return view
}

func (s *Store) ListEventOccurrences(_ context.Context, principal core.Principal, query core.EventListQuery) ([]core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.EventOccurrence{}
	for _, stored := range s.eventOccurrences {
		if stored.TenantID != principal.TenantID {
			continue
		}
		if stored.StartsAt.Before(query.From) || !stored.StartsAt.Before(query.To) {
			continue
		}
		result = append(result, s.eventOccurrenceView(stored, principal))
	}
	sort.Slice(result, func(left, right int) bool {
		if !result[left].StartsAt.Equal(result[right].StartsAt) {
			return result[left].StartsAt.Before(result[right].StartsAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func (s *Store) lockedEventOccurrence(principal core.Principal, occurrenceID string) (*eventOccurrence, error) {
	stored := s.eventOccurrences[occurrenceID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return nil, core.E(core.CodeNotFound, "event not found", nil)
	}
	if stored.Status != "scheduled" {
		return nil, core.E(core.CodeInvalidState, "event is not open for changes", nil)
	}
	return stored, nil
}

func (s *Store) RsvpToEvent(_ context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	studentID, err := s.eventStudent(principal)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	stored, err := s.lockedEventOccurrence(principal, command.OccurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	s.expireAndCascade(principal.TenantID, stored.ID, command.Now, command.OfferTTL)
	if s.eventRsvps[stored.ID] == nil {
		s.eventRsvps[stored.ID] = map[string]*eventRsvp{}
	}
	if existing := s.eventRsvps[stored.ID][studentID]; existing != nil && existing.Status == "confirmed" {
		return s.eventOccurrenceView(stored, principal), nil
	}
	if offer := s.pendingOfferFor(stored.ID); offer != nil && offer.StudentID == studentID {
		resolved := command.Now
		offer.Status = "confirmed"
		offer.ResolvedAt = &resolved
	} else if s.heldEventSeats(stored.ID) >= stored.Capacity {
		return core.EventOccurrence{}, core.E(core.CodeConflict, "event is full — join the waitlist", nil)
	}
	confirmed := command.Now
	s.eventRsvps[stored.ID][studentID] = &eventRsvp{Status: "confirmed", ConfirmedAt: &confirmed}
	s.removeFromWaitlist(stored.ID, studentID)
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "EventRsvpConfirmed",
		"event_occurrence", stored.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "EventRsvpConfirmed", stored.ID, command.Now)
	return s.eventOccurrenceView(stored, principal), nil
}

func (s *Store) removeFromWaitlist(occurrenceID, studentID string) bool {
	queue := s.eventWaitlists[occurrenceID]
	for index, entry := range queue {
		if entry.StudentID == studentID {
			s.eventWaitlists[occurrenceID] = append(queue[:index], queue[index+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) CancelEventRsvp(_ context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	studentID, err := s.eventStudent(principal)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	stored, err := s.lockedEventOccurrence(principal, command.OccurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	existing := s.eventRsvps[stored.ID][studentID]
	if existing == nil || existing.Status != "confirmed" {
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "no confirmed RSVP to cancel", nil)
	}
	cancelled := command.Now
	existing.Status = "cancelled"
	existing.CancelledAt = &cancelled
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "EventRsvpCancelled",
		"event_occurrence", stored.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "EventRsvpCancelled", stored.ID, command.Now)
	s.expireAndCascade(principal.TenantID, stored.ID, command.Now, command.OfferTTL)
	return s.eventOccurrenceView(stored, principal), nil
}

func (s *Store) JoinEventWaitlist(_ context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	studentID, err := s.eventStudent(principal)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	stored, err := s.lockedEventOccurrence(principal, command.OccurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	s.expireAndCascade(principal.TenantID, stored.ID, command.Now, command.OfferTTL)
	if existing := s.eventRsvps[stored.ID][studentID]; existing != nil && existing.Status == "confirmed" {
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "RSVP is already confirmed", nil)
	}
	if offer := s.pendingOfferFor(stored.ID); offer != nil && offer.StudentID == studentID {
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "a spot offer is already waiting for your decision", nil)
	}
	for _, entry := range s.eventWaitlists[stored.ID] {
		if entry.StudentID == studentID {
			return s.eventOccurrenceView(stored, principal), nil
		}
	}
	if s.heldEventSeats(stored.ID) < stored.Capacity {
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "seats are available — RSVP directly", nil)
	}
	s.eventWaitlists[stored.ID] = append(s.eventWaitlists[stored.ID], &waitlistEntry{
		StudentID: studentID, JoinedAt: command.Now,
	})
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "EventWaitlistJoined",
		"event_occurrence", stored.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "EventWaitlistJoined", stored.ID, command.Now)
	return s.eventOccurrenceView(stored, principal), nil
}

func (s *Store) LeaveEventWaitlist(_ context.Context, command core.EventSeatCommand) (core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	studentID, err := s.eventStudent(principal)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	stored, err := s.lockedEventOccurrence(principal, command.OccurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	if !s.removeFromWaitlist(stored.ID, studentID) {
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "not on the waitlist", nil)
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "EventWaitlistLeft",
		"event_occurrence", stored.ID, "allow", "", command.Now, nil)
	return s.eventOccurrenceView(stored, principal), nil
}

func (s *Store) callerSpotOffer(principal core.Principal, offerID string) (*spotOffer, string, error) {
	studentID, err := s.eventStudent(principal)
	if err != nil {
		return nil, "", err
	}
	offer := s.spotOffers[offerID]
	if offer == nil || offer.TenantID != principal.TenantID || offer.StudentID != studentID {
		return nil, "", core.E(core.CodeNotFound, "spot offer not found", nil)
	}
	return offer, studentID, nil
}

func (s *Store) ConfirmSpotOffer(_ context.Context, command core.SpotOfferDecisionCommand) (core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	offer, studentID, err := s.callerSpotOffer(principal, command.OfferID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	stored, err := s.lockedEventOccurrence(principal, offer.OccurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	if offer.Status != "pending" || !offer.ExpiresAt.After(command.Now) {
		s.expireAndCascade(principal.TenantID, stored.ID, command.Now, command.OfferTTL)
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "spot offer has expired", nil)
	}
	resolved := command.Now
	offer.Status = "confirmed"
	offer.ResolvedAt = &resolved
	if s.eventRsvps[stored.ID] == nil {
		s.eventRsvps[stored.ID] = map[string]*eventRsvp{}
	}
	confirmed := command.Now
	s.eventRsvps[stored.ID][studentID] = &eventRsvp{Status: "confirmed", ConfirmedAt: &confirmed}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "SpotOfferConfirmed",
		"spot_offer", offer.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "SpotOfferConfirmed", offer.ID, command.Now)
	return s.eventOccurrenceView(stored, principal), nil
}

func (s *Store) DeclineSpotOffer(_ context.Context, command core.SpotOfferDecisionCommand) (core.EventOccurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	offer, _, err := s.callerSpotOffer(principal, command.OfferID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	stored, err := s.lockedEventOccurrence(principal, offer.OccurrenceID)
	if err != nil {
		return core.EventOccurrence{}, err
	}
	if offer.Status != "pending" {
		return core.EventOccurrence{}, core.E(core.CodeInvalidState, "spot offer is already resolved", nil)
	}
	resolved := command.Now
	offer.Status = "declined"
	offer.ResolvedAt = &resolved
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "SpotOfferDeclined",
		"spot_offer", offer.ID, "allow", "", command.Now, nil)
	s.expireAndCascade(principal.TenantID, stored.ID, command.Now, command.OfferTTL)
	return s.eventOccurrenceView(stored, principal), nil
}

func newMemoryID(prefix string) (string, error) {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(buffer), nil
}

func (s *Store) GetEventSeries(_ context.Context, principal core.Principal, seriesID string) (core.EventSeries, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.eventSeriesMap[seriesID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.EventSeries{}, core.E(core.CodeNotFound, "event series not found", nil)
	}
	return s.eventSeriesView(stored), nil
}
