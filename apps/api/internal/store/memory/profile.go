package memory

import (
	"context"
	"sort"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 account profile — parity with PostgreSQL (ACC-01/02).

func (s *Store) profileViewLocked(principal core.Principal) (core.ProfileView, error) {
	account := s.accounts[principal.AccountID]
	if account == nil || account.TenantID != principal.TenantID || account.Status != "active" {
		return core.ProfileView{}, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	view := core.ProfileView{
		AccountID:  account.ID,
		FullName:   account.FullName,
		TenantName: s.tenants[account.TenantID],
		Roles:      []core.Role{},
	}
	for role := range account.Roles {
		view.Roles = append(view.Roles, role)
	}
	sort.Slice(view.Roles, func(left, right int) bool {
		return view.Roles[left] < view.Roles[right]
	})
	for phone, accountID := range s.accountPhone {
		if accountID == account.ID {
			view.Phone = phone
		}
	}
	return view, nil
}

func (s *Store) ProfileView(_ context.Context, principal core.Principal) (core.ProfileView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.profileViewLocked(principal)
}

func (s *Store) UpdateProfile(_ context.Context, command core.UpdateProfileCommand) (core.ProfileView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	account := s.accounts[principal.AccountID]
	if account == nil || account.TenantID != principal.TenantID || account.Status != "active" {
		return core.ProfileView{}, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	account.FullName = command.FullName
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "ProfileUpdated",
		"account", principal.AccountID, "allow", "", command.Now, nil)
	return s.profileViewLocked(principal)
}
