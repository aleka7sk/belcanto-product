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

// TestPostgreSQLRepertoireHistory proves the StudentSong journey on real
// SQL: sequence-keyed append-only history, CAS on stage changes, and
// trigger-rejected deletion.
func TestPostgreSQLRepertoireHistory(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x71}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgrp", TenantName: "Belcanto PG Repertoire",
		FullName: "PG Repertoire Owner", Phone: "+77009000001",
		Operator: "pg-repertoire-operator", Reason: "repertoire integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77009000001",
		Password: "Pg-repertoire-pass-123!", IdempotencyKey: "pgrp-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77009000001", "Pg-repertoire-pass-123!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Repertoire Teacher", Phone: "+77009000002", Role: core.RoleTeacher,
		Operator: "pg-repertoire-operator", Reason: "repertoire integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77009000002",
		Password: "Pg-repertoire-teach-123!", IdempotencyKey: "pgrp-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77009000002", "Pg-repertoire-teach-123!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Repertoire Student", Phone: "+77009000101", EnrollmentReference: "PGRP-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pgrp-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	song, err := service.AddStudentSong(ctx, teacher, app.AddStudentSongInput{
		StudentID: created.StudentID, Title: "Easy On Me", Artist: "Adele",
		IdempotencyKey: "pgrp-add",
	})
	if err != nil {
		t.Fatalf("add song: %v", err)
	}
	moved, err := service.ChangeSongStage(ctx, teacher, app.ChangeSongStageInput{
		SongID: song.ID, Stage: "learning", StageNote: "Куплет 1",
		ExpectedVersion: song.Version, IdempotencyKey: "pgrp-move",
	})
	if err != nil || moved.Stage != "learning" || len(moved.History) != 2 {
		t.Fatalf("change stage = %#v, %v", moved, err)
	}
	if _, err := service.ChangeSongStage(ctx, teacher, app.ChangeSongStageInput{
		SongID: song.ID, Stage: "interpretation", ExpectedVersion: song.Version,
		IdempotencyKey: "pgrp-stale",
	}); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale CAS = %v, want CONFLICT", err)
	}

	if _, err := pool.Exec(ctx, `
		DELETE FROM student_song_stage_history WHERE song_id = $1
	`, song.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("history DELETE = %v, want immutability rejection", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE student_song_stage_history SET to_stage = 'stage_ready' WHERE song_id = $1
	`, song.ID); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("history UPDATE = %v, want immutability rejection", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM student_songs WHERE id = $1`, song.ID); err == nil ||
		!strings.Contains(err.Error(), "immutable") {
		t.Fatalf("song DELETE = %v, want immutability rejection", err)
	}
}
