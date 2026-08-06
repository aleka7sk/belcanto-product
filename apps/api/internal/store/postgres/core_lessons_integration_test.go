package postgres_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

// TestPostgreSQLSeriesLifecycle proves the weekly series lifecycle on
// real SQL: occurrences copy the series format and room into the Lesson
// view, pausing blocks generation, ending is terminal, and scheduled
// Lessons stay untouched throughout.
func TestPostgreSQLSeriesLifecycle(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x74}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgser", TenantName: "Belcanto PG Series",
		FullName: "PG Series Owner", Phone: "+77001300001",
		Operator: "pg-series-operator", Reason: "series integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77001300001",
		Password: "Pg-series-owner-1!", IdempotencyKey: "pgser-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77001300001", "Pg-series-owner-1!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Series Teacher", Phone: "+77001300002", Role: core.RoleTeacher,
		Operator: "pg-series-operator", Reason: "series integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77001300002",
		Password: "Pg-series-teacher-1!", IdempotencyKey: "pgser-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77001300002", "Pg-series-teacher-1!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Series Student", Phone: "+77001300101", EnrollmentReference: "PGSER-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pgser-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	room, err := service.CreateRoom(ctx, owner, app.CreateRoomInput{Name: "Зал «Ария»"})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	series, err := service.CreateCoreLessonSeries(ctx, owner, app.CreateCoreLessonSeriesInput{
		Format: "individual", Title: "Вокальная техника",
		TeacherAccountID: teacher.AccountID, RoomID: room.ID,
		Weekday: 0, StartMinutes: 600, DurationMinutes: 45,
		EffectiveFrom:  time.Now().UTC().Format("2006-01-02"),
		StudentIDs:     []string{created.StudentID},
		IdempotencyKey: "pgser-series",
	})
	if err != nil {
		t.Fatalf("create series: %v", err)
	}
	generation, err := service.GenerateSeriesOccurrences(ctx, owner, series.ID, 2, "pgser-gen")
	if err != nil || generation.CreatedCount == 0 {
		t.Fatalf("generate occurrences = %#v, %v", generation, err)
	}
	lesson, err := service.GetLesson(ctx, owner, generation.OccurrenceIDs[0])
	if err != nil {
		t.Fatalf("read occurrence: %v", err)
	}
	if lesson.Format != "individual" || lesson.Location != "Зал «Ария»" {
		t.Fatalf("occurrence view = format %q, location %q", lesson.Format, lesson.Location)
	}

	paused, err := service.ChangeCoreLessonSeriesStatus(ctx, owner, app.ChangeSeriesStatusInput{
		SeriesID: series.ID, Status: "paused", ExpectedVersion: series.Version,
		IdempotencyKey: "pgser-pause",
	})
	if err != nil || paused.Status != "paused" {
		t.Fatalf("pause = %#v, %v", paused, err)
	}
	if _, err := service.GenerateSeriesOccurrences(ctx, owner, series.ID, 4, "pgser-gen-paused"); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("generation on paused = %v, want INVALID_STATE", err)
	}
	ended, err := service.ChangeCoreLessonSeriesStatus(ctx, owner, app.ChangeSeriesStatusInput{
		SeriesID: series.ID, Status: "ended", ExpectedVersion: paused.Version,
		IdempotencyKey: "pgser-end",
	})
	if err != nil || ended.Status != "ended" {
		t.Fatalf("end = %#v, %v", ended, err)
	}
	if _, err := service.ChangeCoreLessonSeriesStatus(ctx, owner, app.ChangeSeriesStatusInput{
		SeriesID: series.ID, Status: "active", ExpectedVersion: ended.Version,
		IdempotencyKey: "pgser-reopen",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("reopen ended = %v, want INVALID_STATE", err)
	}
	var scheduled int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM core_lesson_occurrences WHERE series_id = $1 AND status = 'scheduled'
	`, series.ID).Scan(&scheduled); err != nil || scheduled != generation.CreatedCount {
		t.Fatalf("scheduled occurrences after lifecycle = %d, %v", scheduled, err)
	}
}
