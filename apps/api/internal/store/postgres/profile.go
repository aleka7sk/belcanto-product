package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 account profile (Figma Page 32: ACC-01/02).

func (s *Store) ProfileView(ctx context.Context, principal core.Principal) (core.ProfileView, error) {
	view := core.ProfileView{AccountID: principal.AccountID}
	err := s.pool.QueryRow(ctx, `
		SELECT p.full_name, t.name, COALESCE(ali.normalized_value, '')
		FROM accounts a
		JOIN people p ON p.tenant_id = a.tenant_id AND p.id = a.person_id
		JOIN tenants t ON t.id = a.tenant_id
		LEFT JOIN account_login_identifiers ali
		  ON ali.account_id = a.id AND ali.identifier_type = 'phone' AND ali.status = 'confirmed'
		WHERE a.tenant_id = $1 AND a.id = $2 AND a.status = 'active'
	`, principal.TenantID, principal.AccountID).Scan(&view.FullName, &view.TenantName, &view.Phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.ProfileView{}, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	if err != nil {
		return core.ProfileView{}, fmt.Errorf("read profile: %w", err)
	}
	roles, err := rolesForAccount(ctx, s.pool, principal.TenantID, principal.AccountID)
	if err != nil {
		return core.ProfileView{}, err
	}
	view.Roles = roles
	return view, nil
}

func (s *Store) UpdateProfile(ctx context.Context, command core.UpdateProfileCommand) (core.ProfileView, error) {
	principal := command.Principal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		updated, err := tx.Exec(ctx, `
			UPDATE people p
			SET full_name = $3
			FROM accounts a
			WHERE a.tenant_id = $1 AND a.id = $2 AND a.status = 'active'
			  AND p.tenant_id = a.tenant_id AND p.id = a.person_id
		`, principal.TenantID, principal.AccountID, command.FullName)
		if err != nil {
			return fmt.Errorf("update profile name: %w", err)
		}
		if updated.RowsAffected() == 0 {
			return core.E(core.CodeUnauthenticated, "account is inactive", nil)
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "ProfileUpdated", targetType: "account", targetID: principal.AccountID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return core.ProfileView{}, err
	}
	return s.ProfileView(ctx, principal)
}
