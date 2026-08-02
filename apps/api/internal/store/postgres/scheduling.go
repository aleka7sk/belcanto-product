package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) ListStudents(ctx context.Context, principal core.Principal, asOf, now time.Time) ([]core.StudentDirectoryItem, error) {
	result := make([]core.StudentDirectoryItem, 0)
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		owner, err := hasActiveRole(ctx, tx, principal.TenantID, principal.AccountID, core.RoleOwner)
		if err != nil {
			return err
		}
		administrator, err := hasActiveRole(ctx, tx, principal.TenantID, principal.AccountID, core.RoleAdministrator)
		if err != nil {
			return err
		}
		teacher, err := hasActiveRole(ctx, tx, principal.TenantID, principal.AccountID, core.RoleTeacher)
		if err != nil {
			return err
		}
		manager := owner || administrator
		if !manager && !teacher {
			return core.E(core.CodeForbidden, "student directory permission is required", nil)
		}
		rows, err := tx.Query(ctx, `
			SELECT s.id, student_person.full_name, teacher_account.id, teacher_person.full_name,
			       CASE
			         WHEN teacher_account.status = 'active' AND EXISTS (
			           SELECT 1
			           FROM role_grants teacher_role
			           WHERE teacher_role.tenant_id = teacher_account.tenant_id
			             AND teacher_role.account_id = teacher_account.id
			             AND teacher_role.role_type = 'Teacher'
			             AND teacher_role.status = 'active'
			         ) THEN 'active'
			         ELSE 'inactive'
			       END,
			       (
			         SELECT max(timeline_assignment.version)
			         FROM teacher_assignments timeline_assignment
			         WHERE timeline_assignment.tenant_id = s.tenant_id
			           AND timeline_assignment.student_id = s.id
			       )
			FROM students s
			JOIN accounts student_account
			  ON student_account.tenant_id = s.tenant_id AND student_account.id = s.account_id
			 AND student_account.status IN ('pending_activation', 'active')
			JOIN role_grants student_role
			  ON student_role.tenant_id = s.tenant_id AND student_role.account_id = s.account_id
			 AND student_role.role_type = 'Student' AND student_role.scope_type = 'student'
			 AND student_role.scope_id = s.id AND student_role.status = 'active'
			JOIN people student_person
			  ON student_person.tenant_id = s.tenant_id AND student_person.id = s.person_id
			JOIN teacher_assignments ta
			  ON ta.tenant_id = s.tenant_id AND ta.student_id = s.id AND ta.status = 'active'
			 AND ta.effective_from <= $3
			 AND (ta.effective_until IS NULL OR $3 < ta.effective_until)
			JOIN accounts teacher_account
			  ON teacher_account.tenant_id = ta.tenant_id AND teacher_account.id = ta.teacher_account_id
			JOIN people teacher_person
			  ON teacher_person.tenant_id = teacher_account.tenant_id AND teacher_person.id = teacher_account.person_id
			WHERE s.tenant_id = $1 AND s.status = 'active'
			  AND ($4::boolean OR ta.teacher_account_id = $2)
			ORDER BY student_person.full_name, s.id
		`, principal.TenantID, principal.AccountID, asOf, manager)
		if err != nil {
			return fmt.Errorf("list Student directory: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var item core.StudentDirectoryItem
			if err := rows.Scan(
				&item.StudentID,
				&item.FullName,
				&item.PrimaryTeacher.AccountID,
				&item.PrimaryTeacher.FullName,
				&item.PrimaryTeacher.Status,
				&item.PrimaryTeacherAssignmentVersion,
			); err != nil {
				return fmt.Errorf("scan Student directory item: %w", err)
			}
			result = append(result, item)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate Student directory: %w", err)
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "StudentDirectoryListed", targetType: "student_directory",
			targetID: "students", decision: "allow", at: now,
		})
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "StudentDirectoryListed", targetType: "student_directory",
			targetID: "students", reason: "student_directory_not_allowed", at: now,
		})
	}
	return result, err
}

func (s *Store) ScheduleLesson(ctx context.Context, command core.ScheduleLessonCommand) (core.Lesson, error) {
	var result core.Lesson
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, teacher, err := lessonCreateAuthority(ctx, tx, command.TenantID, command.ActorAccountID)
		if err != nil {
			return err
		}
		if !manager && !teacher {
			return core.E(core.CodeForbidden, "Lesson scheduling permission is required", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "schedule_lesson", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.Lesson](claim)
			return err
		}
		if !command.StartsAt.After(command.Now) {
			return core.E(core.CodeInvalidState, "Lesson must start in the future", nil)
		}
		activeTeacher, err := hasActiveRole(ctx, tx, command.TenantID, command.TeacherAccountID, core.RoleTeacher)
		if err != nil {
			return err
		}
		if !activeTeacher {
			return core.E(core.CodeInvalidInput, "Teacher is not active in this school", nil)
		}
		if teacher && !manager && command.TeacherAccountID != command.ActorAccountID {
			return core.E(core.CodeForbidden, "Teacher can only schedule Lessons for self", nil)
		}
		studentIDs, err := uniqueSorted(command.StudentIDs, "Student ids must be unique")
		if err != nil {
			return err
		}
		if len(studentIDs) == 0 {
			return core.E(core.CodeInvalidInput, "at least one Student is required", nil)
		}
		if err := lockLessonScheduleSubjects(ctx, tx, command.TenantID, command.TeacherAccountID, studentIDs); err != nil {
			return err
		}
		if err := lockAssignmentSubjects(ctx, tx, command.TenantID, studentIDs); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `
			SELECT s.id, current_assignment.teacher_account_id
			FROM students s
			JOIN accounts a
			  ON a.tenant_id = s.tenant_id AND a.id = s.account_id
			 AND a.status IN ('pending_activation', 'active')
			JOIN role_grants rg
			  ON rg.tenant_id = s.tenant_id AND rg.account_id = s.account_id
			 AND rg.role_type = 'Student' AND rg.scope_type = 'student'
			 AND rg.scope_id = s.id AND rg.status = 'active'
			LEFT JOIN LATERAL (
				SELECT ta.teacher_account_id
				FROM teacher_assignments ta
				WHERE ta.tenant_id = s.tenant_id AND ta.student_id = s.id
				  AND ta.status = 'active' AND ta.effective_from <= $3
				  AND (ta.effective_until IS NULL OR $3 < ta.effective_until)
				ORDER BY ta.effective_from DESC, ta.id DESC
				LIMIT 1
			) current_assignment ON true
			WHERE s.tenant_id = $1 AND s.id = ANY($2::text[]) AND s.status = 'active'
			ORDER BY s.id
		`, command.TenantID, studentIDs, command.StartsAt)
		if err != nil {
			return fmt.Errorf("validate Lesson Students: %w", err)
		}
		validated := make(map[string]string, len(studentIDs))
		for rows.Next() {
			var studentID string
			var assignedTeacherID *string
			if err := rows.Scan(&studentID, &assignedTeacherID); err != nil {
				rows.Close()
				return fmt.Errorf("scan Lesson Student validation: %w", err)
			}
			if assignedTeacherID != nil {
				validated[studentID] = *assignedTeacherID
			} else {
				validated[studentID] = ""
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("iterate Lesson Student validation: %w", err)
		}
		rows.Close()
		if len(validated) != len(studentIDs) {
			return core.E(core.CodeInvalidInput, "Student is not active in this school", nil)
		}
		if teacher && !manager {
			for _, studentID := range studentIDs {
				if validated[studentID] != command.ActorAccountID {
					return core.E(core.CodeForbidden, "Teacher can only schedule Students assigned at Lesson start", nil)
				}
			}
		}
		conflict, err := lessonScheduleConflict(ctx, tx, command.TenantID, command.StartsAt, command.DurationMinutes, command.TeacherAccountID, studentIDs, nil)
		if err != nil {
			return err
		}
		if conflict {
			return core.E(core.CodeConflict, "Teacher or Student has an overlapping Lesson", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO lessons (
				id, tenant_id, title, starts_at, duration_minutes, location,
				teacher_account_id, status, version, created_by_account_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, 'scheduled', 0, $8, $9, $9)
		`, command.LessonID, command.TenantID, command.Title, command.StartsAt,
			command.DurationMinutes, command.Location, command.TeacherAccountID,
			command.ActorAccountID, command.Now); err != nil {
			return mapSchedulingWriteError(err, "Lesson conflicts with existing data")
		}
		for _, studentID := range studentIDs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO lesson_participants (tenant_id, lesson_id, student_id, added_at)
				VALUES ($1, $2, $3, $4)
			`, command.TenantID, command.LessonID, studentID, command.Now); err != nil {
				return mapSchedulingWriteError(err, "Lesson participant conflicts with existing data")
			}
		}
		result, err = readLesson(ctx, tx, command.TenantID, command.LessonID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "schedule_lesson", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "LessonScheduled", targetType: "lesson", targetID: command.LessonID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"teacherAccountId": command.TeacherAccountID, "studentIds": studentIDs}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "LessonScheduled", "lesson", command.LessonID, map[string]any{
			"lessonId": command.LessonID, "teacherAccountId": command.TeacherAccountID,
			"studentIds": studentIDs, "startsAt": command.StartsAt,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "LessonScheduled", targetType: "lesson", targetID: command.LessonID,
			reason: "lesson_create_not_allowed", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return result, err
}

func (s *Store) ListLessons(ctx context.Context, principal core.Principal, query core.LessonListQuery, _ time.Time) ([]core.Lesson, error) {
	scope, err := s.lessonReadScope(ctx, principal)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT l.id
		FROM lessons l
		WHERE l.tenant_id = $1 AND l.starts_at >= $2 AND l.starts_at < $3
		  AND ($4 = '' OR l.teacher_account_id = $4)
		  AND ($5 = '' OR EXISTS (
			SELECT 1 FROM lesson_participants filtered_participant
			WHERE filtered_participant.tenant_id = l.tenant_id
			  AND filtered_participant.lesson_id = l.id
			  AND filtered_participant.student_id = $5
		  ))
		  AND (
			$6::boolean
			OR ($7 <> '' AND l.teacher_account_id = $7)
			OR ($8 <> '' AND EXISTS (
				SELECT 1
				FROM lesson_participants scoped_participant
				WHERE scoped_participant.tenant_id = l.tenant_id
				  AND scoped_participant.lesson_id = l.id
				  AND scoped_participant.student_id = $8
			))
		  )
		ORDER BY l.starts_at, l.id
	`, principal.TenantID, query.From, query.To, query.TeacherAccountID, query.StudentID,
		scope.manager, scope.teacherID, scope.studentID)
	if err != nil {
		return nil, fmt.Errorf("list Lessons: %w", err)
	}
	defer rows.Close()
	lessonIDs := make([]string, 0)
	for rows.Next() {
		var lessonID string
		if err := rows.Scan(&lessonID); err != nil {
			return nil, fmt.Errorf("scan Lesson id: %w", err)
		}
		lessonIDs = append(lessonIDs, lessonID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Lesson ids: %w", err)
	}
	rows.Close()
	result := make([]core.Lesson, 0, len(lessonIDs))
	for _, lessonID := range lessonIDs {
		item, err := readLesson(ctx, s.pool, principal.TenantID, lessonID)
		if err != nil {
			return nil, err
		}
		if lessonVisible(scope, item) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Store) GetLesson(ctx context.Context, principal core.Principal, lessonID string, _ time.Time) (core.Lesson, error) {
	scope, err := s.lessonReadScope(ctx, principal)
	if err != nil {
		return core.Lesson{}, err
	}
	result, err := readLesson(ctx, s.pool, principal.TenantID, lessonID)
	if err != nil {
		return core.Lesson{}, err
	}
	if !lessonVisible(scope, result) {
		return core.Lesson{}, core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	return result, nil
}

func (s *Store) ReplaceLessonTeachers(ctx context.Context, command core.ReplaceLessonTeachersCommand) (core.LessonTeacherReplacementResult, error) {
	var result core.LessonTeacherReplacementResult
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, err := lessonManagementAuthority(ctx, tx, command.TenantID, command.ActorAccountID)
		if err != nil {
			return err
		}
		if !manager {
			return core.E(core.CodeForbidden, "Lesson Teacher replacement permission is required", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "replace_lesson_teachers", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.LessonTeacherReplacementResult](claim)
			return err
		}
		activeTeacher, err := hasActiveRole(ctx, tx, command.TenantID, command.NewTeacherAccountID, core.RoleTeacher)
		if err != nil {
			return err
		}
		if !activeTeacher {
			return core.E(core.CodeInvalidInput, "new Teacher is not active in this school", nil)
		}
		if err := lockLessonScheduleSubjects(ctx, tx, command.TenantID, command.NewTeacherAccountID, nil); err != nil {
			return err
		}
		if len(command.Targets) == 0 {
			return core.E(core.CodeInvalidInput, "at least one Lesson is required", nil)
		}
		targets := append([]core.ReplaceLessonTeacherTarget(nil), command.Targets...)
		sort.Slice(targets, func(left, right int) bool { return targets[left].LessonID < targets[right].LessonID })
		plans := make([]lessonReplacementPlan, 0, len(targets))
		seen := make(map[string]struct{}, len(targets))
		for _, target := range targets {
			if _, duplicate := seen[target.LessonID]; duplicate {
				return core.E(core.CodeInvalidInput, "Lesson ids must be unique", nil)
			}
			seen[target.LessonID] = struct{}{}
			var plan lessonReplacementPlan
			plan.lessonID = target.LessonID
			err := tx.QueryRow(ctx, `
				SELECT teacher_account_id, starts_at, duration_minutes, status, version
				FROM lessons
				WHERE tenant_id = $1 AND id = $2
				FOR UPDATE
			`, command.TenantID, target.LessonID).Scan(&plan.previousTeacherID, &plan.startsAt, &plan.durationMinutes, &plan.status, &plan.version)
			if errors.Is(err, pgx.ErrNoRows) {
				return core.E(core.CodeNotFound, "Lesson not found", nil)
			}
			if err != nil {
				return fmt.Errorf("lock Lesson Teacher replacement: %w", err)
			}
			if plan.status != string(core.LessonScheduled) || !plan.startsAt.After(command.Now) {
				return core.E(core.CodeInvalidState, "only scheduled future Lessons can change Teacher", nil)
			}
			if plan.version != target.ExpectedVersion {
				return core.E(core.CodeConflict, "Lesson version is stale", nil)
			}
			if plan.previousTeacherID != target.ExpectedPreviousTeacherAccountID {
				return core.E(core.CodeConflict, "Lesson previous Teacher is stale", nil)
			}
			if plan.previousTeacherID == command.NewTeacherAccountID {
				return core.E(core.CodeInvalidState, "new Teacher is already assigned to Lesson", nil)
			}
			plans = append(plans, plan)
		}
		targetIDs := make([]string, len(plans))
		for index, plan := range plans {
			targetIDs[index] = plan.lessonID
		}
		for index, plan := range plans {
			conflict, err := lessonScheduleConflict(ctx, tx, command.TenantID, plan.startsAt, plan.durationMinutes, command.NewTeacherAccountID, nil, targetIDs)
			if err != nil {
				return err
			}
			if conflict {
				return core.E(core.CodeConflict, "new Teacher has an overlapping Lesson", nil)
			}
			for otherIndex := index + 1; otherIndex < len(plans); otherIndex++ {
				if lessonIntervalsOverlap(plan.startsAt, plan.durationMinutes, plans[otherIndex].startsAt, plans[otherIndex].durationMinutes) {
					return core.E(core.CodeConflict, "selected Lessons overlap for the new Teacher", nil)
				}
			}
		}
		result = core.LessonTeacherReplacementResult{UpdatedCount: len(plans), Lessons: make([]core.Lesson, 0, len(plans))}
		for _, plan := range plans {
			if _, err := tx.Exec(ctx, `
				UPDATE lessons
				SET teacher_account_id = $3, version = version + 1, updated_at = $4
				WHERE tenant_id = $1 AND id = $2
			`, command.TenantID, plan.lessonID, command.NewTeacherAccountID, command.Now); err != nil {
				return fmt.Errorf("replace Lesson Teacher: %w", err)
			}
			updated, err := readLesson(ctx, tx, command.TenantID, plan.lessonID)
			if err != nil {
				return err
			}
			result.Lessons = append(result.Lessons, updated)
			metadata := map[string]any{
				"previousTeacherAccountId": plan.previousTeacherID,
				"newTeacherAccountId":      command.NewTeacherAccountID,
				"version":                  updated.Version,
			}
			if err := appendAudit(ctx, tx, auditInput{
				tenantID: command.TenantID, actorID: command.ActorAccountID,
				action: "LessonTeacherReplaced", targetType: "lesson", targetID: plan.lessonID,
				decision: "allow", reason: "temporary_teacher_continuity", idempotencyKey: command.IdempotencyKey,
				metadata: metadata, at: command.Now,
			}); err != nil {
				return err
			}
			if err := appendOutbox(ctx, tx, command.TenantID, "LessonTeacherReplaced", "lesson", plan.lessonID, metadata, command.Now); err != nil {
				return err
			}
		}
		return completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "replace_lesson_teachers", command.IdempotencyKey, result, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "LessonTeacherReplaced", targetType: "lesson_batch", targetID: "lessons",
			reason: "lesson_teacher_replace_not_allowed", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return result, err
}

func (s *Store) ReassignPrimaryTeachers(ctx context.Context, command core.ReassignPrimaryTeachersCommand) (core.PrimaryTeacherReassignmentResult, error) {
	var result core.PrimaryTeacherReassignmentResult
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, err := lessonManagementAuthority(ctx, tx, command.TenantID, command.ActorAccountID)
		if err != nil {
			return err
		}
		if !manager {
			return core.E(core.CodeForbidden, "primary Teacher reassignment permission is required", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "reassign_primary_teachers", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.PrimaryTeacherReassignmentResult](claim)
			return err
		}
		effectiveFrom := command.EffectiveFrom
		switch command.EffectiveMode {
		case core.PrimaryTeacherEffectiveImmediate:
			effectiveFrom = command.Now
		case core.PrimaryTeacherEffectiveScheduled:
			if !effectiveFrom.After(command.Now) {
				return core.E(core.CodeInvalidInput, "scheduled effectiveFrom must be in the future", nil)
			}
		default:
			return core.E(core.CodeInvalidInput, "effectiveMode must be immediate or scheduled", nil)
		}
		activeTeacher, err := hasActiveRole(ctx, tx, command.TenantID, command.NewTeacherAccountID, core.RoleTeacher)
		if err != nil {
			return err
		}
		if !activeTeacher {
			return core.E(core.CodeInvalidInput, "new Teacher is not active in this school", nil)
		}
		if len(command.Targets) == 0 {
			return core.E(core.CodeInvalidInput, "at least one Student is required", nil)
		}
		targets := append([]core.PrimaryTeacherReassignmentTarget(nil), command.Targets...)
		sort.Slice(targets, func(left, right int) bool { return targets[left].StudentID < targets[right].StudentID })
		studentIDs := make([]string, len(targets))
		for index, target := range targets {
			if index > 0 && target.StudentID == targets[index-1].StudentID {
				return core.E(core.CodeInvalidInput, "Student ids must be unique", nil)
			}
			studentIDs[index] = target.StudentID
		}
		if err := lockAssignmentSubjects(ctx, tx, command.TenantID, studentIDs); err != nil {
			return err
		}
		plans := make([]primaryTeacherPlan, 0, len(targets))
		for _, target := range targets {
			var eligible bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1
					FROM students s
					JOIN accounts a ON a.tenant_id = s.tenant_id AND a.id = s.account_id
					JOIN role_grants rg
					  ON rg.tenant_id = s.tenant_id AND rg.account_id = s.account_id
					 AND rg.role_type = 'Student' AND rg.scope_type = 'student'
					 AND rg.scope_id = s.id AND rg.status = 'active'
					WHERE s.tenant_id = $1 AND s.id = $2 AND s.status = 'active'
					  AND a.status IN ('pending_activation', 'active')
				)
			`, command.TenantID, target.StudentID).Scan(&eligible); err != nil {
				return fmt.Errorf("validate reassignment Student: %w", err)
			}
			if !eligible {
				return core.E(core.CodeInvalidInput, "Student is not active in this school", nil)
			}
			var plan primaryTeacherPlan
			plan.target = target
			err := tx.QueryRow(ctx, `
				SELECT id, teacher_account_id, effective_from, version
				FROM teacher_assignments
				WHERE tenant_id = $1 AND student_id = $2 AND status = 'active'
				  AND effective_from <= $3
				  AND (effective_until IS NULL OR $3 < effective_until)
				ORDER BY effective_from DESC, id DESC
				LIMIT 1
				FOR UPDATE
			`, command.TenantID, target.StudentID, effectiveFrom).Scan(
				&plan.previousAssignmentID,
				&plan.previousTeacherID,
				&plan.previousEffectiveFrom,
				&plan.previousVersion,
			)
			if errors.Is(err, pgx.ErrNoRows) {
				return core.E(core.CodeInvalidState, "Student has no primary Teacher at effectiveFrom", nil)
			}
			if err != nil {
				return fmt.Errorf("lock primary Teacher assignment: %w", err)
			}
			if err := tx.QueryRow(ctx, `
				SELECT COALESCE(max(version), -1)
				FROM teacher_assignments
				WHERE tenant_id = $1 AND student_id = $2
			`, command.TenantID, target.StudentID).Scan(&plan.timelineVersion); err != nil {
				return fmt.Errorf("read primary Teacher assignment timeline version: %w", err)
			}
			if plan.timelineVersion != target.ExpectedAssignmentVersion {
				return core.E(core.CodeConflict, "primary Teacher assignment version is stale", nil)
			}
			if plan.previousTeacherID == command.NewTeacherAccountID {
				return core.E(core.CodeInvalidState, "new Teacher is already the Student primary Teacher", nil)
			}
			plan.newVersion = plan.timelineVersion + 1
			plans = append(plans, plan)
		}
		result = core.PrimaryTeacherReassignmentResult{ReassignedCount: len(plans), Assignments: make([]core.PrimaryTeacherReassignment, 0, len(plans))}
		for _, plan := range plans {
			if _, err := tx.Exec(ctx, `
				UPDATE teacher_assignments
				SET status = 'ended', ended_at = $4, effective_until = effective_from
				WHERE tenant_id = $1 AND student_id = $2 AND status = 'active'
				  AND effective_from >= $3
			`, command.TenantID, plan.target.StudentID, effectiveFrom, command.Now); err != nil {
				return fmt.Errorf("supersede future primary Teacher assignments: %w", err)
			}
			if plan.previousEffectiveFrom.Before(effectiveFrom) {
				if _, err := tx.Exec(ctx, `
					UPDATE teacher_assignments
					SET effective_until = $3
					WHERE tenant_id = $1 AND id = $2 AND status = 'active'
				`, command.TenantID, plan.previousAssignmentID, effectiveFrom); err != nil {
					return fmt.Errorf("close previous primary Teacher assignment interval: %w", err)
				}
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO teacher_assignments (
					id, tenant_id, student_id, teacher_account_id, status,
					assigned_by_account_id, assigned_at, effective_from, version
				) VALUES ($1, $2, $3, $4, 'active', $5, $6, $7, $8)
			`, plan.target.AssignmentID, command.TenantID, plan.target.StudentID,
				command.NewTeacherAccountID, command.ActorAccountID, command.Now,
				effectiveFrom, plan.newVersion); err != nil {
				return mapSchedulingWriteError(err, "primary Teacher assignment conflicts with another change")
			}
			assignment := core.PrimaryTeacherReassignment{
				StudentID: plan.target.StudentID, PreviousTeacherAccountID: plan.previousTeacherID,
				NewTeacherAccountID: command.NewTeacherAccountID, EffectiveFrom: effectiveFrom,
				Version: plan.newVersion,
			}
			result.Assignments = append(result.Assignments, assignment)
			metadata := map[string]any{
				"previousTeacherAccountId": plan.previousTeacherID,
				"newTeacherAccountId":      command.NewTeacherAccountID,
				"effectiveFrom":            effectiveFrom,
				"version":                  plan.newVersion,
			}
			if err := appendAudit(ctx, tx, auditInput{
				tenantID: command.TenantID, actorID: command.ActorAccountID,
				action: "StudentPrimaryTeacherReassigned", targetType: "student", targetID: plan.target.StudentID,
				decision: "allow", reason: "primary_teacher_continuity", idempotencyKey: command.IdempotencyKey,
				metadata: metadata, at: command.Now,
			}); err != nil {
				return err
			}
			if err := appendOutbox(ctx, tx, command.TenantID, "StudentPrimaryTeacherReassigned", "student", plan.target.StudentID, metadata, command.Now); err != nil {
				return err
			}
		}
		return completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "reassign_primary_teachers", command.IdempotencyKey, result, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "StudentPrimaryTeacherReassigned", targetType: "student_batch", targetID: "students",
			reason: "primary_teacher_reassign_not_allowed", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return result, err
}

type lessonReplacementPlan struct {
	lessonID          string
	previousTeacherID string
	startsAt          time.Time
	durationMinutes   int
	status            string
	version           int64
}

type primaryTeacherPlan struct {
	target                core.PrimaryTeacherReassignmentTarget
	previousAssignmentID  string
	previousTeacherID     string
	previousEffectiveFrom time.Time
	previousVersion       int64
	timelineVersion       int64
	newVersion            int64
}

type lessonReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func readLesson(ctx context.Context, reader lessonReader, tenantID, lessonID string) (core.Lesson, error) {
	var result core.Lesson
	var status string
	var location *string
	err := reader.QueryRow(ctx, `
		SELECT l.id, l.title, l.starts_at, l.duration_minutes, l.location,
		       l.teacher_account_id, teacher_person.full_name, l.status, l.version
		FROM lessons l
		JOIN accounts teacher_account
		  ON teacher_account.tenant_id = l.tenant_id AND teacher_account.id = l.teacher_account_id
		JOIN people teacher_person
		  ON teacher_person.tenant_id = teacher_account.tenant_id AND teacher_person.id = teacher_account.person_id
		WHERE l.tenant_id = $1 AND l.id = $2
	`, tenantID, lessonID).Scan(
		&result.ID,
		&result.Title,
		&result.StartsAt,
		&result.DurationMinutes,
		&location,
		&result.Teacher.AccountID,
		&result.Teacher.FullName,
		&status,
		&result.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.Lesson{}, core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	if err != nil {
		return core.Lesson{}, fmt.Errorf("read Lesson: %w", err)
	}
	if location != nil {
		result.Location = *location
	}
	result.Status = core.LessonStatus(status)
	result.Students = make([]core.LessonStudent, 0)
	rows, err := reader.Query(ctx, `
		SELECT s.id, p.full_name
		FROM lesson_participants lp
		JOIN students s ON s.tenant_id = lp.tenant_id AND s.id = lp.student_id
		JOIN people p ON p.tenant_id = s.tenant_id AND p.id = s.person_id
		WHERE lp.tenant_id = $1 AND lp.lesson_id = $2
		ORDER BY p.full_name, s.id
	`, tenantID, lessonID)
	if err != nil {
		return core.Lesson{}, fmt.Errorf("read Lesson Students: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var student core.LessonStudent
		if err := rows.Scan(&student.StudentID, &student.FullName); err != nil {
			return core.Lesson{}, fmt.Errorf("scan Lesson Student: %w", err)
		}
		result.Students = append(result.Students, student)
	}
	if err := rows.Err(); err != nil {
		return core.Lesson{}, fmt.Errorf("iterate Lesson Students: %w", err)
	}
	return result, nil
}

type lessonScope struct {
	manager   bool
	teacherID string
	studentID string
}

func (s *Store) lessonReadScope(ctx context.Context, principal core.Principal) (lessonScope, error) {
	roles, err := rolesForAccount(ctx, s.pool, principal.TenantID, principal.AccountID)
	if err != nil {
		return lessonScope{}, err
	}
	scope := lessonScope{
		manager: principalHasRole(roles, core.RoleOwner) || principalHasRole(roles, core.RoleAdministrator),
	}
	if principalHasRole(roles, core.RoleTeacher) {
		scope.teacherID = principal.AccountID
	}
	if principalHasRole(roles, core.RoleStudent) {
		err := s.pool.QueryRow(ctx, `
			SELECT scope_id
			FROM role_grants
			WHERE tenant_id = $1 AND account_id = $2
			  AND role_type = 'Student' AND scope_type = 'student' AND status = 'active'
			ORDER BY granted_at DESC, id DESC
			LIMIT 1
		`, principal.TenantID, principal.AccountID).Scan(&scope.studentID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return lessonScope{}, fmt.Errorf("read Student Lesson scope: %w", err)
		}
	}
	if !scope.manager && scope.teacherID == "" && scope.studentID == "" {
		return lessonScope{}, core.E(core.CodeForbidden, "Lesson read permission is required", nil)
	}
	return scope, nil
}

func lessonVisible(scope lessonScope, lesson core.Lesson) bool {
	if scope.manager || (scope.teacherID != "" && lesson.Teacher.AccountID == scope.teacherID) {
		return true
	}
	for _, student := range lesson.Students {
		if student.StudentID == scope.studentID {
			return true
		}
	}
	return false
}

func lessonCreateAuthority(ctx context.Context, tx pgx.Tx, tenantID, actorID string) (bool, bool, error) {
	manager, err := lessonManagementAuthority(ctx, tx, tenantID, actorID)
	if err != nil {
		return false, false, err
	}
	teacher, err := hasActiveRole(ctx, tx, tenantID, actorID, core.RoleTeacher)
	return manager, teacher, err
}

func lessonManagementAuthority(ctx context.Context, tx pgx.Tx, tenantID, actorID string) (bool, error) {
	owner, err := hasActiveRole(ctx, tx, tenantID, actorID, core.RoleOwner)
	if err != nil || owner {
		return owner, err
	}
	return hasActiveRole(ctx, tx, tenantID, actorID, core.RoleAdministrator)
}

func lockAssignmentSubjects(ctx context.Context, tx pgx.Tx, tenantID string, studentIDs []string) error {
	for _, studentID := range studentIDs {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, advisoryLockKey("primary-teacher-assignment", tenantID, studentID)); err != nil {
			return fmt.Errorf("lock primary Teacher assignment: %w", err)
		}
	}
	return nil
}

func lockLessonScheduleSubjects(ctx context.Context, tx pgx.Tx, tenantID, teacherAccountID string, studentIDs []string) error {
	subjects := make([]string, 0, len(studentIDs)+1)
	subjects = append(subjects, "teacher:"+teacherAccountID)
	for _, studentID := range studentIDs {
		subjects = append(subjects, "student:"+studentID)
	}
	sort.Strings(subjects)
	for _, subject := range subjects {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, advisoryLockKey("lesson-schedule", tenantID, subject)); err != nil {
			return fmt.Errorf("lock Lesson schedule subject: %w", err)
		}
	}
	return nil
}

func lessonScheduleConflict(ctx context.Context, tx pgx.Tx, tenantID string, startsAt time.Time, durationMinutes int, teacherAccountID string, studentIDs, excludedLessonIDs []string) (bool, error) {
	if studentIDs == nil {
		studentIDs = []string{}
	}
	if excludedLessonIDs == nil {
		excludedLessonIDs = []string{}
	}
	var conflict bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM lessons l
			WHERE l.tenant_id = $1 AND l.status = 'scheduled'
			  AND l.starts_at < $3
			  AND l.starts_at + l.duration_minutes * interval '1 minute' > $2
			  AND NOT (l.id = ANY($6::text[]))
			  AND (
				l.teacher_account_id = $4
				OR EXISTS (
					SELECT 1
					FROM lesson_participants lp
					WHERE lp.tenant_id = l.tenant_id AND lp.lesson_id = l.id
					  AND lp.student_id = ANY($5::text[])
				)
			  )
		)
	`, tenantID, startsAt, startsAt.Add(time.Duration(durationMinutes)*time.Minute), teacherAccountID, studentIDs, excludedLessonIDs).Scan(&conflict)
	if err != nil {
		return false, fmt.Errorf("check overlapping Lesson: %w", err)
	}
	return conflict, nil
}

func lessonIntervalsOverlap(firstStart time.Time, firstDuration int, secondStart time.Time, secondDuration int) bool {
	return firstStart.Before(secondStart.Add(time.Duration(secondDuration)*time.Minute)) &&
		secondStart.Before(firstStart.Add(time.Duration(firstDuration)*time.Minute))
}

func uniqueSorted(values []string, message string) ([]string, error) {
	result := append([]string(nil), values...)
	sort.Strings(result)
	for index := 1; index < len(result); index++ {
		if result[index] == result[index-1] {
			return nil, core.E(core.CodeInvalidInput, message, nil)
		}
	}
	return result, nil
}

func mapSchedulingWriteError(err error, conflictMessage string) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && (postgresError.Code == "23505" || postgresError.Code == "23P01") {
		return core.E(core.CodeConflict, conflictMessage, err)
	}
	return err
}
