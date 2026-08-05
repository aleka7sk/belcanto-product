package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 policies, privacy and data rights (Figma Page 32: ACC-10..18).

func (s *Store) ListPolicies(ctx context.Context, principal core.Principal) ([]core.PolicyVersion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pv.id, pv.kind, pv.version, pv.title, pv.body_ref, pv.effective_from, pa.accepted_at
		FROM policy_versions pv
		LEFT JOIN policy_acceptances pa
		  ON pa.tenant_id = pv.tenant_id AND pa.policy_version_id = pv.id
		 AND pa.account_id = $2
		WHERE pv.tenant_id = $1
		ORDER BY pv.kind, pv.effective_from DESC
	`, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, fmt.Errorf("list policy versions: %w", err)
	}
	defer rows.Close()
	policies := []core.PolicyVersion{}
	for rows.Next() {
		var policy core.PolicyVersion
		var acceptedAt *time.Time
		if err := rows.Scan(&policy.ID, &policy.Kind, &policy.Version, &policy.Title,
			&policy.BodyRef, &policy.EffectiveFrom, &acceptedAt); err != nil {
			return nil, fmt.Errorf("scan policy version: %w", err)
		}
		policy.AcceptedAt = acceptedAt
		policies = append(policies, policy)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate policy versions: %w", err)
	}
	return policies, nil
}

func (s *Store) AcceptPolicy(ctx context.Context, command core.AcceptPolicyCommand) error {
	principal := command.Principal
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM policy_versions WHERE tenant_id = $1 AND id = $2
			)
		`, principal.TenantID, command.PolicyVersionID).Scan(&exists); err != nil {
			return fmt.Errorf("check policy version: %w", err)
		}
		if !exists {
			return core.E(core.CodeNotFound, "policy version was not found", nil)
		}
		inserted, err := tx.Exec(ctx, `
			INSERT INTO policy_acceptances (id, tenant_id, account_id, policy_version_id, accepted_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tenant_id, account_id, policy_version_id) DO NOTHING
		`, command.AcceptanceID, principal.TenantID, principal.AccountID,
			command.PolicyVersionID, command.Now)
		if err != nil {
			return mapWriteError(err, "policy acceptance could not be recorded")
		}
		if inserted.RowsAffected() == 0 {
			return nil
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "PolicyAccepted", targetType: "policy_version", targetID: command.PolicyVersionID,
			decision: "allow", at: command.Now,
		})
	})
}

func scanPrivacySettings(row pgx.Row) (core.PrivacySettings, error) {
	var settings core.PrivacySettings
	err := row.Scan(
		&settings.CommunityProfileVisible,
		&settings.AchievementsVisible,
		&settings.StaffMessagesAllowed,
		&settings.MentionsAllowed,
		&settings.PushPreview,
		&settings.Version,
	)
	return settings, err
}

func defaultPrivacySettings() core.PrivacySettings {
	return core.PrivacySettings{
		CommunityProfileVisible: true,
		AchievementsVisible:     true,
		StaffMessagesAllowed:    true,
		MentionsAllowed:         true,
		PushPreview:             "hidden",
		Version:                 0,
	}
}

func (s *Store) PrivacySettings(ctx context.Context, principal core.Principal) (core.PrivacySettings, error) {
	settings, err := scanPrivacySettings(s.pool.QueryRow(ctx, `
		SELECT community_profile_visible, achievements_visible, staff_messages_allowed,
		       mentions_allowed, push_preview, version
		FROM privacy_settings
		WHERE tenant_id = $1 AND account_id = $2
	`, principal.TenantID, principal.AccountID))
	if errors.Is(err, pgx.ErrNoRows) {
		return defaultPrivacySettings(), nil
	}
	if err != nil {
		return core.PrivacySettings{}, fmt.Errorf("read privacy settings: %w", err)
	}
	return settings, nil
}

func (s *Store) UpdatePrivacySettings(ctx context.Context, command core.UpdatePrivacySettingsCommand) (core.PrivacySettings, error) {
	principal := command.Principal
	var updated core.PrivacySettings
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		current, err := scanPrivacySettings(tx.QueryRow(ctx, `
			SELECT community_profile_visible, achievements_visible, staff_messages_allowed,
			       mentions_allowed, push_preview, version
			FROM privacy_settings
			WHERE tenant_id = $1 AND account_id = $2
			FOR UPDATE
		`, principal.TenantID, principal.AccountID))
		if errors.Is(err, pgx.ErrNoRows) {
			current = defaultPrivacySettings()
		} else if err != nil {
			return fmt.Errorf("lock privacy settings: %w", err)
		}
		if current.Version != command.ExpectedVersion {
			return core.E(core.CodeConflict, "privacy settings changed since they were loaded", nil)
		}
		next := command.Settings
		next.Version = current.Version + 1
		if _, err := tx.Exec(ctx, `
			INSERT INTO privacy_settings (
				tenant_id, account_id, community_profile_visible, achievements_visible,
				staff_messages_allowed, mentions_allowed, push_preview, version, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (tenant_id, account_id)
			DO UPDATE SET community_profile_visible = EXCLUDED.community_profile_visible,
			              achievements_visible = EXCLUDED.achievements_visible,
			              staff_messages_allowed = EXCLUDED.staff_messages_allowed,
			              mentions_allowed = EXCLUDED.mentions_allowed,
			              push_preview = EXCLUDED.push_preview,
			              version = EXCLUDED.version,
			              updated_at = EXCLUDED.updated_at
		`, principal.TenantID, principal.AccountID,
			next.CommunityProfileVisible, next.AchievementsVisible,
			next.StaffMessagesAllowed, next.MentionsAllowed,
			next.PushPreview, next.Version, command.Now); err != nil {
			return fmt.Errorf("write privacy settings: %w", err)
		}
		updated = next
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "PrivacySettingsUpdated", targetType: "account", targetID: principal.AccountID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.PrivacySettings{}, err
	}
	return updated, nil
}

func (s *Store) CreateDataExport(ctx context.Context, command core.CreateDataExportCommand) (core.DataExportRequest, error) {
	principal := command.Principal
	var request core.DataExportRequest
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
			advisoryLockKey("data-export", principal.TenantID, principal.AccountID)); err != nil {
			return fmt.Errorf("lock data export subject: %w", err)
		}
		var open bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM data_export_requests
				WHERE tenant_id = $1 AND account_id = $2 AND status IN ('requested', 'processing')
			)
		`, principal.TenantID, principal.AccountID).Scan(&open); err != nil {
			return fmt.Errorf("check open data exports: %w", err)
		}
		if open {
			return core.E(core.CodeConflict, "a data export is already in progress", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO data_export_requests (id, tenant_id, account_id, status, requested_at)
			VALUES ($1, $2, $3, 'requested', $4)
		`, command.ExportID, principal.TenantID, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "data export could not be created")
		}
		request = core.DataExportRequest{ID: command.ExportID, Status: "requested", RequestedAt: command.Now}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "DataExportRequested", targetType: "data_export", targetID: command.ExportID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "DataExportRequested",
			"data_export", command.ExportID,
			map[string]any{"exportId": command.ExportID}, command.Now)
	})
	if err != nil {
		return core.DataExportRequest{}, err
	}
	return request, nil
}

func (s *Store) ListDataExports(ctx context.Context, principal core.Principal) ([]core.DataExportRequest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, status, requested_at, ready_at, expires_at
		FROM data_export_requests
		WHERE tenant_id = $1 AND account_id = $2
		ORDER BY requested_at DESC
		LIMIT 10
	`, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, fmt.Errorf("list data exports: %w", err)
	}
	defer rows.Close()
	exports := []core.DataExportRequest{}
	for rows.Next() {
		var export core.DataExportRequest
		if err := rows.Scan(&export.ID, &export.Status, &export.RequestedAt, &export.ReadyAt, &export.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan data export: %w", err)
		}
		exports = append(exports, export)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate data exports: %w", err)
	}
	return exports, nil
}

func scanDeletionRequest(row pgx.Row) (core.DeletionRequest, error) {
	var request core.DeletionRequest
	err := row.Scan(&request.ID, &request.Status, &request.RequestedAt, &request.CancelledAt)
	return request, err
}

func (s *Store) DeletionRequest(ctx context.Context, principal core.Principal) (core.DeletionRequest, error) {
	request, err := scanDeletionRequest(s.pool.QueryRow(ctx, `
		SELECT id, status, requested_at, cancelled_at
		FROM account_deletion_requests
		WHERE tenant_id = $1 AND account_id = $2 AND status IN ('requested', 'pending_review')
	`, principal.TenantID, principal.AccountID))
	if errors.Is(err, pgx.ErrNoRows) {
		return core.DeletionRequest{}, core.E(core.CodeNotFound, "no deletion request is open", nil)
	}
	if err != nil {
		return core.DeletionRequest{}, fmt.Errorf("read deletion request: %w", err)
	}
	return request, nil
}

func (s *Store) CreateDeletionRequest(ctx context.Context, command core.CreateDeletionRequestCommand) (core.DeletionRequest, error) {
	principal := command.Principal
	var request core.DeletionRequest
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
			advisoryLockKey("deletion-request", principal.TenantID, principal.AccountID)); err != nil {
			return fmt.Errorf("lock deletion subject: %w", err)
		}
		var open bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM account_deletion_requests
				WHERE tenant_id = $1 AND account_id = $2 AND status IN ('requested', 'pending_review')
			)
		`, principal.TenantID, principal.AccountID).Scan(&open); err != nil {
			return fmt.Errorf("check open deletion requests: %w", err)
		}
		if open {
			return core.E(core.CodeConflict, "a deletion request is already open", nil)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO account_deletion_requests (id, tenant_id, account_id, status, requested_at)
			VALUES ($1, $2, $3, 'pending_review', $4)
		`, command.RequestID, principal.TenantID, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "deletion request could not be created")
		}
		request = core.DeletionRequest{ID: command.RequestID, Status: "pending_review", RequestedAt: command.Now}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "DeletionRequested", targetType: "deletion_request", targetID: command.RequestID,
			decision: "allow", at: command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "DeletionRequested",
			"deletion_request", command.RequestID,
			map[string]any{"requestId": command.RequestID}, command.Now)
	})
	if err != nil {
		return core.DeletionRequest{}, err
	}
	return request, nil
}

func (s *Store) CancelDeletionRequest(ctx context.Context, command core.CancelDeletionRequestCommand) (core.DeletionRequest, error) {
	principal := command.Principal
	var request core.DeletionRequest
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		row, err := scanDeletionRequest(tx.QueryRow(ctx, `
			SELECT id, status, requested_at, cancelled_at
			FROM account_deletion_requests
			WHERE tenant_id = $1 AND account_id = $2 AND status IN ('requested', 'pending_review')
			FOR UPDATE
		`, principal.TenantID, principal.AccountID))
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeInvalidState, "no deletion request is open", nil)
		}
		if err != nil {
			return fmt.Errorf("lock deletion request: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE account_deletion_requests
			SET status = 'cancelled', cancelled_at = $2
			WHERE id = $1
		`, row.ID, command.Now); err != nil {
			return fmt.Errorf("cancel deletion request: %w", err)
		}
		cancelledAt := command.Now
		request = core.DeletionRequest{ID: row.ID, Status: "cancelled", RequestedAt: row.RequestedAt, CancelledAt: &cancelledAt}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "DeletionRequestCancelled", targetType: "deletion_request", targetID: row.ID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.DeletionRequest{}, err
	}
	return request, nil
}
