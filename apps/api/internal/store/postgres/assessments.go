package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.4 assessments (domain/assessment.md, Approved). An assessment is
// the assigned Teacher's professional observation: Draft -> Published
// -> Superseded plus Withdrawn with a mandatory reason. Published
// content never rewrites — a correction is a new linked version that
// carries the evidence forward. Nothing deletes.

// assessmentAuthorAuthority: assessing is a pedagogical right — the
// Student's active assigned Teacher only (management reads via
// visibility, it does not author).
func assessmentAuthorAuthority(ctx context.Context, tx pgx.Tx, tenantID, actorID, studentID string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM students WHERE tenant_id = $1 AND id = $2)
	`, tenantID, studentID).Scan(&exists); err != nil {
		return fmt.Errorf("check assessment student: %w", err)
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
		return fmt.Errorf("check assessment teacher: %w", err)
	}
	if !assigned {
		return core.E(core.CodeForbidden, "assessments are written by the Student's assigned Teacher", nil)
	}
	return nil
}

const assessmentColumns = `
	a.id, a.student_id, a.author_account_id, person.full_name, a.author_role,
	a.assessment_type, a.context_type, a.context_id, to_char(a.assessment_date, 'YYYY-MM-DD'),
	a.summary, a.strengths, a.development_areas, a.recommendations,
	a.confidence, a.visibility, a.related_song_id, a.related_goal_id, a.areas,
	a.status, a.superseded_by_id, a.withdrawal_reason, a.published_at,
	a.version, a.created_at`

const assessmentJoins = `
	FROM assessments a
	JOIN accounts author ON author.tenant_id = a.tenant_id AND author.id = a.author_account_id
	JOIN people person ON person.tenant_id = author.tenant_id AND person.id = author.person_id`

type assessmentRowScanner interface {
	Scan(dest ...any) error
}

func scanAssessment(row assessmentRowScanner) (core.Assessment, error) {
	var result core.Assessment
	var contextID, confidence, songID, goalID, supersededBy, withdrawalReason *string
	err := row.Scan(
		&result.ID, &result.StudentID, &result.Author.AccountID, &result.Author.FullName,
		&result.AuthorRole, &result.Type, &result.ContextType, &contextID,
		&result.AssessmentDate, &result.Summary, &result.Strengths,
		&result.DevelopmentAreas, &result.Recommendations,
		&confidence, &result.Visibility, &songID, &goalID, &result.Areas,
		&result.Status, &supersededBy, &withdrawalReason, &result.PublishedAt,
		&result.Version, &result.CreatedAt,
	)
	if err != nil {
		return core.Assessment{}, err
	}
	if contextID != nil {
		result.ContextID = *contextID
	}
	if confidence != nil {
		result.Confidence = *confidence
	}
	if songID != nil {
		result.RelatedSongID = *songID
	}
	if goalID != nil {
		result.RelatedGoalID = *goalID
	}
	if supersededBy != nil {
		result.SupersededByID = *supersededBy
	}
	if withdrawalReason != nil {
		result.WithdrawalReason = *withdrawalReason
	}
	result.CreatedAt = result.CreatedAt.UTC()
	if result.PublishedAt != nil {
		published := result.PublishedAt.UTC()
		result.PublishedAt = &published
	}
	result.Evidence = []core.AssessmentEvidence{}
	return result, nil
}

func readAssessment(ctx context.Context, reader lessonReader, tenantID, assessmentID string) (core.Assessment, error) {
	row := reader.QueryRow(ctx, `
		SELECT `+assessmentColumns+assessmentJoins+`
		WHERE a.tenant_id = $1 AND a.id = $2
	`, tenantID, assessmentID)
	result, err := scanAssessment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.Assessment{}, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	if err != nil {
		return core.Assessment{}, fmt.Errorf("read assessment: %w", err)
	}
	rows, err := reader.Query(ctx, `
		SELECT id, kind, note, reference_id, added_at
		FROM assessment_evidence
		WHERE tenant_id = $1 AND assessment_id = $2
		ORDER BY added_at, id
	`, tenantID, assessmentID)
	if err != nil {
		return core.Assessment{}, fmt.Errorf("read assessment evidence: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var entry core.AssessmentEvidence
		var reference *string
		if err := rows.Scan(&entry.ID, &entry.Kind, &entry.Note, &reference, &entry.AddedAt); err != nil {
			return core.Assessment{}, fmt.Errorf("scan assessment evidence: %w", err)
		}
		if reference != nil {
			entry.ReferenceID = *reference
		}
		entry.AddedAt = entry.AddedAt.UTC()
		result.Evidence = append(result.Evidence, entry)
	}
	if err := rows.Err(); err != nil {
		return core.Assessment{}, fmt.Errorf("iterate assessment evidence: %w", err)
	}
	return result, nil
}

func (s *Store) assessmentViewerScope(ctx context.Context, principal core.Principal, studentID string) (assigned bool, manager bool, isSelf bool, err error) {
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		authority, authorityErr := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if authorityErr != nil {
			return authorityErr
		}
		manager = authority
		return tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM teacher_assignments
				WHERE tenant_id = $1 AND student_id = $2 AND teacher_account_id = $3
				  AND status = 'active'
			)
		`, principal.TenantID, studentID, principal.AccountID).Scan(&assigned)
	}); err != nil {
		return false, false, false, err
	}
	ownStudentID, err := studentIDForAccount(ctx, s.pool, principal.TenantID, principal.AccountID)
	if err != nil {
		return false, false, false, err
	}
	return assigned, manager, ownStudentID != "" && ownStudentID == studentID, nil
}

func writeAssessmentContent(ctx context.Context, tx pgx.Tx, command core.CreateAssessmentCommand, status string, publish bool) error {
	content := command.Content
	publishedAt := any(nil)
	if publish {
		publishedAt = command.Now
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO assessments (
			id, tenant_id, student_id, author_account_id, author_role,
			assessment_type, context_type, context_id, assessment_date,
			summary, strengths, development_areas, recommendations,
			confidence, visibility, related_song_id, related_goal_id, areas,
			status, published_at, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'Teacher', $5, $6, NULLIF($7, ''), $8::date,
		          $9, $10, $11, $12, NULLIF($13, ''), $14, NULLIF($15, ''), NULLIF($16, ''), $17,
		          $18, $19, 0, $20, $20)
	`, command.AssessmentID, command.Principal.TenantID, command.StudentID,
		command.Principal.AccountID, content.Type, content.ContextType, content.ContextID,
		content.AssessmentDate, content.Summary, content.Strengths, content.DevelopmentAreas,
		content.Recommendations, content.Confidence, content.Visibility,
		content.RelatedSongID, content.RelatedGoalID, content.Areas,
		status, publishedAt, command.Now)
	if err != nil {
		return mapWriteError(err, "assessment conflicts with existing data")
	}
	return nil
}

func (s *Store) CreateAssessment(ctx context.Context, command core.CreateAssessmentCommand) (core.Assessment, error) {
	principal := command.Principal
	var result core.Assessment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := assessmentAuthorAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_assessment", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.Assessment](claim)
			return err
		}
		if err := writeAssessmentContent(ctx, tx, command, "draft", false); err != nil {
			return err
		}
		result, err = readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_assessment", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AssessmentDraftCreated", targetType: "assessment", targetID: command.AssessmentID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"studentId": command.StudentID, "contextType": command.Content.ContextType},
			at:       command.Now,
		})
	})
	if err != nil {
		return core.Assessment{}, err
	}
	return result, nil
}

func (s *Store) UpdateAssessmentDraft(ctx context.Context, command core.UpdateAssessmentDraftCommand) (core.Assessment, error) {
	principal := command.Principal
	var result core.Assessment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "update_assessment", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.Assessment](claim)
			return err
		}
		var authorID, status string
		var version int64
		err = tx.QueryRow(ctx, `
			SELECT author_account_id, status, version FROM assessments
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.AssessmentID).Scan(&authorID, &status, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "assessment not found", nil)
		}
		if err != nil {
			return fmt.Errorf("read assessment for update: %w", err)
		}
		if authorID != principal.AccountID {
			return core.E(core.CodeForbidden, "only the author edits a draft", nil)
		}
		if status != "draft" {
			return core.E(core.CodeInvalidState, "published assessment content is immutable; supersede it instead", nil)
		}
		if version != command.ExpectedVersion {
			return core.E(core.CodeConflict, "assessment was changed by someone else", nil)
		}
		content := command.Content
		if _, err := tx.Exec(ctx, `
			UPDATE assessments SET
				assessment_type = $3, context_type = $4, context_id = NULLIF($5, ''),
				assessment_date = $6::date, summary = $7, strengths = $8,
				development_areas = $9, recommendations = $10,
				confidence = NULLIF($11, ''), visibility = $12,
				related_song_id = NULLIF($13, ''), related_goal_id = NULLIF($14, ''),
				areas = $15, version = version + 1, updated_at = $16
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.AssessmentID, content.Type, content.ContextType,
			content.ContextID, content.AssessmentDate, content.Summary, content.Strengths,
			content.DevelopmentAreas, content.Recommendations, content.Confidence,
			content.Visibility, content.RelatedSongID, content.RelatedGoalID,
			content.Areas, command.Now); err != nil {
			return mapWriteError(err, "assessment conflicts with existing data")
		}
		result, err = readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "update_assessment", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AssessmentDraftUpdated", targetType: "assessment", targetID: command.AssessmentID,
			decision: "allow", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.Assessment{}, err
	}
	return result, nil
}

func assessmentEvidenceGate(ctx context.Context, tx pgx.Tx, tenantID, actorID, assessmentID string) error {
	var authorID, status string
	err := tx.QueryRow(ctx, `
		SELECT author_account_id, status FROM assessments
		WHERE tenant_id = $1 AND id = $2
		FOR UPDATE
	`, tenantID, assessmentID).Scan(&authorID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.E(core.CodeNotFound, "assessment not found", nil)
	}
	if err != nil {
		return fmt.Errorf("read assessment for evidence: %w", err)
	}
	if authorID != actorID {
		return core.E(core.CodeForbidden, "only the author manages draft evidence", nil)
	}
	if status != "draft" {
		return core.E(core.CodeInvalidState, "evidence is edited while the assessment is a draft", nil)
	}
	return nil
}

func (s *Store) AddAssessmentEvidence(ctx context.Context, command core.AddAssessmentEvidenceCommand) (core.Assessment, error) {
	principal := command.Principal
	var result core.Assessment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "add_assessment_evidence", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.Assessment](claim)
			return err
		}
		if err := assessmentEvidenceGate(ctx, tx, principal.TenantID, principal.AccountID, command.AssessmentID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO assessment_evidence (id, tenant_id, assessment_id, kind, note, reference_id, added_at)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7)
		`, command.EvidenceID, principal.TenantID, command.AssessmentID,
			command.Kind, command.Note, command.ReferenceID, command.Now); err != nil {
			return mapWriteError(err, "evidence conflicts with existing data")
		}
		result, err = readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "add_assessment_evidence", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AssessmentEvidenceAdded", targetType: "assessment", targetID: command.AssessmentID,
			decision: "allow", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.Assessment{}, err
	}
	return result, nil
}

func (s *Store) RemoveAssessmentEvidence(ctx context.Context, command core.RemoveAssessmentEvidenceCommand) (core.Assessment, error) {
	principal := command.Principal
	var result core.Assessment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "remove_assessment_evidence", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.Assessment](claim)
			return err
		}
		if err := assessmentEvidenceGate(ctx, tx, principal.TenantID, principal.AccountID, command.AssessmentID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
			DELETE FROM assessment_evidence
			WHERE tenant_id = $1 AND assessment_id = $2 AND id = $3
		`, principal.TenantID, command.AssessmentID, command.EvidenceID)
		if err != nil {
			return mapWriteError(err, "evidence conflicts with existing data")
		}
		if tag.RowsAffected() == 0 {
			return core.E(core.CodeNotFound, "evidence not found", nil)
		}
		result, err = readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "remove_assessment_evidence", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AssessmentEvidenceRemoved", targetType: "assessment", targetID: command.AssessmentID,
			decision: "allow", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.Assessment{}, err
	}
	return result, nil
}

func (s *Store) PublishAssessment(ctx context.Context, command core.PublishAssessmentCommand) (core.Assessment, error) {
	principal := command.Principal
	var result core.Assessment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "publish_assessment", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.Assessment](claim)
			return err
		}
		current, err := readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if current.Author.AccountID != principal.AccountID {
			return core.E(core.CodeForbidden, "only the author publishes an assessment", nil)
		}
		if current.Status != "draft" {
			return core.E(core.CodeInvalidState, "only a draft publishes", nil)
		}
		if current.Version != command.ExpectedVersion {
			return core.E(core.CodeConflict, "assessment was changed by someone else", nil)
		}
		// A published assessment carries substance (business rule):
		// a summary plus at least one observation block or evidence row.
		if current.Summary == "" ||
			(current.Strengths == "" && current.DevelopmentAreas == "" &&
				current.Recommendations == "" && len(current.Evidence) == 0) {
			return core.E(core.CodeInvalidState, "a published assessment needs a summary and at least one observation, strength, development area or recommendation", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE assessments
			SET status = 'published', published_at = $3, version = version + 1, updated_at = $3
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.AssessmentID, command.Now); err != nil {
			return mapWriteError(err, "assessment conflicts with existing data")
		}
		result, err = readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "publish_assessment", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AssessmentPublished", targetType: "assessment", targetID: command.AssessmentID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"visibility": result.Visibility}, at: command.Now,
		}); err != nil {
			return err
		}
		if result.Visibility == "student_visible" {
			return appendOutbox(ctx, tx, principal.TenantID, "AssessmentPublished", "assessment", command.AssessmentID,
				map[string]any{"assessmentId": command.AssessmentID, "studentId": result.StudentID}, command.Now)
		}
		return nil
	})
	if err != nil {
		return core.Assessment{}, err
	}
	return result, nil
}

func (s *Store) SupersedeAssessment(ctx context.Context, command core.SupersedeAssessmentCommand) ([]core.Assessment, error) {
	principal := command.Principal
	var chain []core.Assessment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "supersede_assessment", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			chain, err = decodeReplay[[]core.Assessment](claim)
			return err
		}
		current, err := readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if current.Author.AccountID != principal.AccountID {
			return core.E(core.CodeForbidden, "only the author supersedes an assessment", nil)
		}
		if current.Status != "published" {
			return core.E(core.CodeInvalidState, "only a published assessment can be superseded", nil)
		}
		createCommand := core.CreateAssessmentCommand{
			Principal: principal, AssessmentID: command.NewAssessmentID,
			StudentID: current.StudentID, Content: command.Content, Now: command.Now,
		}
		if err := writeAssessmentContent(ctx, tx, createCommand, "published", true); err != nil {
			return err
		}
		// The observations still ground the correction: evidence rows
		// travel to the replacement under version-scoped identifiers.
		if _, err := tx.Exec(ctx, `
			INSERT INTO assessment_evidence (id, tenant_id, assessment_id, kind, note, reference_id, added_at)
			SELECT $3 || '.' || row_number() OVER (ORDER BY added_at, id), tenant_id, $3, kind, note, reference_id, added_at
			FROM assessment_evidence
			WHERE tenant_id = $1 AND assessment_id = $2
		`, principal.TenantID, command.AssessmentID, command.NewAssessmentID); err != nil {
			return mapWriteError(err, "evidence conflicts with existing data")
		}
		if _, err := tx.Exec(ctx, `
			UPDATE assessments
			SET status = 'superseded', superseded_by_id = $3, version = version + 1, updated_at = $4
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.AssessmentID, command.NewAssessmentID, command.Now); err != nil {
			return mapWriteError(err, "assessment conflicts with existing data")
		}
		replaced, err := readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		replacement, err := readAssessment(ctx, tx, principal.TenantID, command.NewAssessmentID)
		if err != nil {
			return err
		}
		chain = []core.Assessment{replaced, replacement}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "supersede_assessment", command.IdempotencyKey, chain, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AssessmentSuperseded", targetType: "assessment", targetID: command.AssessmentID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"supersededById": command.NewAssessmentID}, at: command.Now,
		}); err != nil {
			return err
		}
		if replacement.Visibility == "student_visible" {
			return appendOutbox(ctx, tx, principal.TenantID, "AssessmentPublished", "assessment", command.NewAssessmentID,
				map[string]any{"assessmentId": command.NewAssessmentID, "studentId": replacement.StudentID}, command.Now)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return chain, nil
}

func (s *Store) WithdrawAssessment(ctx context.Context, command core.WithdrawAssessmentCommand) (core.Assessment, error) {
	principal := command.Principal
	var result core.Assessment
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "withdraw_assessment", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.Assessment](claim)
			return err
		}
		current, err := readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		manager, err := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if err != nil {
			return err
		}
		if current.Author.AccountID != principal.AccountID && !manager {
			return core.E(core.CodeForbidden, "withdrawal is the author's or the school's action", nil)
		}
		if current.Status != "draft" && current.Status != "published" {
			return core.E(core.CodeInvalidState, "only a draft or published assessment withdraws", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE assessments
			SET status = 'withdrawn', withdrawal_reason = $3, version = version + 1, updated_at = $4
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.AssessmentID, command.Reason, command.Now); err != nil {
			return mapWriteError(err, "assessment conflicts with existing data")
		}
		result, err = readAssessment(ctx, tx, principal.TenantID, command.AssessmentID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "withdraw_assessment", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AssessmentWithdrawn", targetType: "assessment", targetID: command.AssessmentID,
			decision: "allow", reason: command.Reason,
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.Assessment{}, err
	}
	return result, nil
}

func (s *Store) ListStudentAssessments(ctx context.Context, principal core.Principal, studentID string) ([]core.Assessment, error) {
	assigned, manager, isSelf, err := s.assessmentViewerScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+assessmentColumns+assessmentJoins+`
		WHERE a.tenant_id = $1 AND a.student_id = $2
		ORDER BY a.assessment_date DESC, a.created_at DESC, a.id
	`, principal.TenantID, studentID)
	if err != nil {
		return nil, fmt.Errorf("list assessments: %w", err)
	}
	defer rows.Close()
	candidates := []core.Assessment{}
	for rows.Next() {
		view, scanErr := scanAssessment(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan assessment: %w", scanErr)
		}
		candidates = append(candidates, view)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assessments: %w", err)
	}
	result := []core.Assessment{}
	for _, view := range candidates {
		isAuthor := view.Author.AccountID == principal.AccountID
		if !core.AssessmentVisible(view, isAuthor, assigned, manager, isSelf) {
			continue
		}
		full, readErr := readAssessment(ctx, s.pool, principal.TenantID, view.ID)
		if readErr != nil {
			return nil, readErr
		}
		result = append(result, full)
	}
	return result, nil
}

func (s *Store) GetAssessment(ctx context.Context, principal core.Principal, assessmentID string) (core.Assessment, error) {
	view, err := readAssessment(ctx, s.pool, principal.TenantID, assessmentID)
	if err != nil {
		return core.Assessment{}, err
	}
	assigned, manager, isSelf, err := s.assessmentViewerScope(ctx, principal, view.StudentID)
	if err != nil {
		return core.Assessment{}, err
	}
	if !core.AssessmentVisible(view, view.Author.AccountID == principal.AccountID, assigned, manager, isSelf) {
		return core.Assessment{}, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	return view, nil
}
