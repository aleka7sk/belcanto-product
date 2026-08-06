package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/notify"
)

// TestActivityDelivery pins the notification flow: the worker drains the
// outbox into per-recipient activity, delivery is idempotent, reads mark
// up to a moment, preferences default on per category, and a broken
// event retries with backoff until dead-letter — never blocking the
// rest of the queue.
func TestActivityDelivery(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000001101", "ENR-1101")
	const studentPassword = "Activity-student-pass-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000001101", Password: studentPassword,
		IdempotencyKey: "act-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000001101", studentPassword)
	directory, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(directory) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := directory[len(directory)-1].StudentID

	lesson, err := fixture.service.ScheduleLesson(ctx, fixture.owner, app.ScheduleLessonInput{
		Title: "Вокал · активность", StartsAt: fixture.clock.Now().Add(time.Hour),
		DurationMinutes: 45, TeacherAccountID: fixture.teacher.AccountID,
		StudentIDs: []string{studentID}, IdempotencyKey: "act-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}
	if _, err := fixture.service.SaveJournalDraft(ctx, fixture.teacher, app.JournalDraftInput{
		OccurrenceID: lesson.ID, StudentID: studentID,
		WhatWorked: "Опора", CurrentFocus: "Верх", NextStep: "Легато",
	}); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	fixture.clock.Advance(2 * time.Hour)
	if _, err := fixture.service.PublishJournal(ctx, fixture.teacher, app.PublishJournalInput{
		OccurrenceID: lesson.ID, StudentID: studentID, IdempotencyKey: "act-publish",
	}); err != nil {
		t.Fatalf("publish journal: %v", err)
	}
	if _, err := fixture.service.MarkAttendance(ctx, fixture.teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendanceAbsent,
		Note: "Болеет", IdempotencyKey: "act-absent",
	}); err != nil {
		t.Fatalf("mark absence: %v", err)
	}

	worker := notify.NewWorker(fixture.store, notify.Options{
		Clock: func() time.Time { return fixture.clock.Now() },
	})
	processed, failed, err := worker.DrainOnce(ctx)
	if err != nil || failed != 0 || processed == 0 {
		t.Fatalf("drain = %d processed, %d failed, %v", processed, failed, err)
	}
	// Idempotent: a second drain has nothing pending.
	processed, _, err = worker.DrainOnce(ctx)
	if err != nil || processed != 0 {
		t.Fatalf("second drain = %d processed, %v", processed, err)
	}

	feed, err := fixture.service.ActivityFeed(ctx, student)
	if err != nil {
		t.Fatalf("student feed: %v", err)
	}
	if feed.UnreadCount == 0 || len(feed.Entries) == 0 {
		t.Fatalf("student feed = %#v", feed)
	}
	sawJournal := false
	for _, entry := range feed.Entries {
		if entry.Kind == "JournalPublished" && entry.Category == "learning" {
			sawJournal = true
		}
		if entry.Kind == "AttendanceAbsenceRecorded" {
			t.Fatalf("student received an admin-only event: %#v", entry)
		}
	}
	if !sawJournal {
		t.Fatalf("journal publication missing from the student feed: %#v", feed.Entries)
	}

	marked, err := fixture.service.MarkActivityRead(ctx, student, fixture.clock.Now())
	if err != nil || marked == 0 {
		t.Fatalf("mark read = %d, %v", marked, err)
	}
	feed, err = fixture.service.ActivityFeed(ctx, student)
	if err != nil || feed.UnreadCount != 0 {
		t.Fatalf("feed after read = %#v, %v", feed, err)
	}

	preferences, err := fixture.service.NotificationPreferences(ctx, student)
	if err != nil || len(preferences) != len(core.NotificationCategories) {
		t.Fatalf("preferences = %#v, %v", preferences, err)
	}
	for _, preference := range preferences {
		if !preference.PushEnabled {
			t.Fatalf("default preference must be enabled: %#v", preference)
		}
	}
	updated, err := fixture.service.UpdateNotificationPreference(ctx, student, "community", false)
	if err != nil {
		t.Fatalf("update preference: %v", err)
	}
	for _, preference := range updated {
		if preference.Category == "community" && preference.PushEnabled {
			t.Fatalf("community preference did not persist: %#v", updated)
		}
	}
	if _, err := fixture.service.UpdateNotificationPreference(ctx, student, "marketing", true); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("unknown category = %v, want INVALID_INPUT", err)
	}

	// A poisoned event retries with backoff and finally dead-letters
	// without blocking anything else.
	fixture.store.AppendBrokenOutboxEvent(fixture.owner.TenantID, "HomeworkSubmitted", fixture.clock.Now())
	for attempt := 0; attempt < 10; attempt++ {
		if _, _, err := worker.DrainOnce(ctx); err != nil {
			t.Fatalf("drain with poison: %v", err)
		}
		fixture.clock.Advance(2 * time.Hour)
	}
	pending, err := fixture.store.PendingOutboxEvents(ctx, 10, fixture.clock.Now())
	if err != nil || len(pending) != 0 {
		t.Fatalf("poisoned event still pending after max attempts: %#v, %v", pending, err)
	}
}
