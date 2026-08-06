package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 student repertoire (aggregate StudentSong). The student's active
// assigned Teacher or an Administrator moves a song along the journey;
// stage history is append-only (DB trigger). The stage is a named step,
// never a computed readiness (SongReadiness is a separate future
// aggregate) and never a score (DEC-006).

func repertoireMarkerAuthority(ctx context.Context, tx pgx.Tx, tenantID, actorID, studentID string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM students WHERE tenant_id = $1 AND id = $2
		)
	`, tenantID, studentID).Scan(&exists); err != nil {
		return fmt.Errorf("check repertoire student: %w", err)
	}
	if !exists {
		return core.E(core.CodeNotFound, "Student not found", nil)
	}
	var assigned bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM teacher_assignments
			WHERE tenant_id = $1 AND student_id = $2 AND teacher_account_id = $3
			  AND status = 'active'
		)
	`, tenantID, studentID, actorID).Scan(&assigned); err != nil {
		return fmt.Errorf("check repertoire teacher: %w", err)
	}
	if assigned {
		return nil
	}
	admin, err := hasActiveRole(ctx, tx, tenantID, actorID, core.RoleAdministrator)
	if err != nil {
		return err
	}
	if !admin {
		return core.E(core.CodeForbidden, "repertoire is led by the Student's Teacher or an Administrator", nil)
	}
	return nil
}

func readStudentSong(ctx context.Context, reader lessonReader, tenantID, songID string) (core.StudentSong, error) {
	var song core.StudentSong
	var artist, stageNote *string
	err := reader.QueryRow(ctx, `
		SELECT s.id, s.student_id, s.title, s.artist, s.stage, s.stage_note,
		       s.assigned_by_account_id, person.full_name,
		       s.version, s.created_at, s.updated_at
		FROM student_songs s
		JOIN accounts account ON account.tenant_id = s.tenant_id AND account.id = s.assigned_by_account_id
		JOIN people person ON person.tenant_id = account.tenant_id AND person.id = account.person_id
		WHERE s.tenant_id = $1 AND s.id = $2
	`, tenantID, songID).Scan(
		&song.ID, &song.StudentID, &song.Title, &artist, &song.Stage, &stageNote,
		&song.AssignedBy.AccountID, &song.AssignedBy.FullName,
		&song.Version, &song.CreatedAt, &song.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.StudentSong{}, core.E(core.CodeNotFound, "song not found", nil)
	}
	if err != nil {
		return core.StudentSong{}, fmt.Errorf("read student song: %w", err)
	}
	if artist != nil {
		song.Artist = *artist
	}
	if stageNote != nil {
		song.StageNote = *stageNote
	}
	song.CreatedAt = song.CreatedAt.UTC()
	song.UpdatedAt = song.UpdatedAt.UTC()
	song.History = make([]core.SongStageChange, 0)
	rows, err := reader.Query(ctx, `
		SELECT COALESCE(from_stage, ''), to_stage, COALESCE(note, ''), changed_at
		FROM student_song_stage_history
		WHERE tenant_id = $1 AND song_id = $2
		ORDER BY seq DESC
	`, tenantID, songID)
	if err != nil {
		return core.StudentSong{}, fmt.Errorf("read song history: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var change core.SongStageChange
		if err := rows.Scan(&change.FromStage, &change.ToStage, &change.Note, &change.ChangedAt); err != nil {
			return core.StudentSong{}, fmt.Errorf("scan song history: %w", err)
		}
		change.ChangedAt = change.ChangedAt.UTC()
		song.History = append(song.History, change)
	}
	if err := rows.Err(); err != nil {
		return core.StudentSong{}, fmt.Errorf("iterate song history: %w", err)
	}
	return song, nil
}

func (s *Store) AddStudentSong(ctx context.Context, command core.AddStudentSongCommand) (core.StudentSong, error) {
	principal := command.Principal
	var song core.StudentSong
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := repertoireMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "add_student_song", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			song, err = decodeReplay[core.StudentSong](claim)
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO student_songs (
				id, tenant_id, student_id, title, artist, stage, stage_note,
				assigned_by_account_id, version, created_at, updated_at
			) VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, NULLIF($7, ''), $8, 1, $9, $9)
		`, command.SongID, principal.TenantID, command.StudentID, command.Title,
			command.Artist, command.Stage, command.StageNote, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "song conflicts with existing data")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO student_song_stage_history (
				tenant_id, song_id, seq, changed_at, from_stage, to_stage, note, changed_by_account_id
			) VALUES ($1, $2, 1, $3, NULL, $4, NULLIF($5, ''), $6)
		`, principal.TenantID, command.SongID, command.Now, command.Stage,
			command.StageNote, principal.AccountID); err != nil {
			return mapWriteError(err, "song history conflicts with existing data")
		}
		song, err = readStudentSong(ctx, tx, principal.TenantID, command.SongID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "add_student_song", command.IdempotencyKey, song, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "SongAdded", targetType: "student_song", targetID: command.SongID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"studentId": command.StudentID, "stage": command.Stage},
			at:       command.Now,
		})
	})
	if err != nil {
		return core.StudentSong{}, err
	}
	return song, nil
}

func (s *Store) ChangeSongStage(ctx context.Context, command core.ChangeSongStageCommand) (core.StudentSong, error) {
	principal := command.Principal
	var song core.StudentSong
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "change_song_stage", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			song, err = decodeReplay[core.StudentSong](claim)
			return err
		}
		var studentID, currentStage string
		var version int
		err = tx.QueryRow(ctx, `
			SELECT student_id, stage, version FROM student_songs
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.SongID).Scan(&studentID, &currentStage, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "song not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock student song: %w", err)
		}
		if err := repertoireMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, studentID); err != nil {
			return err
		}
		if command.ExpectedVersion != version {
			return core.E(core.CodeConflict, "the song changed; reload and retry", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO student_song_stage_history (
				tenant_id, song_id, seq, changed_at, from_stage, to_stage, note, changed_by_account_id
			) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8)
		`, principal.TenantID, command.SongID, version+1, command.Now, currentStage,
			command.Stage, command.StageNote, principal.AccountID); err != nil {
			return mapWriteError(err, "song history conflicts with existing data")
		}
		if _, err := tx.Exec(ctx, `
			UPDATE student_songs
			SET stage = $3, stage_note = NULLIF($4, ''), updated_at = $5, version = version + 1
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.SongID, command.Stage, command.StageNote, command.Now); err != nil {
			return mapWriteError(err, "song conflicts with existing data")
		}
		song, err = readStudentSong(ctx, tx, principal.TenantID, command.SongID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "change_song_stage", command.IdempotencyKey, song, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "SongStageChanged", targetType: "student_song", targetID: command.SongID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"from": currentStage, "to": command.Stage},
			at:       command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "SongStageChanged", "student_song", command.SongID,
			map[string]any{"songId": command.SongID, "studentId": studentID, "stage": command.Stage}, command.Now)
	})
	if err != nil {
		return core.StudentSong{}, err
	}
	return song, nil
}

func (s *Store) ListStudentSongs(ctx context.Context, principal core.Principal, studentID string) ([]core.StudentSong, error) {
	manager, isSelf, err := s.journalViewerScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	isTeacher, err := s.evidenceTeacherScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	if !manager && !isSelf && !isTeacher {
		return nil, core.E(core.CodeForbidden, "repertoire is visible to the Student and assigned staff", nil)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM student_songs
		WHERE tenant_id = $1 AND student_id = $2
		ORDER BY updated_at DESC
		LIMIT 100
	`, principal.TenantID, studentID)
	if err != nil {
		return nil, fmt.Errorf("list student songs: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan song id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate song ids: %w", err)
	}
	rows.Close()
	result := make([]core.StudentSong, 0, len(ids))
	for _, id := range ids {
		song, err := readStudentSong(ctx, s.pool, principal.TenantID, id)
		if err != nil {
			return nil, err
		}
		result = append(result, song)
	}
	return result, nil
}
