package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 goals and achievements. A goal is reframed, never «failed»; a
// completed or cancelled goal is immutable (DB trigger). Definitions
// are a published catalog; awards are evidence-backed and revocation
// preserves the original.

const studentGoalColumns = `
	g.id, g.student_id, g.criterion, COALESCE(g.description, ''),
	COALESCE(g.related_song_id, ''), COALESCE(g.related_skill_area, ''),
	g.status, COALESCE(g.completion_note, ''), COALESCE(g.cancel_reason, ''),
	COALESCE(g.replaced_by_goal_id, ''),
	g.created_by_account_id, person.full_name,
	g.version, g.created_at, g.updated_at`

const studentGoalJoins = `
	FROM student_goals g
	JOIN accounts account ON account.tenant_id = g.tenant_id AND account.id = g.created_by_account_id
	JOIN people person ON person.tenant_id = account.tenant_id AND person.id = account.person_id`

func scanStudentGoal(row pgx.Row) (core.StudentGoal, error) {
	var goal core.StudentGoal
	err := row.Scan(
		&goal.ID, &goal.StudentID, &goal.Criterion, &goal.Description,
		&goal.RelatedSongID, &goal.RelatedSkillArea,
		&goal.Status, &goal.CompletionNote, &goal.CancelReason, &goal.ReplacedByGoalID,
		&goal.CreatedBy.AccountID, &goal.CreatedBy.FullName,
		&goal.Version, &goal.CreatedAt, &goal.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.StudentGoal{}, core.E(core.CodeNotFound, "goal not found", nil)
	}
	if err != nil {
		return core.StudentGoal{}, fmt.Errorf("read student goal: %w", err)
	}
	goal.CreatedAt = goal.CreatedAt.UTC()
	goal.UpdatedAt = goal.UpdatedAt.UTC()
	return goal, nil
}

func readStudentGoal(ctx context.Context, reader lessonReader, tenantID, goalID string) (core.StudentGoal, error) {
	return scanStudentGoal(reader.QueryRow(ctx, `
		SELECT `+studentGoalColumns+studentGoalJoins+`
		WHERE g.tenant_id = $1 AND g.id = $2
	`, tenantID, goalID))
}

func (s *Store) CreateGoal(ctx context.Context, command core.CreateGoalCommand) (core.StudentGoal, error) {
	principal := command.Principal
	var goal core.StudentGoal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := repertoireMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_goal", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			goal, err = decodeReplay[core.StudentGoal](claim)
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO student_goals (
				id, tenant_id, student_id, criterion, description,
				related_song_id, related_skill_area, status,
				created_by_account_id, version, created_at, updated_at
			) VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''),
			          'active', $8, 1, $9, $9)
		`, command.GoalID, principal.TenantID, command.StudentID, command.Criterion,
			command.Description, command.RelatedSongID, command.RelatedSkillArea,
			principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "goal conflicts with existing data")
		}
		goal, err = readStudentGoal(ctx, tx, principal.TenantID, command.GoalID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_goal", command.IdempotencyKey, goal, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "GoalCreated", targetType: "student_goal", targetID: command.GoalID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"studentId": command.StudentID},
			at:       command.Now,
		})
	})
	if err != nil {
		return core.StudentGoal{}, err
	}
	return goal, nil
}

func (s *Store) CompleteGoal(ctx context.Context, command core.CompleteGoalCommand) (core.StudentGoal, error) {
	principal := command.Principal
	var goal core.StudentGoal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "complete_goal", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			goal, err = decodeReplay[core.StudentGoal](claim)
			return err
		}
		var studentID, status string
		var version int
		err = tx.QueryRow(ctx, `
			SELECT student_id, status, version FROM student_goals
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.GoalID).Scan(&studentID, &status, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "goal not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock goal: %w", err)
		}
		if err := repertoireMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, studentID); err != nil {
			return err
		}
		if status != core.GoalStatusActive {
			return core.E(core.CodeInvalidState, "only an active goal completes", nil)
		}
		if command.ExpectedVersion != version {
			return core.E(core.CodeConflict, "the goal changed; reload and retry", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE student_goals
			SET status = 'completed', completion_note = $3, updated_at = $4, version = version + 1
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.GoalID, command.CompletionNote, command.Now); err != nil {
			return mapWriteError(err, "goal conflicts with existing data")
		}
		goal, err = readStudentGoal(ctx, tx, principal.TenantID, command.GoalID)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "complete_goal", command.IdempotencyKey, goal, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "GoalCompleted", targetType: "student_goal", targetID: command.GoalID,
			decision: "allow", idempotencyKey: command.IdempotencyKey, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "GoalCompleted", "student_goal", command.GoalID,
			map[string]any{"goalId": command.GoalID, "studentId": studentID}, command.Now)
	})
	if err != nil {
		return core.StudentGoal{}, err
	}
	return goal, nil
}

func (s *Store) ReframeGoal(ctx context.Context, command core.ReframeGoalCommand) ([]core.StudentGoal, error) {
	principal := command.Principal
	var goals []core.StudentGoal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "reframe_goal", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			goals, err = decodeReplay[[]core.StudentGoal](claim)
			return err
		}
		var studentID, status string
		var version int
		err = tx.QueryRow(ctx, `
			SELECT student_id, status, version FROM student_goals
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.GoalID).Scan(&studentID, &status, &version)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "goal not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock goal: %w", err)
		}
		if err := repertoireMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, studentID); err != nil {
			return err
		}
		if status != core.GoalStatusActive {
			return core.E(core.CodeInvalidState, "only an active goal is reframed", nil)
		}
		if command.ExpectedVersion != version {
			return core.E(core.CodeConflict, "the goal changed; reload and retry", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE student_goals
			SET status = 'cancelled', cancel_reason = $3, updated_at = $4, version = version + 1
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.GoalID, command.Reason, command.Now); err != nil {
			return mapWriteError(err, "goal conflicts with existing data")
		}
		if command.NewCriterion != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO student_goals (
					id, tenant_id, student_id, criterion, description, status,
					created_by_account_id, version, created_at, updated_at
				) VALUES ($1, $2, $3, $4, NULLIF($5, ''), 'active', $6, 1, $7, $7)
			`, command.NewGoalID, principal.TenantID, studentID, command.NewCriterion,
				command.NewDescription, principal.AccountID, command.Now); err != nil {
				return mapWriteError(err, "replacement goal conflicts with existing data")
			}
			if _, err := tx.Exec(ctx, `
				UPDATE student_goals
				SET replaced_by_goal_id = $3
				WHERE tenant_id = $1 AND id = $2
			`, principal.TenantID, command.GoalID, command.NewGoalID); err != nil {
				return mapWriteError(err, "goal conflicts with existing data")
			}
		}
		cancelled, err := readStudentGoal(ctx, tx, principal.TenantID, command.GoalID)
		if err != nil {
			return err
		}
		goals = []core.StudentGoal{cancelled}
		if command.NewCriterion != "" {
			replacement, err := readStudentGoal(ctx, tx, principal.TenantID, command.NewGoalID)
			if err != nil {
				return err
			}
			goals = append(goals, replacement)
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "reframe_goal", command.IdempotencyKey, goals, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "GoalReframed", targetType: "student_goal", targetID: command.GoalID,
			decision: "allow", reason: command.Reason,
			idempotencyKey: command.IdempotencyKey,
			metadata:       map[string]any{"replacedBy": command.NewGoalID},
			at:             command.Now,
		})
	})
	if err != nil {
		return nil, err
	}
	return goals, nil
}

func (s *Store) ListStudentGoals(ctx context.Context, principal core.Principal, studentID string) ([]core.StudentGoal, error) {
	manager, isSelf, err := s.journalViewerScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	isTeacher, err := s.evidenceTeacherScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	if !manager && !isSelf && !isTeacher {
		return nil, core.E(core.CodeForbidden, "goals are visible to the Student and assigned staff", nil)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+studentGoalColumns+studentGoalJoins+`
		WHERE g.tenant_id = $1 AND g.student_id = $2
		ORDER BY g.updated_at DESC
		LIMIT 100
	`, principal.TenantID, studentID)
	if err != nil {
		return nil, fmt.Errorf("list student goals: %w", err)
	}
	defer rows.Close()
	result := []core.StudentGoal{}
	for rows.Next() {
		goal, err := scanStudentGoal(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, goal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student goals: %w", err)
	}
	return result, nil
}

const achievementDefinitionColumns = `
	id, name, description, category, COALESCE(evidence_requirement, ''),
	status, definition_version, created_at, retired_at`

func scanAchievementDefinition(row pgx.Row) (core.AchievementDefinition, error) {
	var definition core.AchievementDefinition
	err := row.Scan(&definition.ID, &definition.Name, &definition.Description,
		&definition.Category, &definition.EvidenceRequirement, &definition.Status,
		&definition.DefinitionVersion, &definition.CreatedAt, &definition.RetiredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.AchievementDefinition{}, core.E(core.CodeNotFound, "achievement definition not found", nil)
	}
	if err != nil {
		return core.AchievementDefinition{}, fmt.Errorf("read achievement definition: %w", err)
	}
	definition.CreatedAt = definition.CreatedAt.UTC()
	if definition.RetiredAt != nil {
		utc := definition.RetiredAt.UTC()
		definition.RetiredAt = &utc
	}
	return definition, nil
}

func achievementCatalogAuthority(ctx context.Context, tx pgx.Tx, tenantID, actorID string) error {
	authority, err := lessonManagementAuthority(ctx, tx, tenantID, actorID)
	if err != nil {
		return err
	}
	if !authority {
		return core.E(core.CodeForbidden, "the achievement catalog is managed by the school", nil)
	}
	return nil
}

func (s *Store) CreateAchievementDefinition(ctx context.Context, command core.CreateAchievementDefinitionCommand) (core.AchievementDefinition, error) {
	principal := command.Principal
	var definition core.AchievementDefinition
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := achievementCatalogAuthority(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_achievement_definition", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			definition, err = decodeReplay[core.AchievementDefinition](claim)
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO achievement_definitions (
				id, tenant_id, name, description, category, evidence_requirement,
				status, definition_version, created_by_account_id, created_at
			) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), 'published', 1, $7, $8)
		`, command.DefinitionID, principal.TenantID, command.Name, command.Description,
			command.Category, command.EvidenceRequirement, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "achievement definition conflicts with existing data")
		}
		definition, err = scanAchievementDefinition(tx.QueryRow(ctx, `
			SELECT `+achievementDefinitionColumns+` FROM achievement_definitions
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.DefinitionID))
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_achievement_definition", command.IdempotencyKey, definition, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AchievementDefinitionPublished", targetType: "achievement_definition",
			targetID: command.DefinitionID, decision: "allow",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.AchievementDefinition{}, err
	}
	return definition, nil
}

func (s *Store) RetireAchievementDefinition(ctx context.Context, command core.RetireAchievementDefinitionCommand) (core.AchievementDefinition, error) {
	principal := command.Principal
	var definition core.AchievementDefinition
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := achievementCatalogAuthority(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "retire_achievement_definition", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			definition, err = decodeReplay[core.AchievementDefinition](claim)
			return err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE achievement_definitions
			SET status = 'retired', retired_at = $3
			WHERE tenant_id = $1 AND id = $2 AND status = 'published'
		`, principal.TenantID, command.DefinitionID, command.Now)
		if err != nil {
			return mapWriteError(err, "achievement definition conflicts with existing data")
		}
		if tag.RowsAffected() == 0 {
			return core.E(core.CodeInvalidState, "only a published definition retires", nil)
		}
		definition, err = scanAchievementDefinition(tx.QueryRow(ctx, `
			SELECT `+achievementDefinitionColumns+` FROM achievement_definitions
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.DefinitionID))
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "retire_achievement_definition", command.IdempotencyKey, definition, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AchievementDefinitionRetired", targetType: "achievement_definition",
			targetID: command.DefinitionID, decision: "allow",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.AchievementDefinition{}, err
	}
	return definition, nil
}

func (s *Store) ListAchievementDefinitions(ctx context.Context, principal core.Principal) ([]core.AchievementDefinition, error) {
	if err := activeAccountExistsPool(ctx, s.pool, principal.TenantID, principal.AccountID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+achievementDefinitionColumns+` FROM achievement_definitions
		WHERE tenant_id = $1
		ORDER BY category, name
		LIMIT 200
	`, principal.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list achievement definitions: %w", err)
	}
	defer rows.Close()
	result := []core.AchievementDefinition{}
	for rows.Next() {
		definition, err := scanAchievementDefinition(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, definition)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievement definitions: %w", err)
	}
	return result, nil
}

func activeAccountExistsPool(ctx context.Context, reader lessonReader, tenantID, accountID string) error {
	var active bool
	if err := reader.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM accounts
			WHERE tenant_id = $1 AND id = $2 AND status = 'active'
		)
	`, tenantID, accountID).Scan(&active); err != nil {
		return fmt.Errorf("check active account: %w", err)
	}
	if !active {
		return core.E(core.CodeForbidden, "an active account is required", nil)
	}
	return nil
}

const achievementAwardColumns = `
	a.id, a.definition_id, d.name, d.category, a.student_id, a.evidence_note,
	a.status, COALESCE(a.revoke_reason, ''), a.revoked_at,
	a.awarded_by_account_id, person.full_name, a.awarded_at, a.definition_version`

const achievementAwardJoins = `
	FROM achievement_awards a
	JOIN achievement_definitions d ON d.tenant_id = a.tenant_id AND d.id = a.definition_id
	JOIN accounts account ON account.tenant_id = a.tenant_id AND account.id = a.awarded_by_account_id
	JOIN people person ON person.tenant_id = account.tenant_id AND person.id = account.person_id`

func scanAchievementAward(row pgx.Row) (core.AchievementAward, error) {
	var award core.AchievementAward
	err := row.Scan(&award.ID, &award.DefinitionID, &award.DefinitionName, &award.Category,
		&award.StudentID, &award.EvidenceNote, &award.Status, &award.RevokeReason,
		&award.RevokedAt, &award.AwardedBy.AccountID, &award.AwardedBy.FullName,
		&award.AwardedAt, &award.DefinitionVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.AchievementAward{}, core.E(core.CodeNotFound, "award not found", nil)
	}
	if err != nil {
		return core.AchievementAward{}, fmt.Errorf("read achievement award: %w", err)
	}
	award.AwardedAt = award.AwardedAt.UTC()
	if award.RevokedAt != nil {
		utc := award.RevokedAt.UTC()
		award.RevokedAt = &utc
	}
	return award, nil
}

func (s *Store) AwardAchievement(ctx context.Context, command core.AwardAchievementCommand) (core.AchievementAward, error) {
	principal := command.Principal
	var award core.AchievementAward
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := repertoireMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "award_achievement", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			award, err = decodeReplay[core.AchievementAward](claim)
			return err
		}
		var definitionStatus string
		var definitionVersion int
		err = tx.QueryRow(ctx, `
			SELECT status, definition_version FROM achievement_definitions
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.DefinitionID).Scan(&definitionStatus, &definitionVersion)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "achievement definition not found", nil)
		}
		if err != nil {
			return fmt.Errorf("read definition for award: %w", err)
		}
		if definitionStatus != "published" {
			return core.E(core.CodeInvalidState, "a retired definition does not create new awards", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO achievement_awards (
				id, tenant_id, definition_id, definition_version, student_id,
				evidence_note, status, awarded_by_account_id, awarded_at
			) VALUES ($1, $2, $3, $4, $5, $6, 'awarded', $7, $8)
		`, command.AwardID, principal.TenantID, command.DefinitionID, definitionVersion,
			command.StudentID, command.EvidenceNote, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "achievement award conflicts with existing data")
		}
		award, err = scanAchievementAward(tx.QueryRow(ctx, `
			SELECT `+achievementAwardColumns+achievementAwardJoins+`
			WHERE a.tenant_id = $1 AND a.id = $2
		`, principal.TenantID, command.AwardID))
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "award_achievement", command.IdempotencyKey, award, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AchievementAwarded", targetType: "achievement_award", targetID: command.AwardID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"definitionId": command.DefinitionID, "studentId": command.StudentID},
			at:       command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "AchievementAwarded", "achievement_award", command.AwardID,
			map[string]any{"awardId": command.AwardID, "studentId": command.StudentID}, command.Now)
	})
	if err != nil {
		return core.AchievementAward{}, err
	}
	return award, nil
}

func (s *Store) RevokeAchievement(ctx context.Context, command core.RevokeAchievementCommand) (core.AchievementAward, error) {
	principal := command.Principal
	var award core.AchievementAward
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "revoke_achievement", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			award, err = decodeReplay[core.AchievementAward](claim)
			return err
		}
		var studentID, status string
		err = tx.QueryRow(ctx, `
			SELECT student_id, status FROM achievement_awards
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.AwardID).Scan(&studentID, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "award not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock award: %w", err)
		}
		if err := repertoireMarkerAuthority(ctx, tx, principal.TenantID, principal.AccountID, studentID); err != nil {
			return err
		}
		if status != "awarded" {
			return core.E(core.CodeInvalidState, "the award is already revoked", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE achievement_awards
			SET status = 'revoked', revoke_reason = $3, revoked_at = $4
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.AwardID, command.Reason, command.Now); err != nil {
			return mapWriteError(err, "achievement award conflicts with existing data")
		}
		award, err = scanAchievementAward(tx.QueryRow(ctx, `
			SELECT `+achievementAwardColumns+achievementAwardJoins+`
			WHERE a.tenant_id = $1 AND a.id = $2
		`, principal.TenantID, command.AwardID))
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "revoke_achievement", command.IdempotencyKey, award, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "AchievementRevoked", targetType: "achievement_award", targetID: command.AwardID,
			decision: "allow", reason: command.Reason,
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.AchievementAward{}, err
	}
	return award, nil
}

func (s *Store) ListStudentAwards(ctx context.Context, principal core.Principal, studentID string) ([]core.AchievementAward, error) {
	manager, isSelf, err := s.journalViewerScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	isTeacher, err := s.evidenceTeacherScope(ctx, principal, studentID)
	if err != nil {
		return nil, err
	}
	if !manager && !isSelf && !isTeacher {
		return nil, core.E(core.CodeForbidden, "achievements are visible to the Student and assigned staff", nil)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+achievementAwardColumns+achievementAwardJoins+`
		WHERE a.tenant_id = $1 AND a.student_id = $2
		ORDER BY a.awarded_at DESC, a.id
		LIMIT 100
	`, principal.TenantID, studentID)
	if err != nil {
		return nil, fmt.Errorf("list student awards: %w", err)
	}
	defer rows.Close()
	result := []core.AchievementAward{}
	for rows.Next() {
		award, err := scanAchievementAward(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, award)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student awards: %w", err)
	}
	return result, nil
}
