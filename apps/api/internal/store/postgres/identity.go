package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/jackc/pgx/v5"
)

func (s *Store) BootstrapOwner(ctx context.Context, command core.BootstrapOwnerCommand) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, command.TenantID); err != nil {
			return fmt.Errorf("lock owner bootstrap: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO tenants (id, name, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO NOTHING
		`, command.TenantID, command.TenantName, command.Now); err != nil {
			return fmt.Errorf("create tenant: %w", err)
		}

		var ownerExists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM role_grants
				WHERE tenant_id = $1 AND role_type = 'Owner' AND status = 'active'
			)
		`, command.TenantID).Scan(&ownerExists); err != nil {
			return fmt.Errorf("check existing Owner: %w", err)
		}
		if ownerExists {
			return core.E(core.CodeConflict, "owner is already bootstrapped", nil)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO people (id, tenant_id, full_name, created_at)
			VALUES ($1, $2, $3, $4)
		`, command.PersonID, command.TenantID, command.FullName, command.Now); err != nil {
			return mapWriteError(err, "owner identity already exists")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO school_memberships (id, tenant_id, person_id, created_at)
			VALUES ($1, $2, $3, $4)
		`, command.MembershipID, command.TenantID, command.PersonID, command.Now); err != nil {
			return fmt.Errorf("create Owner membership: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO accounts (
				id, tenant_id, person_id, status, created_at, updated_at
			) VALUES ($1, $2, $3, 'pending_activation', $4, $4)
		`, command.AccountID, command.TenantID, command.PersonID, command.Now); err != nil {
			return fmt.Errorf("create Owner account: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO account_login_identifiers (
				account_id, tenant_id, identifier_type, normalized_value, status, created_at
			) VALUES ($1, $2, 'phone', $3, 'reserved', $4)
		`, command.AccountID, command.TenantID, command.Phone, command.Now); err != nil {
			return mapWriteError(err, "login identifier is unavailable")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_grants (
				id, tenant_id, account_id, role_type, scope_type, scope_id,
				status, granted_by, granted_at
			) VALUES ($1, $2, $3, 'Owner', 'tenant', $2, 'active', NULL, $4)
		`, command.RoleGrantID, command.TenantID, command.AccountID, command.Now); err != nil {
			return fmt.Errorf("grant Owner role: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO activation_invitations (
				id, tenant_id, account_id, kind, token_digest,
				status, issued_at, expires_at
			) VALUES ($1, $2, $3, 'owner_bootstrap', $4, 'issued', $5, $6)
		`, command.InvitationID, command.TenantID, command.AccountID, command.TokenDigest,
			command.Now, command.ExpiresAt); err != nil {
			return fmt.Errorf("create Owner activation invitation: %w", err)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, operatorID: command.Operator, action: "OwnerBootstrapCreated",
			targetType: "account", targetID: command.AccountID, decision: "allow",
			reason: command.Reason, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "OwnerBootstrapCreated", "account", command.AccountID, map[string]any{
			"accountId": command.AccountID,
		}, command.Now)
	})
	if err != nil {
		return mapWriteError(err, "owner bootstrap conflicts with existing identity data")
	}
	return nil
}

func (s *Store) BootstrapStaff(ctx context.Context, command core.BootstrapStaffCommand) error {
	if command.Role != core.RoleAdministrator && command.Role != core.RoleTeacher {
		return core.E(core.CodeInvalidInput, "staff bootstrap role must be Administrator or Teacher", nil)
	}
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, command.TenantID+"\x00"+command.Phone); err != nil {
			return fmt.Errorf("lock staff bootstrap: %w", err)
		}
		owner, err := hasActiveRole(ctx, tx, command.TenantID, command.OwnerAccountID, core.RoleOwner)
		if err != nil {
			return err
		}
		if !owner {
			return core.E(core.CodeForbidden, "active Owner authorization is required", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO people (id, tenant_id, full_name, created_at)
			VALUES ($1, $2, $3, $4)
		`, command.PersonID, command.TenantID, command.FullName, command.Now); err != nil {
			return mapWriteError(err, "staff identity already exists")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO school_memberships (id, tenant_id, person_id, created_at)
			VALUES ($1, $2, $3, $4)
		`, command.MembershipID, command.TenantID, command.PersonID, command.Now); err != nil {
			return fmt.Errorf("create staff membership: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO accounts (
				id, tenant_id, person_id, status, created_at, updated_at
			) VALUES ($1, $2, $3, 'pending_activation', $4, $4)
		`, command.AccountID, command.TenantID, command.PersonID, command.Now); err != nil {
			return fmt.Errorf("create pending staff account: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO account_login_identifiers (
				account_id, tenant_id, identifier_type, normalized_value, status, created_at
			) VALUES ($1, $2, 'phone', $3, 'reserved', $4)
		`, command.AccountID, command.TenantID, command.Phone, command.Now); err != nil {
			return mapWriteError(err, "login identifier is unavailable")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_grants (
				id, tenant_id, account_id, role_type, scope_type, scope_id,
				status, granted_by, granted_at
			) VALUES ($1, $2, $3, $4, 'tenant', $2, 'active', $5, $6)
		`, command.RoleGrantID, command.TenantID, command.AccountID, string(command.Role),
			command.OwnerAccountID, command.Now); err != nil {
			return fmt.Errorf("grant staff role: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO activation_invitations (
				id, tenant_id, account_id, kind, token_digest, status,
				issued_by_account_id, issued_at, expires_at
			) VALUES ($1, $2, $3, 'staff_activation', $4, 'issued', $5, $6, $7)
		`, command.InvitationID, command.TenantID, command.AccountID, command.TokenDigest,
			command.OwnerAccountID, command.Now, command.ExpiresAt); err != nil {
			return fmt.Errorf("create staff activation invitation: %w", err)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, operatorID: command.Operator,
			action: "StaffBootstrapCreated", targetType: "account",
			targetID: command.AccountID, decision: "allow", reason: command.Reason,
			metadata: map[string]any{"role": command.Role, "ownerAccountId": command.OwnerAccountID}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "StaffBootstrapCreated", "account", command.AccountID, map[string]any{
			"accountId": command.AccountID,
			"role":      command.Role,
		}, command.Now)
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: command.TenantID, operatorID: command.Operator,
			action: "StaffBootstrapCreated", targetType: "account",
			targetID: command.AccountID, reason: "owner_required",
			metadata: map[string]any{"ownerAccountId": command.OwnerAccountID}, at: command.Now,
		})
	}
	return mapWriteError(err, "staff bootstrap conflicts with existing identity data")
}

func (s *Store) ReissueBootstrapInvitation(ctx context.Context, command core.ReissueBootstrapInvitationCommand) (core.BootstrapInvitationResult, error) {
	var result core.BootstrapInvitationResult
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := lockActivationSubject(ctx, tx, command.TenantID, command.AccountID, ""); err != nil {
			return err
		}
		var accountStatus string
		err := tx.QueryRow(ctx, `
			SELECT status
			FROM accounts
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, command.TenantID, command.AccountID).Scan(&accountStatus)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "pending bootstrap account not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock bootstrap recovery account: %w", err)
		}
		if accountStatus != "pending_activation" {
			return core.E(core.CodeInvalidState, "account is not pending activation", nil)
		}
		var kind string
		var issuerAccountID *string
		err = tx.QueryRow(ctx, `
			SELECT kind, issued_by_account_id
			FROM activation_invitations
			WHERE tenant_id = $1 AND account_id = $2
			  AND kind IN ('owner_bootstrap', 'staff_activation')
			ORDER BY issued_at DESC, id DESC
			LIMIT 1
			FOR UPDATE
		`, command.TenantID, command.AccountID).Scan(&kind, &issuerAccountID)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidState, "original bootstrap invitation not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock original bootstrap invitation: %w", err)
		}
		role := core.RoleOwner
		if kind == "staff_activation" {
			role = core.RoleAdministrator
		}
		var validRole bool
		if kind == "staff_activation" {
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM role_grants
					WHERE tenant_id = $1 AND account_id = $2
					  AND role_type IN ('Administrator', 'Teacher') AND status = 'active'
				)
			`, command.TenantID, command.AccountID).Scan(&validRole); err != nil {
				return fmt.Errorf("check staff recovery role: %w", err)
			}
		} else {
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM role_grants
					WHERE tenant_id = $1 AND account_id = $2
					  AND role_type = $3 AND status = 'active'
				)
			`, command.TenantID, command.AccountID, string(role)).Scan(&validRole); err != nil {
				return fmt.Errorf("check Owner recovery role: %w", err)
			}
		}
		if !validRole {
			return core.E(core.CodeInvalidState, "account is not a bootstrap or staff account", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO activation_invitations (
				id, tenant_id, account_id, kind, token_digest, status,
				issued_by_account_id, issued_at, expires_at
			) VALUES ($1, $2, $3, $4, $5, 'issued', $6, $7, $8)
		`, command.InvitationID, command.TenantID, command.AccountID, kind,
			command.TokenDigest, issuerAccountID, command.Now, command.ExpiresAt); err != nil {
			return mapWriteError(err, "bootstrap invitation recovery conflicts with existing data")
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_invitations
			SET status = 'superseded'
			WHERE tenant_id = $1 AND account_id = $2 AND id <> $3
			  AND kind = $4 AND status = 'issued'
		`, command.TenantID, command.AccountID, command.InvitationID, kind); err != nil {
			return fmt.Errorf("supersede bootstrap invitations: %w", err)
		}
		result = core.BootstrapInvitationResult{
			InvitationID: command.InvitationID, AccountID: command.AccountID,
			Kind: kind, Status: "issued", ExpiresAt: command.ExpiresAt,
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: command.TenantID, operatorID: command.Operator,
			action: "BootstrapActivationInvitationReissued", targetType: "invitation",
			targetID: command.InvitationID, decision: "allow", reason: command.Reason,
			metadata: map[string]any{"accountId": command.AccountID, "kind": kind}, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, command.TenantID, "BootstrapActivationInvitationReissued", "invitation", command.InvitationID, map[string]any{
			"invitationId": command.InvitationID, "accountId": command.AccountID, "kind": kind,
		}, command.Now)
	})
	return result, err
}

func (s *Store) PreviewActivation(ctx context.Context, digest []byte, now time.Time) (core.ActivationPreview, error) {
	var preview core.ActivationPreview
	var phone string
	err := s.pool.QueryRow(ctx, `
		SELECT i.id, i.kind, p.full_name, li.normalized_value, i.expires_at
		FROM activation_invitations i
		JOIN accounts a ON a.tenant_id = i.tenant_id AND a.id = i.account_id
		JOIN people p ON p.tenant_id = a.tenant_id AND p.id = a.person_id
		JOIN account_login_identifiers li
		  ON li.tenant_id = a.tenant_id AND li.account_id = a.id
		 AND li.identifier_type = 'phone'
		WHERE i.token_digest = $1 AND i.status = 'issued' AND i.expires_at > $2
		  AND a.status = 'pending_activation' AND li.status = 'reserved'
	`, digest, now).Scan(
		&preview.InvitationID,
		&preview.Kind,
		&preview.DisplayName,
		&phone,
		&preview.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.ActivationPreview{}, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	if err != nil {
		return core.ActivationPreview{}, fmt.Errorf("preview activation: %w", err)
	}
	preview.MaskedPhone = security.MaskPhone(phone)
	return preview, nil
}

func (s *Store) ValidateActivation(ctx context.Context, command core.ActivationValidationCommand) (bool, error) {
	var status string
	var expiresAt time.Time
	var accountStatus string
	var expectedPhone string
	var storedKey *string
	var storedFingerprint []byte
	err := s.pool.QueryRow(ctx, `
		SELECT i.status, i.expires_at, a.status, li.normalized_value,
		       i.consumed_idempotency_key, i.consumed_payload_fingerprint
		FROM activation_invitations i
		JOIN accounts a ON a.tenant_id = i.tenant_id AND a.id = i.account_id
		JOIN account_login_identifiers li
		  ON li.tenant_id = a.tenant_id AND li.account_id = a.id
		 AND li.identifier_type = 'phone'
		WHERE i.token_digest = $1
	`, command.TokenDigest).Scan(
		&status, &expiresAt, &accountStatus, &expectedPhone,
		&storedKey, &storedFingerprint,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	if err != nil {
		return false, fmt.Errorf("validate activation: %w", err)
	}
	if status == "consumed" && storedKey != nil && *storedKey == command.IdempotencyKey {
		if !security.EqualDigest(storedFingerprint, command.PayloadFingerprint) {
			return false, core.E(core.CodeConflict, "Idempotency-Key was reused with a different payload", nil)
		}
		return true, nil
	}
	if status != "issued" || !expiresAt.After(command.Now) || accountStatus != "pending_activation" || expectedPhone != command.Phone {
		return false, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	return false, nil
}

func (s *Store) CompleteActivation(ctx context.Context, command core.ActivationCompleteCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var subjectTenantID string
		var subjectAccountID string
		var subjectStudentID *string
		err := tx.QueryRow(ctx, `
			SELECT tenant_id, account_id, student_id
			FROM activation_invitations
			WHERE token_digest = $1
		`, command.TokenDigest).Scan(&subjectTenantID, &subjectAccountID, &subjectStudentID)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}
		if err != nil {
			return fmt.Errorf("locate activation subject: %w", err)
		}
		studentID := ""
		if subjectStudentID != nil {
			studentID = *subjectStudentID
		}
		if err := lockActivationSubject(ctx, tx, subjectTenantID, subjectAccountID, studentID); err != nil {
			return err
		}

		var invitationID string
		var tenantID string
		var accountID string
		var status string
		var expiresAt time.Time
		var storedKey *string
		var storedFingerprint []byte
		err = tx.QueryRow(ctx, `
			SELECT id, tenant_id, account_id, status, expires_at,
			       consumed_idempotency_key, consumed_payload_fingerprint
			FROM activation_invitations
			WHERE token_digest = $1
			FOR UPDATE
		`, command.TokenDigest).Scan(
			&invitationID,
			&tenantID,
			&accountID,
			&status,
			&expiresAt,
			&storedKey,
			&storedFingerprint,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}
		if err != nil {
			return fmt.Errorf("lock activation invitation: %w", err)
		}
		if status == "consumed" && storedKey != nil && *storedKey == command.IdempotencyKey {
			if !security.EqualDigest(storedFingerprint, command.PayloadFingerprint) {
				return core.E(core.CodeConflict, "Idempotency-Key was reused with a different payload", nil)
			}
			return nil
		}
		if status != "issued" || !expiresAt.After(command.Now) {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}

		var accountStatus string
		var expectedPhone string
		err = tx.QueryRow(ctx, `
			SELECT a.status, li.normalized_value
			FROM accounts a
			JOIN account_login_identifiers li
			  ON li.tenant_id = a.tenant_id AND li.account_id = a.id
			 AND li.identifier_type = 'phone'
			WHERE a.tenant_id = $1 AND a.id = $2
			FOR UPDATE OF a, li
		`, tenantID, accountID).Scan(&accountStatus, &expectedPhone)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}
		if err != nil {
			return fmt.Errorf("lock activation account: %w", err)
		}
		if accountStatus != "pending_activation" || expectedPhone != command.Phone {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO password_credentials (
				account_id, password_hash, algorithm, created_at, updated_at
			) VALUES ($1, $2, 'argon2id', $3, $3)
		`, accountID, command.PasswordHash, command.Now); err != nil {
			return fmt.Errorf("create password credential: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE accounts
			SET status = 'active', activated_at = $3, updated_at = $3, version = version + 1
			WHERE tenant_id = $1 AND id = $2 AND status = 'pending_activation'
		`, tenantID, accountID, command.Now); err != nil {
			return fmt.Errorf("activate account: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE account_login_identifiers
			SET status = 'confirmed', confirmed_at = $3
			WHERE tenant_id = $1 AND account_id = $2 AND identifier_type = 'phone'
		`, tenantID, accountID, command.Now); err != nil {
			return fmt.Errorf("confirm login identifier: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_invitations
			SET status = 'consumed', consumed_at = $3,
			    consumed_idempotency_key = $4, consumed_payload_fingerprint = $5
			WHERE tenant_id = $1 AND id = $2
		`, tenantID, invitationID, command.Now, command.IdempotencyKey, command.PayloadFingerprint); err != nil {
			return fmt.Errorf("consume activation invitation: %w", err)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: tenantID, actorID: accountID, action: "AccountActivated",
			targetType: "account", targetID: accountID, decision: "allow",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, tenantID, "AccountActivated", "account", accountID, map[string]any{
			"accountId": accountID,
		}, command.Now)
	})
}

func (s *Store) CredentialByPhone(ctx context.Context, phone string) (core.CredentialRecord, error) {
	var record core.CredentialRecord
	err := s.pool.QueryRow(ctx, `
		SELECT a.id, a.tenant_id, li.normalized_value, pc.password_hash, a.status
		FROM account_login_identifiers li
		JOIN accounts a ON a.tenant_id = li.tenant_id AND a.id = li.account_id
		JOIN password_credentials pc ON pc.account_id = a.id
		WHERE li.identifier_type = 'phone' AND li.normalized_value = $1
		  AND li.status = 'confirmed'
	`, phone).Scan(&record.AccountID, &record.TenantID, &record.Phone, &record.PasswordHash, &record.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.CredentialRecord{}, core.E(core.CodeNotFound, "credential not found", nil)
	}
	if err != nil {
		return core.CredentialRecord{}, fmt.Errorf("read credential by phone: %w", err)
	}
	record.Roles, err = rolesForAccount(ctx, s.pool, record.TenantID, record.AccountID)
	if err != nil {
		return core.CredentialRecord{}, err
	}
	return record, nil
}

func (s *Store) CredentialByAccount(ctx context.Context, accountID string) (core.CredentialRecord, error) {
	var record core.CredentialRecord
	err := s.pool.QueryRow(ctx, `
		SELECT a.id, a.tenant_id, li.normalized_value, pc.password_hash, a.status
		FROM accounts a
		JOIN account_login_identifiers li
		  ON li.tenant_id = a.tenant_id AND li.account_id = a.id
		 AND li.identifier_type = 'phone'
		JOIN password_credentials pc ON pc.account_id = a.id
		WHERE a.id = $1
	`, accountID).Scan(&record.AccountID, &record.TenantID, &record.Phone, &record.PasswordHash, &record.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.CredentialRecord{}, core.E(core.CodeNotFound, "credential not found", nil)
	}
	if err != nil {
		return core.CredentialRecord{}, fmt.Errorf("read credential by account: %w", err)
	}
	record.Roles, err = rolesForAccount(ctx, s.pool, record.TenantID, record.AccountID)
	if err != nil {
		return core.CredentialRecord{}, err
	}
	return record, nil
}

func (s *Store) CreateSession(ctx context.Context, accountID, tenantID string, material core.SessionMaterial) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var active bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM accounts
				WHERE tenant_id = $1 AND id = $2 AND status = 'active'
			)
		`, tenantID, accountID).Scan(&active); err != nil {
			return fmt.Errorf("check session account: %w", err)
		}
		if !active {
			return core.E(core.CodeUnauthenticated, "account cannot create a session", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO sessions (
				id, family_id, tenant_id, account_id, access_digest, refresh_digest,
				status, access_expires_at, refresh_expires_at, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8, $9)
		`, material.SessionID, material.FamilyID, tenantID, accountID,
			material.AccessDigest, material.RefreshDigest, material.AccessExpiresAt,
			material.RefreshExpiresAt, material.CreatedAt); err != nil {
			return fmt.Errorf("create session: %w", err)
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: tenantID, actorID: accountID, action: "SessionCreated",
			targetType: "session", targetID: material.SessionID, decision: "allow",
			at: material.CreatedAt,
		})
	})
}

func (s *Store) PrincipalByAccessDigest(ctx context.Context, digest []byte, now time.Time) (core.Principal, error) {
	var principal core.Principal
	err := s.pool.QueryRow(ctx, `
		SELECT s.account_id, s.tenant_id, s.id
		FROM sessions s
		JOIN accounts a ON a.tenant_id = s.tenant_id AND a.id = s.account_id
		WHERE s.access_digest = $1 AND s.status = 'active'
		  AND s.access_expires_at > $2 AND a.status = 'active'
	`, digest, now).Scan(&principal.AccountID, &principal.TenantID, &principal.SessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.Principal{}, core.E(core.CodeUnauthenticated, "session is inactive", nil)
	}
	if err != nil {
		return core.Principal{}, fmt.Errorf("read access session: %w", err)
	}
	principal.Roles, err = rolesForAccount(ctx, s.pool, principal.TenantID, principal.AccountID)
	if err != nil {
		return core.Principal{}, err
	}
	return principal, nil
}

func (s *Store) RotateSession(ctx context.Context, oldRefreshDigest []byte, material core.SessionMaterial, now time.Time) (string, string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("begin refresh rotation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var oldID string
	var familyID string
	var tenantID string
	var accountID string
	var status string
	var refreshExpiresAt time.Time
	var accountStatus string
	err = tx.QueryRow(ctx, `
		SELECT s.id, s.family_id, s.tenant_id, s.account_id, s.status,
		       s.refresh_expires_at, a.status
		FROM sessions s
		JOIN accounts a ON a.tenant_id = s.tenant_id AND a.id = s.account_id
		WHERE s.refresh_digest = $1
		FOR UPDATE OF s, a
	`, oldRefreshDigest).Scan(
		&oldID,
		&familyID,
		&tenantID,
		&accountID,
		&status,
		&refreshExpiresAt,
		&accountStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", core.E(core.CodeUnauthenticated, "refresh token not found", nil)
	}
	if err != nil {
		return "", "", fmt.Errorf("lock refresh session: %w", err)
	}

	if status != "active" || !refreshExpiresAt.After(now) || accountStatus != "active" {
		if _, err := tx.Exec(ctx, `
			UPDATE sessions
			SET status = 'revoked', revoked_at = COALESCE(revoked_at, $2)
			WHERE tenant_id = $3 AND account_id = $4 AND family_id = $1
			  AND status <> 'revoked'
		`, familyID, now, tenantID, accountID); err != nil {
			return "", "", fmt.Errorf("revoke reused session family: %w", err)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: tenantID, actorID: accountID, action: "RefreshTokenReuseDetected",
			targetType: "session_family", targetID: familyID, decision: "deny",
			reason: "inactive_or_reused_refresh_token", at: now,
		}); err != nil {
			return "", "", err
		}
		if err := tx.Commit(ctx); err != nil {
			return "", "", fmt.Errorf("commit reused session revocation: %w", err)
		}
		return "", "", core.E(core.CodeUnauthenticated, "refresh token cannot be reused", nil)
	}

	material.FamilyID = familyID
	if _, err := tx.Exec(ctx, `
		INSERT INTO sessions (
			id, family_id, tenant_id, account_id, access_digest, refresh_digest,
			status, access_expires_at, refresh_expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8, $9)
	`, material.SessionID, familyID, tenantID, accountID, material.AccessDigest,
		material.RefreshDigest, material.AccessExpiresAt, material.RefreshExpiresAt,
		material.CreatedAt); err != nil {
		return "", "", fmt.Errorf("create replacement session: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET status = 'replaced', replaced_by_id = $3
		WHERE id = $1 AND family_id = $2 AND status = 'active'
	`, oldID, familyID, material.SessionID); err != nil {
		return "", "", fmt.Errorf("replace refresh session: %w", err)
	}
	if err := appendAudit(ctx, tx, auditInput{
		tenantID: tenantID, actorID: accountID, action: "SessionRefreshed",
		targetType: "session", targetID: material.SessionID, decision: "allow",
		at: material.CreatedAt,
	}); err != nil {
		return "", "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("commit refresh rotation: %w", err)
	}
	return accountID, tenantID, nil
}

func (s *Store) RevokeSession(ctx context.Context, accessDigest []byte, now time.Time) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var tenantID string
		var accountID string
		var sessionID string
		var familyID string
		err := tx.QueryRow(ctx, `
			SELECT tenant_id, account_id, id, family_id
			FROM sessions
			WHERE access_digest = $1
			FOR UPDATE
		`, accessDigest).Scan(&tenantID, &accountID, &sessionID, &familyID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("revoke session: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE sessions
			SET status = 'revoked', revoked_at = COALESCE(revoked_at, $4)
			WHERE tenant_id = $1 AND account_id = $2 AND family_id = $3
			  AND status <> 'revoked'
		`, tenantID, accountID, familyID, now); err != nil {
			return fmt.Errorf("revoke session family: %w", err)
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: tenantID, actorID: accountID, action: "SessionRevoked",
			targetType: "session_family", targetID: familyID, decision: "allow",
			metadata: map[string]any{"sessionId": sessionID}, at: now,
		})
	})
}
