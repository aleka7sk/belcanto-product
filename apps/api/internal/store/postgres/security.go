package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// securityAuditActions is the closed set of audit actions projected into the
// account's Security Center feed (Page 32, ACC-05/09). The projection stays
// privacy-safe: it exposes action, decision, reason code and target
// identifiers only, never metadata payloads.
var securityAuditActions = []string{
	"SessionCreated",
	"SessionRefreshed",
	"SessionRevoked",
	"RefreshTokenReuseDetected",
	"AccountActivated",
	"PasswordResetRequested",
	"PasswordResetCompleted",
	"OtherSessionsRevoked",
	"ContactChangeStarted",
	"ContactVerified",
	"TwofaEnrolled",
	"TwofaDisabled",
	"TwofaChallengeFailed",
	"PolicyAccepted",
	"PrivacySettingsUpdated",
	"DataExportRequested",
	"DeletionRequested",
	"DeletionRequestCancelled",
}

func (s *Store) ListSessions(ctx context.Context, principal core.Principal, now time.Time) ([]core.SessionDevice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, COALESCE(device_label, ''), COALESCE(platform, ''), created_at, last_seen_at
		FROM sessions
		WHERE tenant_id = $1 AND account_id = $2 AND status = 'active'
		  AND refresh_expires_at > $3
		ORDER BY created_at DESC, id
	`, principal.TenantID, principal.AccountID, now)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()
	devices := []core.SessionDevice{}
	for rows.Next() {
		var device core.SessionDevice
		var lastSeen *time.Time
		if err := rows.Scan(&device.SessionID, &device.DeviceLabel, &device.Platform, &device.CreatedAt, &lastSeen); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		device.LastSeenAt = lastSeen
		device.Current = device.SessionID == principal.SessionID
		devices = append(devices, device)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}
	return devices, nil
}

func (s *Store) RevokeSessionByID(ctx context.Context, command core.RevokeSessionByIDCommand) error {
	principal := command.Principal
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var familyID string
		var status string
		err := tx.QueryRow(ctx, `
			SELECT family_id, status FROM sessions
			WHERE tenant_id = $1 AND account_id = $2 AND id = $3
			FOR UPDATE
		`, principal.TenantID, principal.AccountID, command.SessionID).Scan(&familyID, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "session was not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock session for revocation: %w", err)
		}
		var currentFamily string
		err = tx.QueryRow(ctx, `
			SELECT family_id FROM sessions
			WHERE tenant_id = $1 AND account_id = $2 AND id = $3
		`, principal.TenantID, principal.AccountID, principal.SessionID).Scan(&currentFamily)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("read current session family: %w", err)
		}
		if err == nil && familyID == currentFamily {
			return core.E(core.CodeInvalidState, "sign out to end the current session", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE sessions
			SET status = 'revoked', revoked_at = COALESCE(revoked_at, $4)
			WHERE tenant_id = $1 AND account_id = $2 AND family_id = $3
			  AND status <> 'revoked'
		`, principal.TenantID, principal.AccountID, familyID, command.Now); err != nil {
			return fmt.Errorf("revoke session family: %w", err)
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "SessionRevoked", targetType: "session_family", targetID: familyID,
			decision: "allow", metadata: map[string]any{"sessionId": command.SessionID},
			at: command.Now,
		})
	})
}

func (s *Store) RevokeOtherSessions(ctx context.Context, command core.RevokeOtherSessionsCommand) (int, error) {
	principal := command.Principal
	revoked := 0
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var currentFamily string
		err := tx.QueryRow(ctx, `
			SELECT family_id FROM sessions
			WHERE tenant_id = $1 AND account_id = $2 AND id = $3
			FOR UPDATE
		`, principal.TenantID, principal.AccountID, principal.SessionID).Scan(&currentFamily)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeUnauthenticated, "session is inactive", nil)
		}
		if err != nil {
			return fmt.Errorf("read current session family: %w", err)
		}
		rows, err := tx.Query(ctx, `
			SELECT family_id FROM sessions
			WHERE tenant_id = $1 AND account_id = $2 AND family_id <> $3
			  AND status = 'active'
			ORDER BY family_id
			FOR UPDATE
		`, principal.TenantID, principal.AccountID, currentFamily)
		if err != nil {
			return fmt.Errorf("list other session families: %w", err)
		}
		seen := map[string]bool{}
		families := []string{}
		for rows.Next() {
			var familyID string
			if err := rows.Scan(&familyID); err != nil {
				rows.Close()
				return fmt.Errorf("scan session family: %w", err)
			}
			if !seen[familyID] {
				seen[familyID] = true
				families = append(families, familyID)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate session families: %w", err)
		}
		for _, familyID := range families {
			if _, err := tx.Exec(ctx, `
				UPDATE sessions
				SET status = 'revoked', revoked_at = COALESCE(revoked_at, $4)
				WHERE tenant_id = $1 AND account_id = $2 AND family_id = $3
				  AND status <> 'revoked'
			`, principal.TenantID, principal.AccountID, familyID, command.Now); err != nil {
				return fmt.Errorf("revoke other session family: %w", err)
			}
		}
		revoked = len(families)
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "OtherSessionsRevoked", targetType: "account", targetID: principal.AccountID,
			decision: "allow", metadata: map[string]any{"revokedFamilies": revoked},
			at: command.Now,
		})
	})
	if err != nil {
		return 0, err
	}
	return revoked, nil
}

func (s *Store) ListSecurityEvents(ctx context.Context, principal core.Principal, query core.SecurityEventsQuery) ([]core.SecurityEvent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, action, decision, COALESCE(reason_code, ''), target_type, target_id, recorded_at
		FROM audit_records
		WHERE tenant_id = $1 AND actor_account_id = $2
		  AND action = ANY($3)
		  AND ($4::bigint = 0 OR id < $4)
		ORDER BY id DESC
		LIMIT $5
	`, principal.TenantID, principal.AccountID, securityAuditActions, query.BeforeID, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("list security events: %w", err)
	}
	defer rows.Close()
	events := []core.SecurityEvent{}
	for rows.Next() {
		var event core.SecurityEvent
		if err := rows.Scan(&event.ID, &event.Action, &event.Decision, &event.ReasonCode, &event.TargetType, &event.TargetID, &event.RecordedAt); err != nil {
			return nil, fmt.Errorf("scan security event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate security events: %w", err)
	}
	return events, nil
}

func (s *Store) CreatePasswordReset(ctx context.Context, command core.CreatePasswordResetCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
			advisoryLockKey("password-reset", command.Phone)); err != nil {
			return fmt.Errorf("lock password reset subject: %w", err)
		}
		var tenantID string
		var accountID string
		err := tx.QueryRow(ctx, `
			SELECT a.tenant_id, a.id
			FROM account_login_identifiers ali
			JOIN accounts a ON a.id = ali.account_id
			WHERE ali.normalized_value = $1 AND ali.status = 'confirmed'
			  AND a.status = 'active'
		`, command.Phone).Scan(&tenantID, &accountID)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "account was not found", nil)
		}
		if err != nil {
			return fmt.Errorf("resolve reset account: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE password_reset_tokens
			SET superseded_at = $3
			WHERE tenant_id = $1 AND account_id = $2
			  AND consumed_at IS NULL AND superseded_at IS NULL
		`, tenantID, accountID, command.Now); err != nil {
			return fmt.Errorf("supersede open password resets: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO password_reset_tokens (id, tenant_id, account_id, token_digest, expires_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, command.ResetID, tenantID, accountID, command.TokenDigest, command.ExpiresAt, command.Now); err != nil {
			return mapWriteError(err, "password reset could not be created")
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: tenantID, actorID: accountID,
			action: "PasswordResetRequested", targetType: "password_reset", targetID: command.ResetID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, tenantID, "PasswordResetRequested", "password_reset", command.ResetID,
			map[string]any{"resetId": command.ResetID}, command.Now)
	})
}

func (s *Store) CompletePasswordReset(ctx context.Context, command core.CompletePasswordResetCommand) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var resetID string
		var tenantID string
		var accountID string
		var expiresAt time.Time
		var consumedAt *time.Time
		var supersededAt *time.Time
		err := tx.QueryRow(ctx, `
			SELECT id, tenant_id, account_id, expires_at, consumed_at, superseded_at
			FROM password_reset_tokens
			WHERE token_digest = $1
			FOR UPDATE
		`, command.TokenDigest).Scan(&resetID, &tenantID, &accountID, &expiresAt, &consumedAt, &supersededAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeUnauthenticated, "recovery link is invalid or expired", nil)
		}
		if err != nil {
			return fmt.Errorf("lock password reset: %w", err)
		}
		if consumedAt != nil || supersededAt != nil || !expiresAt.After(command.Now) {
			if err := appendAudit(ctx, tx, auditInput{
				tenantID: tenantID, actorID: accountID,
				action: "PasswordResetCompleted", targetType: "password_reset", targetID: resetID,
				decision: "deny", reason: "inactive_or_expired_reset_token", at: command.Now,
			}); err != nil {
				return err
			}
			return core.E(core.CodeUnauthenticated, "recovery link is invalid or expired", nil)
		}
		var accountStatus string
		if err := tx.QueryRow(ctx, `
			SELECT status FROM accounts WHERE tenant_id = $1 AND id = $2 FOR UPDATE
		`, tenantID, accountID).Scan(&accountStatus); err != nil {
			return fmt.Errorf("lock reset account: %w", err)
		}
		if accountStatus != "active" {
			return core.E(core.CodeUnauthenticated, "recovery link is invalid or expired", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE password_reset_tokens SET consumed_at = $2 WHERE id = $1
		`, resetID, command.Now); err != nil {
			return fmt.Errorf("consume password reset: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE password_credentials SET password_hash = $2, updated_at = $3
			WHERE account_id = $1
		`, accountID, command.PasswordHash, command.Now); err != nil {
			return fmt.Errorf("rotate password credential: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE sessions
			SET status = 'revoked', revoked_at = COALESCE(revoked_at, $3)
			WHERE tenant_id = $1 AND account_id = $2 AND status <> 'revoked'
		`, tenantID, accountID, command.Now); err != nil {
			return fmt.Errorf("revoke sessions after password reset: %w", err)
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: tenantID, actorID: accountID,
			action: "PasswordResetCompleted", targetType: "password_reset", targetID: resetID,
			decision: "allow", at: command.Now,
		})
	})
}
