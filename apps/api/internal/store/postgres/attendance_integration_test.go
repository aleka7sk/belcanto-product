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

// TestPostgreSQLAttendanceHistory proves attendance on real SQL: a
// correction demands a reason, the row survives (DELETE rejected by the
// trigger) and the Student's view hides the teacher note.
func TestPostgreSQLAttendanceHistory(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x7f}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgat", TenantName: "Belcanto PG Attendance",
		FullName: "PG Attendance Owner", Phone: "+77008000001",
		Operator: "pg-attendance-operator", Reason: "attendance integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77008000001",
		Password: "Pg-attendance-pass-123!", IdempotencyKey: "pgat-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77008000001", "Pg-attendance-pass-123!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Attendance Teacher", Phone: "+77008000002", Role: core.RoleTeacher,
		Operator: "pg-attendance-operator", Reason: "attendance integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77008000002",
		Password: "Pg-attendance-teach-123!", IdempotencyKey: "pgat-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77008000002", "Pg-attendance-teach-123!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Attendance Student", Phone: "+77008000101", EnrollmentReference: "PGAT-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pgat-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	lesson, err := service.ScheduleLesson(ctx, owner, app.ScheduleLessonInput{
		Title: "Вокал · посещаемость", StartsAt: time.Now().UTC().Add(time.Hour), DurationMinutes: 45,
		TeacherAccountID: teacher.AccountID, StudentIDs: []string{created.StudentID},
		IdempotencyKey: "pgat-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}

	if _, err := service.MarkAttendance(ctx, teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		Status: core.AttendanceLate, LateMinutes: 7, Note: "Приватная заметка",
		IdempotencyKey: "pgat-late",
	}); err != nil {
		t.Fatalf("mark late: %v", err)
	}
	if _, err := service.MarkAttendance(ctx, teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		Status: core.AttendancePresent, IdempotencyKey: "pgat-change-bare",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("correction without reason = %v, want INVALID_INPUT", err)
	}
	records, err := service.MarkAttendance(ctx, teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: created.StudentID,
		Status: core.AttendancePresent, ChangeReason: "Опоздание не подтвердилось",
		IdempotencyKey: "pgat-change",
	})
	if err != nil || records[0].Status != core.AttendancePresent {
		t.Fatalf("corrected attendance = %#v, %v", records, err)
	}

	if _, err := pool.Exec(ctx, `
		DELETE FROM core_lesson_attendance WHERE occurrence_id = $1
	`, lesson.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("attendance DELETE = %v, want immutability rejection", err)
	}
}
