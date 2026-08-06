package postgres_test

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

// TestPostgreSQLEventLastSeatRace proves the atomic-last-seat property
// against the real schema: two students race for the final seat and
// exactly one wins; the loser gets CONFLICT, joins the waitlist, and the
// winner's cancellation hands the seat back as a spot offer.
func TestPostgreSQLEventLastSeatRace(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, _ := isolatedPool(t, ctx, databaseURL)
	if err := migrations.Up(ctx, pool); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	store := postgres.New(pool)
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x7b}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgev", TenantName: "Belcanto PG Events",
		FullName: "PG Events Owner", Phone: "+77005000001",
		Operator: "pg-events-operator", Reason: "events integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	const ownerPassword = "Pg-events-password-123!"
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77005000001",
		Password: ownerPassword, IdempotencyKey: "pgev-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77005000001", ownerPassword, core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Events Teacher", Phone: "+77005000002", Role: core.RoleTeacher,
		Operator: "pg-events-operator", Reason: "events integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77005000002",
		Password: "Pg-events-teacher-123!", IdempotencyKey: "pgev-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77005000002", "Pg-events-teacher-123!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	students := make([]core.Principal, 0, 2)
	for index := 0; index < 2; index++ {
		phone := []string{"+77005000101", "+77005000102"}[index]
		enrollment := []string{"PGEV-101", "PGEV-102"}[index]
		created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
			FullName: "PG Events Student", Phone: phone, EnrollmentReference: enrollment,
			TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
			AdultConfirmed: true, IdempotencyKey: "pgev-create-" + enrollment,
		})
		if err != nil {
			t.Fatalf("create student %d: %v", index, err)
		}
		if _, err := service.PublishFirstMinute(ctx, teacher, app.PublishFirstMinuteInput{
			StudentID: created.StudentID, WhatWorked: "Worked", CurrentFocus: "Focus",
			NextStep: "Next", ExpectedVersion: 0, IdempotencyKey: "pgev-first-" + enrollment,
		}); err != nil {
			t.Fatalf("publish first minute %d: %v", index, err)
		}
		_, link, err := service.IssueInvitation(ctx, owner, created.StudentID, "pgev-invite-"+enrollment, core.InvitationIssue)
		if err != nil {
			t.Fatalf("invite student %d: %v", index, err)
		}
		if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
			Token: integrationToken(t, link), Phone: phone,
			Password: "Pg-events-student-123!", IdempotencyKey: "pgev-activate-" + enrollment,
		}); err != nil {
			t.Fatalf("activate student %d: %v", index, err)
		}
		outcome, err := service.SignIn(ctx, phone, "Pg-events-student-123!", core.SessionClientInfo{})
		if err != nil || outcome.Tokens == nil {
			t.Fatalf("sign in student %d: %v", index, err)
		}
		principal, err := service.Authenticate(ctx, outcome.Tokens.AccessToken)
		if err != nil {
			t.Fatalf("authenticate student %d: %v", index, err)
		}
		students = append(students, principal)
	}

	category, err := service.CreateEventCategory(ctx, owner, "Концерт")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	event, err := service.CreateEvent(ctx, owner, app.CreateEventInput{
		CategoryID: category.ID, Title: "Отчётный концерт",
		StartsAt: time.Now().UTC().Add(96 * time.Hour), DurationMinutes: 120,
		HostAccountID: teacher.AccountID, Capacity: 1,
		IdempotencyKey: "pgev-event",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	var waitGroup sync.WaitGroup
	results := make([]error, 2)
	for index := 0; index < 2; index++ {
		waitGroup.Add(1)
		go func(slot int) {
			defer waitGroup.Done()
			_, results[slot] = service.RsvpToEvent(ctx, students[slot], event.ID)
		}(index)
	}
	waitGroup.Wait()

	winners, conflicts := 0, 0
	loser := -1
	for index, raceErr := range results {
		switch {
		case raceErr == nil:
			winners++
		case core.IsCode(raceErr, core.CodeConflict):
			conflicts++
			loser = index
		default:
			t.Fatalf("unexpected race outcome %d: %v", index, raceErr)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("last-seat race: winners=%d conflicts=%d, want exactly 1/1", winners, conflicts)
	}

	if _, err := service.JoinEventWaitlist(ctx, students[loser], event.ID); err != nil {
		t.Fatalf("loser joins waitlist: %v", err)
	}
	winner := 1 - loser
	if _, err := service.CancelEventRsvp(ctx, students[winner], event.ID); err != nil {
		t.Fatalf("winner cancels: %v", err)
	}
	views, err := service.ListEvents(ctx, students[loser],
		time.Now().UTC(), time.Now().UTC().Add(7*24*time.Hour))
	if err != nil || len(views) != 1 {
		t.Fatalf("event catalog = %#v, %v", views, err)
	}
	if views[0].MyOffer == nil {
		t.Fatalf("cancellation did not cascade a spot offer: %#v", views[0])
	}
	confirmed, err := service.ConfirmSpotOffer(ctx, students[loser], views[0].MyOffer.ID)
	if err != nil {
		t.Fatalf("confirm cascaded offer: %v", err)
	}
	if confirmed.ConfirmedCount != 1 || confirmed.MyRsvp != "confirmed" {
		t.Fatalf("confirmed view = %#v", confirmed)
	}
}
