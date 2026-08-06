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

// TestPostgreSQLHomeworkHistory proves the approved homework invariants
// at the database: completed homework and practice history reject any
// UPDATE or DELETE, tasks lock after assignment, and the whole
// assigned → submitted → reviewed → completed loop runs on real SQL.
func TestPostgreSQLHomeworkHistory(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x7e}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pghw", TenantName: "Belcanto PG Homework",
		FullName: "PG Homework Owner", Phone: "+77007000001",
		Operator: "pg-homework-operator", Reason: "homework integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77007000001",
		Password: "Pg-homework-password-123!", IdempotencyKey: "pghw-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77007000001", "Pg-homework-password-123!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Homework Teacher", Phone: "+77007000002", Role: core.RoleTeacher,
		Operator: "pg-homework-operator", Reason: "homework integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77007000002",
		Password: "Pg-homework-teacher-123!", IdempotencyKey: "pghw-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77007000002", "Pg-homework-teacher-123!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Homework Student", Phone: "+77007000101", EnrollmentReference: "PGHW-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pghw-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if _, err := service.PublishFirstMinute(ctx, teacher, app.PublishFirstMinuteInput{
		StudentID: created.StudentID, WhatWorked: "Ровное дыхание",
		CurrentFocus: "Верхний регистр", NextStep: "Легато",
		IdempotencyKey: "pghw-first-minute",
	}); err != nil {
		t.Fatalf("publish first minute: %v", err)
	}
	_, invitationLink, err := service.IssueInvitation(ctx, owner, created.StudentID, "pghw-invite", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue invitation: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, invitationLink), Phone: "+77007000101",
		Password: "Pg-homework-student-123!", IdempotencyKey: "pghw-activate-student",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	studentOutcome, err := service.SignIn(ctx, "+77007000101", "Pg-homework-student-123!", core.SessionClientInfo{})
	if err != nil || studentOutcome.Tokens == nil {
		t.Fatalf("sign in student: %v", err)
	}
	student, err := service.Authenticate(ctx, studentOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate student: %v", err)
	}

	lesson, err := service.ScheduleLesson(ctx, owner, app.ScheduleLessonInput{
		Title: "Вокал · практика", StartsAt: time.Now().UTC().Add(time.Hour), DurationMinutes: 45,
		TeacherAccountID: teacher.AccountID, StudentIDs: []string{created.StudentID},
		IdempotencyKey: "pghw-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}

	payload := bytes.Repeat([]byte{0x5A}, 48)
	media, err := service.CreateMedia(ctx, student, app.CreateMediaInput{
		Kind: "audio", ContentType: "audio/m4a", ByteSize: int64(len(payload)),
		IdempotencyKey: "pghw-media",
	})
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	if _, err := service.AppendMediaChunk(ctx, student, media.ID, 0, payload); err != nil {
		t.Fatalf("upload media: %v", err)
	}

	homework, err := service.CreateHomework(ctx, teacher, app.CreateHomeworkInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		Goal: "Три повтора в 75% темпа",
		Tasks: []core.HomeworkTaskInput{
			{Title: "Разминка", RecommendedMinutes: 3},
		},
		Assign: true, IdempotencyKey: "pghw-homework",
	})
	if err != nil {
		t.Fatalf("create homework: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		DELETE FROM homework_assignments WHERE id = $1
	`, homework.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("homework DELETE = %v, want immutability rejection", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM homework_tasks WHERE homework_id = $1
	`, homework.ID); err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("assigned task DELETE = %v, want draft-only rejection", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE homework_tasks SET title = 'переписано' WHERE homework_id = $1
	`, homework.ID); err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("assigned task UPDATE = %v, want status-only rejection", err)
	}

	if _, err := service.StartHomework(ctx, student, homework.ID, "pghw-start"); err != nil {
		t.Fatalf("start homework: %v", err)
	}
	current, err := service.GetHomework(ctx, student, homework.ID)
	if err != nil {
		t.Fatalf("read homework: %v", err)
	}
	submitted, err := service.SubmitHomework(ctx, student, app.SubmitHomeworkInput{
		HomeworkID: homework.ID, Note: "Две попытки", MediaIDs: []string{media.ID},
		ExpectedVersion: current.Version, IdempotencyKey: "pghw-submit",
	})
	if err != nil {
		t.Fatalf("submit homework: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE practice_submissions SET note = 'переписано' WHERE homework_id = $1
	`, homework.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("submission UPDATE = %v, want immutability rejection", err)
	}

	accepted, err := service.ReviewHomework(ctx, teacher, app.ReviewHomeworkInput{
		HomeworkID: homework.ID, Decision: core.FeedbackDecisionAccepted,
		Body: "Принято", EvidenceArea: "Дыхание", EvidenceNote: "Фраза целиком",
		ExpectedVersion: submitted.Version, IdempotencyKey: "pghw-review",
	})
	if err != nil || accepted.Status != core.HomeworkStatusCompleted {
		t.Fatalf("accept homework = %#v, %v", accepted, err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE homework_assignments SET goal = 'переписано' WHERE id = $1
	`, homework.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("completed homework UPDATE = %v, want immutability rejection", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM practice_feedback WHERE homework_id = $1
	`, homework.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("feedback DELETE = %v, want immutability rejection", err)
	}

	evidence, err := service.ListProgressEvidence(ctx, student, created.StudentID)
	if err != nil || len(evidence) != 1 || evidence[0].SourceKind != "practice" {
		t.Fatalf("practice evidence = %#v, %v", evidence, err)
	}

	access, err := service.SignMediaAccess(ctx, teacher, media.ID)
	if err != nil {
		t.Fatalf("sign media access: %v", err)
	}
	token := access.URL[strings.Index(access.URL, "token=")+len("token="):]
	content, contentType, err := service.MediaContentByToken(ctx, media.ID, token)
	if err != nil || contentType != "audio/m4a" || !bytes.Equal(content, payload) {
		t.Fatalf("media content = %d bytes, %q, %v", len(content), contentType, err)
	}
}
