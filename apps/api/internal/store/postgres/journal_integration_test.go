package postgres_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

// TestPostgreSQLJournalImmutability proves DEC-007 at the database:
// published journal versions and evidence reject any UPDATE or DELETE,
// and the correction path appends instead of rewriting.
func TestPostgreSQLJournalImmutability(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x7d}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgjr", TenantName: "Belcanto PG Journal",
		FullName: "PG Journal Owner", Phone: "+77006000001",
		Operator: "pg-journal-operator", Reason: "journal integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77006000001",
		Password: "Pg-journal-password-123!", IdempotencyKey: "pgjr-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77006000001", "Pg-journal-password-123!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Journal Teacher", Phone: "+77006000002", Role: core.RoleTeacher,
		Operator: "pg-journal-operator", Reason: "journal integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77006000002",
		Password: "Pg-journal-teacher-123!", IdempotencyKey: "pgjr-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77006000002", "Pg-journal-teacher-123!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Journal Student", Phone: "+77006000101", EnrollmentReference: "PGJR-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pgjr-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	lesson, err := service.ScheduleLesson(ctx, owner, app.ScheduleLessonInput{
		Title: "Вокал · интеграция", StartsAt: time.Now().UTC().Add(time.Minute), DurationMinutes: 45,
		TeacherAccountID: teacher.AccountID, StudentIDs: []string{created.StudentID},
		IdempotencyKey: "pgjr-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE core_lesson_occurrences SET starts_at = now() - interval '1 hour' WHERE id = $1
	`, lesson.ID); err != nil {
		t.Fatalf("backdate lesson: %v", err)
	}

	if _, err := service.SaveJournalDraft(ctx, teacher, app.JournalDraftInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		WhatWorked: "Ровное дыхание", CurrentFocus: "Верхний регистр", NextStep: "Легато",
	}); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	published, err := service.PublishJournal(ctx, teacher, app.PublishJournalInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		Evidence:       []core.EvidenceInput{{Area: "Дыхание", Note: "Фраза 8 тактов"}},
		IdempotencyKey: "pgjr-publish",
	})
	if err != nil {
		t.Fatalf("publish journal: %v", err)
	}
	if published.CurrentVersion != 1 {
		t.Fatalf("published version = %d, want 1", published.CurrentVersion)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE lesson_journal_versions SET what_worked = 'переписано' WHERE journal_id = $1
	`, published.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("version UPDATE = %v, want immutability rejection", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM lesson_journal_versions WHERE journal_id = $1
	`, published.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("version DELETE = %v, want immutability rejection", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM progress_evidence WHERE student_id = $1`, created.StudentID); err == nil ||
		!strings.Contains(err.Error(), "immutable") {
		t.Fatalf("evidence DELETE = %v, want immutability rejection", err)
	}

	if _, err := service.SaveJournalDraft(ctx, teacher, app.JournalDraftInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		WhatWorked: "Ровное дыхание", CurrentFocus: "Верхний регистр", NextStep: "Легато и стаккато",
	}); err != nil {
		t.Fatalf("save correction draft: %v", err)
	}
	corrected, err := service.PublishJournal(ctx, teacher, app.PublishJournalInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		CorrectionNote: "Уточнение после разбора", IdempotencyKey: "pgjr-correct",
	})
	if err != nil {
		t.Fatalf("publish correction: %v", err)
	}
	if corrected.CurrentVersion != 2 || len(corrected.Versions) != 2 {
		t.Fatalf("corrected journal = %#v", corrected)
	}
}
