package app_test

import (
	"context"
	"testing"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestAssessmentLifecycle pins domain/assessment.md: only the assigned
// Teacher authors; a draft needs substance to publish; published
// content is immutable and corrects through a linked superseding
// version that carries the evidence; withdrawal keeps the record with
// a mandatory reason; visibility gates every read.
func TestAssessmentLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000001401", "ENR-1401")
	const studentPassword = "Assessment-student-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000001401", Password: studentPassword,
		IdempotencyKey: "asmt-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000001401", studentPassword)
	directory, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(directory) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := directory[len(directory)-1].StudentID

	content := app.AssessmentContentInput{
		Type: "formative", ContextType: "lesson",
		AssessmentDate: "2026-08-01",
		Summary:        "Во фразах до восьми долей опора сохраняется устойчиво.",
		Strengths:      "Удерживает окончание фразы без потери опоры.",
		Visibility:     "student_visible",
		Areas:          "Дыхание и опора",
	}

	// Assessing is a pedagogical right: an administrator does not author.
	if _, err := fixture.service.CreateAssessment(ctx, fixture.admin, app.CreateAssessmentInput{
		StudentID: studentID, Content: content, IdempotencyKey: "asmt-admin",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("administrator authoring = %v, want FORBIDDEN", err)
	}
	if _, err := fixture.service.CreateAssessment(ctx, fixture.teacher, app.CreateAssessmentInput{
		StudentID: studentID,
		Content: app.AssessmentContentInput{
			Type: "formative", ContextType: "", AssessmentDate: "2026-08-01",
			Visibility: "student_visible",
		},
		IdempotencyKey: "asmt-no-context",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("assessment without context = %v, want INVALID_INPUT", err)
	}

	draft, err := fixture.service.CreateAssessment(ctx, fixture.teacher, app.CreateAssessmentInput{
		StudentID: studentID, Content: content, IdempotencyKey: "asmt-create",
	})
	if err != nil || draft.Status != "draft" {
		t.Fatalf("create draft = %#v, %v", draft, err)
	}

	// The student never sees a draft.
	if list, err := fixture.service.ListStudentAssessments(ctx, student, studentID); err != nil || len(list) != 0 {
		t.Fatalf("student sees drafts = %d, %v", len(list), err)
	}
	// The administrator sees drafts (administrative permission).
	if list, err := fixture.service.ListStudentAssessments(ctx, fixture.admin, studentID); err != nil || len(list) != 1 {
		t.Fatalf("admin draft view = %d, %v", len(list), err)
	}

	withEvidence, err := fixture.service.AddAssessmentEvidence(ctx, fixture.teacher, app.AddAssessmentEvidenceInput{
		AssessmentID: draft.ID, Kind: "observation",
		Note:           "Три повтора подряд без подсказки на уроке 30 июля.",
		IdempotencyKey: "asmt-evidence",
	})
	if err != nil || len(withEvidence.Evidence) != 1 {
		t.Fatalf("add evidence = %#v, %v", withEvidence, err)
	}

	published, err := fixture.service.PublishAssessment(ctx, fixture.teacher, app.PublishAssessmentInput{
		AssessmentID: draft.ID, ExpectedVersion: withEvidence.Version,
		IdempotencyKey: "asmt-publish",
	})
	if err != nil || published.Status != "published" || published.PublishedAt == nil {
		t.Fatalf("publish = %#v, %v", published, err)
	}

	// Published content is immutable — the draft update path refuses.
	if _, err := fixture.service.UpdateAssessmentDraft(ctx, fixture.teacher, app.UpdateAssessmentDraftInput{
		AssessmentID: draft.ID, Content: content,
		ExpectedVersion: published.Version, IdempotencyKey: "asmt-edit-published",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("editing published = %v, want INVALID_STATE", err)
	}

	// The student sees the published student_visible assessment.
	visible, err := fixture.service.ListStudentAssessments(ctx, student, studentID)
	if err != nil || len(visible) != 1 || visible[0].Summary == "" {
		t.Fatalf("student view = %#v, %v", visible, err)
	}

	// Correction: a superseding version carries the evidence forward.
	correction := content
	correction.Summary = "Опора устойчива и в полном куплете; фиксирую перенос в новую песню."
	chain, err := fixture.service.SupersedeAssessment(ctx, fixture.teacher, app.SupersedeAssessmentInput{
		AssessmentID: draft.ID, Content: correction, IdempotencyKey: "asmt-supersede",
	})
	if err != nil || len(chain) != 2 {
		t.Fatalf("supersede = %#v, %v", chain, err)
	}
	if chain[0].Status != "superseded" || chain[0].SupersededByID != chain[1].ID {
		t.Fatalf("superseded link = %#v", chain[0])
	}
	if chain[1].Status != "published" || len(chain[1].Evidence) != 1 {
		t.Fatalf("replacement = %#v", chain[1])
	}

	// The history stays: the student sees both versions.
	history, err := fixture.service.ListStudentAssessments(ctx, student, studentID)
	if err != nil || len(history) != 2 {
		t.Fatalf("student history = %d, %v", len(history), err)
	}

	// teacher_only notes never reach the student or the administrator.
	internalContent := content
	internalContent.Visibility = "teacher_only"
	internalContent.Summary = "Внутренняя заметка о методике работы."
	internalDraft, err := fixture.service.CreateAssessment(ctx, fixture.teacher, app.CreateAssessmentInput{
		StudentID: studentID, Content: internalContent, IdempotencyKey: "asmt-internal",
	})
	if err != nil {
		t.Fatalf("create internal note: %v", err)
	}
	if _, err := fixture.service.PublishAssessment(ctx, fixture.teacher, app.PublishAssessmentInput{
		AssessmentID: internalDraft.ID, ExpectedVersion: internalDraft.Version,
		IdempotencyKey: "asmt-internal-publish",
	}); err != nil {
		t.Fatalf("publish internal note: %v", err)
	}
	studentView, err := fixture.service.ListStudentAssessments(ctx, student, studentID)
	if err != nil || len(studentView) != 2 {
		t.Fatalf("student sees internal notes = %d, %v", len(studentView), err)
	}
	adminView, err := fixture.service.ListStudentAssessments(ctx, fixture.admin, studentID)
	for _, entry := range adminView {
		if entry.ID == internalDraft.ID {
			t.Fatalf("administrator sees a published teacher_only note: %#v", entry)
		}
	}
	if err != nil {
		t.Fatalf("admin list: %v", err)
	}
	if _, err := fixture.service.GetAssessment(ctx, student, internalDraft.ID); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("student opening internal note = %v, want NOT_FOUND", err)
	}

	// Withdrawal keeps the record with a mandatory reason.
	if _, err := fixture.service.WithdrawAssessment(ctx, fixture.teacher, internalDraft.ID, "", "asmt-withdraw-empty"); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("withdrawal without reason = %v, want INVALID_INPUT", err)
	}
	withdrawn, err := fixture.service.WithdrawAssessment(ctx, fixture.teacher, internalDraft.ID,
		"Записано не тому ученику", "asmt-withdraw")
	if err != nil || withdrawn.Status != "withdrawn" || withdrawn.WithdrawalReason == "" {
		t.Fatalf("withdraw = %#v, %v", withdrawn, err)
	}
	if _, err := fixture.service.PublishAssessment(ctx, fixture.teacher, app.PublishAssessmentInput{
		AssessmentID: internalDraft.ID, ExpectedVersion: withdrawn.Version,
		IdempotencyKey: "asmt-republish-withdrawn",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("republishing withdrawn = %v, want INVALID_STATE", err)
	}
}
