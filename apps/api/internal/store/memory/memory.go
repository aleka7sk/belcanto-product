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

type teacherAssignment struct {
	ID               string
	TenantID         string
	StudentID        string
	TeacherAccountID string
	EffectiveFrom    time.Time
	EffectiveUntil   *time.Time
	Version          int64
	Status           string
}

type lesson struct {
	ID               string
	TenantID         string
	SeriesID         string
	Title            string
	StartsAt         time.Time
	DurationMinutes  int
	Location         string
	TeacherAccountID string
	StudentIDs       []string
	Status           core.LessonStatus
	Version          int64
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
	LastSeenAt *time.Time
}

type idempotencyRecord struct {
	Fingerprint []byte
	Response    []byte
	Completed   bool
}

type AuditRecord struct {
	ID           int64
	TargetType   string
	TenantID     string
	ActorID      string
	OperatorID   string
	DelegationID string
	Action       string
	TargetID     string
	Decision     string
	Reason       string
	Metadata     map[string]any
	RecordedAt   time.Time
}

type OutboxRecord struct {
	ID            int64
	TenantID      string
	EventType     string
	AggregateType string
	AggregateID   string
	Payload       []byte
	RecordedAt    time.Time
	Status        string
	AttemptCount  int
	NextAttemptAt *time.Time
	LastError     string
}

type Store struct {
	mu sync.Mutex

	tenants            map[string]string
	accounts           map[string]*account
	accountPhone       map[string]string
	delegations        map[string]*delegation
	students           map[string]*student
	assignments        map[string][]*teacherAssignment
	logicalTimes       map[string]time.Time
	lessons            map[string]*lesson
	rooms              map[string]*room
	lessonSeries       map[string]*coreLessonSeries
	eventCategories    map[string]*eventCategory
	eventSeriesMap     map[string]*eventSeries
	eventOccurrences   map[string]*eventOccurrence
	eventRsvps         map[string]map[string]*eventRsvp
	eventWaitlists     map[string][]*waitlistEntry
	rescheduleRequests map[string]*rescheduleRequest
	journals           map[string]*lessonJournal
	evidence           []*progressEvidenceRecord
	homework           map[string]*homeworkRecord
	mediaObjects       map[string]*mediaObject
	attendance         map[string]*attendanceRecord
	songs              map[string]*studentSong
	goals              map[string]*studentGoal
	achievementDefs    map[string]*achievementDefinition
	awards             map[string]*achievementAward
	activity           []*activityEntry
	notificationPrefs  map[string]bool
	spotOffers         map[string]*spotOffer
	enrollments        map[string]string
	firstMinutes       map[string][]core.FirstMinute
	invitations        map[string]*invitation
	inviteDigest       map[string]string
	sessions           map[string]*session
	accessIndex        map[string]string
	refreshIndex       map[string]string
	resets             map[string]*passwordReset
	resetDigest        map[string]string
	contacts           map[string]*verifiedContact
	contactVerifs      map[string]*contactVerification
	twofaSecrets       map[string]*twofaSecret
	recoveryCodes      []*recoveryCode
	twofaChallenges    map[string]*twofaChallenge
	challengeDigest    map[string]string
	activationProg     map[string]*activationProgress
	policies           map[string]*policyVersion
	acceptances        map[string]*policyAcceptance
	privacy            map[string]*privacyRecord
	exports            map[string]*dataExport
	deletions          map[string]*deletionRequest
	idempotency        map[string]*idempotencyRecord
	audit              []AuditRecord
	outbox             []OutboxRecord
}

func New() *Store {
	return &Store{
		tenants:            make(map[string]string),
		accounts:           make(map[string]*account),
		accountPhone:       make(map[string]string),
		delegations:        make(map[string]*delegation),
		students:           make(map[string]*student),
		assignments:        make(map[string][]*teacherAssignment),
		logicalTimes:       make(map[string]time.Time),
		lessons:            make(map[string]*lesson),
		rooms:              make(map[string]*room),
		lessonSeries:       make(map[string]*coreLessonSeries),
		eventCategories:    make(map[string]*eventCategory),
		eventSeriesMap:     make(map[string]*eventSeries),
		eventOccurrences:   make(map[string]*eventOccurrence),
		eventRsvps:         make(map[string]map[string]*eventRsvp),
		eventWaitlists:     make(map[string][]*waitlistEntry),
		rescheduleRequests: make(map[string]*rescheduleRequest),
		journals:           make(map[string]*lessonJournal),
		homework:           make(map[string]*homeworkRecord),
		mediaObjects:       make(map[string]*mediaObject),
		attendance:         make(map[string]*attendanceRecord),
		songs:              make(map[string]*studentSong),
		goals:              make(map[string]*studentGoal),
		achievementDefs:    make(map[string]*achievementDefinition),
		awards:             make(map[string]*achievementAward),
		notificationPrefs:  make(map[string]bool),
		spotOffers:         make(map[string]*spotOffer),
		enrollments:        make(map[string]string),
		firstMinutes:       make(map[string][]core.FirstMinute),
		invitations:        make(map[string]*invitation),
		inviteDigest:       make(map[string]string),
		sessions:           make(map[string]*session),
		accessIndex:        make(map[string]string),
		refreshIndex:       make(map[string]string),
		resets:             make(map[string]*passwordReset),
		resetDigest:        make(map[string]string),
		contacts:           make(map[string]*verifiedContact),
		contactVerifs:      make(map[string]*contactVerification),
		twofaSecrets:       make(map[string]*twofaSecret),
		twofaChallenges:    make(map[string]*twofaChallenge),
		challengeDigest:    make(map[string]string),
		activationProg:     make(map[string]*activationProgress),
		policies:           make(map[string]*policyVersion),
		acceptances:        make(map[string]*policyAcceptance),
		privacy:            make(map[string]*privacyRecord),
		exports:            make(map[string]*dataExport),
		deletions:          make(map[string]*deletionRequest),
		idempotency:        make(map[string]*idempotencyRecord),
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
	s.appendSecurityAudit(tenantID, accountID, "SessionCreated", "session", material.SessionID, "allow", "", material.CreatedAt, nil)
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
			s.appendSecurityAudit(old.TenantID, old.AccountID, "RefreshTokenReuseDetected",
				"session_family", old.Material.FamilyID, "deny", "inactive_or_reused_refresh_token", now, nil)
		}
		return "", "", core.E(core.CodeUnauthenticated, "refresh token cannot be reused", nil)
	}
	acct := s.accounts[old.AccountID]
	if acct == nil || acct.Status != "active" {
		s.revokeFamily(old.TenantID, old.AccountID, old.Material.FamilyID, now)
		return "", "", core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	material.FamilyID = old.Material.FamilyID
	if material.DeviceLabel == "" {
		material.DeviceLabel = old.Material.DeviceLabel
	}
	if material.Platform == "" {
		material.Platform = old.Material.Platform
	}
	old.Status = "replaced"
	old.ReplacedBy = material.SessionID
	lastSeen := now
	stored := &session{Material: cloneSession(material), AccountID: old.AccountID, TenantID: old.TenantID, Status: "active", LastSeenAt: &lastSeen}
	s.sessions[material.SessionID] = stored
	s.accessIndex[hex.EncodeToString(material.AccessDigest)] = material.SessionID
	s.refreshIndex[hex.EncodeToString(material.RefreshDigest)] = material.SessionID
	s.appendSecurityAudit(old.TenantID, old.AccountID, "SessionRefreshed", "session", material.SessionID, "allow", "", now, nil)
	return old.AccountID, old.TenantID, nil
}

func (s *Store) RevokeSession(_ context.Context, accessDigest []byte, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sessionID := s.accessIndex[hex.EncodeToString(accessDigest)]; sessionID != "" {
		if stored := s.sessions[sessionID]; stored != nil {
			s.revokeFamily(stored.TenantID, stored.AccountID, stored.Material.FamilyID, now)
			s.appendSecurityAudit(stored.TenantID, stored.AccountID, "SessionRevoked",
				"session_family", stored.Material.FamilyID, "allow", "", now,
				map[string]any{"sessionId": sessionID})
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
	operationAt := s.nextAssignmentOperationTime(command.Now, []string{command.StudentID})
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
	s.assignments[command.StudentID] = append(s.assignments[command.StudentID], &teacherAssignment{
		ID: command.TeacherAssignmentID, TenantID: command.TenantID, StudentID: command.StudentID,
		TeacherAccountID: command.TeacherAccountID, EffectiveFrom: operationAt, Version: 0, Status: "active",
	})
	s.enrollments[command.TenantID+"\x00"+command.EnrollmentReference] = command.StudentID
	result := core.StudentResult{StudentID: command.StudentID, AccountID: command.AccountID, OnboardingState: "awaiting_first_minute"}
	if err := s.completeIdempotency("create_student", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.StudentResult{}, err
	}
	s.recordAssignmentOperationTime(operationAt, []string{command.StudentID})
	s.appendAudit(command.TenantID, command.ActorAccountID, delegationID, "StudentCreated", command.StudentID, "allow", "", operationAt)
	s.appendOutbox(command.TenantID, "StudentCreated", command.StudentID, operationAt)
	return result, nil
}

func (s *Store) PublishFirstMinute(_ context.Context, command core.PublishFirstMinuteCommand) (core.FirstMinute, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	studentRecord := s.students[command.StudentID]
	if studentRecord == nil || studentRecord.TenantID != command.TenantID {
		return core.FirstMinute{}, core.E(core.CodeNotFound, "student not found", nil)
	}
	operationAt := s.nextAssignmentOperationTime(command.Now, []string{command.StudentID})
	assignment := s.assignmentAt(command.StudentID, operationAt)
	if assignment == nil || assignment.TeacherAccountID != command.ActorAccountID || !s.hasRole(command.ActorAccountID, command.TenantID, core.RoleTeacher) {
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
		NextStep: command.NextStep, PublishedAt: operationAt,
	}
	s.firstMinutes[command.StudentID] = append(s.firstMinutes[command.StudentID], result)
	if err := s.completeIdempotency("publish_first_minute", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.FirstMinute{}, err
	}
	s.recordAssignmentOperationTime(operationAt, []string{command.StudentID})
	s.appendAudit(command.TenantID, command.ActorAccountID, "", "FirstBelcantoMinutePublished", command.StudentID, "allow", "", operationAt)
	s.appendOutbox(command.TenantID, "FirstBelcantoMinutePublished", command.StudentID, operationAt)
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
	view.Permissions = append(view.Permissions, core.LessonPermissionSetForRoles(view.Roles)...)
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
	administrator := s.hasRole(principal.AccountID, principal.TenantID, core.RoleAdministrator)
	allowed := owner || (administrator && role == core.RoleTeacher)
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
	projectionAt := s.currentAssignmentProjectionTime(principal.TenantID, now)
	result := make([]core.StudentOnboardingItem, 0)
	for _, studentRecord := range s.students {
		assignment := s.assignmentAt(studentRecord.ID, projectionAt)
		if studentRecord.TenantID != principal.TenantID || assignment == nil || (teacherOnly && assignment.TeacherAccountID != principal.AccountID) {
			continue
		}
		accountRecord := s.accounts[studentRecord.AccountID]
		if accountRecord == nil {
			continue
		}
		item := core.StudentOnboardingItem{
			StudentID: studentRecord.ID, FullName: studentRecord.FullName,
			EnrollmentReference: studentRecord.EnrollmentReference,
			TeacherAccountID:    assignment.TeacherAccountID,
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

func (s *Store) ListStudents(_ context.Context, principal core.Principal, asOf, now time.Time) ([]core.StudentDirectoryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	accountRecord := s.activeAccount(principal.AccountID, principal.TenantID)
	if accountRecord == nil {
		return nil, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	manageAll := accountRecord.Roles[core.RoleOwner] != "" || accountRecord.Roles[core.RoleAdministrator] != ""
	teacherOnly := !manageAll && accountRecord.Roles[core.RoleTeacher] != ""
	if !manageAll && !teacherOnly {
		s.appendAudit(principal.TenantID, principal.AccountID, "", "StudentDirectoryListed", "students", "deny", "student_directory_not_allowed", now)
		return nil, core.E(core.CodeForbidden, "student directory permission is required", nil)
	}
	projectionAt := asOf
	if projectionAt.IsZero() {
		projectionAt = s.currentAssignmentProjectionTime(principal.TenantID, now)
	}
	result := make([]core.StudentDirectoryItem, 0)
	for _, studentRecord := range s.students {
		if studentRecord.TenantID != principal.TenantID || !s.activeStudent(studentRecord) {
			continue
		}
		assignment := s.assignmentAt(studentRecord.ID, projectionAt)
		if assignment == nil || (teacherOnly && assignment.TeacherAccountID != principal.AccountID) {
			continue
		}
		teacherRecord := s.accounts[assignment.TeacherAccountID]
		if teacherRecord == nil || teacherRecord.TenantID != principal.TenantID {
			continue
		}
		teacherStatus := core.AssignedTeacherInactive
		if teacherRecord.Status == "active" && teacherRecord.Roles[core.RoleTeacher] != "" {
			teacherStatus = core.AssignedTeacherActive
		}
		timelineVersion := int64(-1)
		for _, candidate := range s.assignments[studentRecord.ID] {
			if candidate.Version > timelineVersion {
				timelineVersion = candidate.Version
			}
		}
		result = append(result, core.StudentDirectoryItem{
			StudentID: studentRecord.ID,
			FullName:  studentRecord.FullName,
			PrimaryTeacher: core.AssignedTeacherSummary{
				AccountID: teacherRecord.ID,
				FullName:  teacherRecord.FullName,
				Status:    teacherStatus,
			},
			PrimaryTeacherAssignmentVersion: timelineVersion,
		})
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].FullName == result[right].FullName {
			return result[left].StudentID < result[right].StudentID
		}
		return result[left].FullName < result[right].FullName
	})
	s.appendAudit(principal.TenantID, principal.AccountID, "", "StudentDirectoryListed", "students", "allow", "", now)
	return result, nil
}

func (s *Store) ScheduleLesson(_ context.Context, command core.ScheduleLessonCommand) (core.Lesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(command.ActorAccountID, command.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	teacher := actor != nil && actor.Roles[core.RoleTeacher] != ""
	if !manager && !teacher {
		s.appendAudit(command.TenantID, command.ActorAccountID, "", "LessonScheduled", command.LessonID, "deny", "lesson_create_not_allowed", command.Now)
		return core.Lesson{}, core.E(core.CodeForbidden, "Lesson scheduling permission is required", nil)
	}
	if response, ok, err := s.replay("schedule_lesson", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.Lesson{}, err
		}
		var result core.Lesson
		if err := json.Unmarshal(response, &result); err != nil {
			return core.Lesson{}, core.E(core.CodeInternal, "decode idempotent Lesson result", err)
		}
		return result, nil
	}
	if !command.StartsAt.After(command.Now) {
		return core.Lesson{}, core.E(core.CodeInvalidState, "Lesson must start in the future", nil)
	}
	teacherRecord := s.activeAccount(command.TeacherAccountID, command.TenantID)
	if teacherRecord == nil || teacherRecord.Roles[core.RoleTeacher] == "" {
		return core.Lesson{}, core.E(core.CodeInvalidInput, "Teacher is not active in this school", nil)
	}
	if teacher && !manager && command.TeacherAccountID != command.ActorAccountID {
		return core.Lesson{}, core.E(core.CodeForbidden, "Teacher can only schedule Lessons for self", nil)
	}
	if len(command.StudentIDs) == 0 {
		return core.Lesson{}, core.E(core.CodeInvalidInput, "at least one Student is required", nil)
	}
	studentIDs := append([]string(nil), command.StudentIDs...)
	sort.Strings(studentIDs)
	for _, studentID := range studentIDs {
		studentRecord := s.students[studentID]
		if studentRecord == nil || studentRecord.TenantID != command.TenantID || !s.activeStudent(studentRecord) {
			return core.Lesson{}, core.E(core.CodeInvalidInput, "Student is not active in this school", nil)
		}
		if teacher && !manager {
			assignment := s.assignmentAt(studentID, command.StartsAt)
			if assignment == nil || assignment.TeacherAccountID != command.ActorAccountID {
				return core.Lesson{}, core.E(core.CodeForbidden, "Teacher can only schedule Students assigned at Lesson start", nil)
			}
		}
	}
	if s.lessonScheduleConflict(command.TenantID, command.StartsAt, command.DurationMinutes, command.TeacherAccountID, studentIDs, nil) {
		return core.Lesson{}, core.E(core.CodeConflict, "Teacher or Student has an overlapping Lesson", nil)
	}
	stored := &lesson{
		ID: command.LessonID, TenantID: command.TenantID, Title: command.Title,
		StartsAt: command.StartsAt, DurationMinutes: command.DurationMinutes, Location: command.Location,
		TeacherAccountID: command.TeacherAccountID, StudentIDs: studentIDs,
		Status: core.LessonScheduled, Version: 0, CreatedBy: command.ActorAccountID,
		CreatedAt: command.Now, UpdatedAt: command.Now,
	}
	s.lessons[stored.ID] = stored
	result := s.lessonView(stored)
	if err := s.completeIdempotency("schedule_lesson", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.Lesson{}, err
	}
	s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "LessonScheduled", stored.ID, "allow", "", command.Now, map[string]any{
		"teacherAccountId": stored.TeacherAccountID,
		"studentIds":       append([]string(nil), stored.StudentIDs...),
	})
	s.appendOutbox(command.TenantID, "LessonScheduled", stored.ID, command.Now)
	return result, nil
}

func (s *Store) ListLessons(_ context.Context, principal core.Principal, query core.LessonListQuery, now time.Time) ([]core.Lesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	if actor == nil {
		return nil, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	if actor.Roles[core.RoleOwner] == "" && actor.Roles[core.RoleAdministrator] == "" && actor.Roles[core.RoleTeacher] == "" && actor.Roles[core.RoleStudent] == "" {
		return nil, core.E(core.CodeForbidden, "Lesson read permission is required", nil)
	}
	result := make([]core.Lesson, 0)
	for _, stored := range s.lessons {
		if stored.TenantID != principal.TenantID || stored.StartsAt.Before(query.From) || !stored.StartsAt.Before(query.To) {
			continue
		}
		if query.StudentID != "" && !containsString(stored.StudentIDs, query.StudentID) {
			continue
		}
		if query.TeacherAccountID != "" && stored.TeacherAccountID != query.TeacherAccountID {
			continue
		}
		if !s.lessonReadable(actor, stored) {
			continue
		}
		result = append(result, s.lessonViewForActor(actor, stored))
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].StartsAt.Equal(result[right].StartsAt) {
			return result[left].ID < result[right].ID
		}
		return result[left].StartsAt.Before(result[right].StartsAt)
	})
	return result, nil
}

func (s *Store) GetLesson(_ context.Context, principal core.Principal, lessonID string, _ time.Time) (core.Lesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	if actor == nil {
		return core.Lesson{}, core.E(core.CodeUnauthenticated, "account is inactive", nil)
	}
	stored := s.lessons[lessonID]
	if stored == nil || stored.TenantID != principal.TenantID || !s.lessonReadable(actor, stored) {
		return core.Lesson{}, core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	return s.lessonViewForActor(actor, stored), nil
}

func (s *Store) ReplaceLessonTeachers(_ context.Context, command core.ReplaceLessonTeachersCommand) (core.LessonTeacherReplacementResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(command.ActorAccountID, command.TenantID)
	if actor == nil || (actor.Roles[core.RoleOwner] == "" && actor.Roles[core.RoleAdministrator] == "") {
		s.appendAudit(command.TenantID, command.ActorAccountID, "", "LessonTeacherReplaced", "lessons", "deny", "lesson_teacher_replace_not_allowed", command.Now)
		return core.LessonTeacherReplacementResult{}, core.E(core.CodeForbidden, "Lesson Teacher replacement permission is required", nil)
	}
	if response, ok, err := s.replay("replace_lesson_teachers", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.LessonTeacherReplacementResult{}, err
		}
		var result core.LessonTeacherReplacementResult
		if err := json.Unmarshal(response, &result); err != nil {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeInternal, "decode idempotent Lesson Teacher replacement", err)
		}
		return result, nil
	}
	newTeacher := s.activeAccount(command.NewTeacherAccountID, command.TenantID)
	if newTeacher == nil || newTeacher.Roles[core.RoleTeacher] == "" {
		return core.LessonTeacherReplacementResult{}, core.E(core.CodeInvalidInput, "new Teacher is not active in this school", nil)
	}
	selected := make([]*lesson, len(command.Targets))
	seen := make(map[string]struct{}, len(command.Targets))
	for index, target := range command.Targets {
		if _, duplicate := seen[target.LessonID]; duplicate {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeInvalidInput, "Lesson ids must be unique", nil)
		}
		seen[target.LessonID] = struct{}{}
		stored := s.lessons[target.LessonID]
		if stored == nil || stored.TenantID != command.TenantID {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeNotFound, "Lesson not found", nil)
		}
		if stored.Status != core.LessonScheduled || !stored.StartsAt.After(command.Now) {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeInvalidState, "only scheduled future Lessons can change Teacher", nil)
		}
		if stored.Version != target.ExpectedVersion {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeConflict, "Lesson version is stale", nil)
		}
		if stored.TeacherAccountID != target.ExpectedPreviousTeacherAccountID {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeConflict, "Lesson previous Teacher is stale", nil)
		}
		if stored.TeacherAccountID == command.NewTeacherAccountID {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeInvalidState, "new Teacher is already assigned to Lesson", nil)
		}
		selected[index] = stored
	}
	targetIDs := make(map[string]struct{}, len(selected))
	for _, stored := range selected {
		targetIDs[stored.ID] = struct{}{}
	}
	for index, stored := range selected {
		if s.lessonScheduleConflict(command.TenantID, stored.StartsAt, stored.DurationMinutes, command.NewTeacherAccountID, nil, targetIDs) {
			return core.LessonTeacherReplacementResult{}, core.E(core.CodeConflict, "new Teacher has an overlapping Lesson", nil)
		}
		for otherIndex := index + 1; otherIndex < len(selected); otherIndex++ {
			if lessonIntervalsOverlap(stored.StartsAt, stored.DurationMinutes, selected[otherIndex].StartsAt, selected[otherIndex].DurationMinutes) {
				return core.LessonTeacherReplacementResult{}, core.E(core.CodeConflict, "selected Lessons overlap for the new Teacher", nil)
			}
		}
	}
	result := core.LessonTeacherReplacementResult{UpdatedCount: len(selected), Lessons: make([]core.Lesson, 0, len(selected))}
	for _, stored := range selected {
		previousTeacherID := stored.TeacherAccountID
		stored.TeacherAccountID = command.NewTeacherAccountID
		stored.Version++
		stored.UpdatedAt = command.Now
		result.Lessons = append(result.Lessons, s.lessonView(stored))
		s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "LessonTeacherReplaced", stored.ID, "allow", "temporary_teacher_continuity", command.Now, map[string]any{
			"previousTeacherAccountId": previousTeacherID,
			"newTeacherAccountId":      command.NewTeacherAccountID,
			"version":                  stored.Version,
		})
		s.appendOutbox(command.TenantID, "LessonTeacherReplaced", stored.ID, command.Now)
	}
	if err := s.completeIdempotency("replace_lesson_teachers", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.LessonTeacherReplacementResult{}, err
	}
	return result, nil
}

func (s *Store) ReassignPrimaryTeachers(_ context.Context, command core.ReassignPrimaryTeachersCommand) (core.PrimaryTeacherReassignmentResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(command.ActorAccountID, command.TenantID)
	if actor == nil || (actor.Roles[core.RoleOwner] == "" && actor.Roles[core.RoleAdministrator] == "") {
		s.appendAudit(command.TenantID, command.ActorAccountID, "", "StudentPrimaryTeacherReassigned", "students", "deny", "primary_teacher_reassign_not_allowed", command.Now)
		return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeForbidden, "primary Teacher reassignment permission is required", nil)
	}
	if response, ok, err := s.replay("reassign_primary_teachers", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.PrimaryTeacherReassignmentResult{}, err
		}
		var result core.PrimaryTeacherReassignmentResult
		if err := json.Unmarshal(response, &result); err != nil {
			return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInternal, "decode idempotent primary Teacher reassignment", err)
		}
		return result, nil
	}
	switch command.EffectiveMode {
	case core.PrimaryTeacherEffectiveImmediate, core.PrimaryTeacherEffectiveScheduled:
	default:
		return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInvalidInput, "effectiveMode must be immediate or scheduled", nil)
	}
	newTeacher := s.activeAccount(command.NewTeacherAccountID, command.TenantID)
	if newTeacher == nil || newTeacher.Roles[core.RoleTeacher] == "" {
		return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInvalidInput, "new Teacher is not active in this school", nil)
	}
	studentIDs := make([]string, len(command.Targets))
	for index, target := range command.Targets {
		studentIDs[index] = target.StudentID
	}
	operationAt := s.nextAssignmentOperationTime(command.Now, studentIDs)
	effectiveFrom := command.EffectiveFrom
	switch command.EffectiveMode {
	case core.PrimaryTeacherEffectiveImmediate:
		effectiveFrom = operationAt
	case core.PrimaryTeacherEffectiveScheduled:
		if !effectiveFrom.After(operationAt) {
			return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInvalidInput, "scheduled effectiveFrom must be in the future", nil)
		}
	}
	type reassignmentPlan struct {
		target   core.PrimaryTeacherReassignmentTarget
		previous *teacherAssignment
		version  int64
	}
	plans := make([]reassignmentPlan, len(command.Targets))
	seen := make(map[string]struct{}, len(command.Targets))
	for index, target := range command.Targets {
		if _, duplicate := seen[target.StudentID]; duplicate {
			return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInvalidInput, "Student ids must be unique", nil)
		}
		seen[target.StudentID] = struct{}{}
		studentRecord := s.students[target.StudentID]
		if studentRecord == nil || studentRecord.TenantID != command.TenantID || !s.activeStudent(studentRecord) {
			return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInvalidInput, "Student is not active in this school", nil)
		}
		previous := s.assignmentAt(target.StudentID, effectiveFrom)
		if previous == nil {
			return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInvalidState, "Student has no primary Teacher at effectiveFrom", nil)
		}
		timelineVersion := int64(-1)
		for _, candidate := range s.assignments[target.StudentID] {
			if candidate.Version > timelineVersion {
				timelineVersion = candidate.Version
			}
		}
		if timelineVersion != target.ExpectedAssignmentVersion {
			return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeConflict, "primary Teacher assignment version is stale", nil)
		}
		if previous.TeacherAccountID == command.NewTeacherAccountID {
			return core.PrimaryTeacherReassignmentResult{}, core.E(core.CodeInvalidState, "new Teacher is already the Student primary Teacher", nil)
		}
		version := timelineVersion + 1
		plans[index] = reassignmentPlan{target: target, previous: previous, version: version}
	}
	result := core.PrimaryTeacherReassignmentResult{ReassignedCount: len(plans), Assignments: make([]core.PrimaryTeacherReassignment, 0, len(plans))}
	for _, plan := range plans {
		for _, candidate := range s.assignments[plan.target.StudentID] {
			if candidate.Status == "active" && !candidate.EffectiveFrom.Before(effectiveFrom) {
				candidate.Status = "ended"
				until := candidate.EffectiveFrom
				candidate.EffectiveUntil = &until
			}
		}
		if plan.previous.Status == "active" && plan.previous.EffectiveFrom.Before(effectiveFrom) {
			until := effectiveFrom
			plan.previous.EffectiveUntil = &until
		} else if plan.previous.EffectiveFrom.Equal(effectiveFrom) {
			plan.previous.Status = "ended"
			until := effectiveFrom
			plan.previous.EffectiveUntil = &until
		}
		stored := &teacherAssignment{
			ID: plan.target.AssignmentID, TenantID: command.TenantID, StudentID: plan.target.StudentID,
			TeacherAccountID: command.NewTeacherAccountID, EffectiveFrom: effectiveFrom,
			Version: plan.version, Status: "active",
		}
		s.assignments[plan.target.StudentID] = append(s.assignments[plan.target.StudentID], stored)
		if command.EffectiveMode == core.PrimaryTeacherEffectiveImmediate {
			s.students[plan.target.StudentID].TeacherAccountID = command.NewTeacherAccountID
		}
		assignmentResult := core.PrimaryTeacherReassignment{
			StudentID: plan.target.StudentID, PreviousTeacherAccountID: plan.previous.TeacherAccountID,
			NewTeacherAccountID: command.NewTeacherAccountID, EffectiveFrom: effectiveFrom,
			Version: stored.Version,
		}
		result.Assignments = append(result.Assignments, assignmentResult)
		s.appendAuditMetadata(command.TenantID, command.ActorAccountID, "StudentPrimaryTeacherReassigned", plan.target.StudentID, "allow", "primary_teacher_continuity", operationAt, map[string]any{
			"previousTeacherAccountId": plan.previous.TeacherAccountID,
			"newTeacherAccountId":      command.NewTeacherAccountID,
			"effectiveFrom":            effectiveFrom,
			"version":                  stored.Version,
		})
		s.appendOutbox(command.TenantID, "StudentPrimaryTeacherReassigned", plan.target.StudentID, operationAt)
	}
	if err := s.completeIdempotency("reassign_primary_teachers", command.TenantID, command.ActorAccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.PrimaryTeacherReassignmentResult{}, err
	}
	s.recordAssignmentOperationTime(operationAt, studentIDs)
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

// SeedActiveStudentForTest is intentionally outside the production Store
// interface. It creates a fully activated Student projection for role-scope tests.
func (s *Store) SeedActiveStudentForTest(tenantID, accountID, studentID, personID, fullName, phone, teacherAccountID string, assignedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.accountPhone[phone]; exists {
		return fmt.Errorf("phone already exists")
	}
	if !s.hasRole(teacherAccountID, tenantID, core.RoleTeacher) {
		return fmt.Errorf("teacher is not active")
	}
	s.accounts[accountID] = &account{
		ID: accountID, TenantID: tenantID, PersonID: personID, FullName: fullName,
		Phone: phone, Status: "active", Roles: map[core.Role]string{core.RoleStudent: studentID},
	}
	s.accountPhone[phone] = accountID
	s.students[studentID] = &student{
		ID: studentID, TenantID: tenantID, PersonID: personID, AccountID: accountID,
		FullName: fullName, EnrollmentReference: "TEST-" + studentID, TeacherAccountID: teacherAccountID,
	}
	s.assignments[studentID] = append(s.assignments[studentID], &teacherAssignment{
		ID: "assignment_" + studentID, TenantID: tenantID, StudentID: studentID,
		TeacherAccountID: teacherAccountID, EffectiveFrom: assignedAt, Status: "active",
	})
	s.recordAssignmentOperationTime(assignedAt, []string{studentID})
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
	assignment := s.assignmentAt(studentID, time.Now().UTC())
	if assignment == nil {
		return fmt.Errorf("teacher assignment not found")
	}
	assignment.TeacherAccountID = teacherAccountID
	studentRecord.TeacherAccountID = teacherAccountID
	return nil
}

// SetAccountStatusForTest is intentionally outside the production Store
// interface. It supports deterministic inactive-staff continuity tests.
func (s *Store) SetAccountStatusForTest(tenantID, accountID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	accountRecord := s.accounts[accountID]
	if accountRecord == nil || accountRecord.TenantID != tenantID {
		return fmt.Errorf("account not found")
	}
	if status != "active" && status != "suspended" {
		return fmt.Errorf("unsupported account status")
	}
	accountRecord.Status = status
	return nil
}

func (s *Store) AuditRecords() []AuditRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]AuditRecord, len(s.audit))
	copy(result, s.audit)
	return result
}

func (s *Store) OutboxRecords() []OutboxRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]OutboxRecord, len(s.outbox))
	copy(result, s.outbox)
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

func (s *Store) activeAccount(accountID, tenantID string) *account {
	acct := s.accounts[accountID]
	if acct == nil || acct.TenantID != tenantID || acct.Status != "active" {
		return nil
	}
	return acct
}

func (s *Store) activeStudent(studentRecord *student) bool {
	if studentRecord == nil {
		return false
	}
	acct := s.accounts[studentRecord.AccountID]
	return acct != nil && acct.TenantID == studentRecord.TenantID &&
		(acct.Status == "pending_activation" || acct.Status == "active") &&
		acct.Roles[core.RoleStudent] == studentRecord.ID
}

func (s *Store) assignmentAt(studentID string, at time.Time) *teacherAssignment {
	var selected *teacherAssignment
	for _, candidate := range s.assignments[studentID] {
		if candidate.Status != "active" || candidate.EffectiveFrom.After(at) || (candidate.EffectiveUntil != nil && !at.Before(*candidate.EffectiveUntil)) {
			continue
		}
		if selected == nil || selected.EffectiveFrom.Before(candidate.EffectiveFrom) {
			selected = candidate
		}
	}
	return selected
}

func (s *Store) nextAssignmentOperationTime(requestAt time.Time, studentIDs []string) time.Time {
	operationAt := requestAt.UTC()
	for _, studentID := range studentIDs {
		previous := s.logicalTimes[studentID]
		if candidate := previous.Add(time.Microsecond); candidate.After(operationAt) {
			operationAt = candidate
		}
	}
	return operationAt
}

func (s *Store) recordAssignmentOperationTime(operationAt time.Time, studentIDs []string) {
	operationAt = operationAt.UTC()
	for _, studentID := range studentIDs {
		if operationAt.After(s.logicalTimes[studentID]) {
			s.logicalTimes[studentID] = operationAt
		}
	}
}

func (s *Store) currentAssignmentProjectionTime(tenantID string, appNow time.Time) time.Time {
	projectionAt := appNow.UTC()
	for studentID, operationAt := range s.logicalTimes {
		studentRecord := s.students[studentID]
		if studentRecord != nil && studentRecord.TenantID == tenantID && operationAt.After(projectionAt) {
			projectionAt = operationAt
		}
	}
	return projectionAt
}

func (s *Store) lessonScheduleConflict(tenantID string, startsAt time.Time, durationMinutes int, teacherAccountID string, studentIDs []string, excludedLessonIDs map[string]struct{}) bool {
	for _, stored := range s.lessons {
		if stored.TenantID != tenantID || stored.Status != core.LessonScheduled {
			continue
		}
		if _, excluded := excludedLessonIDs[stored.ID]; excluded {
			continue
		}
		if !lessonIntervalsOverlap(startsAt, durationMinutes, stored.StartsAt, stored.DurationMinutes) {
			continue
		}
		if stored.TeacherAccountID == teacherAccountID {
			return true
		}
		for _, studentID := range studentIDs {
			if containsString(stored.StudentIDs, studentID) {
				return true
			}
		}
	}
	return false
}

func lessonIntervalsOverlap(firstStart time.Time, firstDuration int, secondStart time.Time, secondDuration int) bool {
	return firstStart.Before(secondStart.Add(time.Duration(secondDuration)*time.Minute)) &&
		secondStart.Before(firstStart.Add(time.Duration(firstDuration)*time.Minute))
}

func (s *Store) lessonReadable(actor *account, stored *lesson) bool {
	if actor == nil || stored == nil || actor.TenantID != stored.TenantID {
		return false
	}
	if actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "" {
		return true
	}
	if actor.Roles[core.RoleTeacher] != "" && stored.TeacherAccountID == actor.ID {
		return true
	}
	studentID := actor.Roles[core.RoleStudent]
	return studentID != "" && containsString(stored.StudentIDs, studentID)
}

func (s *Store) lessonView(stored *lesson) core.Lesson {
	teacherRecord := s.accounts[stored.TeacherAccountID]
	teacherName := ""
	if teacherRecord != nil {
		teacherName = teacherRecord.FullName
	}
	students := make([]core.LessonStudent, 0, len(stored.StudentIDs))
	for _, studentID := range stored.StudentIDs {
		studentRecord := s.students[studentID]
		if studentRecord == nil {
			continue
		}
		students = append(students, core.LessonStudent{StudentID: studentRecord.ID, FullName: studentRecord.FullName})
	}
	sort.Slice(students, func(left, right int) bool {
		if students[left].FullName == students[right].FullName {
			return students[left].StudentID < students[right].StudentID
		}
		return students[left].FullName < students[right].FullName
	})
	return core.Lesson{
		ID: stored.ID, Title: stored.Title, StartsAt: stored.StartsAt,
		DurationMinutes: stored.DurationMinutes, Location: stored.Location,
		Teacher:  core.TeacherSummary{AccountID: stored.TeacherAccountID, FullName: teacherName},
		Students: students, Status: stored.Status, Version: stored.Version,
	}
}

func (s *Store) lessonViewForActor(actor *account, stored *lesson) core.Lesson {
	result := s.lessonView(stored)
	if actor == nil ||
		actor.Roles[core.RoleOwner] != "" ||
		actor.Roles[core.RoleAdministrator] != "" ||
		(actor.Roles[core.RoleTeacher] != "" && stored.TeacherAccountID == actor.ID) {
		return result
	}
	studentID := actor.Roles[core.RoleStudent]
	for _, student := range result.Students {
		if student.StudentID == studentID {
			result.Students = []core.LessonStudent{student}
			return result
		}
	}
	result.Students = []core.LessonStudent{}
	return result
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
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
	s.audit = append(s.audit, AuditRecord{ID: int64(len(s.audit) + 1), TenantID: tenantID, ActorID: actorID, DelegationID: delegationID, Action: action, TargetID: targetID, Decision: decision, Reason: reason, RecordedAt: at})
}

func (s *Store) appendAuditMetadata(tenantID, actorID, action, targetID, decision, reason string, at time.Time, metadata map[string]any) {
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	s.audit = append(s.audit, AuditRecord{
		ID:       int64(len(s.audit) + 1),
		TenantID: tenantID, ActorID: actorID, Action: action, TargetID: targetID,
		Decision: decision, Reason: reason, Metadata: cloned, RecordedAt: at,
	})
}

func (s *Store) appendOperatorAudit(tenantID, operatorID, action, targetID, decision, reason string, at time.Time) {
	s.audit = append(s.audit, AuditRecord{ID: int64(len(s.audit) + 1), TenantID: tenantID, OperatorID: operatorID, Action: action, TargetID: targetID, Decision: decision, Reason: reason, RecordedAt: at})
}

func (s *Store) appendOutbox(tenantID, eventType, aggregateID string, at time.Time) {
	s.appendOutboxPayload(tenantID, eventType, "", aggregateID, nil, at)
}

func (s *Store) appendOutboxPayload(tenantID, eventType, aggregateType, aggregateID string, payload map[string]any, at time.Time) {
	encoded, _ := json.Marshal(payload)
	s.outbox = append(s.outbox, OutboxRecord{
		ID: int64(len(s.outbox) + 1), TenantID: tenantID, EventType: eventType,
		AggregateType: aggregateType, AggregateID: aggregateID,
		Payload: encoded, RecordedAt: at, Status: "pending",
	})
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
