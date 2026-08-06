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

// TestPostgreSQLAssessments proves the assessment invariants on real
// SQL: the guard trigger freezes published content, deletes are
// rejected outright, draft evidence stays editable until publication,
// and the supersede chain carries evidence under new identifiers.
func TestPostgreSQLAssessments(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x75}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgasmt", TenantName: "Belcanto PG Assessments",
		FullName: "PG Assessment Owner", Phone: "+77001400001",
		Operator: "pg-asmt-operator", Reason: "assessment integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77001400001",
		Password: "Pg-asmt-owner-1!", IdempotencyKey: "pgasmt-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77001400001", "Pg-asmt-owner-1!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Assessment Teacher", Phone: "+77001400002", Role: core.RoleTeacher,
		Operator: "pg-asmt-operator", Reason: "assessment integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77001400002",
		Password: "Pg-asmt-teacher-1!", IdempotencyKey: "pgasmt-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77001400002", "Pg-asmt-teacher-1!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Assessment Student", Phone: "+77001400101", EnrollmentReference: "PGASMT-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pgasmt-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	content := app.AssessmentContentInput{
		Type: "formative", ContextType: "lesson",
		AssessmentDate: "2026-08-01",
		Summary:        "Опора устойчива в коротких фразах.",
		Strengths:      "Стабильное окончание фразы.",
		Visibility:     "student_visible",
	}
	draft, err := service.CreateAssessment(ctx, teacher, app.CreateAssessmentInput{
		StudentID: created.StudentID, Content: content, IdempotencyKey: "pgasmt-create",
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	withEvidence, err := service.AddAssessmentEvidence(ctx, teacher, app.AddAssessmentEvidenceInput{
		AssessmentID: draft.ID, Kind: "observation",
		Note: "Три повтора без подсказки.", IdempotencyKey: "pgasmt-evidence",
	})
	if err != nil || len(withEvidence.Evidence) != 1 {
		t.Fatalf("add evidence = %#v, %v", withEvidence, err)
	}
	// Draft evidence removal is allowed before publication.
	trimmed, err := service.RemoveAssessmentEvidence(ctx, teacher, draft.ID,
		withEvidence.Evidence[0].ID, "pgasmt-remove-evidence")
	if err != nil || len(trimmed.Evidence) != 0 {
		t.Fatalf("remove draft evidence = %#v, %v", trimmed, err)
	}
	restored, err := service.AddAssessmentEvidence(ctx, teacher, app.AddAssessmentEvidenceInput{
		AssessmentID: draft.ID, Kind: "observation",
		Note: "Полный куплет на уроке 1 августа.", IdempotencyKey: "pgasmt-evidence-2",
	})
	if err != nil {
		t.Fatalf("re-add evidence: %v", err)
	}
	published, err := service.PublishAssessment(ctx, teacher, app.PublishAssessmentInput{
		AssessmentID: draft.ID, ExpectedVersion: restored.Version,
		IdempotencyKey: "pgasmt-publish",
	})
	if err != nil || published.Status != "published" {
		t.Fatalf("publish = %#v, %v", published, err)
	}

	// The SQL guard freezes published content against any direct write.
	if _, err := pool.Exec(ctx, `
		UPDATE assessments SET summary = 'подмена' WHERE id = $1
	`, draft.ID); err == nil {
		t.Fatal("direct rewrite of published content must be rejected")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM assessments WHERE id = $1`, draft.ID); err == nil {
		t.Fatal("deleting an assessment must be rejected")
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM assessment_evidence WHERE assessment_id = $1
	`, draft.ID); err == nil {
		t.Fatal("deleting published evidence must be rejected")
	}

	correction := content
	correction.Summary = "Опора устойчива и в полном куплете."
	chain, err := service.SupersedeAssessment(ctx, teacher, app.SupersedeAssessmentInput{
		AssessmentID: draft.ID, Content: correction, IdempotencyKey: "pgasmt-supersede",
	})
	if err != nil || len(chain) != 2 {
		t.Fatalf("supersede = %#v, %v", chain, err)
	}
	if chain[0].Status != "superseded" || chain[1].Status != "published" || len(chain[1].Evidence) != 1 {
		t.Fatalf("chain = %#v", chain)
	}
	var evidenceCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM assessment_evidence WHERE assessment_id = $1
	`, chain[1].ID).Scan(&evidenceCount); err != nil || evidenceCount != 1 {
		t.Fatalf("carried evidence rows = %d, %v", evidenceCount, err)
	}
}
