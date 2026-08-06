package postgres_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

type schedulingOffsetClock struct {
	offset time.Duration
}

func (c schedulingOffsetClock) Now() time.Time {
	return time.Now().UTC().Add(c.offset)
}

func TestPostgreSQLInternalSchedulingAndTeacherContinuity(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, _ := isolatedPool(t, ctx, databaseURL)
	if err := migrations.Up(ctx, pool); err != nil {
		t.Fatalf("apply scheduling migration: %v", err)
	}
	store := postgres.New(pool)
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x64}, 32))
	if err != nil {
		t.Fatalf("new scheduling token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})
	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pg_scheduling", TenantName: "Belcanto Scheduling PostgreSQL",
		FullName: "Scheduling Owner", Phone: "+77003000001",
		Operator: "pg-scheduling-operator", Reason: "scheduling integration fixture",
	})
	if err != nil {
		t.Fatalf("bootstrap scheduling Owner: %v", err)
	}
	activateIntegration(t, ctx, service, ownerLink, "+77003000001", "Owner-password-123!", "activate-pg-scheduling-owner")
	owner := integrationPrincipal(t, ctx, service, "+77003000001", "Owner-password-123!")
	administrator := bootstrapSchedulingStaff(t, ctx, service, owner, core.RoleAdministrator, "Scheduling Administrator", "+77003000002", "Admin-password-123!", "admin")
	firstTeacher := bootstrapSchedulingStaff(t, ctx, service, owner, core.RoleTeacher, "Scheduling Teacher One", "+77003000003", "Teacher-password-123!", "teacher-one")
	secondTeacher := bootstrapSchedulingStaff(t, ctx, service, owner, core.RoleTeacher, "Scheduling Teacher Two", "+77003000004", "Teacher-password-123!", "teacher-two")
	thirdTeacher := bootstrapSchedulingStaff(t, ctx, service, owner, core.RoleTeacher, "Scheduling Teacher Three", "+77003000005", "Teacher-password-123!", "teacher-three")

	firstStudent := createSchedulingStudent(t, ctx, service, owner, "+77003000101", "PG-SCHEDULE-101", firstTeacher.AccountID, "pg-scheduling-student-one")
	secondStudent := createSchedulingStudent(t, ctx, service, owner, "+77003000102", "PG-SCHEDULE-102", firstTeacher.AccountID, "pg-scheduling-student-two")
	thirdStudent := createSchedulingStudent(t, ctx, service, owner, "+77003000103", "PG-SCHEDULE-103", firstTeacher.AccountID, "pg-scheduling-student-three")

	directory, err := service.ListStudents(ctx, administrator, app.ListStudentsInput{})
	if err != nil || len(directory) != 3 {
		t.Fatalf("ordinary Administrator pending Student directory = %#v, %v", directory, err)
	}
	for _, studentID := range []string{firstStudent.StudentID, secondStudent.StudentID, thirdStudent.StudentID} {
		item := pgDirectoryStudent(t, directory, studentID)
		if item.PrimaryTeacher.AccountID != firstTeacher.AccountID || item.PrimaryTeacherAssignmentVersion != 0 {
			t.Fatalf("initial Student directory item = %#v", item)
		}
	}

	// Application clocks must not place assignment facts in the database future,
	// and a default/current projection must remain readable if the API clock is
	// behind PostgreSQL. Explicit projections retain their literal boundary.
	var databaseBefore time.Time
	if err := pool.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&databaseBefore); err != nil {
		t.Fatalf("read database time before skewed Student creation: %v", err)
	}
	aheadService := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour, Clock: schedulingOffsetClock{offset: 5 * time.Minute},
	})
	skewedStudent := createSchedulingStudent(t, ctx, aheadService, owner, "+77003000106", "PG-SCHEDULE-106", firstTeacher.AccountID, "pg-scheduling-clock-skew-student")
	var databaseAfter, skewedAssignedAt, skewedEffectiveFrom time.Time
	if err := pool.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&databaseAfter); err != nil {
		t.Fatalf("read database time after skewed Student creation: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT assigned_at, effective_from
		FROM teacher_assignments
		WHERE tenant_id = $1 AND student_id = $2 AND version = 0
	`, owner.TenantID, skewedStudent.StudentID).Scan(&skewedAssignedAt, &skewedEffectiveFrom); err != nil {
		t.Fatalf("read skewed Student assignment time: %v", err)
	}
	if skewedAssignedAt.Before(databaseBefore) || skewedAssignedAt.After(databaseAfter) || !skewedAssignedAt.Equal(skewedEffectiveFrom) {
		t.Fatalf("skewed Student assignment time = assigned %s, effective %s, database window %s..%s", skewedAssignedAt, skewedEffectiveFrom, databaseBefore, databaseAfter)
	}
	behindService := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour, Clock: schedulingOffsetClock{offset: -5 * time.Minute},
	})
	currentDirectory, err := behindService.ListStudents(ctx, administrator, app.ListStudentsInput{})
	if err != nil || pgDirectoryStudent(t, currentDirectory, skewedStudent.StudentID).PrimaryTeacher.AccountID != firstTeacher.AccountID {
		t.Fatalf("database-backed current projection with lagging API clock = %#v, %v", currentDirectory, err)
	}
	explicitBeforeAssignment, err := behindService.ListStudents(ctx, administrator, app.ListStudentsInput{AsOf: skewedEffectiveFrom.Add(-time.Microsecond)})
	if err != nil {
		t.Fatalf("explicit pre-assignment projection: %v", err)
	}
	for _, item := range explicitBeforeAssignment {
		if item.StudentID == skewedStudent.StudentID {
			t.Fatalf("explicit pre-assignment projection was clamped to current: %#v", item)
		}
	}
	currentQueue, err := behindService.ListStudentOnboarding(ctx, owner)
	if err != nil {
		t.Fatalf("database-backed onboarding projection with lagging API clock: %v", err)
	}
	var skewedQueueTeacher string
	for _, item := range currentQueue {
		if item.StudentID == skewedStudent.StudentID {
			skewedQueueTeacher = item.TeacherAccountID
			break
		}
	}
	if skewedQueueTeacher != firstTeacher.AccountID {
		t.Fatalf("database-backed onboarding projection omitted skewed assignment: %#v", currentQueue)
	}

	now := time.Now().UTC()
	firstLesson := schedulePostgreSQLLesson(t, ctx, service, administrator, "pg-scheduling-first", now.Add(4*time.Hour), firstTeacher.AccountID, firstStudent.StudentID)
	secondLesson := schedulePostgreSQLLesson(t, ctx, service, administrator, "pg-scheduling-second", now.Add(5*time.Hour), firstTeacher.AccountID, secondStudent.StudentID)
	permanentLesson := schedulePostgreSQLLesson(t, ctx, service, administrator, "pg-scheduling-permanent", now.Add(6*time.Hour), firstTeacher.AccountID, firstStudent.StudentID)
	groupLesson, err := service.ScheduleLesson(ctx, administrator, app.ScheduleLessonInput{
		Title: "PostgreSQL group privacy", StartsAt: now.Add(7 * time.Hour), DurationMinutes: 60,
		TeacherAccountID: firstTeacher.AccountID,
		StudentIDs:       []string{firstStudent.StudentID, secondStudent.StudentID},
		IdempotencyKey:   "pg-scheduling-group-privacy",
	})
	if err != nil {
		t.Fatalf("schedule PostgreSQL group Lesson: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE accounts
		SET status = 'active', activated_at = clock_timestamp(), updated_at = clock_timestamp(), version = version + 1
		WHERE tenant_id = $1 AND id = $2
	`, owner.TenantID, firstStudent.AccountID); err != nil {
		t.Fatalf("activate group Lesson Student fixture: %v", err)
	}
	studentPrincipal := core.Principal{TenantID: owner.TenantID, AccountID: firstStudent.AccountID}
	studentLessons, err := service.ListLessons(ctx, studentPrincipal, app.ListLessonsInput{
		From: now.Add(3 * time.Hour), To: now.Add(8 * time.Hour),
	})
	if err != nil {
		t.Fatalf("list Student Lessons: %v", err)
	}
	var studentGroupLesson core.Lesson
	for _, lesson := range studentLessons {
		if lesson.ID == groupLesson.ID {
			studentGroupLesson = lesson
			break
		}
	}
	if studentGroupLesson.ID == "" || len(studentGroupLesson.Students) != 1 || studentGroupLesson.Students[0].StudentID != firstStudent.StudentID {
		t.Fatalf("Student group Lesson list leaked peer participants = %#v", studentGroupLesson)
	}
	studentGroupLesson, err = service.GetLesson(ctx, studentPrincipal, groupLesson.ID)
	if err != nil || len(studentGroupLesson.Students) != 1 || studentGroupLesson.Students[0].StudentID != firstStudent.StudentID {
		t.Fatalf("Student group Lesson detail leaked peer participants = %#v, %v", studentGroupLesson, err)
	}
	managerGroupLesson, err := service.GetLesson(ctx, owner, groupLesson.ID)
	if err != nil || len(managerGroupLesson.Students) != 2 {
		t.Fatalf("manager group Lesson projection = %#v, %v", managerGroupLesson, err)
	}

	_, err = service.ReplaceLessonTeachers(ctx, firstTeacher, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: firstLesson.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: firstTeacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "pg-teacher-replace-forbidden",
	})
	if !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("Teacher replacement = %v", err)
	}
	_, err = service.ReassignPrimaryTeachers(ctx, firstTeacher, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: now.Add(time.Hour), IdempotencyKey: "pg-teacher-reassign-forbidden",
	})
	if !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("Teacher reassignment = %v", err)
	}

	_, err = service.ReplaceLessonTeachers(ctx, administrator, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{
			{LessonID: firstLesson.ID, ExpectedVersion: 0, ExpectedPreviousTeacherAccountID: firstTeacher.AccountID},
			{LessonID: secondLesson.ID, ExpectedVersion: 99, ExpectedPreviousTeacherAccountID: firstTeacher.AccountID},
		},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "pg-replace-atomic-stale",
	})
	if !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale replacement batch = %v", err)
	}
	assertPostgreSQLLessonTeacher(t, ctx, pool, firstLesson.ID, firstTeacher.AccountID, 0)

	replaced, err := service.ReplaceLessonTeachers(ctx, administrator, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: firstLesson.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: firstTeacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "pg-replace-selected",
	})
	if err != nil || replaced.UpdatedCount != 1 || replaced.Lessons[0].Teacher.AccountID != secondTeacher.AccountID || replaced.Lessons[0].Version != 1 {
		t.Fatalf("selected replacement = %#v, %v", replaced, err)
	}
	replacedReplay, err := service.ReplaceLessonTeachers(ctx, administrator, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: firstLesson.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: firstTeacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "pg-replace-selected",
	})
	if err != nil || !reflect.DeepEqual(replacedReplay, replaced) {
		t.Fatalf("replacement replay = %#v, %v", replacedReplay, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE core_lesson_occurrences SET starts_at = $2 WHERE id = $1`, secondLesson.ID, time.Now().UTC().Add(-time.Hour)); err != nil {
		t.Fatalf("prepare past Lesson fixture: %v", err)
	}
	_, err = service.ReplaceLessonTeachers(ctx, administrator, app.ReplaceLessonTeachersInput{
		Lessons: []core.ReplaceLessonTeacherTarget{{LessonID: secondLesson.ID, ExpectedVersion: 0,
			ExpectedPreviousTeacherAccountID: firstTeacher.AccountID}},
		NewTeacherAccountID: secondTeacher.AccountID,
		IdempotencyKey:      "pg-replace-past",
	})
	if !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("past Lesson replacement = %v", err)
	}

	reassigned, err := service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey: "pg-reassign-immediate",
	})
	if err != nil || reassigned.ReassignedCount != 1 || reassigned.Assignments[0].Version != 1 || reassigned.Assignments[0].EffectiveFrom.IsZero() {
		t.Fatalf("immediate reassignment = %#v, %v", reassigned, err)
	}
	reassignedReplay, err := service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey: "pg-reassign-immediate",
	})
	if err != nil || !reflect.DeepEqual(reassignedReplay, reassigned) {
		t.Fatalf("reassignment replay = %#v, %v", reassignedReplay, err)
	}
	unchanged, err := service.GetLesson(ctx, owner, permanentLesson.ID)
	if err != nil || unchanged.Teacher.AccountID != firstTeacher.AccountID || unchanged.Version != 0 {
		t.Fatalf("permanent reassignment mutated Lesson = %#v, %v", unchanged, err)
	}
	_, err = service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: time.Now().UTC().Add(time.Hour), IdempotencyKey: "pg-reassign-stale",
	})
	if !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale assignment version = %v", err)
	}
	_, err = service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 1}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: time.Now().UTC().Add(-2 * time.Minute), IdempotencyKey: "pg-reassign-past",
	})
	if !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("retroactive reassignment = %v", err)
	}

	_, err = service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students: []core.PrimaryTeacherReassignmentTarget{
			{StudentID: firstStudent.StudentID, ExpectedAssignmentVersion: 1},
			{StudentID: thirdStudent.StudentID, ExpectedAssignmentVersion: 99},
		},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: time.Now().UTC().Add(time.Hour), IdempotencyKey: "pg-reassign-atomic-stale",
	})
	if !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale reassignment batch = %v", err)
	}
	directory, err = service.ListStudents(ctx, administrator, app.ListStudentsInput{})
	if err != nil {
		t.Fatalf("directory after failed reassignment batch: %v", err)
	}
	if item := pgDirectoryStudent(t, directory, firstStudent.StudentID); item.PrimaryTeacher.AccountID != secondTeacher.AccountID || item.PrimaryTeacherAssignmentVersion != 1 {
		t.Fatalf("failed reassignment batch mutated valid Student = %#v", item)
	}

	futureEffective := time.Now().UTC().Add(12 * time.Hour)
	_, err = service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: secondStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: secondTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: futureEffective, IdempotencyKey: "pg-reassign-future",
	})
	if err != nil {
		t.Fatalf("future reassignment: %v", err)
	}
	directory, err = service.ListStudents(ctx, administrator, app.ListStudentsInput{})
	if err != nil {
		t.Fatalf("directory after future reassignment: %v", err)
	}
	if item := pgDirectoryStudent(t, directory, secondStudent.StudentID); item.PrimaryTeacher.AccountID != firstTeacher.AccountID || item.PrimaryTeacherAssignmentVersion != 1 {
		t.Fatalf("future timeline concurrency projection = %#v", item)
	}
	_, err = service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: secondStudent.StudentID, ExpectedAssignmentVersion: 0}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom: futureEffective.Add(-time.Hour), IdempotencyKey: "pg-stale-before-known-future",
	})
	if !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale command canceled known future timeline = %v", err)
	}
	projected, err := service.ListStudents(ctx, secondTeacher, app.ListStudentsInput{AsOf: futureEffective.Add(time.Minute)})
	if err != nil || pgDirectoryStudent(t, projected, secondStudent.StudentID).PrimaryTeacher.AccountID != secondTeacher.AccountID {
		t.Fatalf("incoming Teacher future directory = %#v, %v", projected, err)
	}
	if _, err := service.ScheduleLesson(ctx, firstTeacher, pgLessonInput("pg-old-before-boundary", futureEffective.Add(-90*time.Minute), firstTeacher.AccountID, secondStudent.StudentID)); err != nil {
		t.Fatalf("old Teacher before boundary: %v", err)
	}
	if _, err := service.ScheduleLesson(ctx, firstTeacher, pgLessonInput("pg-old-after-boundary", futureEffective.Add(time.Minute), firstTeacher.AccountID, secondStudent.StudentID)); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("old Teacher after boundary = %v", err)
	}
	if _, err := service.ScheduleLesson(ctx, secondTeacher, pgLessonInput("pg-new-before-boundary", futureEffective.Add(-90*time.Minute), secondTeacher.AccountID, secondStudent.StudentID)); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("new Teacher before boundary = %v", err)
	}
	if _, err := service.ScheduleLesson(ctx, secondTeacher, pgLessonInput("pg-new-after-boundary", futureEffective.Add(time.Minute), secondTeacher.AccountID, secondStudent.StudentID)); err != nil {
		t.Fatalf("new Teacher after boundary: %v", err)
	}

	assignmentLockTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin First Minute assignment lock harness: %v", err)
	}
	assignmentLockKey := integrationAdvisoryLockKey("primary-teacher-assignment", owner.TenantID, thirdStudent.StudentID)
	if _, err := assignmentLockTx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, assignmentLockKey); err != nil {
		_ = assignmentLockTx.Rollback(ctx)
		t.Fatalf("hold primary Teacher assignment lock: %v", err)
	}
	firstMinuteDone := make(chan error, 1)
	go func() {
		_, publishErr := service.PublishFirstMinute(ctx, firstTeacher, app.PublishFirstMinuteInput{
			StudentID: thirdStudent.StudentID, WhatWorked: "Lock proof", CurrentFocus: "Serialized assignment",
			NextStep: "Continue safely", ExpectedVersion: 0, IdempotencyKey: "pg-first-minute-assignment-lock",
		})
		firstMinuteDone <- publishErr
	}()
	select {
	case publishErr := <-firstMinuteDone:
		_ = assignmentLockTx.Rollback(ctx)
		t.Fatalf("First Minute bypassed held assignment lock: %v", publishErr)
	case <-time.After(75 * time.Millisecond):
	}
	if err := assignmentLockTx.Commit(ctx); err != nil {
		t.Fatalf("release primary Teacher assignment lock: %v", err)
	}
	if publishErr := <-firstMinuteDone; publishErr != nil {
		t.Fatalf("First Minute after assignment lock release: %v", publishErr)
	}

	// Deterministically model an older immediate-reassignment request that is
	// delayed before the assignment subject lock while a newer First Minute
	// publication acquires that lock and commits first. The later mutation must
	// not use the request timestamp to cut the former Teacher's interval
	// retroactively.
	temporalStudent := createSchedulingStudent(t, ctx, service, owner, "+77003000104", "PG-SCHEDULE-104", firstTeacher.AccountID, "pg-scheduling-temporal-student")
	temporalAssignmentID, err := security.NewID("assignment")
	if err != nil {
		t.Fatalf("generate temporal reassignment id: %v", err)
	}
	temporalFingerprint := bytes.Repeat([]byte{0x71}, 32)
	temporalReassignmentKey := "pg-temporal-first-minute-reassignment"
	temporalRequestAt := time.Now().UTC()
	temporalCommand := core.ReassignPrimaryTeachersCommand{
		TenantID: owner.TenantID, ActorAccountID: administrator.AccountID,
		Targets: []core.PrimaryTeacherReassignmentTarget{{
			StudentID: temporalStudent.StudentID, ExpectedAssignmentVersion: 0,
			AssignmentID: temporalAssignmentID,
		}},
		NewTeacherAccountID: secondTeacher.AccountID,
		EffectiveMode:       core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey:      temporalReassignmentKey,
		PayloadFingerprint:  temporalFingerprint,
		Now:                 temporalRequestAt,
	}
	temporalBlocker, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin temporal idempotency blocker: %v", err)
	}
	if _, err := temporalBlocker.Exec(ctx, `
		INSERT INTO idempotency_records (
			tenant_id, actor_account_id, operation_scope, idempotency_key,
			payload_fingerprint, status, created_at
		) VALUES ($1, $2, 'reassign_primary_teachers', $3, $4, 'processing', $5)
	`, owner.TenantID, administrator.AccountID, temporalReassignmentKey, temporalFingerprint, temporalRequestAt); err != nil {
		_ = temporalBlocker.Rollback(ctx)
		t.Fatalf("block temporal reassignment idempotency claim: %v", err)
	}
	type reassignmentOutcome struct {
		result core.PrimaryTeacherReassignmentResult
		err    error
	}
	temporalReassignmentDone := make(chan reassignmentOutcome, 1)
	go func() {
		result, reassignErr := store.ReassignPrimaryTeachers(ctx, temporalCommand)
		temporalReassignmentDone <- reassignmentOutcome{result: result, err: reassignErr}
	}()
	select {
	case outcome := <-temporalReassignmentDone:
		_ = temporalBlocker.Rollback(ctx)
		t.Fatalf("temporal reassignment bypassed held idempotency claim: %#v, %v", outcome.result, outcome.err)
	case <-time.After(75 * time.Millisecond):
	}
	temporalFirstMinuteInput := app.PublishFirstMinuteInput{
		StudentID: temporalStudent.StudentID, WhatWorked: "Temporal worked",
		CurrentFocus: "Serialized continuity", NextStep: "Transfer safely",
		ExpectedVersion: 0, IdempotencyKey: "pg-temporal-first-minute",
	}
	temporalFirstMinute, err := service.PublishFirstMinute(ctx, firstTeacher, temporalFirstMinuteInput)
	if err != nil {
		_ = temporalBlocker.Rollback(ctx)
		t.Fatalf("publish temporal First Minute: %v", err)
	}
	temporalFirstMinuteReplay, err := service.PublishFirstMinute(ctx, firstTeacher, temporalFirstMinuteInput)
	if err != nil || !reflect.DeepEqual(temporalFirstMinuteReplay, temporalFirstMinute) {
		_ = temporalBlocker.Rollback(ctx)
		t.Fatalf("temporal First Minute replay = %#v, %v", temporalFirstMinuteReplay, err)
	}
	if err := temporalBlocker.Rollback(ctx); err != nil {
		t.Fatalf("release temporal idempotency blocker: %v", err)
	}
	temporalReassignment := <-temporalReassignmentDone
	if temporalReassignment.err != nil || temporalReassignment.result.ReassignedCount != 1 {
		t.Fatalf("temporal reassignment after First Minute = %#v, %v", temporalReassignment.result, temporalReassignment.err)
	}
	temporalCutover := temporalReassignment.result.Assignments[0].EffectiveFrom
	if !temporalFirstMinute.PublishedAt.Before(temporalCutover) {
		t.Fatalf("First Minute/cutover order = %s / %s; publication must precede cutover", temporalFirstMinute.PublishedAt, temporalCutover)
	}
	var authorWasEffective bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM teacher_assignments ta
			WHERE ta.tenant_id = $1 AND ta.student_id = $2
			  AND ta.teacher_account_id = $3 AND ta.status = 'active'
			  AND ta.effective_from <= $4
			  AND (ta.effective_until IS NULL OR $4 < ta.effective_until)
		)
	`, owner.TenantID, temporalStudent.StudentID, firstTeacher.AccountID, temporalFirstMinute.PublishedAt).Scan(&authorWasEffective); err != nil || !authorWasEffective {
		t.Fatalf("First Minute author effective at publication = %v, %v", authorWasEffective, err)
	}
	temporalReplay, err := store.ReassignPrimaryTeachers(ctx, temporalCommand)
	if err != nil || !reflect.DeepEqual(temporalReplay, temporalReassignment.result) {
		t.Fatalf("temporal reassignment replay = %#v, %v", temporalReplay, err)
	}
	var temporalAssignmentCount, temporalAuditCount, temporalEventCount int
	var assignedAt, auditedAt, eventAt, completedAt time.Time
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM teacher_assignments WHERE tenant_id = $1 AND student_id = $2 AND version = 1),
			(SELECT count(*) FROM audit_records WHERE tenant_id = $1 AND action = 'StudentPrimaryTeacherReassigned' AND target_id = $2 AND decision = 'allow'),
			(SELECT count(*) FROM outbox_events WHERE tenant_id = $1 AND event_type = 'StudentPrimaryTeacherReassigned' AND aggregate_id = $2),
			(SELECT assigned_at FROM teacher_assignments WHERE tenant_id = $1 AND student_id = $2 AND version = 1),
			(SELECT recorded_at FROM audit_records WHERE tenant_id = $1 AND action = 'StudentPrimaryTeacherReassigned' AND target_id = $2 AND decision = 'allow'),
			(SELECT recorded_at FROM outbox_events WHERE tenant_id = $1 AND event_type = 'StudentPrimaryTeacherReassigned' AND aggregate_id = $2),
			(SELECT completed_at FROM idempotency_records WHERE tenant_id = $1 AND actor_account_id = $3 AND operation_scope = 'reassign_primary_teachers' AND idempotency_key = $4)
	`, owner.TenantID, temporalStudent.StudentID, administrator.AccountID, temporalReassignmentKey).Scan(
		&temporalAssignmentCount, &temporalAuditCount, &temporalEventCount, &assignedAt, &auditedAt, &eventAt, &completedAt,
	); err != nil {
		t.Fatalf("read temporal reassignment atomic records: %v", err)
	}
	if temporalAssignmentCount != 1 || temporalAuditCount != 1 || temporalEventCount != 1 {
		t.Fatalf("temporal replay duplicated records: assignment/audit/event = %d/%d/%d", temporalAssignmentCount, temporalAuditCount, temporalEventCount)
	}
	if !assignedAt.Equal(auditedAt) || !assignedAt.Equal(eventAt) || !assignedAt.Equal(completedAt) || assignedAt.After(temporalCutover) {
		t.Fatalf("temporal mutation timestamps = assigned %s, audit %s, event %s, completed %s, cutover %s", assignedAt, auditedAt, eventAt, completedAt, temporalCutover)
	}

	// A scheduled cutover can become stale while waiting for the same subject
	// lock. It must be revalidated after the wait and roll back atomically.
	scheduledWaitStudent := createSchedulingStudent(t, ctx, service, owner, "+77003000105", "PG-SCHEDULE-105", firstTeacher.AccountID, "pg-scheduling-wait-student")
	scheduledLockTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin scheduled reassignment lock harness: %v", err)
	}
	scheduledLockKey := integrationAdvisoryLockKey("primary-teacher-assignment", owner.TenantID, scheduledWaitStudent.StudentID)
	if _, err := scheduledLockTx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, scheduledLockKey); err != nil {
		_ = scheduledLockTx.Rollback(ctx)
		t.Fatalf("hold scheduled reassignment assignment lock: %v", err)
	}
	scheduledAssignmentID, err := security.NewID("assignment")
	if err != nil {
		_ = scheduledLockTx.Rollback(ctx)
		t.Fatalf("generate scheduled-wait reassignment id: %v", err)
	}
	var scheduledRequestAt time.Time
	if err := scheduledLockTx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&scheduledRequestAt); err != nil {
		_ = scheduledLockTx.Rollback(ctx)
		t.Fatalf("read scheduled reassignment database time: %v", err)
	}
	scheduledEffectiveFrom := scheduledRequestAt.Add(350 * time.Millisecond)
	scheduledReassignmentKey := "pg-scheduled-cutover-passed-while-waiting"
	scheduledCommand := core.ReassignPrimaryTeachersCommand{
		TenantID: owner.TenantID, ActorAccountID: administrator.AccountID,
		Targets: []core.PrimaryTeacherReassignmentTarget{{
			StudentID: scheduledWaitStudent.StudentID, ExpectedAssignmentVersion: 0,
			AssignmentID: scheduledAssignmentID,
		}},
		NewTeacherAccountID: secondTeacher.AccountID,
		EffectiveMode:       core.PrimaryTeacherEffectiveScheduled,
		EffectiveFrom:       scheduledEffectiveFrom,
		IdempotencyKey:      scheduledReassignmentKey,
		PayloadFingerprint:  bytes.Repeat([]byte{0x72}, 32),
		Now:                 scheduledRequestAt,
	}
	scheduledWaitDone := make(chan error, 1)
	go func() {
		_, reassignErr := store.ReassignPrimaryTeachers(ctx, scheduledCommand)
		scheduledWaitDone <- reassignErr
	}()
	select {
	case waitErr := <-scheduledWaitDone:
		_ = scheduledLockTx.Rollback(ctx)
		t.Fatalf("scheduled reassignment bypassed held assignment lock: %v", waitErr)
	case <-time.After(75 * time.Millisecond):
	}
	for {
		var effectiveFromPassed bool
		if err := scheduledLockTx.QueryRow(ctx, `SELECT clock_timestamp() >= $1`, scheduledEffectiveFrom).Scan(&effectiveFromPassed); err != nil {
			_ = scheduledLockTx.Rollback(ctx)
			t.Fatalf("wait for scheduled reassignment database boundary: %v", err)
		}
		if effectiveFromPassed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := scheduledLockTx.Commit(ctx); err != nil {
		t.Fatalf("release scheduled reassignment assignment lock: %v", err)
	}
	if waitErr := <-scheduledWaitDone; !core.IsCode(waitErr, core.CodeInvalidInput) {
		t.Fatalf("scheduled reassignment after effectiveFrom passed = %v", waitErr)
	}
	var scheduledTimelineCount, scheduledIdempotencyCount int
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM teacher_assignments WHERE tenant_id = $1 AND student_id = $2),
			(SELECT count(*) FROM idempotency_records WHERE tenant_id = $1 AND actor_account_id = $3 AND operation_scope = 'reassign_primary_teachers' AND idempotency_key = $4)
	`, owner.TenantID, scheduledWaitStudent.StudentID, administrator.AccountID, scheduledReassignmentKey).Scan(
		&scheduledTimelineCount, &scheduledIdempotencyCount,
	); err != nil {
		t.Fatalf("read stale scheduled reassignment state: %v", err)
	}
	if scheduledTimelineCount != 1 || scheduledIdempotencyCount != 0 {
		t.Fatalf("stale scheduled reassignment mutated state: timeline/idempotency = %d/%d", scheduledTimelineCount, scheduledIdempotencyCount)
	}

	type scheduleOutcome struct {
		lesson core.Lesson
		err    error
	}
	studentRaceStart := time.Now().UTC().Add(20 * time.Hour)
	studentRaceDone := make(chan scheduleOutcome, 2)
	for index, teacher := range []core.Principal{firstTeacher, secondTeacher} {
		index, teacher := index, teacher
		go func() {
			lesson, raceErr := service.ScheduleLesson(ctx, administrator, pgLessonInput(
				fmt.Sprintf("pg-student-overlap-race-%d", index), studentRaceStart, teacher.AccountID, thirdStudent.StudentID,
			))
			studentRaceDone <- scheduleOutcome{lesson: lesson, err: raceErr}
		}()
	}
	assertOneSchedulingRaceWinner(t, <-studentRaceDone, <-studentRaceDone)

	teacherRaceStart := time.Now().UTC().Add(22 * time.Hour)
	teacherRaceDone := make(chan scheduleOutcome, 2)
	for index, studentID := range []string{firstStudent.StudentID, thirdStudent.StudentID} {
		index, studentID := index, studentID
		go func() {
			lesson, raceErr := service.ScheduleLesson(ctx, administrator, pgLessonInput(
				fmt.Sprintf("pg-teacher-overlap-race-%d", index), teacherRaceStart, thirdTeacher.AccountID, studentID,
			))
			teacherRaceDone <- scheduleOutcome{lesson: lesson, err: raceErr}
		}()
	}
	assertOneSchedulingRaceWinner(t, <-teacherRaceDone, <-teacherRaceDone)

	if _, err := pool.Exec(ctx, `UPDATE accounts SET status = 'suspended' WHERE tenant_id = $1 AND id = $2`, owner.TenantID, firstTeacher.AccountID); err != nil {
		t.Fatalf("disable departed PostgreSQL Teacher: %v", err)
	}
	directory, err = service.ListStudents(ctx, administrator, app.ListStudentsInput{})
	if err != nil {
		t.Fatalf("manager directory with departed PostgreSQL Teacher: %v", err)
	}
	departed := pgDirectoryStudent(t, directory, thirdStudent.StudentID)
	if departed.PrimaryTeacher.AccountID != firstTeacher.AccountID || departed.PrimaryTeacher.Status != core.AssignedTeacherInactive {
		t.Fatalf("departed PostgreSQL Teacher continuity projection = %#v", departed)
	}
	if _, err := service.ScheduleLesson(ctx, firstTeacher, pgLessonInput("pg-departed-teacher-denied", time.Now().UTC().Add(30*time.Hour), firstTeacher.AccountID, thirdStudent.StudentID)); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("departed PostgreSQL Teacher schedules Lesson = %v", err)
	}
	if _, err := service.ReassignPrimaryTeachers(ctx, administrator, app.ReassignPrimaryTeachersInput{
		Students:            []core.PrimaryTeacherReassignmentTarget{{StudentID: thirdStudent.StudentID, ExpectedAssignmentVersion: departed.PrimaryTeacherAssignmentVersion}},
		NewTeacherAccountID: thirdTeacher.AccountID, EffectiveMode: core.PrimaryTeacherEffectiveImmediate,
		IdempotencyKey: "pg-reassign-from-departed-teacher",
	}); err != nil {
		t.Fatalf("transfer PostgreSQL Student from departed Teacher: %v", err)
	}

	overlapAssignmentID, err := security.NewID("assignment")
	if err != nil {
		t.Fatalf("generate overlapping assignment id: %v", err)
	}
	assertConstraintRejected(t, ctx, pool, `
		INSERT INTO teacher_assignments (
			id, tenant_id, student_id, teacher_account_id, status,
			assigned_by_account_id, assigned_at, effective_from, version
		) VALUES ($1, $2, $3, $4, 'active', $5, $6, $6, 99)
	`, overlapAssignmentID, owner.TenantID, firstStudent.StudentID, thirdTeacher.AccountID, administrator.AccountID, time.Now().UTC())

	var auditCount int
	var neutralReasons bool
	if err := pool.QueryRow(ctx, `
		SELECT count(*), bool_and(reason_code IN ('temporary_teacher_continuity', 'primary_teacher_continuity')) FROM audit_records
		WHERE action IN ('LessonTeacherReplaced', 'StudentPrimaryTeacherReassigned')
		  AND decision = 'allow' AND reason_code IS NOT NULL AND metadata <> '{}'::jsonb
	`).Scan(&auditCount, &neutralReasons); err != nil || auditCount < 2 || !neutralReasons {
		t.Fatalf("continuity audit rows/reasons = %d/%v, %v", auditCount, neutralReasons, err)
	}
	var outboxCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM outbox_events
		WHERE event_type IN ('LessonTeacherReplaced', 'StudentPrimaryTeacherReassigned')
	`).Scan(&outboxCount); err != nil || outboxCount < 2 {
		t.Fatalf("continuity outbox rows = %d, %v", outboxCount, err)
	}
}

func assertOneSchedulingRaceWinner(t *testing.T, first, second struct {
	lesson core.Lesson
	err    error
}) {
	t.Helper()
	if (first.err == nil) == (second.err == nil) {
		t.Fatalf("scheduling race outcomes = %v / %v; want exactly one success", first.err, second.err)
	}
	loser := first
	if loser.err == nil {
		loser = second
	}
	if !core.IsCode(loser.err, core.CodeConflict) {
		t.Fatalf("scheduling race loser = %v, want conflict", loser.err)
	}
}

func bootstrapSchedulingStaff(t *testing.T, ctx context.Context, service *app.Service, owner core.Principal, role core.Role, fullName, phone, password, key string) core.Principal {
	t.Helper()
	link, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: fullName, Phone: phone, Role: role,
		Operator: "pg-scheduling-operator", Reason: "scheduling integration fixture",
	})
	if err != nil {
		t.Fatalf("bootstrap scheduling %s: %v", role, err)
	}
	activateIntegration(t, ctx, service, link, phone, password, "activate-pg-scheduling-"+key)
	return integrationPrincipal(t, ctx, service, phone, password)
}

func createSchedulingStudent(t *testing.T, ctx context.Context, service *app.Service, owner core.Principal, phone, enrollment, teacherAccountID, key string) core.StudentResult {
	t.Helper()
	result, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "Pending " + enrollment, Phone: phone, EnrollmentReference: enrollment,
		TeacherAccountID: teacherAccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("create scheduling Student: %v", err)
	}
	return result
}

func pgLessonInput(key string, startsAt time.Time, teacherAccountID, studentID string) app.ScheduleLessonInput {
	return app.ScheduleLessonInput{
		Title: "PostgreSQL " + key, StartsAt: startsAt, DurationMinutes: 60,
		TeacherAccountID: teacherAccountID, StudentIDs: []string{studentID}, IdempotencyKey: key,
	}
}

func schedulePostgreSQLLesson(t *testing.T, ctx context.Context, service *app.Service, principal core.Principal, key string, startsAt time.Time, teacherAccountID, studentID string) core.Lesson {
	t.Helper()
	result, err := service.ScheduleLesson(ctx, principal, pgLessonInput(key, startsAt, teacherAccountID, studentID))
	if err != nil {
		t.Fatalf("schedule PostgreSQL Lesson %s: %v", key, err)
	}
	return result
}

func pgDirectoryStudent(t *testing.T, directory []core.StudentDirectoryItem, studentID string) core.StudentDirectoryItem {
	t.Helper()
	for _, item := range directory {
		if item.StudentID == studentID {
			return item
		}
	}
	t.Fatalf("Student %s missing from directory %#v", studentID, directory)
	return core.StudentDirectoryItem{}
}

func assertPostgreSQLLessonTeacher(t *testing.T, ctx context.Context, pool *pgxpool.Pool, lessonID, teacherAccountID string, version int64) {
	t.Helper()
	var storedTeacherID string
	var storedVersion int64
	if err := pool.QueryRow(ctx, `SELECT teacher_account_id, version FROM core_lesson_occurrences WHERE id = $1`, lessonID).Scan(&storedTeacherID, &storedVersion); err != nil {
		t.Fatalf("read stored Lesson Teacher: %v", err)
	}
	if storedTeacherID != teacherAccountID || storedVersion != version {
		t.Fatalf("stored Lesson Teacher/version = %s/%d, want %s/%d", storedTeacherID, storedVersion, teacherAccountID, version)
	}
}
