package postgres_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/notify"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

// TestPostgreSQLOutboxDelivery proves the notification flow on real SQL:
// the worker drains real domain events into activity rows, marks the
// outbox delivered, and a poisoned event walks retry columns into
// dead_letter without blocking the queue.
func TestPostgreSQLOutboxDelivery(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x72}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgnt", TenantName: "Belcanto PG Notify",
		FullName: "PG Notify Owner", Phone: "+77001100001",
		Operator: "pg-notify-operator", Reason: "notify integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77001100001",
		Password: "Pg-notify-password-123!", IdempotencyKey: "pgnt-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77001100001", "Pg-notify-password-123!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Notify Teacher", Phone: "+77001100002", Role: core.RoleTeacher,
		Operator: "pg-notify-operator", Reason: "notify integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77001100002",
		Password: "Pg-notify-teacher-123!", IdempotencyKey: "pgnt-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77001100002", "Pg-notify-teacher-123!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Notify Student", Phone: "+77001100101", EnrollmentReference: "PGNT-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pgnt-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if _, err := service.PublishFirstMinute(ctx, teacher, app.PublishFirstMinuteInput{
		StudentID: created.StudentID, WhatWorked: "Опора",
		CurrentFocus: "Верх", NextStep: "Легато",
		IdempotencyKey: "pgnt-first-minute",
	}); err != nil {
		t.Fatalf("publish first minute: %v", err)
	}
	_, invitationLink, err := service.IssueInvitation(ctx, owner, created.StudentID, "pgnt-invite", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue invitation: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, invitationLink), Phone: "+77001100101",
		Password: "Pg-notify-student-123!", IdempotencyKey: "pgnt-activate-student",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	studentOutcome, err := service.SignIn(ctx, "+77001100101", "Pg-notify-student-123!", core.SessionClientInfo{})
	if err != nil || studentOutcome.Tokens == nil {
		t.Fatalf("sign in student: %v", err)
	}
	student, err := service.Authenticate(ctx, studentOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate student: %v", err)
	}

	goal, err := service.CreateGoal(ctx, teacher, app.CreateGoalInput{
		StudentID: created.StudentID, Criterion: "Свободный припев",
		IdempotencyKey: "pgnt-goal",
	})
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	if _, err := service.CompleteGoal(ctx, teacher, app.CompleteGoalInput{
		GoalID: goal.ID, CompletionNote: "Подтверждено записью",
		ExpectedVersion: goal.Version, IdempotencyKey: "pgnt-goal-complete",
	}); err != nil {
		t.Fatalf("complete goal: %v", err)
	}

	worker := notify.NewWorker(store, notify.Options{})
	if _, _, err := worker.DrainOnce(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	feed, err := service.ActivityFeed(ctx, student)
	if err != nil {
		t.Fatalf("student feed: %v", err)
	}
	sawGoal := false
	for _, entry := range feed.Entries {
		if entry.Kind == "GoalCompleted" {
			sawGoal = true
		}
	}
	if !sawGoal {
		t.Fatalf("GoalCompleted missing from feed: %#v", feed.Entries)
	}
	var pendingCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM outbox_events WHERE status = 'pending'
	`).Scan(&pendingCount); err != nil || pendingCount != 0 {
		t.Fatalf("pending after drain = %d, %v", pendingCount, err)
	}

	// Poison: a routed event whose payload cannot resolve a recipient.
	if _, err := pool.Exec(ctx, `
		INSERT INTO outbox_events (tenant_id, event_type, aggregate_type, aggregate_id, payload, recorded_at)
		VALUES ($1, 'HomeworkSubmitted', 'homework', 'hw_broken', '{}'::jsonb, now())
	`, owner.TenantID); err != nil {
		t.Fatalf("insert poison event: %v", err)
	}
	fastWorker := notify.NewWorker(store, notify.Options{MaxAttempts: 2, BackoffBase: time.Millisecond})
	for attempt := 0; attempt < 3; attempt++ {
		if _, _, err := fastWorker.DrainOnce(ctx); err != nil {
			t.Fatalf("drain poison: %v", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
	var status, lastError string
	var attempts int
	if err := pool.QueryRow(ctx, `
		SELECT status, attempt_count, COALESCE(last_error, '')
		FROM outbox_events WHERE aggregate_id = 'hw_broken'
	`).Scan(&status, &attempts, &lastError); err != nil {
		t.Fatalf("read poison state: %v", err)
	}
	if status != "dead_letter" || attempts < 2 || lastError == "" {
		t.Fatalf("poison state = %s / %d attempts / %q", status, attempts, lastError)
	}
}
