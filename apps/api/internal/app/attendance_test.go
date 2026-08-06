package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestAttendanceLifecycle pins the attendance rules of TCH-JOURNAL-01/02
// on the approved Lesson model: a late mark carries minutes, an absence
// carries a note, corrections carry a reason, the Student sees their own
// mark without the teacher note, and an empty group seat has no row.
func TestAttendanceLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000000801", "ENR-801")
	const studentPassword = "Attendance-student-pass-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000000801", Password: studentPassword,
		IdempotencyKey: "att-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000000801", studentPassword)
	directory, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(directory) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := directory[len(directory)-1].StudentID

	lesson, err := fixture.service.ScheduleLesson(ctx, fixture.owner, app.ScheduleLessonInput{
		Title: "Вокал · посещаемость", StartsAt: fixture.clock.Now().Add(time.Hour),
		DurationMinutes: 45, TeacherAccountID: fixture.teacher.AccountID,
		StudentIDs: []string{studentID}, IdempotencyKey: "att-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}

	if _, err := fixture.service.MarkAttendance(ctx, student, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendancePresent,
		IdempotencyKey: "att-student-marks",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student marking attendance = %v, want FORBIDDEN", err)
	}
	if _, err := fixture.service.MarkAttendance(ctx, fixture.teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendanceLate,
		IdempotencyKey: "att-late-no-minutes",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("late without minutes = %v, want INVALID_INPUT", err)
	}
	if _, err := fixture.service.MarkAttendance(ctx, fixture.teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendanceAbsent,
		IdempotencyKey: "att-absent-no-note",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("absence without note = %v, want INVALID_INPUT", err)
	}

	marked, err := fixture.service.MarkAttendance(ctx, fixture.teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendanceLate,
		LateMinutes: 7, Note: "Предупредила заранее", IdempotencyKey: "att-late",
	})
	if err != nil || len(marked) != 1 {
		t.Fatalf("mark late = %#v, %v", marked, err)
	}
	if marked[0].Status != core.AttendanceLate || marked[0].LateMinutes != 7 {
		t.Fatalf("late record = %#v", marked[0])
	}

	if _, err := fixture.service.MarkAttendance(ctx, fixture.teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendancePresent,
		IdempotencyKey: "att-change-bare",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("correction without reason = %v, want INVALID_INPUT", err)
	}
	corrected, err := fixture.service.MarkAttendance(ctx, fixture.teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendancePresent,
		ChangeReason: "Опоздание не подтвердилось", IdempotencyKey: "att-change",
	})
	if err != nil || corrected[0].Status != core.AttendancePresent || corrected[0].LateMinutes != 0 {
		t.Fatalf("corrected record = %#v, %v", corrected, err)
	}

	teacherView, err := fixture.service.ListLessonAttendance(ctx, fixture.teacher, lesson.ID)
	if err != nil || len(teacherView) != 1 {
		t.Fatalf("teacher attendance view = %#v, %v", teacherView, err)
	}
	ownerView, err := fixture.service.ListLessonAttendance(ctx, fixture.owner, lesson.ID)
	if err != nil || len(ownerView) != 1 {
		t.Fatalf("owner attendance view = %#v, %v", ownerView, err)
	}

	absent, err := fixture.service.MarkAttendance(ctx, fixture.teacher, app.MarkAttendanceInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Status: core.AttendanceAbsent,
		Note: "Болеет, родители сообщили", ChangeReason: "Не пришла после переноса",
		IdempotencyKey: "att-absent",
	})
	if err != nil || absent[0].Status != core.AttendanceAbsent {
		t.Fatalf("absence record = %#v, %v", absent, err)
	}
	studentView, err := fixture.service.ListLessonAttendance(ctx, student, lesson.ID)
	if err != nil || len(studentView) != 1 {
		t.Fatalf("student attendance view = %#v, %v", studentView, err)
	}
	if studentView[0].Note != "" {
		t.Fatalf("teacher note leaked to the student: %#v", studentView[0])
	}
}
