package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestRescheduleRequestLifecycle walks flows J/K/L: a participant asks
// to move a lesson, the admin approves and the occurrence moves in
// place; a cancellation request applies the student-initiated status;
// DEC-102 stays open — nothing here computes any consequence.
func TestRescheduleRequestLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000000501", "ENR-501")
	const studentPassword = "Resched-student-pass-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000000501", Password: studentPassword,
		IdempotencyKey: "resched-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000000501", studentPassword)
	studentResult, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(studentResult) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := studentResult[len(studentResult)-1].StudentID

	startsAt := fixture.clock.Now().Add(72 * time.Hour)
	lesson, err := fixture.service.ScheduleLesson(ctx, fixture.owner, app.ScheduleLessonInput{
		Title: "Вокал · разбор произведения", StartsAt: startsAt, DurationMinutes: 45,
		TeacherAccountID: fixture.teacher.AccountID, StudentIDs: []string{studentID},
		IdempotencyKey: "resched-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}

	proposed := startsAt.Add(24 * time.Hour)
	request, err := fixture.service.CreateRescheduleRequest(ctx, student, app.CreateRescheduleRequestInput{
		OccurrenceID: lesson.ID, Kind: "reschedule",
		ProposedStartsAt: &proposed, Reason: "Совпадает со школьной олимпиадой",
		IdempotencyKey: "resched-req-1",
	})
	if err != nil {
		t.Fatalf("create reschedule request: %v", err)
	}
	if request.Status != "pending" || request.ProposedStartsAt == nil {
		t.Fatalf("created request = %#v", request)
	}

	if _, err := fixture.service.CreateRescheduleRequest(ctx, student, app.CreateRescheduleRequestInput{
		OccurrenceID: lesson.ID, Kind: "cancellation",
		Reason: "Вторая попытка", IdempotencyKey: "resched-req-dup",
	}); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("second open request = %v, want CONFLICT", err)
	}

	if _, err := fixture.service.DecideRescheduleRequest(ctx, student, request.ID, app.DecideRescheduleRequestInput{
		Approve: true, ExpectedVersion: request.Version,
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student deciding = %v, want FORBIDDEN", err)
	}

	decided, err := fixture.service.DecideRescheduleRequest(ctx, fixture.owner, request.ID, app.DecideRescheduleRequestInput{
		Approve: true, DecisionNote: "Перенос согласован с преподавателем",
		ExpectedVersion: request.Version,
	})
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}
	if decided.Status != "approved" || decided.DecidedAt == nil {
		t.Fatalf("approved request = %#v", decided)
	}

	moved, err := fixture.service.GetLesson(ctx, fixture.owner, lesson.ID)
	if err != nil {
		t.Fatalf("read moved lesson: %v", err)
	}
	if !moved.StartsAt.Equal(proposed.UTC()) || moved.Version != lesson.Version+1 {
		t.Fatalf("moved lesson = startsAt %s version %d, want %s / %d",
			moved.StartsAt, moved.Version, proposed.UTC(), lesson.Version+1)
	}

	cancelRequest, err := fixture.service.CreateRescheduleRequest(ctx, student, app.CreateRescheduleRequestInput{
		OccurrenceID: lesson.ID, Kind: "cancellation",
		Reason: "Болезнь", IdempotencyKey: "resched-cancel-1",
	})
	if err != nil {
		t.Fatalf("create cancellation request: %v", err)
	}
	approvedCancel, err := fixture.service.DecideRescheduleRequest(ctx, fixture.owner, cancelRequest.ID, app.DecideRescheduleRequestInput{
		Approve: true, ExpectedVersion: cancelRequest.Version,
	})
	if err != nil {
		t.Fatalf("approve cancellation: %v", err)
	}
	if approvedCancel.Status != "approved" {
		t.Fatalf("approved cancellation = %#v", approvedCancel)
	}
	cancelled, err := fixture.service.GetLesson(ctx, fixture.owner, lesson.ID)
	if err != nil {
		t.Fatalf("read cancelled lesson: %v", err)
	}
	if cancelled.Status != core.LessonCancelledStudent {
		t.Fatalf("cancelled lesson status = %s, want cancelled_student (DEC-102: no consequence)", cancelled.Status)
	}

	if _, err := fixture.service.CreateRescheduleRequest(ctx, student, app.CreateRescheduleRequestInput{
		OccurrenceID: lesson.ID, Kind: "cancellation",
		Reason: "Уже отменён", IdempotencyKey: "resched-after-cancel",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("request on cancelled lesson = %v, want INVALID_STATE", err)
	}
}
