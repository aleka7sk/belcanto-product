package app

import (
	"context"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// P.1 policies, privacy and data rights (Figma Page 32: ACC-10..12,
// ACC-14..18; HOF-12). Export and deletion are sensitive operations and
// require fresh authentication; deletion stays a reviewable request while
// DEC-104 is open — nothing here schedules erasure.

func (s *Service) ListPolicies(ctx context.Context, principal core.Principal) ([]core.PolicyVersion, error) {
	policies, err := s.store.ListPolicies(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list policies", err)
	}
	return policies, nil
}

func (s *Service) AcceptPolicy(ctx context.Context, principal core.Principal, policyVersionID string) error {
	normalizedID, err := security.ValidateIdentifier("policyVersionId", policyVersionID, 128)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	acceptanceID, err := security.NewID("polacc")
	if err != nil {
		return core.E(core.CodeInternal, "could not record the acceptance", err)
	}
	if err := s.store.AcceptPolicy(ctx, core.AcceptPolicyCommand{
		Principal:       principal,
		AcceptanceID:    acceptanceID,
		PolicyVersionID: normalizedID,
		Now:             s.clock.Now(),
	}); err != nil {
		return normalizeStoreError("accept policy", err)
	}
	return nil
}

func (s *Service) PrivacySettings(ctx context.Context, principal core.Principal) (core.PrivacySettings, error) {
	settings, err := s.store.PrivacySettings(ctx, principal)
	if err != nil {
		return core.PrivacySettings{}, normalizeStoreError("read privacy settings", err)
	}
	return settings, nil
}

func (s *Service) UpdatePrivacySettings(ctx context.Context, principal core.Principal, settings core.PrivacySettings) (core.PrivacySettings, error) {
	switch settings.PushPreview {
	case "hidden", "title", "full":
	default:
		return core.PrivacySettings{}, core.E(core.CodeInvalidInput, "pushPreview must be hidden, title or full", nil)
	}
	if settings.Version < 0 {
		return core.PrivacySettings{}, core.E(core.CodeInvalidInput, "version must be at least 0", nil)
	}
	updated, err := s.store.UpdatePrivacySettings(ctx, core.UpdatePrivacySettingsCommand{
		Principal:       principal,
		Settings:        settings,
		ExpectedVersion: settings.Version,
		Now:             s.clock.Now(),
	})
	if err != nil {
		return core.PrivacySettings{}, normalizeStoreError("update privacy settings", err)
	}
	return updated, nil
}

func (s *Service) CreateDataExport(ctx context.Context, principal core.Principal, currentPassword string) (core.DataExportRequest, error) {
	if err := s.reauthenticate(ctx, principal, currentPassword); err != nil {
		return core.DataExportRequest{}, err
	}
	exportID, err := security.NewID("export")
	if err != nil {
		return core.DataExportRequest{}, core.E(core.CodeInternal, "could not create the export request", err)
	}
	request, err := s.store.CreateDataExport(ctx, core.CreateDataExportCommand{
		Principal: principal,
		ExportID:  exportID,
		Now:       s.clock.Now(),
	})
	if err != nil {
		return core.DataExportRequest{}, normalizeStoreError("create data export", err)
	}
	return request, nil
}

func (s *Service) ListDataExports(ctx context.Context, principal core.Principal) ([]core.DataExportRequest, error) {
	exports, err := s.store.ListDataExports(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list data exports", err)
	}
	return exports, nil
}

func (s *Service) DeletionRequest(ctx context.Context, principal core.Principal) (core.DeletionRequest, error) {
	request, err := s.store.DeletionRequest(ctx, principal)
	if err != nil {
		return core.DeletionRequest{}, normalizeStoreError("read deletion request", err)
	}
	return request, nil
}

func (s *Service) CreateDeletionRequest(ctx context.Context, principal core.Principal, currentPassword string) (core.DeletionRequest, error) {
	if err := s.reauthenticate(ctx, principal, currentPassword); err != nil {
		return core.DeletionRequest{}, err
	}
	requestID, err := security.NewID("delreq")
	if err != nil {
		return core.DeletionRequest{}, core.E(core.CodeInternal, "could not create the deletion request", err)
	}
	request, err := s.store.CreateDeletionRequest(ctx, core.CreateDeletionRequestCommand{
		Principal: principal,
		RequestID: requestID,
		Now:       s.clock.Now(),
	})
	if err != nil {
		return core.DeletionRequest{}, normalizeStoreError("create deletion request", err)
	}
	return request, nil
}

func (s *Service) CancelDeletionRequest(ctx context.Context, principal core.Principal) (core.DeletionRequest, error) {
	request, err := s.store.CancelDeletionRequest(ctx, core.CancelDeletionRequestCommand{
		Principal: principal,
		Now:       s.clock.Now(),
	})
	if err != nil {
		return core.DeletionRequest{}, normalizeStoreError("cancel deletion request", err)
	}
	return request, nil
}
