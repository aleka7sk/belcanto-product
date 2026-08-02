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

	now := time.Now().UTC()
	firstLesson := schedulePostgreSQLLesson(t, ctx, service, administrator, "pg-scheduling-first", now.Add(4*time.Hour), firstTeacher.AccountID, firstStudent.StudentID)
	secondLesson := schedulePostgreSQLLesson(t, ctx, service, administrator, "pg-scheduling-second", now.Add(5*time.Hour), firstTeacher.AccountID, secondStudent.StudentID)
	permanentLesson := schedulePostgreSQLLesson(t, ctx, service, administrator, "pg-scheduling-permanent", now.Add(6*time.Hour), firstTeacher.AccountID, firstStudent.StudentID)

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
	if _, err := pool.Exec(ctx, `UPDATE lessons SET starts_at = $2 WHERE id = $1`, secondLesson.ID, time.Now().UTC().Add(-time.Hour)); err != nil {
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
	if err := pool.QueryRow(ctx, `SELECT teacher_account_id, version FROM lessons WHERE id = $1`, lessonID).Scan(&storedTeacherID, &storedVersion); err != nil {
		t.Fatalf("read stored Lesson Teacher: %v", err)
	}
	if storedTeacherID != teacherAccountID || storedVersion != version {
		t.Fatalf("stored Lesson Teacher/version = %s/%d, want %s/%d", storedTeacherID, storedVersion, teacherAccountID, version)
	}
}
