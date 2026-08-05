package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// P.1 contacts, activation progress and TOTP 2FA (Figma Page 32).

func (s *Store) ListVerifiedContacts(ctx context.Context, principal core.Principal) ([]core.VerifiedContact, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, contact_kind, normalized_value, verified_at
		FROM verified_contacts
		WHERE tenant_id = $1 AND account_id = $2
		ORDER BY contact_kind
	`, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, fmt.Errorf("list verified contacts: %w", err)
	}
	defer rows.Close()
	contacts := []core.VerifiedContact{}
	for rows.Next() {
		var contact core.VerifiedContact
		var kind string
		if err := rows.Scan(&contact.ID, &kind, &contact.Value, &contact.VerifiedAt); err != nil {
			return nil, fmt.Errorf("scan verified contact: %w", err)
		}
		contact.Kind = core.ContactKind(kind)
		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate verified contacts: %w", err)
	}
	return contacts, nil
}

func (s *Store) StartContactChange(ctx context.Context, command core.StartContactChangeCommand) error {
	principal := command.Principal
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			UPDATE contact_verifications
			SET superseded_at = $3
			WHERE tenant_id = $1 AND account_id = $2 AND purpose = 'contact_change'
			  AND consumed_at IS NULL AND superseded_at IS NULL
		`, principal.TenantID, principal.AccountID, command.Now); err != nil {
			return fmt.Errorf("supersede contact verifications: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO contact_verifications (
				id, tenant_id, account_id, contact_kind, normalized_value, purpose,
				code_digest, expires_at, attempts_remaining, created_at
			) VALUES ($1, $2, $3, $4, $5, 'contact_change', $6, $7, 5, $8)
		`, command.VerificationID, principal.TenantID, principal.AccountID,
			string(command.Kind), command.Value, command.CodeDigest,
			command.ExpiresAt, command.Now); err != nil {
			return mapWriteError(err, "contact verification could not be created")
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "ContactChangeStarted", targetType: "contact_verification",
			targetID: command.VerificationID, decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "ContactVerificationRequested",
			"contact_verification", command.VerificationID,
			map[string]any{"verificationId": command.VerificationID}, command.Now)
	})
}

func consumeOpenVerification(ctx context.Context, tx pgx.Tx, tenantID, accountID, purpose string, codeDigest []byte, now time.Time) (core.ContactKind, string, error) {
	var id string
	var kind string
	var value string
	var storedDigest []byte
	var expiresAt time.Time
	var attempts int
	err := tx.QueryRow(ctx, `
		SELECT id, contact_kind, normalized_value, code_digest, expires_at, attempts_remaining
		FROM contact_verifications
		WHERE tenant_id = $1 AND account_id = $2 AND purpose = $3
		  AND consumed_at IS NULL AND superseded_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`, tenantID, accountID, purpose).Scan(&id, &kind, &value, &storedDigest, &expiresAt, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", core.E(core.CodeInvalidState, "no confirmation is in progress", nil)
	}
	if err != nil {
		return "", "", fmt.Errorf("lock contact verification: %w", err)
	}
	if !expiresAt.After(now) || attempts <= 0 {
		return "", "", core.E(core.CodeInvalidState, "no confirmation is in progress", nil)
	}
	if !security.EqualDigest(storedDigest, codeDigest) {
		if _, err := tx.Exec(ctx, `
			UPDATE contact_verifications
			SET attempts_remaining = attempts_remaining - 1,
			    superseded_at = CASE WHEN attempts_remaining - 1 <= 0 THEN $2 ELSE superseded_at END
			WHERE id = $1
		`, id, now); err != nil {
			return "", "", fmt.Errorf("record failed confirmation attempt: %w", err)
		}
		return "", "", core.E(core.CodeInvalidInput, "confirmation code is incorrect", nil)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE contact_verifications SET consumed_at = $2 WHERE id = $1
	`, id, now); err != nil {
		return "", "", fmt.Errorf("consume contact verification: %w", err)
	}
	return core.ContactKind(kind), value, nil
}

func upsertVerifiedContact(ctx context.Context, tx pgx.Tx, tenantID, accountID string, kind core.ContactKind, value string, now time.Time) (core.VerifiedContact, error) {
	contactID, err := security.NewID("contact")
	if err != nil {
		return core.VerifiedContact{}, fmt.Errorf("mint contact id: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO verified_contacts (id, tenant_id, account_id, contact_kind, normalized_value, verified_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, account_id, contact_kind)
		DO UPDATE SET normalized_value = EXCLUDED.normalized_value, verified_at = EXCLUDED.verified_at
	`, contactID, tenantID, accountID, string(kind), value, now); err != nil {
		return core.VerifiedContact{}, fmt.Errorf("upsert verified contact: %w", err)
	}
	var contact core.VerifiedContact
	var storedKind string
	if err := tx.QueryRow(ctx, `
		SELECT id, contact_kind, normalized_value, verified_at
		FROM verified_contacts
		WHERE tenant_id = $1 AND account_id = $2 AND contact_kind = $3
	`, tenantID, accountID, string(kind)).Scan(&contact.ID, &storedKind, &contact.Value, &contact.VerifiedAt); err != nil {
		return core.VerifiedContact{}, fmt.Errorf("read verified contact: %w", err)
	}
	contact.Kind = core.ContactKind(storedKind)
	if err := appendAudit(ctx, tx, auditInput{
		tenantID: tenantID, actorID: accountID,
		action: "ContactVerified", targetType: "verified_contact", targetID: contact.ID,
		decision: "allow", at: now,
	}); err != nil {
		return core.VerifiedContact{}, err
	}
	return contact, nil
}

func (s *Store) ConfirmContactChange(ctx context.Context, command core.ConfirmContactChangeCommand) (core.VerifiedContact, error) {
	principal := command.Principal
	var contact core.VerifiedContact
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		kind, value, err := consumeOpenVerification(ctx, tx, principal.TenantID, principal.AccountID,
			"contact_change", command.CodeDigest, command.Now)
		if err != nil {
			return err
		}
		contact, err = upsertVerifiedContact(ctx, tx, principal.TenantID, principal.AccountID, kind, value, command.Now)
		return err
	})
	if err != nil {
		return core.VerifiedContact{}, err
	}
	return contact, nil
}

func (s *Store) TwofaStatus(ctx context.Context, principal core.Principal) (core.TwofaStatus, error) {
	status := core.TwofaStatus{}
	var confirmedAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT confirmed_at FROM twofa_totp_secrets
		WHERE tenant_id = $1 AND account_id = $2
	`, principal.TenantID, principal.AccountID).Scan(&confirmedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return core.TwofaStatus{}, fmt.Errorf("read twofa secret: %w", err)
	}
	if err == nil && confirmedAt != nil {
		status.Enabled = true
		status.ConfirmedAt = confirmedAt
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT count(*) FROM twofa_recovery_codes
		WHERE tenant_id = $1 AND account_id = $2 AND used_at IS NULL AND superseded_at IS NULL
	`, principal.TenantID, principal.AccountID).Scan(&status.RecoveryCodesRemaining); err != nil {
		return core.TwofaStatus{}, fmt.Errorf("count recovery codes: %w", err)
	}
	return status, nil
}

func (s *Store) TwofaSecret(ctx context.Context, tenantID, accountID string) (core.TwofaSecretRecord, error) {
	var record core.TwofaSecretRecord
	var confirmedAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT secret_ciphertext, confirmed_at FROM twofa_totp_secrets
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(&record.Ciphertext, &confirmedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.TwofaSecretRecord{}, core.E(core.CodeNotFound, "two-factor authentication is not set up", nil)
	}
	if err != nil {
		return core.TwofaSecretRecord{}, fmt.Errorf("read twofa secret: %w", err)
	}
	record.Confirmed = confirmedAt != nil
	return record, nil
}

func (s *Store) StartTwofaEnrollment(ctx context.Context, command core.StartTwofaEnrollmentCommand) error {
	principal := command.Principal
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var confirmedAt *time.Time
		err := tx.QueryRow(ctx, `
			SELECT confirmed_at FROM twofa_totp_secrets
			WHERE tenant_id = $1 AND account_id = $2
			FOR UPDATE
		`, principal.TenantID, principal.AccountID).Scan(&confirmedAt)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("lock twofa secret: %w", err)
		}
		if err == nil && confirmedAt != nil {
			return core.E(core.CodeConflict, "two-factor authentication is already enabled", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO twofa_totp_secrets (tenant_id, account_id, secret_ciphertext, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $4)
			ON CONFLICT (tenant_id, account_id)
			DO UPDATE SET secret_ciphertext = EXCLUDED.secret_ciphertext,
			              confirmed_at = NULL, updated_at = EXCLUDED.updated_at
		`, principal.TenantID, principal.AccountID, command.SecretCiphertext, command.Now); err != nil {
			return fmt.Errorf("store twofa enrollment: %w", err)
		}
		return nil
	})
}

func replaceRecoveryCodes(ctx context.Context, tx pgx.Tx, tenantID, accountID string, digests [][]byte, now time.Time) error {
	if _, err := tx.Exec(ctx, `
		UPDATE twofa_recovery_codes
		SET superseded_at = $3
		WHERE tenant_id = $1 AND account_id = $2 AND used_at IS NULL AND superseded_at IS NULL
	`, tenantID, accountID, now); err != nil {
		return fmt.Errorf("supersede recovery codes: %w", err)
	}
	for _, digest := range digests {
		if _, err := tx.Exec(ctx, `
			INSERT INTO twofa_recovery_codes (tenant_id, account_id, code_digest, created_at)
			VALUES ($1, $2, $3, $4)
		`, tenantID, accountID, digest, now); err != nil {
			return mapWriteError(err, "recovery code could not be stored")
		}
	}
	return nil
}

func (s *Store) ConfirmTwofaEnrollment(ctx context.Context, command core.ConfirmTwofaEnrollmentCommand) error {
	principal := command.Principal
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var confirmedAt *time.Time
		err := tx.QueryRow(ctx, `
			SELECT confirmed_at FROM twofa_totp_secrets
			WHERE tenant_id = $1 AND account_id = $2
			FOR UPDATE
		`, principal.TenantID, principal.AccountID).Scan(&confirmedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidState, "two-factor enrollment has not started", nil)
		}
		if err != nil {
			return fmt.Errorf("lock twofa secret: %w", err)
		}
		if confirmedAt != nil {
			return core.E(core.CodeConflict, "two-factor authentication is already enabled", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE twofa_totp_secrets SET confirmed_at = $3, updated_at = $3
			WHERE tenant_id = $1 AND account_id = $2
		`, principal.TenantID, principal.AccountID, command.Now); err != nil {
			return fmt.Errorf("confirm twofa enrollment: %w", err)
		}
		if err := replaceRecoveryCodes(ctx, tx, principal.TenantID, principal.AccountID, command.RecoveryDigests, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "TwofaEnrolled", targetType: "account", targetID: principal.AccountID,
			decision: "allow", at: command.Now,
		})
	})
}

func (s *Store) DisableTwofa(ctx context.Context, command core.DisableTwofaCommand) error {
	principal := command.Principal
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `
			DELETE FROM twofa_totp_secrets
			WHERE tenant_id = $1 AND account_id = $2 AND confirmed_at IS NOT NULL
		`, principal.TenantID, principal.AccountID)
		if err != nil {
			return fmt.Errorf("delete twofa secret: %w", err)
		}
		if result.RowsAffected() == 0 {
			return core.E(core.CodeInvalidState, "two-factor authentication is not enabled", nil)
		}
		if err := replaceRecoveryCodes(ctx, tx, principal.TenantID, principal.AccountID, nil, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "TwofaDisabled", targetType: "account", targetID: principal.AccountID,
			decision: "allow", at: command.Now,
		})
	})
}

func (s *Store) CreateTwofaChallenge(ctx context.Context, command core.CreateTwofaChallengeCommand) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO twofa_challenges (
			id, tenant_id, account_id, token_digest, device_label, platform,
			expires_at, attempts_remaining, created_at
		) VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, 5, $8)
	`, command.ChallengeID, command.TenantID, command.AccountID, command.TokenDigest,
		command.DeviceLabel, command.Platform, command.ExpiresAt, command.Now)
	if err != nil {
		return mapWriteError(err, "second-factor challenge could not be created")
	}
	return nil
}

func (s *Store) TwofaChallengeByDigest(ctx context.Context, digest []byte, now time.Time) (core.TwofaChallengeRecord, error) {
	var record core.TwofaChallengeRecord
	err := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, COALESCE(device_label, ''), COALESCE(platform, ''), attempts_remaining
		FROM twofa_challenges
		WHERE token_digest = $1 AND consumed_at IS NULL AND expires_at > $2 AND attempts_remaining > 0
	`, digest, now).Scan(&record.ID, &record.TenantID, &record.AccountID,
		&record.DeviceLabel, &record.Platform, &record.AttemptsRemaining)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.TwofaChallengeRecord{}, core.E(core.CodeUnauthenticated, "second-factor challenge is invalid or expired", nil)
	}
	if err != nil {
		return core.TwofaChallengeRecord{}, fmt.Errorf("read twofa challenge: %w", err)
	}
	return record, nil
}

func (s *Store) ConsumeTwofaChallenge(ctx context.Context, digest []byte, now time.Time) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE twofa_challenges SET consumed_at = $2
		WHERE token_digest = $1 AND consumed_at IS NULL AND expires_at > $2 AND attempts_remaining > 0
	`, digest, now)
	if err != nil {
		return fmt.Errorf("consume twofa challenge: %w", err)
	}
	if result.RowsAffected() == 0 {
		return core.E(core.CodeUnauthenticated, "second-factor challenge is invalid or expired", nil)
	}
	return nil
}

func (s *Store) FailTwofaChallenge(ctx context.Context, digest []byte, now time.Time) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var id string
		var tenantID string
		var accountID string
		var attempts int
		err := tx.QueryRow(ctx, `
			SELECT id, tenant_id, account_id, attempts_remaining
			FROM twofa_challenges
			WHERE token_digest = $1 AND consumed_at IS NULL AND expires_at > $2 AND attempts_remaining > 0
			FOR UPDATE
		`, digest, now).Scan(&id, &tenantID, &accountID, &attempts)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("lock twofa challenge: %w", err)
		}
		attempts--
		if attempts <= 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE twofa_challenges SET attempts_remaining = 0, consumed_at = $2 WHERE id = $1
			`, id, now); err != nil {
				return fmt.Errorf("burn twofa challenge: %w", err)
			}
			return appendAudit(ctx, tx, auditInput{
				tenantID: tenantID, actorID: accountID,
				action: "TwofaChallengeFailed", targetType: "twofa_challenge", targetID: id,
				decision: "deny", reason: "second_factor_attempts_exhausted", at: now,
			})
		}
		if _, err := tx.Exec(ctx, `
			UPDATE twofa_challenges SET attempts_remaining = $2 WHERE id = $1
		`, id, attempts); err != nil {
			return fmt.Errorf("record failed twofa attempt: %w", err)
		}
		return nil
	})
}

func (s *Store) TryConsumeRecoveryCode(ctx context.Context, tenantID, accountID string, digest []byte, now time.Time) (bool, error) {
	result, err := s.pool.Exec(ctx, `
		UPDATE twofa_recovery_codes SET used_at = $4
		WHERE tenant_id = $1 AND account_id = $2 AND code_digest = $3
		  AND used_at IS NULL AND superseded_at IS NULL
	`, tenantID, accountID, digest, now)
	if err != nil {
		return false, fmt.Errorf("consume recovery code: %w", err)
	}
	return result.RowsAffected() > 0, nil
}

type activationRow struct {
	InvitationID string
	TenantID     string
	AccountID    string
	Kind         string
	DisplayName  string
	Phone        string
	ExpiresAt    time.Time
}

func lockActivationRow(ctx context.Context, tx pgx.Tx, digest []byte, now time.Time) (activationRow, error) {
	var row activationRow
	err := tx.QueryRow(ctx, `
		SELECT i.id, i.tenant_id, i.account_id, i.kind, p.full_name, li.normalized_value, i.expires_at
		FROM activation_invitations i
		JOIN accounts a ON a.tenant_id = i.tenant_id AND a.id = i.account_id
		JOIN people p ON p.tenant_id = a.tenant_id AND p.id = a.person_id
		JOIN account_login_identifiers li
		  ON li.tenant_id = a.tenant_id AND li.account_id = a.id
		 AND li.identifier_type = 'phone'
		WHERE i.token_digest = $1 AND i.status = 'issued' AND i.expires_at > $2
		  AND a.status = 'pending_activation'
		FOR UPDATE OF i, a
	`, digest, now).Scan(&row.InvitationID, &row.TenantID, &row.AccountID,
		&row.Kind, &row.DisplayName, &row.Phone, &row.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return activationRow{}, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	if err != nil {
		return activationRow{}, fmt.Errorf("lock activation invitation: %w", err)
	}
	return activationRow{
		InvitationID: row.InvitationID, TenantID: row.TenantID, AccountID: row.AccountID,
		Kind: row.Kind, DisplayName: row.DisplayName, Phone: row.Phone, ExpiresAt: row.ExpiresAt,
	}, nil
}

func ensureActivationProgress(ctx context.Context, tx pgx.Tx, row activationRow, now time.Time) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO activation_progress (invitation_id, tenant_id, account_id, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (invitation_id) DO NOTHING
	`, row.InvitationID, row.TenantID, row.AccountID, now)
	if err != nil {
		return fmt.Errorf("ensure activation progress: %w", err)
	}
	return nil
}

func (s *Store) ActivationProgress(ctx context.Context, digest []byte, now time.Time) (core.ActivationProgressView, error) {
	var view core.ActivationProgressView
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := lockActivationRow(ctx, tx, digest, now)
		if err != nil {
			return err
		}
		if err := ensureActivationProgress(ctx, tx, row, now); err != nil {
			return err
		}
		var passwordSetAt, contactVerifiedAt, twofaEnrolledAt, completedAt *time.Time
		var contactKind, contactValue *string
		if err := tx.QueryRow(ctx, `
			SELECT password_set_at, contact_kind, contact_value, contact_verified_at,
			       twofa_enrolled_at, completed_at
			FROM activation_progress WHERE invitation_id = $1
		`, row.InvitationID).Scan(&passwordSetAt, &contactKind, &contactValue,
			&contactVerifiedAt, &twofaEnrolledAt, &completedAt); err != nil {
			return fmt.Errorf("read activation progress: %w", err)
		}
		view = core.ActivationProgressView{
			InvitationID: row.InvitationID, Kind: row.Kind, DisplayName: row.DisplayName,
			ExpiresAt:       row.ExpiresAt,
			PasswordSet:     passwordSetAt != nil,
			ContactVerified: contactVerifiedAt != nil,
			TwofaEnrolled:   twofaEnrolledAt != nil,
			Completed:       completedAt != nil,
		}
		if contactKind != nil && contactValue != nil {
			view.ContactKind = core.ContactKind(*contactKind)
			view.ContactMasked = security.MaskContact(*contactKind, *contactValue)
		}
		return nil
	})
	if err != nil {
		return core.ActivationProgressView{}, err
	}
	return view, nil
}

func (s *Store) SetActivationPassword(ctx context.Context, command core.SetActivationPasswordCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := lockActivationRow(ctx, tx, command.TokenDigest, command.Now)
		if err != nil {
			return err
		}
		if row.Phone != command.Phone {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}
		if err := ensureActivationProgress(ctx, tx, row, command.Now); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO password_credentials (account_id, password_hash, algorithm, created_at, updated_at)
			VALUES ($1, $2, 'argon2id', $3, $3)
			ON CONFLICT (account_id)
			DO UPDATE SET password_hash = EXCLUDED.password_hash, updated_at = EXCLUDED.updated_at
		`, row.AccountID, command.PasswordHash, command.Now); err != nil {
			return fmt.Errorf("store activation password: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_progress SET password_set_at = $2, updated_at = $2
			WHERE invitation_id = $1
		`, row.InvitationID, command.Now); err != nil {
			return fmt.Errorf("record activation password step: %w", err)
		}
		return nil
	})
}

func (s *Store) StartActivationContact(ctx context.Context, command core.StartActivationContactCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := lockActivationRow(ctx, tx, command.TokenDigest, command.Now)
		if err != nil {
			return err
		}
		if err := ensureActivationProgress(ctx, tx, row, command.Now); err != nil {
			return err
		}
		var passwordSetAt *time.Time
		if err := tx.QueryRow(ctx, `
			SELECT password_set_at FROM activation_progress WHERE invitation_id = $1 FOR UPDATE
		`, row.InvitationID).Scan(&passwordSetAt); err != nil {
			return fmt.Errorf("read activation progress: %w", err)
		}
		if passwordSetAt == nil {
			return core.E(core.CodeInvalidState, "set the password before confirming a contact", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE contact_verifications
			SET superseded_at = $3
			WHERE tenant_id = $1 AND account_id = $2 AND purpose = 'activation_contact'
			  AND consumed_at IS NULL AND superseded_at IS NULL
		`, row.TenantID, row.AccountID, command.Now); err != nil {
			return fmt.Errorf("supersede activation verifications: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO contact_verifications (
				id, tenant_id, account_id, contact_kind, normalized_value, purpose,
				code_digest, expires_at, attempts_remaining, created_at
			) VALUES ($1, $2, $3, $4, $5, 'activation_contact', $6, $7, 5, $8)
		`, command.VerificationID, row.TenantID, row.AccountID,
			string(command.Kind), command.Value, command.CodeDigest,
			command.ExpiresAt, command.Now); err != nil {
			return mapWriteError(err, "contact verification could not be created")
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_progress
			SET contact_kind = $2, contact_value = $3, contact_verified_at = NULL, updated_at = $4
			WHERE invitation_id = $1
		`, row.InvitationID, string(command.Kind), command.Value, command.Now); err != nil {
			return fmt.Errorf("record activation contact step: %w", err)
		}
		return appendOutbox(ctx, tx, row.TenantID, "ContactVerificationRequested",
			"contact_verification", command.VerificationID,
			map[string]any{"verificationId": command.VerificationID}, command.Now)
	})
}

func (s *Store) VerifyActivationContact(ctx context.Context, command core.VerifyActivationContactCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := lockActivationRow(ctx, tx, command.TokenDigest, command.Now)
		if err != nil {
			return err
		}
		if _, _, err := consumeOpenVerification(ctx, tx, row.TenantID, row.AccountID,
			"activation_contact", command.CodeDigest, command.Now); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_progress SET contact_verified_at = $2, updated_at = $2
			WHERE invitation_id = $1
		`, row.InvitationID, command.Now); err != nil {
			return fmt.Errorf("record activation contact verification: %w", err)
		}
		return nil
	})
}

func (s *Store) SetActivationTwofa(ctx context.Context, command core.SetActivationTwofaCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := lockActivationRow(ctx, tx, command.TokenDigest, command.Now)
		if err != nil {
			return err
		}
		var contactVerifiedAt *time.Time
		if err := tx.QueryRow(ctx, `
			SELECT contact_verified_at FROM activation_progress WHERE invitation_id = $1 FOR UPDATE
		`, row.InvitationID).Scan(&contactVerifiedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return core.E(core.CodeInvalidState, "confirm the contact before enrolling a second factor", nil)
			}
			return fmt.Errorf("read activation progress: %w", err)
		}
		if contactVerifiedAt == nil {
			return core.E(core.CodeInvalidState, "confirm the contact before enrolling a second factor", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO twofa_totp_secrets (tenant_id, account_id, secret_ciphertext, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $4)
			ON CONFLICT (tenant_id, account_id)
			DO UPDATE SET secret_ciphertext = EXCLUDED.secret_ciphertext,
			              confirmed_at = NULL, updated_at = EXCLUDED.updated_at
		`, row.TenantID, row.AccountID, command.SecretCiphertext, command.Now); err != nil {
			return fmt.Errorf("store activation twofa secret: %w", err)
		}
		return nil
	})
}

func (s *Store) ActivationTwofaSecret(ctx context.Context, digest []byte, now time.Time) ([]byte, error) {
	var ciphertext []byte
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := lockActivationRow(ctx, tx, digest, now)
		if err != nil {
			return err
		}
		var confirmedAt *time.Time
		queryErr := tx.QueryRow(ctx, `
			SELECT secret_ciphertext, confirmed_at FROM twofa_totp_secrets
			WHERE tenant_id = $1 AND account_id = $2
		`, row.TenantID, row.AccountID).Scan(&ciphertext, &confirmedAt)
		if errors.Is(queryErr, pgx.ErrNoRows) || (queryErr == nil && confirmedAt != nil) {
			return core.E(core.CodeInvalidState, "two-factor enrollment has not started", nil)
		}
		if queryErr != nil {
			return fmt.Errorf("read activation twofa secret: %w", queryErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func (s *Store) ConfirmActivationTwofa(ctx context.Context, command core.ConfirmActivationTwofaCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := lockActivationRow(ctx, tx, command.TokenDigest, command.Now)
		if err != nil {
			return err
		}
		var confirmedAt *time.Time
		queryErr := tx.QueryRow(ctx, `
			SELECT confirmed_at FROM twofa_totp_secrets
			WHERE tenant_id = $1 AND account_id = $2
			FOR UPDATE
		`, row.TenantID, row.AccountID).Scan(&confirmedAt)
		if errors.Is(queryErr, pgx.ErrNoRows) || (queryErr == nil && confirmedAt != nil) {
			return core.E(core.CodeInvalidState, "two-factor enrollment has not started", nil)
		}
		if queryErr != nil {
			return fmt.Errorf("lock activation twofa secret: %w", queryErr)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE twofa_totp_secrets SET confirmed_at = $3, updated_at = $3
			WHERE tenant_id = $1 AND account_id = $2
		`, row.TenantID, row.AccountID, command.Now); err != nil {
			return fmt.Errorf("confirm activation twofa: %w", err)
		}
		if err := replaceRecoveryCodes(ctx, tx, row.TenantID, row.AccountID, command.RecoveryDigests, command.Now); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_progress SET twofa_enrolled_at = $2, updated_at = $2
			WHERE invitation_id = $1
		`, row.InvitationID, command.Now); err != nil {
			return fmt.Errorf("record activation twofa step: %w", err)
		}
		return nil
	})
}

func (s *Store) FinishActivation(ctx context.Context, command core.FinishActivationCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var invitationID string
		var tenantID string
		var accountID string
		var status string
		var expiresAt time.Time
		var accountStatus string
		var phone string
		var storedKey *string
		var storedFingerprint []byte
		err := tx.QueryRow(ctx, `
			SELECT i.id, i.tenant_id, i.account_id, i.status, i.expires_at, a.status,
			       li.normalized_value, i.consumed_idempotency_key, i.consumed_payload_fingerprint
			FROM activation_invitations i
			JOIN accounts a ON a.tenant_id = i.tenant_id AND a.id = i.account_id
			JOIN account_login_identifiers li
			  ON li.tenant_id = a.tenant_id AND li.account_id = a.id
			 AND li.identifier_type = 'phone'
			WHERE i.token_digest = $1
			FOR UPDATE OF i, a
		`, command.TokenDigest).Scan(&invitationID, &tenantID, &accountID, &status,
			&expiresAt, &accountStatus, &phone, &storedKey, &storedFingerprint)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}
		if err != nil {
			return fmt.Errorf("lock activation for finish: %w", err)
		}
		if status == "consumed" && storedKey != nil && *storedKey == command.IdempotencyKey {
			if !security.EqualDigest(storedFingerprint, command.PayloadFingerprint) {
				return core.E(core.CodeConflict, "Idempotency-Key was reused with a different payload", nil)
			}
			return nil
		}
		if status != "issued" || !expiresAt.After(command.Now) ||
			accountStatus != "pending_activation" || phone != command.Phone {
			return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
		}
		var contactKind, contactValue *string
		var passwordSetAt, contactVerifiedAt *time.Time
		progressErr := tx.QueryRow(ctx, `
			SELECT password_set_at, contact_verified_at, contact_kind, contact_value
			FROM activation_progress WHERE invitation_id = $1 FOR UPDATE
		`, invitationID).Scan(&passwordSetAt, &contactVerifiedAt, &contactKind, &contactValue)
		if errors.Is(progressErr, pgx.ErrNoRows) || (progressErr == nil && (passwordSetAt == nil || contactVerifiedAt == nil)) {
			return core.E(core.CodeInvalidState, "complete the required activation steps first", nil)
		}
		if progressErr != nil {
			return fmt.Errorf("read activation progress for finish: %w", progressErr)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE accounts SET status = 'active', activated_at = $3, updated_at = $3, version = version + 1
			WHERE tenant_id = $1 AND id = $2
		`, tenantID, accountID, command.Now); err != nil {
			return fmt.Errorf("activate account: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE account_login_identifiers SET status = 'confirmed', confirmed_at = $3
			WHERE tenant_id = $1 AND account_id = $2 AND identifier_type = 'phone'
		`, tenantID, accountID, command.Now); err != nil {
			return fmt.Errorf("confirm login identifier: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_invitations
			SET status = 'consumed', consumed_at = $2,
			    consumed_idempotency_key = $3, consumed_payload_fingerprint = $4
			WHERE id = $1
		`, invitationID, command.Now, command.IdempotencyKey, command.PayloadFingerprint); err != nil {
			return fmt.Errorf("consume activation invitation: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE activation_progress SET completed_at = $2, updated_at = $2
			WHERE invitation_id = $1
		`, invitationID, command.Now); err != nil {
			return fmt.Errorf("complete activation progress: %w", err)
		}
		if contactKind != nil && contactValue != nil {
			if _, err := upsertVerifiedContact(ctx, tx, tenantID, accountID,
				core.ContactKind(*contactKind), *contactValue, command.Now); err != nil {
				return err
			}
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: tenantID, actorID: accountID,
			action: "AccountActivated", targetType: "account", targetID: accountID,
			decision: "allow", idempotencyKey: command.IdempotencyKey, at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, tenantID, "AccountActivated", "account", accountID,
			map[string]any{"accountId": accountID}, command.Now)
	})
}
