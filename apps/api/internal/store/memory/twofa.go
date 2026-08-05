package memory

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// P.1 contacts, activation progress and TOTP 2FA — parity with PostgreSQL.

type verifiedContact struct {
	ID         string
	TenantID   string
	AccountID  string
	Kind       core.ContactKind
	Value      string
	VerifiedAt time.Time
}

type contactVerification struct {
	ID                string
	TenantID          string
	AccountID         string
	Kind              core.ContactKind
	Value             string
	Purpose           string
	CodeDigest        []byte
	ExpiresAt         time.Time
	AttemptsRemaining int
	ConsumedAt        *time.Time
	SupersededAt      *time.Time
	CreatedAt         time.Time
}

type twofaSecret struct {
	Ciphertext  []byte
	ConfirmedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type recoveryCode struct {
	TenantID     string
	AccountID    string
	Digest       []byte
	UsedAt       *time.Time
	SupersededAt *time.Time
}

type twofaChallenge struct {
	ID                string
	TenantID          string
	AccountID         string
	DeviceLabel       string
	Platform          string
	Digest            []byte
	ExpiresAt         time.Time
	AttemptsRemaining int
	ConsumedAt        *time.Time
}

type activationProgress struct {
	InvitationID      string
	TenantID          string
	AccountID         string
	PasswordSetAt     *time.Time
	ContactKind       core.ContactKind
	ContactValue      string
	ContactVerifiedAt *time.Time
	TwofaEnrolledAt   *time.Time
	CompletedAt       *time.Time
	UpdatedAt         time.Time
}

func secretKey(tenantID, accountID string) string { return tenantID + "\x00" + accountID }

func (s *Store) ListVerifiedContacts(_ context.Context, principal core.Principal) ([]core.VerifiedContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []core.VerifiedContact{}
	for _, contact := range s.contacts {
		if contact.TenantID == principal.TenantID && contact.AccountID == principal.AccountID {
			result = append(result, core.VerifiedContact{
				ID: contact.ID, Kind: contact.Kind, Value: contact.Value, VerifiedAt: contact.VerifiedAt,
			})
		}
	}
	return result, nil
}

func (s *Store) StartContactChange(_ context.Context, command core.StartContactChangeCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	for _, open := range s.contactVerifs {
		if open.AccountID == principal.AccountID && open.Purpose == "contact_change" &&
			open.ConsumedAt == nil && open.SupersededAt == nil {
			superseded := command.Now
			open.SupersededAt = &superseded
		}
	}
	s.contactVerifs[command.VerificationID] = &contactVerification{
		ID: command.VerificationID, TenantID: principal.TenantID, AccountID: principal.AccountID,
		Kind: command.Kind, Value: command.Value, Purpose: "contact_change",
		CodeDigest: cloneBytes(command.CodeDigest), ExpiresAt: command.ExpiresAt,
		AttemptsRemaining: 5, CreatedAt: command.Now,
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "ContactChangeStarted",
		"contact_verification", command.VerificationID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "ContactVerificationRequested", command.VerificationID, command.Now)
	return nil
}

func (s *Store) ConfirmContactChange(_ context.Context, command core.ConfirmContactChangeCommand) (core.VerifiedContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	var open *contactVerification
	for _, candidate := range s.contactVerifs {
		if candidate.AccountID == principal.AccountID && candidate.Purpose == "contact_change" &&
			candidate.ConsumedAt == nil && candidate.SupersededAt == nil {
			open = candidate
		}
	}
	if open == nil || !open.ExpiresAt.After(command.Now) || open.AttemptsRemaining <= 0 {
		return core.VerifiedContact{}, core.E(core.CodeInvalidState, "no confirmation is in progress", nil)
	}
	if !security.EqualDigest(open.CodeDigest, command.CodeDigest) {
		open.AttemptsRemaining--
		if open.AttemptsRemaining == 0 {
			superseded := command.Now
			open.SupersededAt = &superseded
		}
		return core.VerifiedContact{}, core.E(core.CodeInvalidInput, "confirmation code is incorrect", nil)
	}
	consumed := command.Now
	open.ConsumedAt = &consumed
	return s.upsertVerifiedContact(principal.TenantID, principal.AccountID, open.Kind, open.Value, command.Now), nil
}

func (s *Store) upsertVerifiedContact(tenantID, accountID string, kind core.ContactKind, value string, now time.Time) core.VerifiedContact {
	for id, existing := range s.contacts {
		if existing.TenantID == tenantID && existing.AccountID == accountID && existing.Kind == kind {
			delete(s.contacts, id)
		}
	}
	contact := &verifiedContact{
		ID: "contact_" + hex.EncodeToString([]byte(accountID + string(kind)))[:16], TenantID: tenantID,
		AccountID: accountID, Kind: kind, Value: value, VerifiedAt: now,
	}
	s.contacts[contact.ID] = contact
	s.appendSecurityAudit(tenantID, accountID, "ContactVerified",
		"verified_contact", contact.ID, "allow", "", now, nil)
	return core.VerifiedContact{ID: contact.ID, Kind: kind, Value: value, VerifiedAt: now}
}

func (s *Store) TwofaStatus(_ context.Context, principal core.Principal) (core.TwofaStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status := core.TwofaStatus{}
	if secret := s.twofaSecrets[secretKey(principal.TenantID, principal.AccountID)]; secret != nil && secret.ConfirmedAt != nil {
		confirmedAt := *secret.ConfirmedAt
		status.Enabled = true
		status.ConfirmedAt = &confirmedAt
	}
	for _, code := range s.recoveryCodes {
		if code.TenantID == principal.TenantID && code.AccountID == principal.AccountID &&
			code.UsedAt == nil && code.SupersededAt == nil {
			status.RecoveryCodesRemaining++
		}
	}
	return status, nil
}

func (s *Store) TwofaSecret(_ context.Context, tenantID, accountID string) (core.TwofaSecretRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	secret := s.twofaSecrets[secretKey(tenantID, accountID)]
	if secret == nil {
		return core.TwofaSecretRecord{}, core.E(core.CodeNotFound, "two-factor authentication is not set up", nil)
	}
	return core.TwofaSecretRecord{Ciphertext: cloneBytes(secret.Ciphertext), Confirmed: secret.ConfirmedAt != nil}, nil
}

func (s *Store) StartTwofaEnrollment(_ context.Context, command core.StartTwofaEnrollmentCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := secretKey(command.Principal.TenantID, command.Principal.AccountID)
	if existing := s.twofaSecrets[key]; existing != nil && existing.ConfirmedAt != nil {
		return core.E(core.CodeConflict, "two-factor authentication is already enabled", nil)
	}
	s.twofaSecrets[key] = &twofaSecret{
		Ciphertext: cloneBytes(command.SecretCiphertext),
		CreatedAt:  command.Now, UpdatedAt: command.Now,
	}
	return nil
}

func (s *Store) ConfirmTwofaEnrollment(_ context.Context, command core.ConfirmTwofaEnrollmentCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	secret := s.twofaSecrets[secretKey(principal.TenantID, principal.AccountID)]
	if secret == nil {
		return core.E(core.CodeInvalidState, "two-factor enrollment has not started", nil)
	}
	if secret.ConfirmedAt != nil {
		return core.E(core.CodeConflict, "two-factor authentication is already enabled", nil)
	}
	confirmed := command.Now
	secret.ConfirmedAt = &confirmed
	secret.UpdatedAt = command.Now
	s.replaceRecoveryCodes(principal.TenantID, principal.AccountID, command.RecoveryDigests, command.Now)
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "TwofaEnrolled",
		"account", principal.AccountID, "allow", "", command.Now, nil)
	return nil
}

func (s *Store) replaceRecoveryCodes(tenantID, accountID string, digests [][]byte, now time.Time) {
	for _, code := range s.recoveryCodes {
		if code.TenantID == tenantID && code.AccountID == accountID &&
			code.UsedAt == nil && code.SupersededAt == nil {
			superseded := now
			code.SupersededAt = &superseded
		}
	}
	for _, digest := range digests {
		s.recoveryCodes = append(s.recoveryCodes, &recoveryCode{
			TenantID: tenantID, AccountID: accountID, Digest: cloneBytes(digest),
		})
	}
}

func (s *Store) DisableTwofa(_ context.Context, command core.DisableTwofaCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	key := secretKey(principal.TenantID, principal.AccountID)
	secret := s.twofaSecrets[key]
	if secret == nil || secret.ConfirmedAt == nil {
		return core.E(core.CodeInvalidState, "two-factor authentication is not enabled", nil)
	}
	delete(s.twofaSecrets, key)
	s.replaceRecoveryCodes(principal.TenantID, principal.AccountID, nil, command.Now)
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "TwofaDisabled",
		"account", principal.AccountID, "allow", "", command.Now, nil)
	return nil
}

func (s *Store) CreateTwofaChallenge(_ context.Context, command core.CreateTwofaChallengeCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge := &twofaChallenge{
		ID: command.ChallengeID, TenantID: command.TenantID, AccountID: command.AccountID,
		DeviceLabel: command.DeviceLabel, Platform: command.Platform,
		Digest: cloneBytes(command.TokenDigest), ExpiresAt: command.ExpiresAt,
		AttemptsRemaining: 5,
	}
	s.twofaChallenges[challenge.ID] = challenge
	s.challengeDigest[hex.EncodeToString(challenge.Digest)] = challenge.ID
	return nil
}

func (s *Store) openChallengeByDigest(digest []byte, now time.Time) *twofaChallenge {
	id := s.challengeDigest[hex.EncodeToString(digest)]
	challenge := s.twofaChallenges[id]
	if challenge == nil || challenge.ConsumedAt != nil || !challenge.ExpiresAt.After(now) || challenge.AttemptsRemaining <= 0 {
		return nil
	}
	return challenge
}

func (s *Store) TwofaChallengeByDigest(_ context.Context, digest []byte, now time.Time) (core.TwofaChallengeRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge := s.openChallengeByDigest(digest, now)
	if challenge == nil {
		return core.TwofaChallengeRecord{}, core.E(core.CodeUnauthenticated, "second-factor challenge is invalid or expired", nil)
	}
	return core.TwofaChallengeRecord{
		ID: challenge.ID, TenantID: challenge.TenantID, AccountID: challenge.AccountID,
		DeviceLabel: challenge.DeviceLabel, Platform: challenge.Platform,
		AttemptsRemaining: challenge.AttemptsRemaining,
	}, nil
}

func (s *Store) ConsumeTwofaChallenge(_ context.Context, digest []byte, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge := s.openChallengeByDigest(digest, now)
	if challenge == nil {
		return core.E(core.CodeUnauthenticated, "second-factor challenge is invalid or expired", nil)
	}
	consumed := now
	challenge.ConsumedAt = &consumed
	return nil
}

func (s *Store) FailTwofaChallenge(_ context.Context, digest []byte, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge := s.openChallengeByDigest(digest, now)
	if challenge == nil {
		return nil
	}
	challenge.AttemptsRemaining--
	if challenge.AttemptsRemaining <= 0 {
		consumed := now
		challenge.ConsumedAt = &consumed
		s.appendSecurityAudit(challenge.TenantID, challenge.AccountID, "TwofaChallengeFailed",
			"twofa_challenge", challenge.ID, "deny", "second_factor_attempts_exhausted", now, nil)
	}
	return nil
}

func (s *Store) TryConsumeRecoveryCode(_ context.Context, tenantID, accountID string, digest []byte, now time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, code := range s.recoveryCodes {
		if code.TenantID == tenantID && code.AccountID == accountID &&
			code.UsedAt == nil && code.SupersededAt == nil &&
			security.EqualDigest(code.Digest, digest) {
			used := now
			code.UsedAt = &used
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) activationInviteAndAccount(digest []byte, now time.Time) (*invitation, *account, error) {
	invite := s.findInvite(digest)
	if invite == nil {
		return nil, nil, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	acct := s.accounts[invite.AccountID]
	if invite.Status != "issued" || !invite.ExpiresAt.After(now) || acct == nil || acct.Status != "pending_activation" {
		return nil, nil, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	return invite, acct, nil
}

func (s *Store) progressFor(invite *invitation, acct *account, now time.Time) *activationProgress {
	progress := s.activationProg[invite.ID]
	if progress == nil {
		progress = &activationProgress{
			InvitationID: invite.ID, TenantID: invite.TenantID, AccountID: acct.ID, UpdatedAt: now,
		}
		s.activationProg[invite.ID] = progress
	}
	return progress
}

func (s *Store) ActivationProgress(_ context.Context, digest []byte, now time.Time) (core.ActivationProgressView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, acct, err := s.activationInviteAndAccount(digest, now)
	if err != nil {
		return core.ActivationProgressView{}, err
	}
	progress := s.progressFor(invite, acct, now)
	view := core.ActivationProgressView{
		InvitationID: invite.ID, Kind: invite.Kind, DisplayName: acct.FullName,
		ExpiresAt:       invite.ExpiresAt,
		PasswordSet:     progress.PasswordSetAt != nil,
		ContactVerified: progress.ContactVerifiedAt != nil,
		TwofaEnrolled:   progress.TwofaEnrolledAt != nil,
		Completed:       progress.CompletedAt != nil,
	}
	if progress.ContactKind != "" {
		view.ContactKind = progress.ContactKind
		view.ContactMasked = security.MaskContact(string(progress.ContactKind), progress.ContactValue)
	}
	return view, nil
}

func (s *Store) SetActivationPassword(_ context.Context, command core.SetActivationPasswordCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, acct, err := s.activationInviteAndAccount(command.TokenDigest, command.Now)
	if err != nil {
		return err
	}
	if acct.Phone != command.Phone {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	acct.PasswordHash = command.PasswordHash
	progress := s.progressFor(invite, acct, command.Now)
	setAt := command.Now
	progress.PasswordSetAt = &setAt
	progress.UpdatedAt = command.Now
	return nil
}

func (s *Store) StartActivationContact(_ context.Context, command core.StartActivationContactCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, acct, err := s.activationInviteAndAccount(command.TokenDigest, command.Now)
	if err != nil {
		return err
	}
	progress := s.progressFor(invite, acct, command.Now)
	if progress.PasswordSetAt == nil {
		return core.E(core.CodeInvalidState, "set the password before confirming a contact", nil)
	}
	for _, open := range s.contactVerifs {
		if open.AccountID == acct.ID && open.Purpose == "activation_contact" &&
			open.ConsumedAt == nil && open.SupersededAt == nil {
			superseded := command.Now
			open.SupersededAt = &superseded
		}
	}
	s.contactVerifs[command.VerificationID] = &contactVerification{
		ID: command.VerificationID, TenantID: invite.TenantID, AccountID: acct.ID,
		Kind: command.Kind, Value: command.Value, Purpose: "activation_contact",
		CodeDigest: cloneBytes(command.CodeDigest), ExpiresAt: command.ExpiresAt,
		AttemptsRemaining: 5, CreatedAt: command.Now,
	}
	progress.ContactKind = command.Kind
	progress.ContactValue = command.Value
	progress.ContactVerifiedAt = nil
	progress.UpdatedAt = command.Now
	s.appendOutbox(invite.TenantID, "ContactVerificationRequested", command.VerificationID, command.Now)
	return nil
}

func (s *Store) VerifyActivationContact(_ context.Context, command core.VerifyActivationContactCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, acct, err := s.activationInviteAndAccount(command.TokenDigest, command.Now)
	if err != nil {
		return err
	}
	progress := s.progressFor(invite, acct, command.Now)
	var open *contactVerification
	for _, candidate := range s.contactVerifs {
		if candidate.AccountID == acct.ID && candidate.Purpose == "activation_contact" &&
			candidate.ConsumedAt == nil && candidate.SupersededAt == nil {
			open = candidate
		}
	}
	if open == nil || !open.ExpiresAt.After(command.Now) || open.AttemptsRemaining <= 0 {
		return core.E(core.CodeInvalidState, "no confirmation is in progress", nil)
	}
	if !security.EqualDigest(open.CodeDigest, command.CodeDigest) {
		open.AttemptsRemaining--
		if open.AttemptsRemaining == 0 {
			superseded := command.Now
			open.SupersededAt = &superseded
		}
		return core.E(core.CodeInvalidInput, "confirmation code is incorrect", nil)
	}
	consumed := command.Now
	open.ConsumedAt = &consumed
	progress.ContactVerifiedAt = &consumed
	progress.UpdatedAt = command.Now
	return nil
}

func (s *Store) SetActivationTwofa(_ context.Context, command core.SetActivationTwofaCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, acct, err := s.activationInviteAndAccount(command.TokenDigest, command.Now)
	if err != nil {
		return err
	}
	progress := s.progressFor(invite, acct, command.Now)
	if progress.ContactVerifiedAt == nil {
		return core.E(core.CodeInvalidState, "confirm the contact before enrolling a second factor", nil)
	}
	s.twofaSecrets[secretKey(invite.TenantID, acct.ID)] = &twofaSecret{
		Ciphertext: cloneBytes(command.SecretCiphertext),
		CreatedAt:  command.Now, UpdatedAt: command.Now,
	}
	return nil
}

func (s *Store) ActivationTwofaSecret(_ context.Context, digest []byte, now time.Time) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, acct, err := s.activationInviteAndAccount(digest, now)
	if err != nil {
		return nil, err
	}
	secret := s.twofaSecrets[secretKey(invite.TenantID, acct.ID)]
	if secret == nil || secret.ConfirmedAt != nil {
		return nil, core.E(core.CodeInvalidState, "two-factor enrollment has not started", nil)
	}
	return cloneBytes(secret.Ciphertext), nil
}

func (s *Store) ConfirmActivationTwofa(_ context.Context, command core.ConfirmActivationTwofaCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, acct, err := s.activationInviteAndAccount(command.TokenDigest, command.Now)
	if err != nil {
		return err
	}
	progress := s.progressFor(invite, acct, command.Now)
	secret := s.twofaSecrets[secretKey(invite.TenantID, acct.ID)]
	if secret == nil || secret.ConfirmedAt != nil {
		return core.E(core.CodeInvalidState, "two-factor enrollment has not started", nil)
	}
	confirmed := command.Now
	secret.ConfirmedAt = &confirmed
	secret.UpdatedAt = command.Now
	s.replaceRecoveryCodes(invite.TenantID, acct.ID, command.RecoveryDigests, command.Now)
	progress.TwofaEnrolledAt = &confirmed
	progress.UpdatedAt = command.Now
	return nil
}

func (s *Store) FinishActivation(_ context.Context, command core.FinishActivationCommand) error {
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
	if invite.Status != "issued" || !invite.ExpiresAt.After(command.Now) || acct == nil ||
		acct.Status != "pending_activation" || acct.Phone != command.Phone {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	progress := s.progressFor(invite, acct, command.Now)
	if progress.PasswordSetAt == nil || progress.ContactVerifiedAt == nil {
		return core.E(core.CodeInvalidState, "complete the required activation steps first", nil)
	}
	acct.Status = "active"
	invite.Status = "consumed"
	invite.ConsumedIdempotencyKey = command.IdempotencyKey
	invite.ConsumedFingerprint = cloneBytes(command.PayloadFingerprint)
	completed := command.Now
	progress.CompletedAt = &completed
	progress.UpdatedAt = command.Now
	s.upsertVerifiedContact(invite.TenantID, acct.ID, progress.ContactKind, progress.ContactValue, command.Now)
	s.appendAudit(invite.TenantID, acct.ID, "", "AccountActivated", acct.ID, "allow", "", command.Now)
	s.appendOutbox(invite.TenantID, "AccountActivated", acct.ID, command.Now)
	return nil
}
