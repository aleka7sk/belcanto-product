package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/jackc/pgx/v5"
)

func (s *Store) GrantDelegation(ctx context.Context, command core.GrantDelegationCommand) (core.DelegationResult, error) {
	var result core.DelegationResult
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		owner, err := hasActiveRole(ctx, tx, command.TenantID, command.OwnerAccountID, core.RoleOwner)
		if err != nil {
			return err
		}
		if !owner {
			return core.E(core.CodeForbidden, "only Owner can grant privileged access", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.OwnerAccountID, "grant_delegation", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.DelegationResult](claim)
			return err
		}

		if command.OwnerAccountID == command.AdministratorID {
			return core.E(core.CodeInvalidInput, "target must be another active Administrator in the same school", nil)
		}
		administrator, err := hasActiveRole(ctx, tx, command.TenantID, command.AdministratorID, core.RoleAdministrator)
		if err != nil {
			return err
		}
		if !administrator {
			return core.E(core.CodeInvalidInput, "target must be an active Administrator in the same school", nil)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE capability_delegations
			SET status = 'expired', version = version + 1
			WHERE tenant_id = $1 AND grantee_account_id = $2 AND bundle = $3
			  AND status = 'active' AND expires_at IS NOT NULL AND expires_at <= $4
		`, command.TenantID, command.AdministratorID, command.Bundle, command.Now); err != nil {
			return fmt.Errorf("expire old delegation: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO capability_delegations (
				id, tenant_id, grantee_account_id, granted_by_account_id, bundle,
				status, reason, granted_at, expires_at
			) VALUES ($1, $2, $3, $4, $5, 'active', $6, $7, $8)
		`, command.ID, command.TenantID, command.AdministratorID, command.OwnerAccountID,
			command.Bundle, command.Reason, command.Now, command.ExpiresAt)
		if err != nil {
			return mapWriteError(err, "Administrator already has an active delegation")
		}

		result = core.DelegationResult{
			ID:              command.ID,
			AdministratorID: command.AdministratorID,
			Bundle:          command.Bundle,
			Status:          "active",
			GrantedAt:       command.Now,
			ExpiresAt:       command.ExpiresAt,
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.OwnerAccountID, "grant_delegation", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.OwnerAccountID,
			delegationID: command.ID, action: "StudentOnboardingDelegationGranted",
			targetType: "account", targetID: command.AdministratorID, decision: "allow",
			reason: command.Reason, idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"bundle": command.Bundle, "expiresAt": command.ExpiresAt}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "StudentOnboardingDelegationGranted", "delegation", command.ID, map[string]any{
			"delegationId":           command.ID,
			"administratorAccountId": command.AdministratorID,
			"bundle":                 command.Bundle,
			"expiresAt":              command.ExpiresAt,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.OwnerAccountID,
			action: "StudentOnboardingDelegationGranted", targetType: "account",
			targetID: command.AdministratorID, reason: "owner_required",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return result, err
}

func (s *Store) RevokeDelegation(ctx context.Context, command core.RevokeDelegationCommand) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		owner, err := hasActiveRole(ctx, tx, command.TenantID, command.OwnerAccountID, core.RoleOwner)
		if err != nil {
			return err
		}
		if !owner {
			return core.E(core.CodeForbidden, "only Owner can revoke privileged access", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.OwnerAccountID, "revoke_delegation", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			return nil
		}
		var status string
		err = tx.QueryRow(ctx, `
			SELECT status
			FROM capability_delegations
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, command.TenantID, command.DelegationID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "delegation not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock delegation: %w", err)
		}
		if status != "active" {
			return core.E(core.CodeConflict, "delegation is not active", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE capability_delegations
			SET status = 'revoked', revoked_at = $3, revoked_by_account_id = $4,
			    revocation_reason = $5, version = version + 1
			WHERE tenant_id = $1 AND id = $2 AND status = 'active'
		`, command.TenantID, command.DelegationID, command.Now, command.OwnerAccountID, command.Reason); err != nil {
			return fmt.Errorf("revoke delegation: %w", err)
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.OwnerAccountID, "revoke_delegation", command.IdempotencyKey, struct{}{}, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.OwnerAccountID,
			delegationID: command.DelegationID, action: "StudentOnboardingDelegationRevoked",
			targetType: "delegation", targetID: command.DelegationID, decision: "allow",
			reason: command.Reason, idempotencyKey: command.IdempotencyKey, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "StudentOnboardingDelegationRevoked", "delegation", command.DelegationID, map[string]any{
			"delegationId": command.DelegationID,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.OwnerAccountID,
			action: "StudentOnboardingDelegationRevoked", targetType: "delegation",
			targetID: command.DelegationID, reason: "owner_required",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return err
}

func (s *Store) CreateStudent(ctx context.Context, command core.CreateStudentCommand) (core.StudentResult, error) {
	var result core.StudentResult
	var denialDelegationID string
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		delegationID, authorized, err := onboardingAuthority(ctx, tx, command.TenantID, command.ActorAccountID, command.Now)
		denialDelegationID = delegationID
		if err != nil {
			return err
		}
		if !authorized {
			return core.E(core.CodeForbidden, "student onboarding permission is required", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_student", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.StudentResult](claim)
			return err
		}
		teacher, err := hasActiveRole(ctx, tx, command.TenantID, command.TeacherAccountID, core.RoleTeacher)
		if err != nil {
			return err
		}
		if !teacher {
			return core.E(core.CodeInvalidInput, "assigned teacher is not active in this school", nil)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO people (id, tenant_id, full_name, created_at)
			VALUES ($1, $2, $3, $4)
		`, command.PersonID, command.TenantID, command.FullName, command.Now); err != nil {
			return mapWriteError(err, "student identity already exists")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO school_memberships (id, tenant_id, person_id, created_at)
			VALUES ($1, $2, $3, $4)
		`, command.MembershipID, command.TenantID, command.PersonID, command.Now); err != nil {
			return fmt.Errorf("create student membership: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO accounts (
				id, tenant_id, person_id, status, created_at, updated_at
			) VALUES ($1, $2, $3, 'pending_activation', $4, $4)
		`, command.AccountID, command.TenantID, command.PersonID, command.Now); err != nil {
			return fmt.Errorf("create pending student account: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO account_login_identifiers (
				account_id, tenant_id, identifier_type, normalized_value, status, created_at
			) VALUES ($1, $2, 'phone', $3, 'reserved', $4)
		`, command.AccountID, command.TenantID, command.Phone, command.Now); err != nil {
			return mapWriteError(err, "login identifier is unavailable")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO students (
				id, tenant_id, person_id, membership_id, account_id,
				enrollment_reference, status, locale, timezone, adult_confirmed,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8, $9, $10, $10)
		`, command.StudentID, command.TenantID, command.PersonID, command.MembershipID,
			command.AccountID, command.EnrollmentReference, command.Locale, command.Timezone,
			command.AdultConfirmed, command.Now); err != nil {
			return mapWriteError(err, "enrollment reference already exists")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_grants (
				id, tenant_id, account_id, role_type, scope_type, scope_id,
				status, granted_by, granted_at
			) VALUES ($1, $2, $3, 'Student', 'student', $4, 'active', $5, $6)
		`, command.RoleGrantID, command.TenantID, command.AccountID, command.StudentID,
			command.ActorAccountID, command.Now); err != nil {
			return fmt.Errorf("grant Student role: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO teacher_assignments (
				id, tenant_id, student_id, teacher_account_id, status,
				assigned_by_account_id, assigned_at, effective_from
			) VALUES ($1, $2, $3, $4, 'active', $5, $6, $6)
		`, command.TeacherAssignmentID, command.TenantID, command.StudentID,
			command.TeacherAccountID, command.ActorAccountID, command.Now); err != nil {
			return fmt.Errorf("assign student Teacher: %w", err)
		}

		result = core.StudentResult{
			StudentID:       command.StudentID,
			AccountID:       command.AccountID,
			OnboardingState: "awaiting_first_minute",
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "create_student", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			delegationID: delegationID, action: "StudentCreated", targetType: "student",
			targetID: command.StudentID, decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"teacherAccountId": command.TeacherAccountID}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "StudentCreated", "student", command.StudentID, map[string]any{
			"studentId":        command.StudentID,
			"accountId":        command.AccountID,
			"teacherAccountId": command.TeacherAccountID,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			delegationID: denialDelegationID, action: "StudentCreated", targetType: "student",
			targetID: command.StudentID, reason: "student_create_not_allowed",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return result, err
}

func (s *Store) PublishFirstMinute(ctx context.Context, command core.PublishFirstMinuteCommand) (core.FirstMinute, error) {
	var result core.FirstMinute
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := lockAssignmentSubjects(ctx, tx, command.TenantID, []string{command.StudentID}); err != nil {
			return err
		}
		var version int64
		err := tx.QueryRow(ctx, `
			SELECT version
			FROM students
			WHERE tenant_id = $1 AND id = $2 AND status = 'active'
			FOR UPDATE
		`, command.TenantID, command.StudentID).Scan(&version)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "student not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock student for first minute: %w", err)
		}
		var assigned bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM teacher_assignments ta
				JOIN accounts a ON a.tenant_id = ta.tenant_id AND a.id = ta.teacher_account_id
				JOIN role_grants rg
				  ON rg.tenant_id = a.tenant_id AND rg.account_id = a.id
				WHERE ta.tenant_id = $1 AND ta.student_id = $2
				  AND ta.teacher_account_id = $3 AND ta.status = 'active'
				  AND ta.effective_from <= $4
				  AND (ta.effective_until IS NULL OR $4 < ta.effective_until)
				  AND a.status = 'active' AND rg.role_type = 'Teacher' AND rg.status = 'active'
			)
		`, command.TenantID, command.StudentID, command.ActorAccountID, command.Now).Scan(&assigned); err != nil {
			return fmt.Errorf("check assigned Teacher: %w", err)
		}
		if !assigned {
			return core.E(core.CodeForbidden, "only the assigned Teacher can publish this first minute", nil)
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "publish_first_minute", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.FirstMinute](claim)
			return err
		}
		if version != command.ExpectedVersion {
			return core.E(core.CodeConflict, "student version is stale", nil)
		}

		result = core.FirstMinute{
			StudentID:    command.StudentID,
			Revision:     version + 1,
			WhatWorked:   command.WhatWorked,
			CurrentFocus: command.CurrentFocus,
			NextStep:     command.NextStep,
			PublishedAt:  command.Now,
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO first_minute_revisions (
				id, tenant_id, student_id, revision, what_worked, current_focus,
				next_step, authored_by_account_id, published_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, command.RevisionID, command.TenantID, command.StudentID, result.Revision,
			command.WhatWorked, command.CurrentFocus, command.NextStep,
			command.ActorAccountID, command.Now); err != nil {
			return fmt.Errorf("insert first-minute revision: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE students
			SET version = $3, updated_at = $4
			WHERE tenant_id = $1 AND id = $2
		`, command.TenantID, command.StudentID, result.Revision, command.Now); err != nil {
			return fmt.Errorf("advance student version: %w", err)
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "publish_first_minute", command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "FirstBelcantoMinutePublished", targetType: "student",
			targetID: command.StudentID, decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"revision": result.Revision}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "FirstBelcantoMinutePublished", "student", command.StudentID, map[string]any{
			"studentId": command.StudentID,
			"revision":  result.Revision,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: "FirstBelcantoMinutePublished", targetType: "student",
			targetID: command.StudentID, reason: "assigned_teacher_required",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return result, err
}

func (s *Store) IssueInvitation(ctx context.Context, command core.IssueInvitationCommand) (core.InvitationResult, error) {
	var result core.InvitationResult
	scope := "issue_invitation:" + string(command.Mode)
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		owner, err := hasActiveRole(ctx, tx, command.TenantID, command.ActorAccountID, core.RoleOwner)
		if err != nil {
			return err
		}
		if !owner {
			return core.E(core.CodeForbidden, "only Owner can manage student invitations", nil)
		}
		if err := lockActivationSubject(ctx, tx, command.TenantID, "", command.StudentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, scope, command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			result, err = decodeReplay[core.InvitationResult](claim)
			return err
		}
		var accountID string
		var accountStatus string
		err = tx.QueryRow(ctx, `
			SELECT s.account_id, a.status
			FROM students s
			JOIN accounts a ON a.tenant_id = s.tenant_id AND a.id = s.account_id
			WHERE s.tenant_id = $1 AND s.id = $2 AND s.status = 'active'
			FOR UPDATE OF s, a
		`, command.TenantID, command.StudentID).Scan(&accountID, &accountStatus)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "student not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock student invitation target: %w", err)
		}
		if accountStatus != "pending_activation" {
			return core.E(core.CodeInvalidState, "student account is not pending activation", nil)
		}
		var firstMinuteExists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM first_minute_revisions
				WHERE tenant_id = $1 AND student_id = $2
			)
		`, command.TenantID, command.StudentID).Scan(&firstMinuteExists); err != nil {
			return fmt.Errorf("check First Belcanto Minute: %w", err)
		}
		if !firstMinuteExists {
			return core.E(core.CodeInvalidState, "First Belcanto Minute must be published before invitation", nil)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE activation_invitations
			SET status = 'superseded'
			WHERE tenant_id = $1 AND student_id = $2 AND status = 'issued'
			  AND expires_at <= $3
		`, command.TenantID, command.StudentID, command.Now); err != nil {
			return fmt.Errorf("expire old student invitations: %w", err)
		}
		var activeID string
		err = tx.QueryRow(ctx, `
			SELECT id
			FROM activation_invitations
			WHERE tenant_id = $1 AND student_id = $2 AND status = 'issued'
			FOR UPDATE
		`, command.TenantID, command.StudentID).Scan(&activeID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("lock active student invitation: %w", err)
		}
		if command.Mode == core.InvitationIssue && activeID != "" {
			return core.E(core.CodeConflict, "an active invitation already exists", nil)
		}
		if command.Mode == core.InvitationReissue && activeID == "" {
			return core.E(core.CodeInvalidState, "an active invitation is required for reissue", nil)
		}
		if command.Mode == core.InvitationReissue && activeID != "" {
			if _, err := tx.Exec(ctx, `
				UPDATE activation_invitations
				SET status = 'superseded'
				WHERE tenant_id = $1 AND id = $2 AND status = 'issued'
			`, command.TenantID, activeID); err != nil {
				return fmt.Errorf("supersede active invitation: %w", err)
			}
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO activation_invitations (
				id, tenant_id, account_id, student_id, kind, token_digest, status,
				issued_by_account_id, issued_at, expires_at
			) VALUES ($1, $2, $3, $4, 'student_activation', $5, 'issued', $6, $7, $8)
		`, command.InvitationID, command.TenantID, accountID, command.StudentID,
			command.TokenDigest, command.ActorAccountID, command.Now, command.ExpiresAt); err != nil {
			return mapWriteError(err, "an active invitation already exists")
		}
		if activeID != "" {
			if _, err := tx.Exec(ctx, `
				UPDATE activation_invitations
				SET superseded_by_id = $3
				WHERE tenant_id = $1 AND id = $2 AND status = 'superseded'
			`, command.TenantID, activeID, command.InvitationID); err != nil {
				return fmt.Errorf("link superseded invitation: %w", err)
			}
		}

		result = core.InvitationResult{
			InvitationID: command.InvitationID,
			StudentID:    command.StudentID,
			Status:       "issued",
			ExpiresAt:    command.ExpiresAt,
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, scope, command.IdempotencyKey, result, command.Now); err != nil {
			return err
		}
		action := "StudentActivationInvitationIssued"
		if command.Mode == core.InvitationReissue {
			action = "StudentActivationInvitationReissued"
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action: action, targetType: "invitation",
			targetID: command.InvitationID, decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"studentId": command.StudentID, "expiresAt": command.ExpiresAt}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, action, "invitation", command.InvitationID, map[string]any{
			"invitationId": command.InvitationID,
			"studentId":    command.StudentID,
			"expiresAt":    command.ExpiresAt,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action:     "StudentActivationInvitationIssued",
			targetType: "student", targetID: command.StudentID,
			reason: "owner_required", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return result, err
}

func (s *Store) RevokeInvitation(ctx context.Context, command core.RevokeInvitationCommand) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		owner, err := hasActiveRole(ctx, tx, command.TenantID, command.ActorAccountID, core.RoleOwner)
		if err != nil {
			return err
		}
		if !owner {
			return core.E(core.CodeForbidden, "only Owner can manage student invitations", nil)
		}
		var accountID string
		var studentID string
		err = tx.QueryRow(ctx, `
			SELECT account_id, student_id
			FROM activation_invitations
			WHERE tenant_id = $1 AND id = $2 AND kind = 'student_activation'
		`, command.TenantID, command.InvitationID).Scan(&accountID, &studentID)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "invitation not found", nil)
		}
		if err != nil {
			return fmt.Errorf("locate invitation subject: %w", err)
		}
		if err := lockActivationSubject(ctx, tx, command.TenantID, accountID, studentID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "revoke_invitation", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			return nil
		}
		var status string
		err = tx.QueryRow(ctx, `
			SELECT status
			FROM activation_invitations
			WHERE tenant_id = $1 AND id = $2 AND kind = 'student_activation'
			FOR UPDATE
		`, command.TenantID, command.InvitationID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "invitation not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock invitation: %w", err)
		}
		if status != "issued" {
			return core.E(core.CodeInvalidState, "only an issued invitation can be revoked", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_invitations
			SET status = 'revoked', revoked_at = $3
			WHERE tenant_id = $1 AND id = $2 AND status = 'issued'
		`, command.TenantID, command.InvitationID, command.Now); err != nil {
			return fmt.Errorf("revoke invitation: %w", err)
		}
		if err := completeIdempotency(ctx, tx, command.TenantID, command.ActorAccountID, "revoke_invitation", command.IdempotencyKey, struct{}{}, command.Now); err != nil {
			return err
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action:     "StudentActivationInvitationRevoked",
			targetType: "invitation", targetID: command.InvitationID, decision: "allow",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "StudentActivationInvitationRevoked", "invitation", command.InvitationID, map[string]any{
			"invitationId": command.InvitationID,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, actorID: command.ActorAccountID,
			action:     "StudentActivationInvitationRevoked",
			targetType: "invitation", targetID: command.InvitationID,
			reason: "owner_required", idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	}
	return err
}

func (s *Store) BootstrapView(ctx context.Context, principal core.Principal, now time.Time) (core.BootstrapView, error) {
	var status string
	if err := s.pool.QueryRow(ctx, `
		SELECT status FROM accounts WHERE tenant_id = $1 AND id = $2
	`, principal.TenantID, principal.AccountID).Scan(&status); errors.Is(err, pgx.ErrNoRows) || status != "active" {
		return core.BootstrapView{}, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	} else if err != nil {
		return core.BootstrapView{}, fmt.Errorf("read bootstrap account: %w", err)
	}

	roles, err := rolesForAccount(ctx, s.pool, principal.TenantID, principal.AccountID)
	if err != nil {
		return core.BootstrapView{}, err
	}
	view := core.BootstrapView{
		AccountID: principal.AccountID, Roles: roles,
		AccessProfiles: []string{}, Permissions: []string{},
	}
	view.Permissions = append(view.Permissions, core.LessonPermissionSetForRoles(roles)...)
	if principalHasRole(roles, core.RoleOwner) {
		view.Permissions = append(view.Permissions, core.OwnerStudentOnboardingPermissionSet()...)
	} else if principalHasRole(roles, core.RoleAdministrator) {
		var delegated bool
		if err := s.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM capability_delegations
				WHERE tenant_id = $1 AND grantee_account_id = $2
				  AND bundle = $3 AND status = 'active'
				  AND (expires_at IS NULL OR expires_at > $4)
			)
		`, principal.TenantID, principal.AccountID, core.StudentOnboardingManagerV1, now).Scan(&delegated); err != nil {
			return core.BootstrapView{}, fmt.Errorf("read effective access profile: %w", err)
		}
		if delegated {
			view.AccessProfiles = append(view.AccessProfiles, core.StudentOnboardingManagerV1)
			view.Permissions = append(view.Permissions, core.StudentOnboardingManagerV1PermissionSet()...)
		}
	}

	var studentID string
	var fullName string
	err = s.pool.QueryRow(ctx, `
		SELECT s.id, p.full_name
		FROM role_grants rg
		JOIN students s
		  ON s.tenant_id = rg.tenant_id AND s.id = rg.scope_id
		 AND s.account_id = rg.account_id
		JOIN people p ON p.tenant_id = s.tenant_id AND p.id = s.person_id
		WHERE rg.tenant_id = $1 AND rg.account_id = $2
		  AND rg.role_type = 'Student' AND rg.scope_type = 'student'
		  AND rg.status = 'active'
		ORDER BY rg.granted_at DESC
		LIMIT 1
	`, principal.TenantID, principal.AccountID).Scan(&studentID, &fullName)
	if errors.Is(err, pgx.ErrNoRows) {
		return view, nil
	}
	if err != nil {
		return core.BootstrapView{}, fmt.Errorf("read student bootstrap identity: %w", err)
	}
	view.StudentID = studentID
	view.FullName = fullName

	var firstMinute core.FirstMinute
	err = s.pool.QueryRow(ctx, `
		SELECT student_id, revision, what_worked, current_focus, next_step, published_at
		FROM first_minute_revisions
		WHERE tenant_id = $1 AND student_id = $2
		ORDER BY revision DESC
		LIMIT 1
	`, principal.TenantID, studentID).Scan(
		&firstMinute.StudentID,
		&firstMinute.Revision,
		&firstMinute.WhatWorked,
		&firstMinute.CurrentFocus,
		&firstMinute.NextStep,
		&firstMinute.PublishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return view, nil
	}
	if err != nil {
		return core.BootstrapView{}, fmt.Errorf("read latest First Belcanto Minute: %w", err)
	}
	view.FirstMinute = &firstMinute
	return view, nil
}

func (s *Store) ListStaff(ctx context.Context, principal core.Principal, role core.Role, now time.Time) ([]core.StaffMember, error) {
	result := make([]core.StaffMember, 0)
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		owner, err := hasActiveRole(ctx, tx, principal.TenantID, principal.AccountID, core.RoleOwner)
		if err != nil {
			return err
		}
		administrator, err := hasActiveRole(ctx, tx, principal.TenantID, principal.AccountID, core.RoleAdministrator)
		if err != nil {
			return err
		}
		allowed := owner || (administrator && role == core.RoleTeacher)
		delegationID := ""
		if !allowed {
			return core.E(core.CodeForbidden, "staff discovery permission is required", nil)
		}

		rows, err := tx.Query(ctx, `
			SELECT a.id, p.full_name
			FROM accounts a
			JOIN people p ON p.tenant_id = a.tenant_id AND p.id = a.person_id
			JOIN role_grants rg ON rg.tenant_id = a.tenant_id AND rg.account_id = a.id
			WHERE a.tenant_id = $1 AND a.status = 'active'
			  AND rg.role_type = $2 AND rg.status = 'active'
			ORDER BY p.full_name, a.id
		`, principal.TenantID, string(role))
		if err != nil {
			return fmt.Errorf("list active staff: %w", err)
		}
		for rows.Next() {
			var member core.StaffMember
			if err := rows.Scan(&member.AccountID, &member.FullName); err != nil {
				rows.Close()
				return fmt.Errorf("scan active staff: %w", err)
			}
			member.AccessProfiles = []string{}
			result = append(result, member)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("iterate active staff: %w", err)
		}
		rows.Close()

		for index := range result {
			roles, err := rolesForAccount(ctx, tx, principal.TenantID, result[index].AccountID)
			if err != nil {
				return err
			}
			result[index].Roles = roles
			if role == core.RoleAdministrator {
				var delegationID string
				var expiresAt *time.Time
				err := tx.QueryRow(ctx, `
					SELECT id, expires_at
					FROM capability_delegations
					WHERE tenant_id = $1 AND grantee_account_id = $2
					  AND bundle = $3 AND status = 'active'
					  AND (expires_at IS NULL OR expires_at > $4)
					ORDER BY granted_at DESC, id DESC
					LIMIT 1
				`, principal.TenantID, result[index].AccountID, core.StudentOnboardingManagerV1, now).Scan(&delegationID, &expiresAt)
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("read staff access profile: %w", err)
				}
				if err == nil {
					result[index].AccessProfiles = append(result[index].AccessProfiles, core.StudentOnboardingManagerV1)
					result[index].OnboardingDelegationID = delegationID
					result[index].OnboardingDelegationExpiresAt = expiresAt
				}
			}
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			delegationID: delegationID, action: "StaffListed", targetType: "staff_role",
			targetID: string(role), decision: "allow", at: now,
		})
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "StaffListed", targetType: "staff_role", targetID: string(role),
			reason: "staff_discovery_not_allowed", at: now,
		})
	}
	return result, err
}

func principalHasRole(roles []core.Role, role core.Role) bool {
	for _, candidate := range roles {
		if candidate == role {
			return true
		}
	}
	return false
}
