package memory

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

type account struct {
	ID           string
	TenantID     string
	PersonID     string
	FullName     string
	Phone        string
	PasswordHash string
	Status       string
	Roles        map[core.Role]string
}

type delegation struct {
	Result    core.DelegationResult
	TenantID  string
	GrantedBy string
	RevokedAt *time.Time
	Reason    string
}

type student struct {
	ID                  string
	TenantID            string
	PersonID            string
	MembershipID        string
	AccountID           string
	FullName            string
	EnrollmentReference string
	TeacherAccountID    string
	Version             int64
}

type invitation struct {
	ID                     string
	TenantID               string
	AccountID              string
	StudentID              string
	Kind                   string
	Digest                 []byte
	Status                 string
	ExpiresAt              time.Time
	ConsumedIdempotencyKey string
	ConsumedFingerprint    []byte
}

type session struct {
	Material   core.SessionMaterial
	AccountID  string
	TenantID   string
	Status     string
	ReplacedBy string
}

type idempotencyRecord struct {
	Fingerprint []byte
	Response    []byte
	Completed   bool
}

type AuditRecord struct {
	TenantID     string
	ActorID      string
	OperatorID   string
	DelegationID string
	Action       string
	TargetID     string
	Decision     string
	Reason       string
	RecordedAt   time.Time
}

type OutboxRecord struct {
	TenantID    string
	EventType   string
	AggregateID string
	RecordedAt  time.Time
}

type Store struct {
	mu sync.Mutex

	tenants      map[string]string
	accounts     map[string]*account
	accountPhone map[string]string
	delegations  map[string]*delegation
	students     map[string]*student
	enrollments  map[string]string
	firstMinutes map[string][]core.FirstMinute
	invitations  map[string]*invitation
	inviteDigest map[string]string
	sessions     map[string]*session
	accessIndex  map[string]string
	refreshIndex map[string]string
	idempotency  map[string]*idempotencyRecord
	audit        []AuditRecord
	outbox       []OutboxRecord
}

func New() *Store {
	return &Store{
		tenants:      make(map[string]string),
		accounts:     make(map[string]*account),
		accountPhone: make(map[string]string),
		delegations:  make(map[string]*delegation),
		students:     make(map[string]*student),
		enrollments:  make(map[string]string),
		firstMinutes: make(map[string][]core.FirstMinute),
		invitations:  make(map[string]*invitation),
		inviteDigest: make(map[string]string),
		sessions:     make(map[string]*session),
		accessIndex:  make(map[string]string),
		refreshIndex: make(map[string]string),
		idempotency:  make(map[string]*idempotencyRecord),
	}
}

func (s *Store) Ready(context.Context) error { return nil }

func (s *Store) BootstrapOwner(_ context.Context, command core.BootstrapOwnerCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, candidate := range s.accounts {
		if candidate.TenantID == command.TenantID && candidate.Roles[core.RoleOwner] != "" {
			return core.E(core.CodeConflict, "owner is already bootstrapped", nil)
		}
	}
	if _, exists := s.accountPhone[command.Phone]; exists {
		return core.E(core.CodeConflict, "login identifier is unavailable", nil)
	}
	s.tenants[command.TenantID] = command.TenantName
	s.accounts[command.AccountID] = &account{
		ID:       command.AccountID,
		TenantID: command.TenantID,
		PersonID: command.PersonID,
		FullName: command.FullName,
		Phone:    command.Phone,
		Status:   "pending_activation",
		Roles:    map[core.Role]string{core.RoleOwner: command.TenantID},
	}
	s.accountPhone[command.Phone] = command.AccountID
	s.invitations[command.InvitationID] = &invitation{
		ID:        command.InvitationID,
		TenantID:  command.TenantID,
		AccountID: command.AccountID,
		Kind:      "owner_bootstrap",
		Digest:    cloneBytes(command.TokenDigest),
		Status:    "issued",
		ExpiresAt: command.ExpiresAt,
	}
	s.inviteDigest[hex.EncodeToString(command.TokenDigest)] = command.InvitationID
	s.appendOperatorAudit(command.TenantID, command.Operator, "OwnerBootstrapCreated", command.AccountID, "allow", command.Reason, command.Now)
	s.appendOutbox(command.TenantID, "OwnerBootstrapCreated", command.AccountID, command.Now)
	return nil
}

func (s *Store) BootstrapStaff(_ context.Context, command core.BootstrapStaffCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if command.Role != core.RoleAdministrator && command.Role != core.RoleTeacher {
		return core.E(core.CodeInvalidInput, "staff bootstrap role must be Administrator or Teacher", nil)
	}
	if !s.hasRole(command.OwnerAccountID, command.TenantID, core.RoleOwner) {
		return core.E(core.CodeForbidden, "active Owner authorization is required", nil)
	}
	if _, exists := s.accountPhone[command.Phone]; exists {
		return core.E(core.CodeConflict, "login identifier is unavailable", nil)
	}
	s.accounts[command.AccountID] = &account{
		ID: command.AccountID, TenantID: command.TenantID, PersonID: command.PersonID,
		FullName: command.FullName, Phone: command.Phone, Status: "pending_activation",
		Roles: map[core.Role]string{command.Role: command.TenantID},
	}
	s.accountPhone[command.Phone] = command.AccountID
	s.invitations[command.InvitationID] = &invitation{
		ID: command.InvitationID, TenantID: command.TenantID,
		AccountID: command.AccountID, Kind: "staff_activation",
		Digest: cloneBytes(command.TokenDigest), Status: "issued", ExpiresAt: command.ExpiresAt,
	}
	s.inviteDigest[hex.EncodeToString(command.TokenDigest)] = command.InvitationID
	s.appendOperatorAudit(command.TenantID, command.Operator, "StaffBootstrapCreated", command.AccountID, "allow", command.Reason, command.Now)
	s.appendOutbox(command.TenantID, "StaffBootstrapCreated", command.AccountID, command.Now)
	return nil
}

func (s *Store) ReissueBootstrapInvitation(_ context.Context, command core.ReissueBootstrapInvitationCommand) (core.BootstrapInvitationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acct := s.accounts[command.AccountID]
	if acct == nil || acct.TenantID != command.TenantID {
		return core.BootstrapInvitationResult{}, core.E(core.CodeNotFound, "pending bootstrap account not found", nil)
	}
	if acct.Status != "pending_activation" {
		return core.BootstrapInvitationResult{}, core.E(core.CodeInvalidState, "account is not pending activation", nil)
	}
	kind := ""
	if acct.Roles[core.RoleOwner] != "" {
		kind = "owner_bootstrap"
	} else if acct.Roles[core.RoleAdministrator] != "" || acct.Roles[core.RoleTeacher] != "" {
		kind = "staff_activation"
	}
	if kind == "" {
		return core.BootstrapInvitationResult{}, core.E(core.CodeInvalidState, "account is not a bootstrap or staff account", nil)
	}
	foundExisting := false
	for _, existing := range s.invitations {
		if existing.TenantID == command.TenantID && existing.AccountID == command.AccountID && existing.Kind == kind {
			foundExisting = true
			if existing.Status == "issued" {
				existing.Status = "superseded"
			}
		}
	}
	if !foundExisting {
		return core.BootstrapInvitationResult{}, core.E(core.CodeInvalidState, "original bootstrap invitation not found", nil)
	}
	stored := &invitation{
		ID: command.InvitationID, TenantID: command.TenantID, AccountID: command.AccountID,
		Kind: kind, Digest: cloneBytes(command.TokenDigest), Status: "issued", ExpiresAt: command.ExpiresAt,
	}
	s.invitations[stored.ID] = stored
	s.inviteDigest[hex.EncodeToString(stored.Digest)] = stored.ID
	result := core.BootstrapInvitationResult{
		InvitationID: stored.ID, AccountID: stored.AccountID, Kind: stored.Kind,
		Status: stored.Status, ExpiresAt: stored.ExpiresAt,
	}
	s.appendOperatorAudit(command.TenantID, command.Operator, "BootstrapActivationInvitationReissued", stored.ID, "allow", command.Reason, command.Now)
	s.appendOutbox(command.TenantID, "BootstrapActivationInvitationReissued", stored.ID, command.Now)
	return result, nil
}

func (s *Store) PreviewActivation(_ context.Context, digest []byte, now time.Time) (core.ActivationPreview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite := s.findInvite(digest)
	if invite == nil || invite.Status != "issued" || !invite.ExpiresAt.After(now) {
		return core.ActivationPreview{}, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	acct := s.accounts[invite.AccountID]
	if acct == nil || acct.Status != "pending_activation" {
		return core.ActivationPreview{}, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	displayName := acct.FullName
	if displayName == "" {
		displayName = "Belcanto user"
	}
	if invite.StudentID != "" {
		if candidate := s.students[invite.StudentID]; candidate != nil {
			displayName = candidate.FullName
		}
	}
	return core.ActivationPreview{
		InvitationID: invite.ID,
		Kind:         invite.Kind,
		DisplayName:  displayName,
		MaskedPhone:  security.MaskPhone(acct.Phone),
		ExpiresAt:    invite.ExpiresAt,
	}, nil
}

func (s *Store) ValidateActivation(_ context.Context, command core.ActivationValidationCommand) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite := s.findInvite(command.TokenDigest)
	if invite == nil {
		return false, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	if invite.Status == "consumed" && invite.ConsumedIdempotencyKey == command.IdempotencyKey {
		if !security.EqualDigest(invite.ConsumedFingerprint, command.PayloadFingerprint) {
			return false, core.E(core.CodeConflict, "Idempotency-Key was reused with a different payload", nil)
		}
		return true, nil
	}
	if invite.Status != "issued" || !invite.ExpiresAt.After(command.Now) {
		return false, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	acct := s.accounts[invite.AccountID]
	if acct == nil || acct.Status != "pending_activation" || acct.Phone != command.Phone {
		return false, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	return false, nil
}

func (s *Store) CompleteActivation(_ context.Context, command core.ActivationCompleteCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite := s.findInvite(command.TokenDigest)
	if invite == nil {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	if invite.Status == "consumed" && invite.ConsumedIdempotencyKey == command.IdempotencyKey {
		if !security.EqualDigest(invite.ConsumedFingerprint, command.PayloadFingerprint) {
			return core.E(core.CodeConflict, "Idempotency-Key was reused with a different payload", nil)
		}
		return nil
	}
	acct := s.accounts[invite.AccountID]
	if invite.Status != "issued" || !invite.ExpiresAt.After(command.Now) || acct == nil || acct.Status != "pending_activation" || acct.Phone != command.Phone {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	acct.PasswordHash = command.PasswordHash
	acct.Status = "active"
	invite.Status = "consumed"
	invite.ConsumedIdempotencyKey = command.IdempotencyKey
	invite.ConsumedFingerprint = cloneBytes(command.PayloadFingerprint)
	s.appendAudit(invite.TenantID, acct.ID, "", "AccountActivated", acct.ID, "allow", "", command.Now)
	s.appendOutbox(invite.TenantID, "AccountActivated", acct.ID, command.Now)
	return nil
}

func (s *Store) CredentialByPhone(_ context.Context, phone string) (core.CredentialRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, exists := s.accountPhone[phone]
	if !exists {
		return core.CredentialRecord{}, core.E(core.CodeNotFound, "credential not found", nil)
	}
	return credentialFromAccount(s.accounts[id]), nil
}

func (s *Store) CredentialByAccount(_ context.Context, accountID string) (core.CredentialRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acct := s.accounts[accountID]
	if acct == nil {
		return core.CredentialRecord{}, core.E(core.CodeNotFound, "credential not found", nil)
	}
	return credentialFromAccount(acct), nil
}

func (s *Store) CreateSession(_ context.Context, accountID, tenantID string, material core.SessionMaterial) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	acct := s.accounts[accountID]
	if acct == nil || acct.TenantID != tenantID || acct.Status != "active" {
		return core.E(core.CodeUnauthenticated, "account cannot create a session", nil)
	}
	stored := &session{Material: cloneSession(material), AccountID: accountID, TenantID: tenantID, Status: "active"}
	s.sessions[material.SessionID] = stored
	s.accessIndex[hex.EncodeToString(material.AccessDigest)] = material.SessionID
	s.refreshIndex[hex.EncodeToString(material.RefreshDigest)] = material.SessionID
	return nil
}

func (s *Store) PrincipalByAccessDigest(_ context.Context, digest []byte, now time.Time) (core.Principal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID, exists := s.accessIndex[hex.EncodeToString(digest)]
	if !exists {
		return core.Principal{}, core.E(core.CodeUnauthenticated, "session not found", nil)
	}
	stored := s.sessions[sessionID]
	if stored == nil {
		return core.Principal{}, core.E(core.CodeUnauthenticated, "session is inactive", nil)
	}
	acct := s.accounts[stored.AccountID]
	if stored.Status != "active" || !stored.Material.AccessExpiresAt.After(now) || acct == nil || acct.Status != "active" {
		return core.Principal{}, core.E(core.CodeUnauthenticated, "session is inactive", nil)
	}
	return core.Principal{AccountID: acct.ID, TenantID: acct.TenantID, SessionID: stored.Material.SessionID, Roles: sortedRoles(acct.Roles)}, nil
}

func (s *Store) RotateSession(_ context.Context, oldRefreshDigest []byte, material core.SessionMaterial, now time.Time) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	oldID, exists := s.refreshIndex[hex.EncodeToString(oldRefreshDigest)]
	if !exists {
		return "", "", core.E(core.CodeUnauthenticated, "refresh token not found", nil)
	}
	old := s.sessions[oldID]
	if old == nil || old.Status != "active" || !old.Material.RefreshExpiresAt.After(now) {
		if old != nil {
			s.revokeFamily(old.TenantID, old.AccountID, old.Material.FamilyID, now)
		}
		return "", "", core.E(core.CodeUnauthenticated, "refresh token cannot be reused", nil)
	}
	acct := s.accounts[old.AccountID]
	if acct == nil || acct.Status != "active" {
		s.revokeFamily(old.TenantID, old.AccountID, old.Material.FamilyID, now)
		return "", "", core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	material.FamilyID = old.Material.FamilyID
	old.Status = "replaced"
	old.ReplacedBy = material.SessionID
	stored := &session{Material: cloneSession(material), AccountID: old.AccountID, TenantID: old.TenantID, Status: "active"}
	s.sessions[material.SessionID] = stored
	s.accessIndex[hex.EncodeToString(material.AccessDigest)] = material.SessionID
	s.refreshIndex[hex.EncodeToString(material.RefreshDigest)] = material.SessionID
	return old.AccountID, old.TenantID, nil
}

func (s *Store) RevokeSession(_ context.Context, accessDigest []byte, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sessionID := s.accessIndex[hex.EncodeToString(accessDigest)]; sessionID != "" {
		if stored := s.sessions[sessionID]; stored != nil {
			s.revokeFamily(stored.TenantID, stored.AccountID, stored.Material.FamilyID, now)
		}
	}
	return nil
}

func (s *Store) GrantDelegation(_ context.Context, command core.GrantDelegationCommand) (core.DelegationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasRole(command.OwnerAccountID, command.TenantID, core.RoleOwner) {
		s.appendAudit(command.TenantID, command.OwnerAccountID, "", "StudentOnboardingDelegationGranted", command.AdministratorID, "deny", "owner_required", command.Now)
		return core.DelegationResult{}, core.E(core.CodeForbidden, "only Owner can grant privileged access", nil)
	}
	if response, ok, err := s.replay("grant_delegation", command.TenantID, command.OwnerAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.DelegationResult{}, err
		}
		var result core.DelegationResult
		if err := json.Unmarshal(response, &result); err != nil {
			return core.DelegationResult{}, core.E(core.CodeInternal, "decode idempotent delegation result", err)
		}
		return result, nil
	}
	if command.OwnerAccountID == command.AdministratorID || !s.hasRole(command.AdministratorID, command.TenantID, core.RoleAdministrator) {
		return core.DelegationResult{}, core.E(core.CodeInvalidInput, "target must be an active Administrator in the same school", nil)
	}
	for _, existing := range s.delegations {
		if existing.TenantID == command.TenantID && existing.Result.AdministratorID == command.AdministratorID && existing.Result.Bundle == command.Bundle && existing.Result.Status == "active" && (existing.Result.ExpiresAt == nil || existing.Result.ExpiresAt.After(command.Now)) {
			return core.DelegationResult{}, core.E(core.CodeConflict, "Administrator already has an active delegation", nil)
		}
	}
	result := core.DelegationResult{
		ID: command.ID, AdministratorID: command.AdministratorID, Bundle: command.Bundle,
		Status: "active", GrantedAt: command.Now, ExpiresAt: command.ExpiresAt,
	}
	s.delegations[command.ID] = &delegation{Result: result, TenantID: command.TenantID, GrantedBy: command.OwnerAccountID, Reason: command.Reason}
	if err := s.completeIdempotency("grant_delegation", command.TenantID, command.OwnerAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.DelegationResult{}, err
	}
	s.appendAudit(command.TenantID, command.OwnerAccountID, command.ID, "StudentOnboardingDelegationGranted", command.AdministratorID, "allow", command.Reason, command.Now)
	s.appendOutbox(command.TenantID, "StudentOnboardingDelegationGranted", command.ID, command.Now)
	return result, nil
}

func (s *Store) RevokeDelegation(_ context.Context, command core.RevokeDelegationCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasRole(command.OwnerAccountID, command.TenantID, core.RoleOwner) {
		return core.E(core.CodeForbidden, "only Owner can revoke privileged access", nil)
	}
	if _, ok, err := s.replay("revoke_delegation", command.TenantID, command.OwnerAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return err
	}
	grant := s.delegations[command.DelegationID]
	if grant == nil || grant.TenantID != command.TenantID {
		return core.E(core.CodeNotFound, "delegation not found", nil)
	}
	if grant.Result.Status == "revoked" {
		return core.E(core.CodeConflict, "delegation is already revoked", nil)
	}
	grant.Result.Status = "revoked"
	grant.RevokedAt = ptrTime(command.Now)
	if err := s.completeIdempotency("revoke_delegation", command.TenantID, command.OwnerAccountID, command.IdempotencyKey, command.PayloadFingerprint, struct{}{}); err != nil {
		return err
	}
	s.appendAudit(command.TenantID, command.OwnerAccountID, command.DelegationID, "StudentOnboardingDelegationRevoked", command.DelegationID, "allow", command.Reason, command.Now)
	s.appendOutbox(command.TenantID, "StudentOnboardingDelegationRevoked", command.DelegationID, command.Now)
	return nil
}

func (s *Store) CreateStudent(_ context.Context, command core.CreateStudentCommand) (core.StudentResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delegationID, authorized := s.onboardingAuthority(command.ActorAccountID, command.TenantID, command.Now)
	if !authorized {
		s.appendAudit(command.TenantID, command.ActorAccountID, delegationID, "StudentCreated", command.StudentID, "deny", "student_create_not_allowed", command.Now)
		return core.StudentResult{}, core.E(core.CodeForbidden, "student onboarding permission is required", nil)
	}
	if response, ok, err := s.replay("create_student", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.StudentResult{}, err
		}
		var result core.StudentResult
		if err := json.Unmarshal(response, &result); err != nil {
			return core.StudentResult{}, core.E(core.CodeInternal, "decode idempotent student result", err)
		}
		return result, nil
	}
	if !s.hasRole(command.TeacherAccountID, command.TenantID, core.RoleTeacher) {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, "assigned teacher is not active in this school", nil)
	}
	if _, exists := s.enrollments[command.TenantID+"\x00"+command.EnrollmentReference]; exists {
		return core.StudentResult{}, core.E(core.CodeConflict, "enrollment reference already exists", nil)
	}
	if _, exists := s.accountPhone[command.Phone]; exists {
		return core.StudentResult{}, core.E(core.CodeConflict, "login identifier is unavailable", nil)
	}
	acct := &account{
		ID: command.AccountID, TenantID: command.TenantID, PersonID: command.PersonID,
		FullName: command.FullName, Phone: command.Phone, Status: "pending_activation",
		Roles: map[core.Role]string{core.RoleStudent: command.StudentID},
	}
	s.accounts[acct.ID] = acct
	s.accountPhone[command.Phone] = acct.ID
	s.students[command.StudentID] = &student{
		ID: command.StudentID, TenantID: command.TenantID, PersonID: command.PersonID,
		MembershipID: command.MembershipID, AccountID: command.AccountID, FullName: command.FullName,
		EnrollmentReference: command.EnrollmentReference, TeacherAccountID: command.TeacherAccountID,
	}
	s.enrollments[command.TenantID+"\x00"+command.EnrollmentReference] = command.StudentID
	result := core.StudentResult{StudentID: command.StudentID, AccountID: command.AccountID, OnboardingState: "awaiting_first_minute"}
	if err := s.completeIdempotency("create_student", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.StudentResult{}, err
	}
	s.appendAudit(command.TenantID, command.ActorAccountID, delegationID, "StudentCreated", command.StudentID, "allow", "", command.Now)
	s.appendOutbox(command.TenantID, "StudentCreated", command.StudentID, command.Now)
	return result, nil
}

func (s *Store) PublishFirstMinute(_ context.Context, command core.PublishFirstMinuteCommand) (core.FirstMinute, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	studentRecord := s.students[command.StudentID]
	if studentRecord == nil || studentRecord.TenantID != command.TenantID {
		return core.FirstMinute{}, core.E(core.CodeNotFound, "student not found", nil)
	}
	if studentRecord.TeacherAccountID != command.ActorAccountID || !s.hasRole(command.ActorAccountID, command.TenantID, core.RoleTeacher) {
		return core.FirstMinute{}, core.E(core.CodeForbidden, "only the assigned Teacher can publish this first minute", nil)
	}
	if response, ok, err := s.replay("publish_first_minute", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.FirstMinute{}, err
		}
		var result core.FirstMinute
		if err := json.Unmarshal(response, &result); err != nil {
			return core.FirstMinute{}, core.E(core.CodeInternal, "decode idempotent first-minute result", err)
		}
		return result, nil
	}
	if studentRecord.Version != command.ExpectedVersion {
		return core.FirstMinute{}, core.E(core.CodeConflict, "student version is stale", nil)
	}
	studentRecord.Version++
	result := core.FirstMinute{
		StudentID: command.StudentID, Revision: studentRecord.Version,
		WhatWorked: command.WhatWorked, CurrentFocus: command.CurrentFocus,
		NextStep: command.NextStep, PublishedAt: command.Now,
	}
	s.firstMinutes[command.StudentID] = append(s.firstMinutes[command.StudentID], result)
	if err := s.completeIdempotency("publish_first_minute", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.FirstMinute{}, err
	}
	s.appendAudit(command.TenantID, command.ActorAccountID, "", "FirstBelcantoMinutePublished", command.StudentID, "allow", "", command.Now)
	s.appendOutbox(command.TenantID, "FirstBelcantoMinutePublished", command.StudentID, command.Now)
	return result, nil
}

func (s *Store) IssueInvitation(_ context.Context, command core.IssueInvitationCommand) (core.InvitationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasRole(command.ActorAccountID, command.TenantID, core.RoleOwner) {
		action := "StudentActivationInvitationIssued"
		if command.Mode == core.InvitationReissue {
			action = "StudentActivationInvitationReissued"
		}
		s.appendAudit(command.TenantID, command.ActorAccountID, "", action, command.StudentID, "deny", "owner_required", command.Now)
		return core.InvitationResult{}, core.E(core.CodeForbidden, "only Owner can manage student invitations", nil)
	}
	if response, ok, err := s.replay("issue_invitation:"+string(command.Mode), command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.InvitationResult{}, err
		}
		var result core.InvitationResult
		if err := json.Unmarshal(response, &result); err != nil {
			return core.InvitationResult{}, core.E(core.CodeInternal, "decode idempotent invitation result", err)
		}
		return result, nil
	}
	studentRecord := s.students[command.StudentID]
	if studentRecord == nil || studentRecord.TenantID != command.TenantID {
		return core.InvitationResult{}, core.E(core.CodeNotFound, "student not found", nil)
	}
	acct := s.accounts[studentRecord.AccountID]
	if acct == nil || acct.Status != "pending_activation" {
		return core.InvitationResult{}, core.E(core.CodeInvalidState, "student account is not pending activation", nil)
	}
	if len(s.firstMinutes[command.StudentID]) == 0 {
		return core.InvitationResult{}, core.E(core.CodeInvalidState, "First Belcanto Minute must be published before invitation", nil)
	}
	var active *invitation
	for _, existing := range s.invitations {
		if existing.StudentID == command.StudentID && existing.Status == "issued" && existing.ExpiresAt.After(command.Now) {
			active = existing
			break
		}
	}
	if command.Mode == core.InvitationIssue && active != nil {
		return core.InvitationResult{}, core.E(core.CodeConflict, "an active invitation already exists", nil)
	}
	if command.Mode == core.InvitationReissue && active == nil {
		return core.InvitationResult{}, core.E(core.CodeInvalidState, "an active invitation is required for reissue", nil)
	}
	if command.Mode == core.InvitationReissue && active != nil {
		active.Status = "superseded"
	}
	stored := &invitation{
		ID: command.InvitationID, TenantID: command.TenantID, AccountID: studentRecord.AccountID,
		StudentID: command.StudentID, Kind: "student_activation", Digest: cloneBytes(command.TokenDigest),
		Status: "issued", ExpiresAt: command.ExpiresAt,
	}
	s.invitations[stored.ID] = stored
	s.inviteDigest[hex.EncodeToString(command.TokenDigest)] = stored.ID
	result := core.InvitationResult{
		InvitationID: stored.ID, StudentID: stored.StudentID, Status: stored.Status,
		ExpiresAt: stored.ExpiresAt,
	}
	if err := s.completeIdempotency("issue_invitation:"+string(command.Mode), command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.InvitationResult{}, err
	}
	s.appendAudit(command.TenantID, command.ActorAccountID, "", "StudentActivationInvitationIssued", stored.ID, "allow", "", command.Now)
	s.appendOutbox(command.TenantID, "StudentActivationInvitationIssued", stored.ID, command.Now)
	return result, nil
}

func (s *Store) RevokeInvitation(_ context.Context, command core.RevokeInvitationCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasRole(command.ActorAccountID, command.TenantID, core.RoleOwner) {
		s.appendAudit(command.TenantID, command.ActorAccountID, "", "StudentActivationInvitationRevoked", command.InvitationID, "deny", "owner_required", command.Now)
		return core.E(core.CodeForbidden, "only Owner can manage student invitations", nil)
	}
	if _, ok, err := s.replay("revoke_invitation", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return err
	}
	invite := s.invitations[command.InvitationID]
	if invite == nil || invite.TenantID != command.TenantID {
		return core.E(core.CodeNotFound, "invitation not found", nil)
	}
	if invite.Status != "issued" {
		return core.E(core.CodeInvalidState, "only an issued invitation can be revoked", nil)
	}
	invite.Status = "revoked"
	if err := s.completeIdempotency("revoke_invitation", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, struct{}{}); err != nil {
		return err
	}
	s.appendAudit(command.TenantID, command.ActorAccountID, "", "StudentActivationInvitationRevoked", invite.ID, "allow", "", command.Now)
	s.appendOutbox(command.TenantID, "StudentActivationInvitationRevoked", invite.ID, command.Now)
	return nil
}

func (s *Store) BootstrapView(_ context.Context, principal core.Principal, now time.Time) (core.BootstrapView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acct := s.accounts[principal.AccountID]
	if acct == nil || acct.TenantID != principal.TenantID || acct.Status != "active" {
		return core.BootstrapView{}, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	view := core.BootstrapView{
		AccountID: acct.ID, Roles: sortedRoles(acct.Roles),
		AccessProfiles: []string{}, Permissions: []string{},
	}
	if s.hasRole(acct.ID, acct.TenantID, core.RoleOwner) {
		view.Permissions = append(view.Permissions, core.OwnerStudentOnboardingPermissionSet()...)
	} else if delegationID, allowed := s.onboardingAuthority(acct.ID, acct.TenantID, now); allowed && delegationID != "" {
		view.AccessProfiles = append(view.AccessProfiles, core.StudentOnboardingManagerV1)
		view.Permissions = append(view.Permissions, core.StudentOnboardingManagerV1PermissionSet()...)
	}
	studentID := acct.Roles[core.RoleStudent]
	if studentID == "" {
		return view, nil
	}
	studentRecord := s.students[studentID]
	if studentRecord == nil || studentRecord.AccountID != acct.ID {
		return core.BootstrapView{}, core.E(core.CodeInternal, "student role scope is inconsistent", nil)
	}
	view.StudentID = studentID
	view.FullName = studentRecord.FullName
	revisions := s.firstMinutes[studentID]
	if len(revisions) > 0 {
		latest := revisions[len(revisions)-1]
		view.FirstMinute = &latest
	}
	return view, nil
}

func (s *Store) ListStaff(_ context.Context, principal core.Principal, role core.Role, now time.Time) ([]core.StaffMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owner := s.hasRole(principal.AccountID, principal.TenantID, core.RoleOwner)
	allowed := owner
	if !owner && role == core.RoleTeacher {
		delegationID, delegated := s.onboardingAuthority(principal.AccountID, principal.TenantID, now)
		allowed = delegated && delegationID != ""
	}
	if !allowed {
		s.appendAudit(principal.TenantID, principal.AccountID, "", "StaffListed", string(role), "deny", "staff_discovery_not_allowed", now)
		return nil, core.E(core.CodeForbidden, "staff discovery permission is required", nil)
	}
	result := make([]core.StaffMember, 0)
	for _, candidate := range s.accounts {
		if candidate.TenantID != principal.TenantID || candidate.Status != "active" || candidate.Roles[role] == "" {
			continue
		}
		member := core.StaffMember{
			AccountID: candidate.ID, FullName: candidate.FullName,
			Roles: sortedRoles(candidate.Roles), AccessProfiles: []string{},
		}
		if role == core.RoleAdministrator {
			if delegationID, delegated := s.onboardingAuthority(candidate.ID, principal.TenantID, now); delegated && delegationID != "" {
				member.AccessProfiles = append(member.AccessProfiles, core.StudentOnboardingManagerV1)
				member.OnboardingDelegationID = delegationID
				member.OnboardingDelegationExpiresAt = s.delegations[delegationID].Result.ExpiresAt
			}
		}
		result = append(result, member)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].FullName == result[right].FullName {
			return result[left].AccountID < result[right].AccountID
		}
		return result[left].FullName < result[right].FullName
	})
	s.appendAudit(principal.TenantID, principal.AccountID, "", "StaffListed", string(role), "allow", "", now)
	return result, nil
}

func (s *Store) ListStudentOnboarding(_ context.Context, principal core.Principal, now time.Time) ([]core.StudentOnboardingItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delegationID, onboardingAllowed := s.onboardingAuthority(principal.AccountID, principal.TenantID, now)
	teacherAllowed := s.hasRole(principal.AccountID, principal.TenantID, core.RoleTeacher)
	if !onboardingAllowed && !teacherAllowed {
		s.appendAudit(principal.TenantID, principal.AccountID, "", "StudentOnboardingListed", "queue", "deny", "student_onboarding_read_not_allowed", now)
		return nil, core.E(core.CodeForbidden, "student onboarding read permission is required", nil)
	}
	teacherOnly := !onboardingAllowed && teacherAllowed
	result := make([]core.StudentOnboardingItem, 0)
	for _, studentRecord := range s.students {
		if studentRecord.TenantID != principal.TenantID || (teacherOnly && studentRecord.TeacherAccountID != principal.AccountID) {
			continue
		}
		accountRecord := s.accounts[studentRecord.AccountID]
		if accountRecord == nil {
			continue
		}
		item := core.StudentOnboardingItem{
			StudentID: studentRecord.ID, FullName: studentRecord.FullName,
			EnrollmentReference: studentRecord.EnrollmentReference,
			TeacherAccountID:    studentRecord.TeacherAccountID,
			StudentVersion:      studentRecord.Version,
			OnboardingState:     core.OnboardingAwaitingFirstMinute,
		}
		if accountRecord.Status != "pending_activation" {
			item.OnboardingState = core.OnboardingActivated
		} else {
			for _, invitationRecord := range s.invitations {
				if invitationRecord.StudentID == studentRecord.ID && invitationRecord.Status == "issued" && invitationRecord.ExpiresAt.After(now) {
					item.OnboardingState = core.OnboardingInvited
					item.InvitationID = invitationRecord.ID
					expiresAt := invitationRecord.ExpiresAt
					item.InvitationExpiresAt = &expiresAt
					break
				}
			}
			if item.OnboardingState != core.OnboardingInvited && len(s.firstMinutes[studentRecord.ID]) > 0 {
				item.OnboardingState = core.OnboardingReadyToInvite
			}
		}
		result = append(result, item)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].FullName == result[right].FullName {
			return result[left].StudentID < result[right].StudentID
		}
		return result[left].FullName < result[right].FullName
	})
	s.appendAudit(principal.TenantID, principal.AccountID, delegationID, "StudentOnboardingListed", "queue", "allow", "", now)
	return result, nil
}

// SeedActiveStaff is intentionally not part of the production Store interface.
// It supports deterministic tests until the separate staff-onboarding slice exists.
func (s *Store) SeedActiveStaff(tenantID, accountID, personID, phone, passwordHash string, roles ...core.Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.accountPhone[phone]; exists {
		return fmt.Errorf("phone already exists")
	}
	roleMap := make(map[core.Role]string, len(roles))
	for _, role := range roles {
		roleMap[role] = tenantID
	}
	s.accounts[accountID] = &account{ID: accountID, TenantID: tenantID, PersonID: personID, FullName: accountID, Phone: phone, PasswordHash: passwordHash, Status: "active", Roles: roleMap}
	s.accountPhone[phone] = accountID
	if _, exists := s.tenants[tenantID]; !exists {
		s.tenants[tenantID] = "Test Belcanto"
	}
	return nil
}

// SetAssignedTeacherForTest is intentionally outside the production Store
// interface. It lets command-boundary tests model a current assignment change.
func (s *Store) SetAssignedTeacherForTest(tenantID, studentID, teacherAccountID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	studentRecord := s.students[studentID]
	if studentRecord == nil || studentRecord.TenantID != tenantID {
		return fmt.Errorf("student not found")
	}
	if !s.hasRole(teacherAccountID, tenantID, core.RoleTeacher) {
		return fmt.Errorf("teacher is not active")
	}
	studentRecord.TeacherAccountID = teacherAccountID
	return nil
}

func (s *Store) AuditRecords() []AuditRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]AuditRecord, len(s.audit))
	copy(result, s.audit)
	return result
}

func (s *Store) onboardingAuthority(accountID, tenantID string, now time.Time) (string, bool) {
	if s.hasRole(accountID, tenantID, core.RoleOwner) {
		return "", true
	}
	if !s.hasRole(accountID, tenantID, core.RoleAdministrator) {
		return "", false
	}
	for id, grant := range s.delegations {
		if grant.TenantID == tenantID && grant.Result.AdministratorID == accountID && grant.Result.Bundle == core.StudentOnboardingManagerV1 && grant.Result.Status == "active" && (grant.Result.ExpiresAt == nil || grant.Result.ExpiresAt.After(now)) {
			return id, true
		}
	}
	return "", false
}

func (s *Store) hasRole(accountID, tenantID string, role core.Role) bool {
	acct := s.accounts[accountID]
	return acct != nil && acct.TenantID == tenantID && acct.Status == "active" && acct.Roles[role] != ""
}

func (s *Store) findInvite(digest []byte) *invitation {
	id := s.inviteDigest[hex.EncodeToString(digest)]
	return s.invitations[id]
}

func (s *Store) revokeFamily(tenantID, accountID, familyID string, _ time.Time) {
	for _, candidate := range s.sessions {
		if candidate.TenantID == tenantID && candidate.AccountID == accountID && candidate.Material.FamilyID == familyID {
			candidate.Status = "revoked"
		}
	}
}

func (s *Store) replay(scope, tenantID, actorAccountID, key string, fingerprint []byte) ([]byte, bool, error) {
	compound := tenantID + "\x00" + actorAccountID + "\x00" + scope + "\x00" + key
	record := s.idempotency[compound]
	if record == nil {
		return nil, false, nil
	}
	if !security.EqualDigest(record.Fingerprint, fingerprint) {
		return nil, false, core.E(core.CodeConflict, "Idempotency-Key was reused with a different payload", nil)
	}
	if !record.Completed {
		return nil, false, core.E(core.CodeConflict, "the idempotent operation is still processing", nil)
	}
	return cloneBytes(record.Response), true, nil
}

func (s *Store) completeIdempotency(scope, tenantID, actorAccountID, key string, fingerprint []byte, response any) error {
	encoded, err := json.Marshal(response)
	if err != nil {
		return core.E(core.CodeInternal, "encode idempotency result", err)
	}
	s.idempotency[tenantID+"\x00"+actorAccountID+"\x00"+scope+"\x00"+key] = &idempotencyRecord{Fingerprint: cloneBytes(fingerprint), Response: encoded, Completed: true}
	return nil
}

func (s *Store) appendAudit(tenantID, actorID, delegationID, action, targetID, decision, reason string, at time.Time) {
	s.audit = append(s.audit, AuditRecord{TenantID: tenantID, ActorID: actorID, DelegationID: delegationID, Action: action, TargetID: targetID, Decision: decision, Reason: reason, RecordedAt: at})
}

func (s *Store) appendOperatorAudit(tenantID, operatorID, action, targetID, decision, reason string, at time.Time) {
	s.audit = append(s.audit, AuditRecord{TenantID: tenantID, OperatorID: operatorID, Action: action, TargetID: targetID, Decision: decision, Reason: reason, RecordedAt: at})
}

func (s *Store) appendOutbox(tenantID, eventType, aggregateID string, at time.Time) {
	s.outbox = append(s.outbox, OutboxRecord{TenantID: tenantID, EventType: eventType, AggregateID: aggregateID, RecordedAt: at})
}

func credentialFromAccount(acct *account) core.CredentialRecord {
	return core.CredentialRecord{AccountID: acct.ID, TenantID: acct.TenantID, Phone: acct.Phone, PasswordHash: acct.PasswordHash, Status: acct.Status, Roles: sortedRoles(acct.Roles)}
}

func sortedRoles(roleMap map[core.Role]string) []core.Role {
	roles := make([]core.Role, 0, len(roleMap))
	for role := range roleMap {
		roles = append(roles, role)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
	return roles
}

func cloneBytes(value []byte) []byte {
	return append([]byte(nil), value...)
}

func cloneSession(value core.SessionMaterial) core.SessionMaterial {
	value.AccessDigest = cloneBytes(value.AccessDigest)
	value.RefreshDigest = cloneBytes(value.RefreshDigest)
	return value
}

func ptrTime(value time.Time) *time.Time { return &value }
