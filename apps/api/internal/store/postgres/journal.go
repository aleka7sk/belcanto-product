package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 lesson journals and progress evidence (DEC-006/007). Drafts are
// teacher-private; publishing snapshots an immutable version (the DB
// trigger enforces immutability) and may attach evidence observations.

func journalTeacherAuthority(ctx context.Context, tx pgx.Tx, tenantID, actorID, occurrenceID, studentID string) error {
	var teacherAccountID string
	err := tx.QueryRow(ctx, `
		SELECT teacher_account_id FROM core_lesson_occurrences
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, occurrenceID).Scan(&teacherAccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	if err != nil {
		return fmt.Errorf("read occurrence teacher: %w", err)
	}
	if teacherAccountID != actorID {
		return core.E(core.CodeForbidden, "only the Lesson's Teacher writes the journal", nil)
	}
	var participates bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM core_lesson_occurrence_participants
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
		)
	`, tenantID, occurrenceID, studentID).Scan(&participates); err != nil {
		return fmt.Errorf("check journal participant: %w", err)
	}
	if !participates {
		return core.E(core.CodeInvalidInput, "Student does not participate in this Lesson", nil)
	}
	return nil
}

func (s *Store) SaveJournalDraft(ctx context.Context, command core.SaveJournalDraftCommand) (core.LessonJournal, error) {
	principal := command.Principal
	var journal core.LessonJournal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := journalTeacherAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.OccurrenceID, command.StudentID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO lesson_journals (
				id, tenant_id, occurrence_id, student_id, teacher_account_id,
				status, current_version, draft_what_worked, draft_current_focus, draft_next_step,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, 'draft', 0, $6, $7, $8, $9, $9)
			ON CONFLICT (tenant_id, occurrence_id, student_id)
			DO UPDATE SET draft_what_worked = EXCLUDED.draft_what_worked,
			              draft_current_focus = EXCLUDED.draft_current_focus,
			              draft_next_step = EXCLUDED.draft_next_step,
			              updated_at = EXCLUDED.updated_at
		`, command.JournalID, principal.TenantID, command.OccurrenceID, command.StudentID,
			principal.AccountID, command.Draft.WhatWorked, command.Draft.CurrentFocus,
			command.Draft.NextStep, command.Now); err != nil {
			return mapWriteError(err, "journal draft conflicts with existing data")
		}
		var err error
		journal, err = readJournal(ctx, tx, principal.TenantID, command.OccurrenceID, command.StudentID, true)
		if err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "JournalDraftSaved", targetType: "lesson_journal", targetID: journal.ID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.LessonJournal{}, err
	}
	return journal, nil
}

func (s *Store) PublishJournal(ctx context.Context, command core.PublishJournalCommand) (core.LessonJournal, error) {
	principal := command.Principal
	var journal core.LessonJournal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := journalTeacherAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.OccurrenceID, command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "publish_journal", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			journal, err = decodeReplay[core.LessonJournal](claim)
			return err
		}
		var startsAt time.Time
		var occurrenceStatus string
		if err := tx.QueryRow(ctx, `
			SELECT starts_at, status FROM core_lesson_occurrences
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.OccurrenceID).Scan(&startsAt, &occurrenceStatus); err != nil {
			return fmt.Errorf("read occurrence for publish: %w", err)
		}
		if startsAt.After(command.Now) {
			return core.E(core.CodeInvalidState, "the journal publishes after the Lesson starts", nil)
		}
		var journalID string
		var currentVersion int
		var draftWorked, draftFocus, draftNext *string
		err = tx.QueryRow(ctx, `
			SELECT id, current_version, draft_what_worked, draft_current_focus, draft_next_step
			FROM lesson_journals
			WHERE tenant_id = $1 AND occurrence_id = $2 AND student_id = $3
			FOR UPDATE
		`, principal.TenantID, command.OccurrenceID, command.StudentID).Scan(
			&journalID, &currentVersion, &draftWorked, &draftFocus, &draftNext)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidState, "save a draft before publishing", nil)
		}
		if err != nil {
			return fmt.Errorf("lock journal for publish: %w", err)
		}
		if draftWorked == nil || draftFocus == nil || draftNext == nil {
			return core.E(core.CodeInvalidState, "the draft must be complete before publishing", nil)
		}
		nextVersion := currentVersion + 1
		if nextVersion > 1 && command.CorrectionNote == "" {
			return core.E(core.CodeInvalidInput, "a correction requires an explicit note (DEC-007)", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO lesson_journal_versions (
				tenant_id, journal_id, version, what_worked, current_focus, next_step,
				correction_note, published_by_account_id, published_at
			) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9)
		`, principal.TenantID, journalID, nextVersion, *draftWorked, *draftFocus, *draftNext,
			command.CorrectionNote, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "journal version conflicts with existing data")
		}
		if _, err := tx.Exec(ctx, `
			UPDATE lesson_journals
			SET status = 'published', current_version = $3,
			    draft_what_worked = NULL, draft_current_focus = NULL, draft_next_step = NULL,
			    updated_at = $4
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, journalID, nextVersion, command.Now); err != nil {
			return fmt.Errorf("promote journal version: %w", err)
		}
		sourceID := fmt.Sprintf("%s:%d", journalID, nextVersion)
		for index, evidence := range command.Evidence {
			if _, err := tx.Exec(ctx, `
				INSERT INTO progress_evidence (
					id, tenant_id, student_id, source_kind, source_id, area, note,
					recorded_by_account_id, recorded_at
				) VALUES ($1, $2, $3, 'lesson_journal', $4, $5, $6, $7, $8)
			`, command.EvidenceIDs[index], principal.TenantID, command.StudentID,
				sourceID, evidence.Area, evidence.Note, principal.AccountID, command.Now); err != nil {
				return mapWriteError(err, "progress evidence conflicts with existing data")
			}
		}
		journal, err = readJournal(ctx, tx, principal.TenantID, command.OccurrenceID, command.StudentID, true)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "publish_journal", command.IdempotencyKey, journal, command.Now); err != nil {
			return err
		}
		action := "JournalPublished"
		if nextVersion > 1 {
			action = "JournalCorrected"
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "lesson_journal", targetID: journalID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"version": nextVersion, "evidenceCount": len(command.Evidence)},
			at:       command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, action, "lesson_journal", journalID,
			map[string]any{"journalId": journalID, "studentId": command.StudentID, "version": nextVersion}, command.Now)
	})
	if err != nil {
		return core.LessonJournal{}, err
	}
	return journal, nil
}

func readJournal(ctx context.Context, reader lessonReader, tenantID, occurrenceID, studentID string, includeDraft bool) (core.LessonJournal, error) {
	var journal core.LessonJournal
	var draftWorked, draftFocus, draftNext *string
	err := reader.QueryRow(ctx, `
		SELECT j.id, j.occurrence_id, j.student_id,
		       j.teacher_account_id, teacher_person.full_name,
		       j.status, j.current_version,
		       j.draft_what_worked, j.draft_current_focus, j.draft_next_step,
		       j.updated_at
		FROM lesson_journals j
		JOIN accounts teacher_account
		  ON teacher_account.tenant_id = j.tenant_id AND teacher_account.id = j.teacher_account_id
		JOIN people teacher_person
		  ON teacher_person.tenant_id = teacher_account.tenant_id
		 AND teacher_person.id = teacher_account.person_id
		WHERE j.tenant_id = $1 AND j.occurrence_id = $2 AND j.student_id = $3
	`, tenantID, occurrenceID, studentID).Scan(
		&journal.ID, &journal.OccurrenceID, &journal.StudentID,
		&journal.Teacher.AccountID, &journal.Teacher.FullName,
		&journal.Status, &journal.CurrentVersion,
		&draftWorked, &draftFocus, &draftNext, &journal.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.LessonJournal{}, core.E(core.CodeNotFound, "journal not found", nil)
	}
	if err != nil {
		return core.LessonJournal{}, fmt.Errorf("read journal: %w", err)
	}
	journal.UpdatedAt = journal.UpdatedAt.UTC()
	if includeDraft && draftWorked != nil && draftFocus != nil && draftNext != nil {
		journal.Draft = &core.JournalDraft{
			WhatWorked: *draftWorked, CurrentFocus: *draftFocus, NextStep: *draftNext,
		}
	}
	journal.Versions = make([]core.JournalVersion, 0)
	rows, err := reader.Query(ctx, `
		SELECT version, what_worked, current_focus, next_step,
		       COALESCE(correction_note, ''), published_at
		FROM lesson_journal_versions
		WHERE tenant_id = $1 AND journal_id = $2
		ORDER BY version DESC
	`, tenantID, journal.ID)
	if err != nil {
		return core.LessonJournal{}, fmt.Errorf("read journal versions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var version core.JournalVersion
		if err := rows.Scan(&version.Version, &version.WhatWorked, &version.CurrentFocus,
			&version.NextStep, &version.CorrectionNote, &version.PublishedAt); err != nil {
			return core.LessonJournal{}, fmt.Errorf("scan journal version: %w", err)
		}
		version.PublishedAt = version.PublishedAt.UTC()
		journal.Versions = append(journal.Versions, version)
	}
	if err := rows.Err(); err != nil {
		return core.LessonJournal{}, fmt.Errorf("iterate journal versions: %w", err)
	}
	return journal, nil
}

// journalViewerScope resolves how the caller may see journals for the
// student: managers and the journal's teacher see drafts; the student
// sees published versions only.
func (s *Store) journalViewerScope(ctx context.Context, principal core.Principal, studentID string) (manager bool, isSelf bool, err error) {
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		authority, authorityErr := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if authorityErr != nil {
			return authorityErr
		}
		manager = authority
		return nil
	}); err != nil {
		return false, false, err
	}
	ownStudentID, err := studentIDForAccount(ctx, s.pool, principal.TenantID, principal.AccountID)
	if err != nil {
		return false, false, err
	}
	return manager, ownStudentID != "" && ownStudentID == studentID, nil
}

func (s *Store) GetJournal(ctx context.Context, principal core.Principal, occurrenceID, studentID string) (core.LessonJournal, error) {
	journal, err := readJournal(ctx, s.pool, principal.TenantID, occurrenceID, studentID, true)
	if err != nil {
		return core.LessonJournal{}, err
	}
	if journal.Teacher.AccountID == principal.AccountID {
		return journal, nil
	}
	manager, isSelf, err := s.journalViewerScope(ctx, principal, studentID)
	if err != nil {
		return core.LessonJournal{}, err
	}
	if manager {
		return journal, nil
	}
	if isSelf {
		if journal.CurrentVersion == 0 {
			return core.LessonJournal{}, core.E(core.CodeNotFound, "journal not found", nil)
		}
		journal.Draft = nil
		return journal, nil
	}
	return core.LessonJournal{}, core.E(core.CodeNotFound, "journal not found", nil)
}

func (s *Store) ListStudentJournals(ctx context.Context, principal core.Principal, studentID string) ([]core.LessonJournal, error) {
	manager, isSelf, err := s.journalViewerScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT occurrence_id FROM lesson_journals
		WHERE tenant_id = $1 AND student_id = $2
		  AND ($3::boolean OR teacher_account_id = $4 OR ($5::boolean AND current_version > 0))
		ORDER BY updated_at DESC
		LIMIT 100
	`, principal.TenantID, studentID, manager, principal.AccountID, isSelf)
	if err != nil {
		return nil, fmt.Errorf("list student journals: %w", err)
	}
	defer rows.Close()
	occurrenceIDs := []string{}
	for rows.Next() {
		var occurrenceID string
		if err := rows.Scan(&occurrenceID); err != nil {
			return nil, fmt.Errorf("scan journal occurrence: %w", err)
		}
		occurrenceIDs = append(occurrenceIDs, occurrenceID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate journal occurrences: %w", err)
	}
	rows.Close()
	result := make([]core.LessonJournal, 0, len(occurrenceIDs))
	for _, occurrenceID := range occurrenceIDs {
		journal, err := readJournal(ctx, s.pool, principal.TenantID, occurrenceID, studentID, manager || !isSelf)
		if err != nil {
			return nil, err
		}
		if isSelf && !manager && journal.Teacher.AccountID != principal.AccountID {
			journal.Draft = nil
		}
		result = append(result, journal)
	}
	return result, nil
}

func (s *Store) ListProgressEvidence(ctx context.Context, principal core.Principal, studentID string) ([]core.ProgressEvidence, error) {
	manager, isSelf, err := s.journalViewerScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	isTeacher, err := s.evidenceTeacherScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	if !manager && !isSelf && !isTeacher {
		return nil, core.E(core.CodeForbidden, "progress is visible to the Student and assigned staff", nil)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, area, note, source_kind, source_id, recorded_at
		FROM progress_evidence
		WHERE tenant_id = $1 AND student_id = $2
		ORDER BY recorded_at DESC, id
		LIMIT 200
	`, principal.TenantID, studentID)
	if err != nil {
		return nil, fmt.Errorf("list progress evidence: %w", err)
	}
	defer rows.Close()
	result := []core.ProgressEvidence{}
	for rows.Next() {
		var evidence core.ProgressEvidence
		if err := rows.Scan(&evidence.ID, &evidence.Area, &evidence.Note,
			&evidence.SourceKind, &evidence.SourceID, &evidence.RecordedAt); err != nil {
			return nil, fmt.Errorf("scan progress evidence: %w", err)
		}
		evidence.RecordedAt = evidence.RecordedAt.UTC()
		result = append(result, evidence)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate progress evidence: %w", err)
	}
	return result, nil
}

func (s *Store) evidenceTeacherScope(ctx context.Context, principal core.Principal, studentID string) (bool, error) {
	var isTeacher bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM lesson_journals
			WHERE tenant_id = $1 AND student_id = $2 AND teacher_account_id = $3
		) OR EXISTS (
			SELECT 1 FROM teacher_assignments
			WHERE tenant_id = $1 AND student_id = $2 AND teacher_account_id = $3
			  AND status = 'active'
		)
	`, principal.TenantID, studentID, principal.AccountID).Scan(&isTeacher); err != nil {
		return false, fmt.Errorf("check evidence teacher scope: %w", err)
	}
	return isTeacher, nil
}
