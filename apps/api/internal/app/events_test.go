package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestEventSeatLifecycle walks the whole seat machine: capacity fill,
// explicit waitlist (never silent), cancellation cascading a spot offer
// to the waitlist head, offer expiry passing the seat to the next
// student, and offer confirmation converting the held seat into an RSVP
// (DEC-003: everything binds to one occurrence; DEC-101: the TTL is
// configuration).
func TestEventSeatLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	students := make([]core.Principal, 0, 3)
	for index := 0; index < 3; index++ {
		phone := []string{"+77000000401", "+77000000402", "+77000000403"}[index]
		enrollment := []string{"ENR-401", "ENR-402", "ENR-403"}[index]
		_, link := readyStudentInvitation(t, fixture, phone, enrollment)
		password := "Event-student-password-1!"
		if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
			Token: tokenFromLink(t, link), Phone: phone, Password: password,
			IdempotencyKey: "event-activate-" + enrollment,
		}); err != nil {
			t.Fatalf("activate student %d: %v", index, err)
		}
		students = append(students, signInPrincipal(t, fixture.service, phone, password))
	}

	category, err := fixture.service.CreateEventCategory(ctx, fixture.owner, "Мастер-класс")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	event, err := fixture.service.CreateEvent(ctx, fixture.owner, app.CreateEventInput{
		CategoryID: category.ID, Title: "Мастер-класс по дыханию",
		StartsAt: fixture.clock.Now().Add(72 * time.Hour), DurationMinutes: 90,
		HostAccountID: fixture.teacher.AccountID, Capacity: 1,
		IdempotencyKey: "event-create",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	first, err := fixture.service.RsvpToEvent(ctx, students[0], event.ID)
	if err != nil {
		t.Fatalf("first RSVP: %v", err)
	}
	if first.ConfirmedCount != 1 || first.MyRsvp != "confirmed" {
		t.Fatalf("first RSVP view = %#v", first)
	}

	if _, err := fixture.service.RsvpToEvent(ctx, students[1], event.ID); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("RSVP on full event = %v, want CONFLICT", err)
	}
	if _, err := fixture.service.JoinEventWaitlist(ctx, students[0], event.ID); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("confirmed student joining waitlist = %v, want INVALID_STATE", err)
	}

	second, err := fixture.service.JoinEventWaitlist(ctx, students[1], event.ID)
	if err != nil {
		t.Fatalf("second joins waitlist: %v", err)
	}
	if second.MyWaitlistPosition != 1 {
		t.Fatalf("second waitlist position = %d, want 1", second.MyWaitlistPosition)
	}
	third, err := fixture.service.JoinEventWaitlist(ctx, students[2], event.ID)
	if err != nil {
		t.Fatalf("third joins waitlist: %v", err)
	}
	if third.MyWaitlistPosition != 2 {
		t.Fatalf("third waitlist position = %d, want 2", third.MyWaitlistPosition)
	}

	afterCancel, err := fixture.service.CancelEventRsvp(ctx, students[0], event.ID)
	if err != nil {
		t.Fatalf("cancel RSVP: %v", err)
	}
	if afterCancel.ConfirmedCount != 0 {
		t.Fatalf("confirmed after cancel = %d, want 0", afterCancel.ConfirmedCount)
	}

	events, err := fixture.service.ListEvents(ctx, students[1],
		fixture.clock.Now(), fixture.clock.Now().Add(7*24*time.Hour))
	if err != nil || len(events) != 1 {
		t.Fatalf("event catalog = %#v, %v", events, err)
	}
	offered := events[0]
	if offered.MyOffer == nil {
		t.Fatalf("waitlist head has no spot offer: %#v", offered)
	}
	if offered.MyWaitlistPosition != 0 {
		t.Fatalf("offered student still waitlisted at %d", offered.MyWaitlistPosition)
	}

	// The seat is held by the pending offer: a direct RSVP by anyone else
	// stays CONFLICT even though confirmed = 0.
	if _, err := fixture.service.RsvpToEvent(ctx, students[2], event.ID); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("RSVP while seat held by offer = %v, want CONFLICT", err)
	}

	// Let the offer expire: the seat cascades to the next student.
	fixture.clock.Advance(25 * time.Hour)
	if _, err := fixture.service.ConfirmSpotOffer(ctx, students[1], offered.MyOffer.ID); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("confirm expired offer = %v, want INVALID_STATE", err)
	}
	events, err = fixture.service.ListEvents(ctx, students[2],
		fixture.clock.Now(), fixture.clock.Now().Add(7*24*time.Hour))
	if err != nil || len(events) != 1 || events[0].MyOffer == nil {
		t.Fatalf("cascaded offer view = %#v, %v", events, err)
	}

	confirmed, err := fixture.service.ConfirmSpotOffer(ctx, students[2], events[0].MyOffer.ID)
	if err != nil {
		t.Fatalf("confirm cascaded offer: %v", err)
	}
	if confirmed.ConfirmedCount != 1 || confirmed.MyRsvp != "confirmed" {
		t.Fatalf("confirmed view = %#v", confirmed)
	}
}
