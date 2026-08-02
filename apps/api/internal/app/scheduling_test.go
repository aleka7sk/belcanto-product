package app_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

func TestInternalSchedulingIsIndependentFromOnboardingDelegationAndIncludesPendingStudents(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	view, err := fixture.service.BootstrapView(ctx, fixture.admin)
	if err != nil {
		t.Fatalf("ordinary Administrator bootstrap: %v", err)
	}
	wantPermissions := core.LessonPermissionSetForRoles([]core.Role{core.RoleAdministrator})
	if !reflect.DeepEqual(view.Permissions, wantPermissions) || len(view.AccessProfiles) != 0 {
		t.Fatalf("ordinary Administrator scheduling access = %#v", view)
	}
	teachers, err := fixture.service.ListStaff(ctx, fixture.admin, core.RoleTeacher)
	if err != nil || len(teachers) != 1 || teachers[0].AccountID != fixture.teacher.AccountID {
		t.Fatalf("ordinary Administrator Teacher directory = %#v, %v", teachers, err)
	}

	student, err := fixture.service.CreateStudent(ctx, fixture.owner, studentInput(
		"schedule-pending-student", "+77000001101", "SCHEDULE-1101", fixture.teacher.AccountID,
	))
	if err != nil {
		t.Fatalf("Owner creates pending Student: %v", err)
	}
	directory, err := fixture.service.ListStudents(ctx, fixture.admin, app.ListStudentsInput{})
	if err != nil || len(directory) != 1 || directory[0].StudentID != student.StudentID {
		t.Fatalf("pending Student scheduling directory = %#v, %v", directory, err)
	}
	lesson, err := fixture.service.ScheduleLesson(ctx, fixture.admin, app.ScheduleLessonInput{
		Title: "Первый урок", StartsAt: fixture.clock.Now().Add(2 * time.Hour),
		DurationMinutes: 60, Location: "Класс 1", TeacherAccountID: fixture.teacher.AccountID,
		StudentIDs: []string{student.StudentID}, IdempotencyKey: "admin-schedules-pending",
	})
	if err != nil {
		t.Fatalf("ordinary Administrator schedules pending Student: %v", err)
	}
	if lesson.Teacher.AccountID != fixture.teacher.AccountID || len(lesson.Students) != 1 || lesson.Students[0].StudentID != student.StudentID {
		t.Fatalf("scheduled pending Student Lesson = %#v", lesson)
	}
	replay, err := fixture.service.ScheduleLesson(ctx, fixture.admin, app.ScheduleLessonInput{
		Title: "Первый урок", StartsAt: fixture.clock.Now().Add(2 * time.Hour),
		DurationMinutes: 60, Location: "Класс 1", TeacherAccountID: fixture.teacher.AccountID,
		StudentIDs: []string{student.StudentID}, IdempotencyKey: "admin-schedules-pending",
	})
	if err != nil || replay.ID != lesson.ID {
		t.Fatalf("schedule idempotency replay = %#v, %v", replay, err)
	}
}

func TestAssignmentMutationsUseMonotonicLogicalTimeAndCurrentProjection(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	secondTeacher := seedTeacher(t, fixture, "acct_logical_teacher_2", "+77000001104")
	studentResult, err := fixture.service.CreateStudent(ctx, fixture.owner, studentInput(
		"logical-clock-student", "+77000001105", "LOGICAL-1105", fixture.teacher.AccountID,
	))
	if err != nil {
		t.Fatalf("create logical-clock Student: %v", err)
	}

	fixture.clock.Advance(2 * time.Minute)
	firstMinute, err := fixture.service.PublishFirstMinute(ctx, fixture.teacher, app.PublishFirstMinuteInput{
		StudentID: studentResult.StudentID, WhatWorked: "Monotonic publication",
		CurrentFocus: "Serialized time", NextStep: "Transfer safely",
		ExpectedVersion: 0, IdempotencyKey: "logical-clock-first-minute",
	})
	if err != nil {
		t.Fatalf("publish logical-clock First Minute: %v", err)
	}

	// Model a request issued before the publication but handled afterward. The
	// in-memory store must preserve the same serialized ordering as PostgreSQL.
	fixture.clock.Advance(-time.Minute)
	reassigned, err := fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students: []core.PrimaryTeacherReassignmentTarget{{
			StudentID: studentResult.StudentID, ExpectedAssignmentVersion: 0,
		}},
		NewTeacherAccountID: secondTeacher.AccountID,
		EffectiveMode:       core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey:      "logical-clock-reassignment",
	})
	if err != nil || reassigned.ReassignedCount != 1 {
		t.Fatalf("logical-clock reassignment = %#v, %v", reassigned, err)
	}
	cutover := reassigned.Assignments[0].EffectiveFrom
	if !firstMinute.PublishedAt.Before(cutover) {
		t.Fatalf("logical First Minute/cutover order = %s / %s", firstMinute.PublishedAt, cutover)
	}

	current, err := fixture.service.ListStudents(ctx, fixture.admin, app.ListStudentsInput{})
	if err != nil || directoryStudent(t, current, studentResult.StudentID).PrimaryTeacher.AccountID != secondTeacher.AccountID {
		t.Fatalf("default current projection after clock rollback = %#v, %v", current, err)
	}
	historical, err := fixture.service.ListStudents(ctx, fixture.admin, app.ListStudentsInput{AsOf: firstMinute.PublishedAt})
	if err != nil || directoryStudent(t, historical, studentResult.StudentID).PrimaryTeacher.AccountID != fixture.teacher.AccountID {
		t.Fatalf("explicit pre-cutover projection = %#v, %v", historical, err)
	}
	queue, err := fixture.service.ListStudentOnboarding(ctx, fixture.owner)
	if err != nil || len(queue) != 1 || queue[0].TeacherAccountID != secondTeacher.AccountID {
		t.Fatalf("current onboarding projection after clock rollback = %#v, %v", queue, err)
	}
}

func TestTeacherContinuityCommandsAreAtomicVersionedTemporalAndAudited(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	secondTeacher := seedTeacher(t, fixture, "acct_schedule_teacher_2", "+77000001102")
	thirdTeacher := seedTeacher(t, fixture, "acct_schedule_teacher_3", "+77000001103")
	firstStudent, err := fixture.service.CreateStudent(ctx, fixture.owner, studentInput(
		"continuity-student-1", "+77000001111", "CONTINUITY-1111", fixture.teacher.AccountID,
	))
	if err != nil {
		t.Fatalf("create first continuity Student: %v", err)
	}
	secondStudent, err := fixture.service.CreateStudent(ctx, fixture.owner, studentInput(
		"continuity-student-2", "+77000001112", "CONTINUITY-1112", fixture.teacher.AccountID,
	))
	if err != nil {
		t.Fatalf("create second continuity Student: %v", err)
	}
	thirdStudent, err := fixture.service.CreateStudent(ctx, fixture.owner, studentInput(
		"continuity-student-3", "+77000001113", "CONTINUITY-1113", fixture.teacher.AccountID,
	))
	if err != nil {
		t.Fatalf("create third continuity Student: %v", err)
	}

	pastCandidate := scheduleForTest(t, fixture, fixture.admin, "continuity-past", fixture.clock.Now().Add(time.Hour), fixture.teacher.AccountID, firstStudent.StudentID)
	atomicFirst := scheduleForTest(t, fixture, fixture.admin, "continuity-atomic-1", fixture.clock.Now().Add(4*time.Hour), fixture.teacher.AccountID, firstStudent.StudentID)
	atomicSecond := scheduleForTest(t, fixture, fixture.admin, "continuity-atomic-2", fixture.clock.Now().Add(5*time.Hour), fixture.teacher.AccountID, secondStudent.StudentID)
	unchangedByPermanent := scheduleForTest(t, fixture, fixture.admin, "continuity-permanent", fixture.clock.Now().Add(6*time.Hour), fixture.teacher.AccountID, firstStudent.StudentID)

	_, err = fixture.service.ReplaceLessonTeachers(ctx, fixture.teacher, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: atomicFirst.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "teacher-replace-denied",
	})
	assertCode(t, err, core.CodeForbidden)
	_, err = fixture.service.ReassignPrimaryTeachers(ctx, fixture.teacher, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: fixture.clock.Now().Add(time.Hour), IdempotencyKey: "teacher-reassign-denied",
	})
	assertCode(t, err, core.CodeForbidden)

	_, err = fixture.service.ReplaceLessonTeachers(ctx, fixture.admin, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{
			{LessonID: atomicFirst.ID, ExpectedVersion: 0, ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID},
			{LessonID: atomicSecond.ID, ExpectedVersion: 99, ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID},
		},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "replace-atomic-stale",
	})
	assertCode(t, err, core.CodeConflict)
	stillFirst, err := fixture.service.GetLesson(ctx, fixture.owner, atomicFirst.ID)
	if err != nil || stillFirst.Teacher.AccountID != fixture.teacher.AccountID || stillFirst.Version != 0 {
		t.Fatalf("failed replacement batch mutated valid target = %#v, %v", stillFirst, err)
	}

	replaced, err := fixture.service.ReplaceLessonTeachers(ctx, fixture.admin, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: atomicFirst.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "replace-selected-future",
	})
	if err != nil || replaced.UpdatedCount != 1 || replaced.Lessons[0].Version != 1 || replaced.Lessons[0].Teacher.AccountID != secondTeacher.AccountID {
		t.Fatalf("selected future replacement = %#v, %v", replaced, err)
	}
	replacedReplay, err := fixture.service.ReplaceLessonTeachers(ctx, fixture.admin, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: atomicFirst.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "replace-selected-future",
	})
	if err != nil || !reflect.DeepEqual(replacedReplay, replaced) {
		t.Fatalf("replacement replay = %#v, %v", replacedReplay, err)
	}

	fixture.clock.Advance(2 * time.Hour)
	_, err = fixture.service.ReplaceLessonTeachers(ctx, fixture.admin, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: pastCandidate.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "replace-past-denied",
	})
	assertCode(t, err, core.CodeInvalidState)

	reassigned, err := fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey: "reassign-immediate",
	})
	if err != nil || reassigned.ReassignedCount != 1 || !reassigned.Assignments[0].EffectiveFrom.Equal(fixture.clock.Now()) || reassigned.Assignments[0].Version != 1 {
		t.Fatalf("bounded immediate reassignment = %#v, %v", reassigned, err)
	}
	directory, err := fixture.service.ListStudents(ctx, fixture.admin, app.ListStudentsInput{})
	if err != nil {
		t.Fatalf("directory after reassignment: %v", err)
	}
	firstDirectory := directoryStudent(t, directory, firstStudent.StudentID)
	if firstDirectory.PrimaryTeacher.AccountID != secondTeacher.AccountID || firstDirectory.PrimaryTeacherAssignmentVersion != 1 {
		t.Fatalf("directory after reassignment = %#v", firstDirectory)
	}
	permanentLesson, err := fixture.service.GetLesson(ctx, fixture.owner, unchangedByPermanent.ID)
	if err != nil || permanentLesson.Teacher.AccountID != fixture.teacher.AccountID || permanentLesson.Version != 0 {
		t.Fatalf("permanent reassignment mutated Lesson = %#v, %v", permanentLesson, err)
	}

	fixture.clock.Advance(2 * time.Minute)
	reassignmentReplay, err := fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey: "reassign-immediate",
	})
	if err != nil || !reflect.DeepEqual(reassignmentReplay, reassigned) {
		t.Fatalf("reassignment replay after tolerance window = %#v, %v", reassignmentReplay, err)
	}
	_, err = fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: fixture.clock.Now().Add(time.Hour), IdempotencyKey: "reassign-stale-version",
	})
	assertCode(t, err, core.CodeConflict)
	_, err = fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 1}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: fixture.clock.Now().Add(-2*time.Minute - time.Second), IdempotencyKey: "reassign-retroactive",
	})
	assertCode(t, err, core.CodeInvalidInput)

	_, err = fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students: []core.PrimaryTeacherReassignmentTarget{
			{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 1},
			{StudentID: secondStudent.StudentID, ExpectedAssignmentVersion: 99},
		},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: fixture.clock.Now().Add(time.Hour), IdempotencyKey: "reassign-atomic-stale",
	})
	assertCode(t, err, core.CodeConflict)
	directory, err = fixture.service.ListStudents(ctx, fixture.admin, app.ListStudentsInput{})
	if err != nil {
		t.Fatalf("directory after failed reassignment batch: %v", err)
	}
	if got := directoryStudent(t, directory, firstStudent.StudentID); got.PrimaryTeacher.AccountID != secondTeacher.AccountID || got.PrimaryTeacherAssignmentVersion != 1 {
		t.Fatalf("failed reassignment batch mutated valid Student = %#v", got)
	}

	futureEffective := fixture.clock.Now().Add(8 * time.Hour)
	futureReassignment, err := fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: secondStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: futureEffective, IdempotencyKey: "reassign-future-boundary",
	})
	if err != nil || futureReassignment.Assignments[0].Version != 1 {
		t.Fatalf("future reassignment = %#v, %v", futureReassignment, err)
	}
	_, err = fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: secondStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: futureEffective.Add(-time.Hour), IdempotencyKey: "reassign-stale-before-known-future",
	})
	assertCode(t, err, core.CodeConflict)
	projected, err := fixture.service.ListStudents(ctx, secondTeacher, app.ListStudentsInput{AsOf: futureEffective.Add(time.Minute)})
	if err != nil || directoryStudent(t, projected, secondStudent.StudentID).PrimaryTeacher.AccountID != secondTeacher.AccountID {
		t.Fatalf("incoming Teacher future Student projection = %#v, %v", projected, err)
	}
	projected, err = fixture.service.ListStudents(ctx, fixture.teacher, app.ListStudentsInput{AsOf: futureEffective.Add(time.Minute)})
	if err != nil {
		t.Fatalf("outgoing Teacher future projection: %v", err)
	}
	for _, item := range projected {
		if item.StudentID == secondStudent.StudentID {
			t.Fatalf("outgoing Teacher retained Student after future boundary: %#v", item)
		}
	}
	beforeBoundary := futureEffective.Add(-90 * time.Minute)
	afterBoundary := futureEffective.Add(time.Minute)
	if _, err := fixture.service.ScheduleLesson(ctx, fixture.teacher, lessonInput("old-before-boundary", beforeBoundary, fixture.teacher.AccountID, secondStudent.StudentID)); err != nil {
		t.Fatalf("old Teacher before boundary: %v", err)
	}
	if _, err := fixture.service.ScheduleLesson(ctx, fixture.teacher, lessonInput("old-after-boundary", afterBoundary, fixture.teacher.AccountID, secondStudent.StudentID)); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("old Teacher after boundary = %v", err)
	}
	if _, err := fixture.service.ScheduleLesson(ctx, secondTeacher, lessonInput("new-before-boundary", beforeBoundary, secondTeacher.AccountID, secondStudent.StudentID)); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("new Teacher before boundary = %v", err)
	}
	if _, err := fixture.service.ScheduleLesson(ctx, secondTeacher, lessonInput("new-after-boundary", afterBoundary, secondTeacher.AccountID, secondStudent.StudentID)); err != nil {
		t.Fatalf("new Teacher after boundary: %v", err)
	}

	if err := fixture.store.SetAccountStatusForTest(fixture.owner.TenantID, fixture.teacher.AccountID, "suspended"); err != nil {
		t.Fatalf("disable departed Teacher: %v", err)
	}
	directory, err = fixture.service.ListStudents(ctx, fixture.admin, app.ListStudentsInput{})
	if err != nil {
		t.Fatalf("manager directory with departed Teacher: %v", err)
	}
	departed := directoryStudent(t, directory, thirdStudent.StudentID)
	if departed.PrimaryTeacher.AccountID != fixture.teacher.AccountID || departed.PrimaryTeacher.Status != core.AssignedTeacherInactive {
		t.Fatalf("departed Teacher continuity projection = %#v", departed)
	}
	if _, err := fixture.service.ScheduleLesson(ctx, fixture.teacher, lessonInput("departed-teacher-denied", fixture.clock.Now().Add(20*time.Hour), fixture.teacher.AccountID, thirdStudent.StudentID)); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("departed Teacher schedules Lesson = %v", err)
	}
	_, err = fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: thirdStudent.StudentID, ExpectedAssignmentVersion: departed.PrimaryTeacherAssignmentVersion}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey: "reassign-from-departed-teacher",
	})
	if err != nil {
		t.Fatalf("transfer Student from departed Teacher: %v", err)
	}
	directory, err = fixture.service.ListStudents(ctx, fixture.admin, app.ListStudentsInput{})
	if err != nil || directoryStudent(t, directory, thirdStudent.StudentID).PrimaryTeacher.Status != core.AssignedTeacherActive {
		t.Fatalf("directory after departed Teacher transfer = %#v, %v", directory, err)
	}

	if !hasAuditAction(fixture, "LessonTeacherReplaced") || !hasAuditAction(fixture, "StudentPrimaryTeacherReassigned") {
		t.Fatalf("continuity audit records missing: %#v", fixture.store.AuditRecords())
	}
	if !hasOutboxEvent(fixture, "LessonTeacherReplaced") || !hasOutboxEvent(fixture, "StudentPrimaryTeacherReassigned") {
		t.Fatalf("continuity outbox events missing: %#v", fixture.store.OutboxRecords())
	}
}

func TestSchedulingRejectsOverlapsAndEnforcesRequestBounds(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	secondTeacher := seedTeacher(t, fixture, "acct_overlap_teacher_2", "+77000001131")
	firstStudent, err := fixture.service.CreateStudent(ctx, fixture.owner, studentInput(
		"overlap-student-1", "+77000001132", "OVERLAP-1132", fixture.teacher.AccountID,
	))
	if err != nil {
		t.Fatalf("create first overlap Student: %v", err)
	}
	secondStudent, err := fixture.service.CreateStudent(ctx, fixture.owner, studentInput(
		"overlap-student-2", "+77000001133", "OVERLAP-1133", fixture.teacher.AccountID,
	))
	if err != nil {
		t.Fatalf("create second overlap Student: %v", err)
	}
	start := fixture.clock.Now().Add(2 * time.Hour)
	first := scheduleForTest(t, fixture, fixture.admin, "overlap-first", start, fixture.teacher.AccountID, firstStudent.StudentID)
	conflictingTeacherLesson := scheduleForTest(t, fixture, fixture.admin, "replacement-conflict", start.Add(-30*time.Minute), secondTeacher.AccountID, secondStudent.StudentID)
	if _, err := fixture.service.ScheduleLesson(ctx, fixture.admin, lessonInput("overlap-teacher", start.Add(30*time.Minute), fixture.teacher.AccountID, secondStudent.StudentID)); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("overlapping Teacher Lesson = %v", err)
	}
	if _, err := fixture.service.ScheduleLesson(ctx, fixture.admin, lessonInput("overlap-student", start.Add(15*time.Minute), secondTeacher.AccountID, firstStudent.StudentID)); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("overlapping Student Lesson = %v", err)
	}
	if _, err := fixture.service.ScheduleLesson(ctx, fixture.admin, lessonInput("adjacent-allowed", start.Add(time.Hour), secondTeacher.AccountID, secondStudent.StudentID)); err != nil {
		t.Fatalf("adjacent Lesson rejected: %v", err)
	}
	if _, err := fixture.service.ReplaceLessonTeachers(ctx, fixture.admin, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: first.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID}}, NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey: "replacement-overlap-denied",
	}); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("replacement overlap = %v (fixture %#v)", err, conflictingTeacherLesson)
	}

	studentIDs := make([]string, 101)
	lessonTargets := make([]core.ReplaceLessonTeacherTarget, 101)
	studentTargets := make([]core.PrimaryTeacherReassignmentTarget, 101)
	for index := range studentIDs {
		studentIDs[index] = fmt.Sprintf("student_bound_%03d", index)
		lessonTargets[index] = core.ReplaceLessonTeacherTarget{
			LessonID: fmt.Sprintf("lesson_bound_%03d", index), ExpectedPreviousTeacherAccountID: fixture.teacher.AccountID,
		}
		studentTargets[index] = core.PrimaryTeacherReassignmentTarget{StudentID: studentIDs[index]}
	}
	if _, err := fixture.service.ScheduleLesson(ctx, fixture.admin, app.ScheduleLessonInput{
		Title: "Too many Students", StartsAt: start.Add(10 * time.Hour), DurationMinutes: 60,
		TeacherAccountID: fixture.teacher.AccountID, StudentIDs: studentIDs, IdempotencyKey: "too-many-students",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("ScheduleLesson maxItems = %v", err)
	}
	if _, err := fixture.service.ReplaceLessonTeachers(ctx, fixture.admin, app.ReplaceLessonTeachersInput{
		Lessons: lessonTargets, NewTeacherAccountID: secondTeacher.AccountID, IdempotencyKey: "too-many-lessons",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("ReplaceLessonTeachers maxItems = %v", err)
	}
	if _, err := fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students: studentTargets, NewTeacherAccountID: secondTeacher.AccountID,
		EffectiveMode: core.PrimaryTeacherEffectiveImmediate, IdempotencyKey: "too-many-reassignments",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("ReassignPrimaryTeachers maxItems = %v", err)
	}
	if _, err := fixture.service.ListLessons(ctx, fixture.admin, app.ListLessonsInput{
		From: fixture.clock.Now(), To: fixture.clock.Now().Add(366*24*time.Hour + time.Second),
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("ListLessons range bound = %v", err)
	}
	if _, err := fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students: []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID}}, NewTeacherAccountID: secondTeacher.AccountID,
		EffectiveMode: core.PrimaryTeacherEffectiveImmediate, EffectiveFrom: start, IdempotencyKey: "invalid-immediate-time",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("immediate reassignment accepted client timestamp = %v", err)
	}
	if _, err := fixture.service.ReassignPrimaryTeachers(ctx, fixture.admin, app.ReassignPrimaryTeachersInput{
		Students: []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID}}, NewTeacherAccountID: secondTeacher.AccountID,
		EffectiveMode: core.PrimaryTeacherEffectiveScheduled, IdempotencyKey: "missing-scheduled-time",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("scheduled reassignment without time = %v", err)
	}
}

func TestLessonReadsAndMutationsRemainRoleScoped(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	for _, seed := range []struct {
		accountID string
		studentID string
		fullName  string
		phone     string
	}{
		{"acct_scope_student_1", "student_scope_1", "Scope Student One", "+77000001121"},
		{"acct_scope_student_2", "student_scope_2", "Scope Student Two", "+77000001122"},
	} {
		if err := fixture.store.SeedActiveStudentForTest(
			fixture.owner.TenantID, seed.accountID, seed.studentID, "person_"+seed.studentID,
			seed.fullName, seed.phone, fixture.teacher.AccountID, fixture.clock.Now().Add(-time.Hour),
		); err != nil {
			t.Fatalf("seed role-scope Student: %v", err)
		}
	}
	firstLesson := scheduleForTest(
		t,
		fixture,
		fixture.owner,
		"scope-first",
		fixture.clock.Now().Add(time.Hour),
		fixture.teacher.AccountID,
		"student_scope_1",
		"student_scope_2",
	)
	secondLesson := scheduleForTest(t, fixture, fixture.owner, "scope-second", fixture.clock.Now().Add(2*time.Hour), fixture.teacher.AccountID, "student_scope_2")
	student := core.Principal{AccountID: "acct_scope_student_1", TenantID: fixture.owner.TenantID, Roles: []core.Role{core.RoleStudent}}
	lessons, err := fixture.service.ListLessons(ctx, student, app.ListLessonsInput{From: fixture.clock.Now(), To: fixture.clock.Now().Add(3 * time.Hour)})
	if err != nil || len(lessons) != 1 || lessons[0].ID != firstLesson.ID || len(lessons[0].Students) != 1 || lessons[0].Students[0].StudentID != "student_scope_1" {
		t.Fatalf("Student role-scoped Lesson list = %#v, %v", lessons, err)
	}
	studentLesson, err := fixture.service.GetLesson(ctx, student, firstLesson.ID)
	if err != nil || len(studentLesson.Students) != 1 || studentLesson.Students[0].StudentID != "student_scope_1" {
		t.Fatalf("Student Lesson detail leaked peer participants = %#v, %v", studentLesson, err)
	}
	managerLesson, err := fixture.service.GetLesson(ctx, fixture.owner, firstLesson.ID)
	if err != nil || len(managerLesson.Students) != 2 {
		t.Fatalf("manager group Lesson projection = %#v, %v", managerLesson, err)
	}
	if _, err := fixture.service.GetLesson(ctx, student, secondLesson.ID); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("Student reads another Student Lesson = %v", err)
	}
	if _, err := fixture.service.ScheduleLesson(ctx, student, lessonInput("student-mutation", fixture.clock.Now().Add(4*time.Hour), fixture.teacher.AccountID, "student_scope_1")); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("Student schedules Lesson = %v", err)
	}
	if _, err := fixture.service.ListStudents(ctx, student, app.ListStudentsInput{}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("Student reads scheduling directory = %v", err)
	}
}

func seedTeacher(t *testing.T, fixture *fixture, accountID, phone string) core.Principal {
	t.Helper()
	if err := fixture.store.SeedActiveStaff(fixture.owner.TenantID, accountID, "person_"+accountID, phone, "", core.RoleTeacher); err != nil {
		t.Fatalf("seed Teacher: %v", err)
	}
	return core.Principal{AccountID: accountID, TenantID: fixture.owner.TenantID, Roles: []core.Role{core.RoleTeacher}}
}

func lessonInput(key string, startsAt time.Time, teacherAccountID string, studentIDs ...string) app.ScheduleLessonInput {
	return app.ScheduleLessonInput{
		Title: "Lesson " + key, StartsAt: startsAt, DurationMinutes: 60,
		TeacherAccountID: teacherAccountID, StudentIDs: studentIDs, IdempotencyKey: key,
	}
}

func scheduleForTest(t *testing.T, fixture *fixture, principal core.Principal, key string, startsAt time.Time, teacherAccountID string, studentIDs ...string) core.Lesson {
	t.Helper()
	result, err := fixture.service.ScheduleLesson(context.Background(), principal, lessonInput(key, startsAt, teacherAccountID, studentIDs...))
	if err != nil {
		t.Fatalf("schedule %s: %v", key, err)
	}
	return result
}

func directoryStudent(t *testing.T, items []core.StudentDirectoryItem, studentID string) core.StudentDirectoryItem {
	t.Helper()
	for _, item := range items {
		if item.StudentID == studentID {
			return item
		}
	}
	t.Fatalf("Student %s missing from directory %#v", studentID, items)
	return core.StudentDirectoryItem{}
}

func hasAuditAction(fixture *fixture, action string) bool {
	for _, record := range fixture.store.AuditRecords() {
		if record.Action == action && record.Decision == "allow" {
			return true
		}
	}
	return false
}

func hasOutboxEvent(fixture *fixture, eventType string) bool {
	for _, record := range fixture.store.OutboxRecords() {
		if record.EventType == eventType {
			return true
		}
	}
	return false
}
