package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 rooms and core lesson series (DEC-002/DEC-004). Series carry the
// weekly recurrence anchor; occurrence generation copies the series
// snapshot (teacher, room, enrollment) into concrete occurrences and is
// idempotent per (series, start time) under the series row lock.

func (s *Store) CreateRoom(ctx context.Context, command core.CreateRoomCommand) (core.Room, error) {
	principal := command.Principal
	var room core.Room
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, err := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if err != nil {
			return err
		}
		if !manager {
			return core.E(core.CodeForbidden, "room management permission is required", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO rooms (id, tenant_id, name, capacity, status, version, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'active', 0, $5, $5)
		`, command.RoomID, principal.TenantID, command.Name, command.Capacity, command.Now); err != nil {
			return mapWriteError(err, "room name is already in use")
		}
		room = core.Room{ID: command.RoomID, Name: command.Name, Capacity: command.Capacity, Status: "active", Version: 0}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "RoomCreated", targetType: "room", targetID: command.RoomID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.Room{}, err
	}
	return room, nil
}

func (s *Store) ListRooms(ctx context.Context, principal core.Principal) ([]core.Room, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, capacity, status, version
		FROM rooms
		WHERE tenant_id = $1 AND status = 'active'
		ORDER BY name, id
	`, principal.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()
	rooms := []core.Room{}
	for rows.Next() {
		var room core.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Capacity, &room.Status, &room.Version); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}
	return rooms, nil
}

func (s *Store) CreateCoreLessonSeries(ctx context.Context, command core.CreateCoreLessonSeriesCommand) (core.CoreLessonSeries, error) {
	var result core.CoreLessonSeries
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, teacher, err := lessonCreateAuthority(ctx, tx, command.TenantID, command.ActorAccountID)
		if err != nil {
			return err
		}
		if !manager && !teacher {
			return core.E(core.CodeForbidden, "Lesson scheduling permission is required", nil)
		}
		if teacher && !manager && command.TeacherAccountID != command.ActorAccountID {
			return core.E(core.CodeForbidden, "Teacher can only create series for self", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_lesson_series", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.CoreLessonSeries](claim)
			return err
		}
		activeTeacher, err := hasActiveRole(ctx, tx, command.TenantID, command.TeacherAccountID, core.RoleTeacher)
		if err != nil {
			return err
		}
		if !activeTeacher {
			return core.E(core.CodeInvalidInput, "Teacher is not active in this school", nil)
		}
		studentIDs, err := uniqueSorted(command.StudentIDs, "Student ids must be unique")
		if err != nil {
			return err
		}
		if command.RoomID != "" {
			var roomActive bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (SELECT 1 FROM rooms WHERE tenant_id = $1 AND id = $2 AND status = 'active')
			`, command.TenantID, command.RoomID).Scan(&roomActive); err != nil {
				return fmt.Errorf("check room: %w", err)
			}
			if !roomActive {
				return core.E(core.CodeInvalidInput, "room is not active in this school", nil)
			}
		}
		var activeStudents int
		if err := tx.QueryRow(ctx, `
			SELECT count(*) FROM students
			WHERE tenant_id = $1 AND id = ANY($2::text[]) AND status = 'active'
		`, command.TenantID, studentIDs).Scan(&activeStudents); err != nil {
			return fmt.Errorf("validate series students: %w", err)
		}
		if activeStudents != len(studentIDs) {
			return core.E(core.CodeInvalidInput, "Student is not active in this school", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO core_lesson_series (
				id, tenant_id, format, title, teacher_account_id, room_id,
				weekday, start_minutes, duration_minutes, effective_from, effective_until,
				status, version, created_by_account_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10::date, NULLIF($11, '')::date,
			          'active', 0, $12, $13, $13)
		`, command.SeriesID, command.TenantID, command.Format, command.Title,
			command.TeacherAccountID, command.RoomID, command.Weekday, command.StartMinutes,
			command.DurationMinutes, command.EffectiveFrom, command.EffectiveUntil,
			command.ActorAccountID, command.Now); err != nil {
			return mapWriteError(err, "series conflicts with existing data")
		}
		for _, studentID := range studentIDs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO core_lesson_series_enrollments (tenant_id, series_id, student_id, added_at)
				VALUES ($1, $2, $3, $4)
			`, command.TenantID, command.SeriesID, studentID, command.Now); err != nil {
				return mapWriteError(err, "series enrollment conflicts with existing data")
			}
		}
		result, err = readCoreLessonSeries(ctx, tx, command.TenantID, command.SeriesID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_lesson_series", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "LessonSeriesCreated", targetType: "lesson_series", targetID: command.SeriesID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"teacherAccountId": command.TeacherAccountID, "format": command.Format},
			at:       command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "LessonSeriesCreated", "lesson_series", command.SeriesID,
			map[string]any{"seriesId": command.SeriesID, "teacherAccountId": command.TeacherAccountID}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "LessonSeriesCreated", targetType: "lesson_series", targetID: command.SeriesID,
			reason: "lesson_create_not_allowed", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	if err != nil {
		return core.CoreLessonSeries{}, err
	}
	return result, nil
}

func (s *Store) ListCoreLessonSeries(ctx context.Context, principal core.Principal) ([]core.CoreLessonSeries, error) {
	scope, err := s.lessonReadScope(ctx, principal)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT cls.id
		FROM core_lesson_series cls
		WHERE cls.tenant_id = $1
		  AND (
			$2::boolean
			OR ($3 <> '' AND cls.teacher_account_id = $3)
			OR ($4 <> '' AND EXISTS (
				SELECT 1 FROM core_lesson_series_enrollments enrollment
				WHERE enrollment.tenant_id = cls.tenant_id
				  AND enrollment.series_id = cls.id
				  AND enrollment.student_id = $4
			))
		  )
		ORDER BY cls.weekday, cls.start_minutes, cls.id
	`, principal.TenantID, scope.manager, scope.teacherID, scope.studentID)
	if err != nil {
		return nil, fmt.Errorf("list lesson series: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan series id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate series ids: %w", err)
	}
	rows.Close()
	result := make([]core.CoreLessonSeries, 0, len(ids))
	for _, id := range ids {
		series, err := readCoreLessonSeries(ctx, s.pool, principal.TenantID, id)
		if err != nil {
			return nil, err
		}
		result = append(result, series)
	}
	return result, nil
}

func (s *Store) GetCoreLessonSeries(ctx context.Context, principal core.Principal, seriesID string) (core.CoreLessonSeries, error) {
	scope, err := s.lessonReadScope(ctx, principal)
	if err != nil {
		return core.CoreLessonSeries{}, err
	}
	series, err := readCoreLessonSeries(ctx, s.pool, principal.TenantID, seriesID)
	if err != nil {
		return core.CoreLessonSeries{}, err
	}
	visible := scope.manager || (scope.teacherID != "" && series.Teacher.AccountID == scope.teacherID)
	if !visible && scope.studentID != "" {
		for _, student := range series.Students {
			if student.StudentID == scope.studentID {
				visible = true
			}
		}
	}
	if !visible {
		return core.CoreLessonSeries{}, core.E(core.CodeNotFound, "lesson series not found", nil)
	}
	return series, nil
}

func (s *Store) GenerateSeriesOccurrences(ctx context.Context, command core.GenerateSeriesOccurrencesCommand) (core.SeriesOccurrenceGenerationResult, error) {
	var result core.SeriesOccurrenceGenerationResult
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, teacher, err := lessonCreateAuthority(ctx, tx, command.TenantID, command.ActorAccountID)
		if err != nil {
			return err
		}
		if !manager && !teacher {
			return core.E(core.CodeForbidden, "Lesson scheduling permission is required", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "generate_series_occurrences", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.SeriesOccurrenceGenerationResult](claim)
			return err
		}
		var seriesFormat, seriesTitle, seriesTeacher string
		var seriesRoom *string
		var seriesStatus string
		var durationMinutes int
		err = tx.QueryRow(ctx, `
			SELECT format, title, teacher_account_id, room_id, duration_minutes, status
			FROM core_lesson_series
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, command.TenantID, command.SeriesID).Scan(
			&seriesFormat, &seriesTitle, &seriesTeacher, &seriesRoom, &durationMinutes, &seriesStatus)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "lesson series not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock lesson series: %w", err)
		}
		if seriesStatus != "active" {
			return core.E(core.CodeInvalidState, "only an active series can generate occurrences", nil)
		}
		if teacher && !manager && seriesTeacher != command.ActorAccountID {
			return core.E(core.CodeForbidden, "Teacher can only generate for own series", nil)
		}
		studentRows, err := tx.Query(ctx, `
			SELECT student_id FROM core_lesson_series_enrollments
			WHERE tenant_id = $1 AND series_id = $2
			ORDER BY student_id
		`, command.TenantID, command.SeriesID)
		if err != nil {
			return fmt.Errorf("read series enrollment: %w", err)
		}
		studentIDs := []string{}
		for studentRows.Next() {
			var studentID string
			if err := studentRows.Scan(&studentID); err != nil {
				studentRows.Close()
				return fmt.Errorf("scan series enrollment: %w", err)
			}
			studentIDs = append(studentIDs, studentID)
		}
		studentRows.Close()
		if err := studentRows.Err(); err != nil {
			return fmt.Errorf("iterate series enrollment: %w", err)
		}
		created := []string{}
		for _, planned := range command.Occurrences {
			var exists bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM core_lesson_occurrences
					WHERE tenant_id = $1 AND series_id = $2 AND starts_at = $3
				)
			`, command.TenantID, command.SeriesID, planned.StartsAt).Scan(&exists); err != nil {
				return fmt.Errorf("check generated occurrence: %w", err)
			}
			if exists {
				continue
			}
			conflict, err := lessonScheduleConflict(ctx, tx, command.TenantID, planned.StartsAt, durationMinutes, seriesTeacher, studentIDs, nil)
			if err != nil {
				return err
			}
			if conflict {
				return core.E(core.CodeConflict, "generated occurrence overlaps an existing Lesson", nil)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO core_lesson_occurrences (
					id, tenant_id, series_id, format, title, starts_at, duration_minutes,
					teacher_account_id, room_id, status, version, created_by_account_id, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'scheduled', 0, $10, $11, $11)
			`, planned.OccurrenceID, command.TenantID, command.SeriesID, seriesFormat, seriesTitle,
				planned.StartsAt, durationMinutes, seriesTeacher, seriesRoom,
				command.ActorAccountID, command.Now); err != nil {
				return mapWriteError(err, "occurrence conflicts with existing data")
			}
			for _, studentID := range studentIDs {
				if _, err := tx.Exec(ctx, `
					INSERT INTO core_lesson_occurrence_participants (tenant_id, occurrence_id, student_id, added_at)
					VALUES ($1, $2, $3, $4)
				`, command.TenantID, planned.OccurrenceID, studentID, command.Now); err != nil {
					return mapWriteError(err, "occurrence participant conflicts with existing data")
				}
			}
			created = append(created, planned.OccurrenceID)
		}
		result = core.SeriesOccurrenceGenerationResult{
			SeriesID: command.SeriesID, CreatedCount: len(created), OccurrenceIDs: created,
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "generate_series_occurrences", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "SeriesOccurrencesGenerated", targetType: "lesson_series", targetID: command.SeriesID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"createdCount": len(created)}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "SeriesOccurrencesGenerated", "lesson_series", command.SeriesID,
			map[string]any{"seriesId": command.SeriesID, "createdCount": len(created)}, command.Now)
	})
	if err != nil {
		return core.SeriesOccurrenceGenerationResult{}, err
	}
	return result, nil
}

func readCoreLessonSeries(ctx context.Context, reader lessonReader, tenantID, seriesID string) (core.CoreLessonSeries, error) {
	var series core.CoreLessonSeries
	var roomID *string
	var effectiveUntil *string
	err := reader.QueryRow(ctx, `
		SELECT cls.id, cls.format, cls.title, cls.teacher_account_id, teacher_person.full_name,
		       cls.room_id, cls.weekday, cls.start_minutes, cls.duration_minutes,
		       to_char(cls.effective_from, 'YYYY-MM-DD'),
		       to_char(cls.effective_until, 'YYYY-MM-DD'),
		       cls.status, cls.version
		FROM core_lesson_series cls
		JOIN accounts teacher_account
		  ON teacher_account.tenant_id = cls.tenant_id AND teacher_account.id = cls.teacher_account_id
		JOIN people teacher_person
		  ON teacher_person.tenant_id = teacher_account.tenant_id AND teacher_person.id = teacher_account.person_id
		WHERE cls.tenant_id = $1 AND cls.id = $2
	`, tenantID, seriesID).Scan(
		&series.ID, &series.Format, &series.Title,
		&series.Teacher.AccountID, &series.Teacher.FullName,
		&roomID, &series.Weekday, &series.StartMinutes, &series.DurationMinutes,
		&series.EffectiveFrom, &effectiveUntil, &series.Status, &series.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.CoreLessonSeries{}, core.E(core.CodeNotFound, "lesson series not found", nil)
	}
	if err != nil {
		return core.CoreLessonSeries{}, fmt.Errorf("read lesson series: %w", err)
	}
	if roomID != nil {
		series.RoomID = *roomID
	}
	if effectiveUntil != nil {
		series.EffectiveUntil = *effectiveUntil
	}
	series.Students = make([]core.LessonStudent, 0)
	rows, err := reader.Query(ctx, `
		SELECT s.id, p.full_name
		FROM core_lesson_series_enrollments enrollment
		JOIN students s ON s.tenant_id = enrollment.tenant_id AND s.id = enrollment.student_id
		JOIN people p ON p.tenant_id = s.tenant_id AND p.id = s.person_id
		WHERE enrollment.tenant_id = $1 AND enrollment.series_id = $2
		ORDER BY p.full_name, s.id
	`, tenantID, seriesID)
	if err != nil {
		return core.CoreLessonSeries{}, fmt.Errorf("read series students: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var student core.LessonStudent
		if err := rows.Scan(&student.StudentID, &student.FullName); err != nil {
			return core.CoreLessonSeries{}, fmt.Errorf("scan series student: %w", err)
		}
		series.Students = append(series.Students, student)
	}
	if err := rows.Err(); err != nil {
		return core.CoreLessonSeries{}, fmt.Errorf("iterate series students: %w", err)
	}
	return series, nil
}
