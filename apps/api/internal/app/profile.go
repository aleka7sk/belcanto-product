package app

import (
	"context"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// P.1 account profile (Figma Page 32: ACC-01/02). Contact values change
// only through the verified contact-change flow; the profile itself owns
// the display name.

func (s *Service) MyProfile(ctx context.Context, principal core.Principal) (core.ProfileView, error) {
	view, err := s.store.ProfileView(ctx, principal)
	if err != nil {
		return core.ProfileView{}, normalizeStoreError("read profile", err)
	}
	return view, nil
}

func (s *Service) UpdateMyProfile(ctx context.Context, principal core.Principal, fullName string) (core.ProfileView, error) {
	normalized, err := security.ValidateText("fullName", fullName, 1, 200)
	if err != nil {
		return core.ProfileView{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	view, err := s.store.UpdateProfile(ctx, core.UpdateProfileCommand{
		Principal: principal,
		FullName:  normalized,
		Now:       s.clock.Now(),
	})
	if err != nil {
		return core.ProfileView{}, normalizeStoreError("update profile", err)
	}
	return view, nil
}
