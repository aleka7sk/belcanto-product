package memory

import (
	"context"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 policies, privacy and data rights — parity with PostgreSQL.

type policyVersion struct {
	ID            string
	TenantID      string
	Kind          string
	Version       string
	Title         string
	BodyRef       string
	EffectiveFrom time.Time
}

type policyAcceptance struct {
	ID              string
	TenantID        string
	AccountID       string
	PolicyVersionID string
	AcceptedAt      time.Time
}

type privacyRecord struct {
	Settings  core.PrivacySettings
	UpdatedAt time.Time
}

type dataExport struct {
	ID          string
	TenantID    string
	AccountID   string
	Status      string
	RequestedAt time.Time
	ReadyAt     *time.Time
	ExpiresAt   *time.Time
}

type deletionRequest struct {
	ID          string
	TenantID    string
	AccountID   string
	Status      string
	RequestedAt time.Time
	CancelledAt *time.Time
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

// SeedPolicyVersionForTest registers a policy version; policy authoring is
// an Owner governance capability that arrives in a later slice.
func (s *Store) SeedPolicyVersionForTest(id, tenantID, kind, version, title, bodyRef string, effectiveFrom time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[id] = &policyVersion{
		ID: id, TenantID: tenantID, Kind: kind, Version: version,
		Title: title, BodyRef: bodyRef, EffectiveFrom: effectiveFrom,
	}
}

func (s *Store) ListPolicies(_ context.Context, principal core.Principal) ([]core.PolicyVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.PolicyVersion{}
	for _, policy := range s.policies {
		if policy.TenantID != principal.TenantID {
			continue
		}
		view := core.PolicyVersion{
			ID: policy.ID, Kind: policy.Kind, Version: policy.Version,
			Title: policy.Title, BodyRef: policy.BodyRef, EffectiveFrom: policy.EffectiveFrom,
		}
		for _, acceptance := range s.acceptances {
			if acceptance.AccountID == principal.AccountID && acceptance.PolicyVersionID == policy.ID {
				acceptedAt := acceptance.AcceptedAt
				view.AcceptedAt = &acceptedAt
			}
		}
		result = append(result, view)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Kind != result[right].Kind {
			return result[left].Kind < result[right].Kind
		}
		return result[left].EffectiveFrom.After(result[right].EffectiveFrom)
	})
	return result, nil
}

func (s *Store) AcceptPolicy(_ context.Context, command core.AcceptPolicyCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	policy := s.policies[command.PolicyVersionID]
	if policy == nil || policy.TenantID != principal.TenantID {
		return core.E(core.CodeNotFound, "policy version was not found", nil)
	}
	for _, acceptance := range s.acceptances {
		if acceptance.AccountID == principal.AccountID && acceptance.PolicyVersionID == policy.ID {
			return nil
		}
	}
	s.acceptances[command.AcceptanceID] = &policyAcceptance{
		ID: command.AcceptanceID, TenantID: principal.TenantID, AccountID: principal.AccountID,
		PolicyVersionID: policy.ID, AcceptedAt: command.Now,
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "PolicyAccepted",
		"policy_version", policy.ID, "allow", "", command.Now, nil)
	return nil
}

func (s *Store) PrivacySettings(_ context.Context, principal core.Principal) (core.PrivacySettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.privacy[secretKey(principal.TenantID, principal.AccountID)]
	if record == nil {
		return defaultPrivacySettings(), nil
	}
	return record.Settings, nil
}

func (s *Store) UpdatePrivacySettings(_ context.Context, command core.UpdatePrivacySettingsCommand) (core.PrivacySettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	key := secretKey(principal.TenantID, principal.AccountID)
	current := defaultPrivacySettings()
	if record := s.privacy[key]; record != nil {
		current = record.Settings
	}
	if current.Version != command.ExpectedVersion {
		return core.PrivacySettings{}, core.E(core.CodeConflict, "privacy settings changed since they were loaded", nil)
	}
	next := command.Settings
	next.Version = current.Version + 1
	s.privacy[key] = &privacyRecord{Settings: next, UpdatedAt: command.Now}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "PrivacySettingsUpdated",
		"account", principal.AccountID, "allow", "", command.Now, nil)
	return next, nil
}

func (s *Store) CreateDataExport(_ context.Context, command core.CreateDataExportCommand) (core.DataExportRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	for _, export := range s.exports {
		if export.AccountID == principal.AccountID &&
			(export.Status == "requested" || export.Status == "processing") {
			return core.DataExportRequest{}, core.E(core.CodeConflict, "a data export is already in progress", nil)
		}
	}
	export := &dataExport{
		ID: command.ExportID, TenantID: principal.TenantID, AccountID: principal.AccountID,
		Status: "requested", RequestedAt: command.Now,
	}
	s.exports[export.ID] = export
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "DataExportRequested",
		"data_export", export.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "DataExportRequested", export.ID, command.Now)
	return core.DataExportRequest{ID: export.ID, Status: export.Status, RequestedAt: export.RequestedAt}, nil
}

func (s *Store) ListDataExports(_ context.Context, principal core.Principal) ([]core.DataExportRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.DataExportRequest{}
	for _, export := range s.exports {
		if export.TenantID != principal.TenantID || export.AccountID != principal.AccountID {
			continue
		}
		view := core.DataExportRequest{ID: export.ID, Status: export.Status, RequestedAt: export.RequestedAt}
		if export.ReadyAt != nil {
			readyAt := *export.ReadyAt
			view.ReadyAt = &readyAt
		}
		if export.ExpiresAt != nil {
			expiresAt := *export.ExpiresAt
			view.ExpiresAt = &expiresAt
		}
		result = append(result, view)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].RequestedAt.After(result[right].RequestedAt)
	})
	if len(result) > 10 {
		result = result[:10]
	}
	return result, nil
}

func (s *Store) openDeletionRequest(principal core.Principal) *deletionRequest {
	for _, request := range s.deletions {
		if request.TenantID == principal.TenantID && request.AccountID == principal.AccountID &&
			(request.Status == "requested" || request.Status == "pending_review") {
			return request
		}
	}
	return nil
}

func deletionView(request *deletionRequest) core.DeletionRequest {
	view := core.DeletionRequest{ID: request.ID, Status: request.Status, RequestedAt: request.RequestedAt}
	if request.CancelledAt != nil {
		cancelledAt := *request.CancelledAt
		view.CancelledAt = &cancelledAt
	}
	return view
}

func (s *Store) DeletionRequest(_ context.Context, principal core.Principal) (core.DeletionRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if open := s.openDeletionRequest(principal); open != nil {
		return deletionView(open), nil
	}
	return core.DeletionRequest{}, core.E(core.CodeNotFound, "no deletion request is open", nil)
}

func (s *Store) CreateDeletionRequest(_ context.Context, command core.CreateDeletionRequestCommand) (core.DeletionRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if open := s.openDeletionRequest(principal); open != nil {
		return core.DeletionRequest{}, core.E(core.CodeConflict, "a deletion request is already open", nil)
	}
	request := &deletionRequest{
		ID: command.RequestID, TenantID: principal.TenantID, AccountID: principal.AccountID,
		Status: "pending_review", RequestedAt: command.Now,
	}
	s.deletions[request.ID] = request
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "DeletionRequested",
		"deletion_request", request.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "DeletionRequested", request.ID, command.Now)
	return deletionView(request), nil
}

func (s *Store) CancelDeletionRequest(_ context.Context, command core.CancelDeletionRequestCommand) (core.DeletionRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	open := s.openDeletionRequest(principal)
	if open == nil {
		return core.DeletionRequest{}, core.E(core.CodeInvalidState, "no deletion request is open", nil)
	}
	cancelled := command.Now
	open.Status = "cancelled"
	open.CancelledAt = &cancelled
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "DeletionRequestCancelled",
		"deletion_request", open.ID, "allow", "", command.Now, nil)
	return deletionView(open), nil
}
