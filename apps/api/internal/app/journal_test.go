package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestJournalLifecycle pins DEC-006/007: the draft is teacher-private,
// publishing makes an immutable version the student sees, a correction
// requires an explicit note and appends the next version, and progress
// arrives as named-area evidence — never a score.
func TestJournalLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000000601", "ENR-601")
	const studentPassword = "Journal-student-pass-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000000601", Password: studentPassword,
		IdempotencyKey: "journal-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000000601", studentPassword)
	directory, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(directory) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := directory[len(directory)-1].StudentID

	startsAt := fixture.clock.Now().Add(2 * time.Hour)
	lesson, err := fixture.service.ScheduleLesson(ctx, fixture.owner, app.ScheduleLessonInput{
		Title: "Вокал · дыхание", StartsAt: startsAt, DurationMinutes: 45,
		TeacherAccountID: fixture.teacher.AccountID, StudentIDs: []string{studentID},
		IdempotencyKey: "journal-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}

	draftInput := app.JournalDraftInput{
		OccurrenceID: lesson.ID, StudentID: studentID,
		WhatWorked:   "Опора устойчива на средних нотах",
		CurrentFocus: "Свобода верхнего регистра",
		NextStep:     "Гаммы легато через переходные ноты",
	}
	if _, err := fixture.service.SaveJournalDraft(ctx, student, draftInput); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student saving draft = %v, want FORBIDDEN", err)
	}
	journal, err := fixture.service.SaveJournalDraft(ctx, fixture.teacher, draftInput)
	if err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if journal.Status != "draft" || journal.Draft == nil || journal.CurrentVersion != 0 {
		t.Fatalf("draft journal = %#v", journal)
	}

	if _, err := fixture.service.GetJournal(ctx, student, lesson.ID, studentID); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("student reading unpublished journal = %v, want NOT_FOUND", err)
	}

	if _, err := fixture.service.PublishJournal(ctx, fixture.teacher, app.PublishJournalInput{
		OccurrenceID: lesson.ID, StudentID: studentID, IdempotencyKey: "journal-early",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("publish before lesson start = %v, want INVALID_STATE", err)
	}

	fixture.clock.Advance(3 * time.Hour)
	published, err := fixture.service.PublishJournal(ctx, fixture.teacher, app.PublishJournalInput{
		OccurrenceID: lesson.ID, StudentID: studentID,
		Evidence: []core.EvidenceInput{
			{Area: "Дыхание", Note: "Держит фразу 8 тактов без потери опоры"},
		},
		IdempotencyKey: "journal-publish",
	})
	if err != nil {
		t.Fatalf("publish journal: %v", err)
	}
	if published.Status != "published" || published.CurrentVersion != 1 ||
		published.Draft != nil || len(published.Versions) != 1 {
		t.Fatalf("published journal = %#v", published)
	}

	studentView, err := fixture.service.GetJournal(ctx, student, lesson.ID, studentID)
	if err != nil {
		t.Fatalf("student reads journal: %v", err)
	}
	if studentView.Draft != nil || len(studentView.Versions) != 1 {
		t.Fatalf("student view leaks drafts: %#v", studentView)
	}

	evidence, err := fixture.service.ListProgressEvidence(ctx, student, studentID)
	if err != nil || len(evidence) != 1 {
		t.Fatalf("student evidence = %#v, %v", evidence, err)
	}
	if evidence[0].Area != "Дыхание" || evidence[0].SourceKind != "lesson_journal" {
		t.Fatalf("evidence shape = %#v", evidence[0])
	}

	correctionDraft := draftInput
	correctionDraft.NextStep = "Гаммы легато и стаккато через переходные ноты"
	if _, err := fixture.service.SaveJournalDraft(ctx, fixture.teacher, correctionDraft); err != nil {
		t.Fatalf("save correction draft: %v", err)
	}
	if _, err := fixture.service.PublishJournal(ctx, fixture.teacher, app.PublishJournalInput{
		OccurrenceID: lesson.ID, StudentID: studentID, IdempotencyKey: "journal-correct-bare",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("correction without note = %v, want INVALID_INPUT (DEC-007)", err)
	}
	corrected, err := fixture.service.PublishJournal(ctx, fixture.teacher, app.PublishJournalInput{
		OccurrenceID: lesson.ID, StudentID: studentID,
		CorrectionNote: "Уточнён следующий шаг после разбора записи",
		IdempotencyKey: "journal-correct",
	})
	if err != nil {
		t.Fatalf("publish correction: %v", err)
	}
	if corrected.CurrentVersion != 2 || len(corrected.Versions) != 2 {
		t.Fatalf("corrected journal = %#v", corrected)
	}
	if corrected.Versions[0].Version != 2 || corrected.Versions[0].CorrectionNote == "" {
		t.Fatalf("newest version = %#v", corrected.Versions[0])
	}
	if corrected.Versions[1].NextStep != "Гаммы легато через переходные ноты" {
		t.Fatal("original published version changed (DEC-007 violated)")
	}

	journals, err := fixture.service.ListStudentJournals(ctx, student, studentID)
	if err != nil || len(journals) != 1 {
		t.Fatalf("student journal list = %#v, %v", journals, err)
	}
}
