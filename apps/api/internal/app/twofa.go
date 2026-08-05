package app

import (
	"context"
	"strings"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// P.1 contacts, TOTP 2FA and the multi-step activation journey
// (Figma Page 32: AUTH-01..05/10, ACC-03/06; HOF-12).

const twofaIssuer = "Belcanto"
const recoveryCodeCount = 10

func (s *Service) secretBoxOrError() (*security.SecretBox, error) {
	if s.twofaBox == nil {
		return nil, core.E(core.CodeInternal, "two-factor secret store is unavailable", nil)
	}
	return s.twofaBox, nil
}

func normalizeContact(kind core.ContactKind, value string) (string, error) {
	switch kind {
	case core.ContactEmail:
		normalized, err := security.NormalizeEmail(value)
		if err != nil {
			return "", core.E(core.CodeInvalidInput, err.Error(), nil)
		}
		return normalized, nil
	case core.ContactPhone:
		normalized, err := security.NormalizePhone(value)
		if err != nil {
			return "", core.E(core.CodeInvalidInput, "phone must be in E.164 format", nil)
		}
		return normalized, nil
	default:
		return "", core.E(core.CodeInvalidInput, "contact kind must be email or phone", nil)
	}
}

func normalizeConfirmationCode(code string) (string, error) {
	normalized := strings.TrimSpace(code)
	if len(normalized) != 6 {
		return "", core.E(core.CodeInvalidInput, "confirmation code must contain 6 digits", nil)
	}
	for _, r := range normalized {
		if r < '0' || r > '9' {
			return "", core.E(core.CodeInvalidInput, "confirmation code must contain 6 digits", nil)
		}
	}
	return normalized, nil
}

// ---- authenticated contact management (ACC-03) ----

func (s *Service) ListVerifiedContacts(ctx context.Context, principal core.Principal) ([]core.VerifiedContact, error) {
	contacts, err := s.store.ListVerifiedContacts(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list verified contacts", err)
	}
	return contacts, nil
}

func (s *Service) StartContactChange(ctx context.Context, principal core.Principal, currentPassword string, kind core.ContactKind, value string) error {
	normalized, err := normalizeContact(kind, value)
	if err != nil {
		return err
	}
	if err := s.reauthenticate(ctx, principal, currentPassword); err != nil {
		return err
	}
	verificationID, err := security.NewID("verify")
	if err != nil {
		return core.E(core.CodeInternal, "could not start contact confirmation", err)
	}
	code := s.tokens.ContactVerificationCode(verificationID)
	now := s.clock.Now()
	if err := s.store.StartContactChange(ctx, core.StartContactChangeCommand{
		Principal:      principal,
		VerificationID: verificationID,
		Kind:           kind,
		Value:          normalized,
		CodeDigest:     s.tokens.Digest(code),
		ExpiresAt:      now.Add(s.contactCodeTTL),
		Now:            now,
	}); err != nil {
		return normalizeStoreError("start contact change", err)
	}
	return nil
}

func (s *Service) ConfirmContactChange(ctx context.Context, principal core.Principal, code string) (core.VerifiedContact, error) {
	normalized, err := normalizeConfirmationCode(code)
	if err != nil {
		return core.VerifiedContact{}, err
	}
	contact, err := s.store.ConfirmContactChange(ctx, core.ConfirmContactChangeCommand{
		Principal:  principal,
		CodeDigest: s.tokens.Digest(normalized),
		Now:        s.clock.Now(),
	})
	if err != nil {
		return core.VerifiedContact{}, normalizeStoreError("confirm contact change", err)
	}
	return contact, nil
}

// ---- authenticated 2FA management (ACC-06) ----

type TwofaEnrollment struct {
	Secret          string `json:"secret"`
	ProvisioningURI string `json:"provisioningUri"`
}

func (s *Service) TwofaStatus(ctx context.Context, principal core.Principal) (core.TwofaStatus, error) {
	status, err := s.store.TwofaStatus(ctx, principal)
	if err != nil {
		return core.TwofaStatus{}, normalizeStoreError("read twofa status", err)
	}
	return status, nil
}

func (s *Service) StartTwofaEnrollment(ctx context.Context, principal core.Principal, currentPassword string) (TwofaEnrollment, error) {
	box, err := s.secretBoxOrError()
	if err != nil {
		return TwofaEnrollment{}, err
	}
	if err := s.reauthenticate(ctx, principal, currentPassword); err != nil {
		return TwofaEnrollment{}, err
	}
	record, err := s.store.CredentialByAccount(ctx, principal.AccountID)
	if err != nil {
		return TwofaEnrollment{}, normalizeStoreError("read enrollment account", err)
	}
	secret, err := security.NewTOTPSecret()
	if err != nil {
		return TwofaEnrollment{}, core.E(core.CodeInternal, "could not create enrollment", err)
	}
	ciphertext, err := box.Seal([]byte(secret))
	if err != nil {
		return TwofaEnrollment{}, core.E(core.CodeInternal, "could not protect enrollment", err)
	}
	if err := s.store.StartTwofaEnrollment(ctx, core.StartTwofaEnrollmentCommand{
		Principal:        principal,
		SecretCiphertext: ciphertext,
		Now:              s.clock.Now(),
	}); err != nil {
		return TwofaEnrollment{}, normalizeStoreError("start twofa enrollment", err)
	}
	return TwofaEnrollment{
		Secret:          secret,
		ProvisioningURI: security.TOTPProvisioningURI(twofaIssuer, record.Phone, secret),
	}, nil
}

func (s *Service) verifyTotpAgainstSecret(ciphertext []byte, code string) (bool, error) {
	box, err := s.secretBoxOrError()
	if err != nil {
		return false, err
	}
	secret, err := box.Open(ciphertext)
	if err != nil {
		return false, core.E(core.CodeInternal, "stored second factor is invalid", err)
	}
	matched, err := security.VerifyTOTP(string(secret), code, s.clock.Now())
	if err != nil {
		return false, core.E(core.CodeInternal, "stored second factor is invalid", err)
	}
	return matched, nil
}

func (s *Service) ConfirmTwofaEnrollment(ctx context.Context, principal core.Principal, code string) ([]string, error) {
	record, err := s.store.TwofaSecret(ctx, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, normalizeStoreError("read twofa enrollment", err)
	}
	if record.Confirmed {
		return nil, core.E(core.CodeConflict, "two-factor authentication is already enabled", nil)
	}
	matched, err := s.verifyTotpAgainstSecret(record.Ciphertext, code)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, core.E(core.CodeInvalidInput, "authenticator code is incorrect", nil)
	}
	codes, err := security.NewRecoveryCodes(recoveryCodeCount)
	if err != nil {
		return nil, core.E(core.CodeInternal, "could not create recovery codes", err)
	}
	digests := make([][]byte, len(codes))
	for index, recovery := range codes {
		digests[index] = s.tokens.Digest(recovery)
	}
	if err := s.store.ConfirmTwofaEnrollment(ctx, core.ConfirmTwofaEnrollmentCommand{
		Principal:       principal,
		RecoveryDigests: digests,
		Now:             s.clock.Now(),
	}); err != nil {
		return nil, normalizeStoreError("confirm twofa enrollment", err)
	}
	return codes, nil
}

func (s *Service) DisableTwofa(ctx context.Context, principal core.Principal, currentPassword, code string) error {
	if err := s.reauthenticate(ctx, principal, currentPassword); err != nil {
		return err
	}
	record, err := s.store.TwofaSecret(ctx, principal.TenantID, principal.AccountID)
	if err != nil {
		return normalizeStoreError("read twofa enrollment", err)
	}
	if !record.Confirmed {
		return core.E(core.CodeInvalidState, "two-factor authentication is not enabled", nil)
	}
	matched, err := s.verifySecondFactor(ctx, principal.TenantID, principal.AccountID, record.Ciphertext, code)
	if err != nil {
		return err
	}
	if !matched {
		return core.E(core.CodeInvalidInput, "second-factor code is incorrect", nil)
	}
	if err := s.store.DisableTwofa(ctx, core.DisableTwofaCommand{
		Principal: principal,
		Now:       s.clock.Now(),
	}); err != nil {
		return normalizeStoreError("disable twofa", err)
	}
	return nil
}

// verifySecondFactor accepts an authenticator code or a one-time recovery
// code (consumed on success).
func (s *Service) verifySecondFactor(ctx context.Context, tenantID, accountID string, ciphertext []byte, code string) (bool, error) {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return false, core.E(core.CodeInvalidInput, "second-factor code is required", nil)
	}
	if len(trimmed) == 6 {
		return s.verifyTotpAgainstSecret(ciphertext, trimmed)
	}
	normalizedRecovery := strings.ToUpper(trimmed)
	consumed, err := s.store.TryConsumeRecoveryCode(ctx, tenantID, accountID,
		s.tokens.Digest(normalizedRecovery), s.clock.Now())
	if err != nil {
		return false, normalizeStoreError("consume recovery code", err)
	}
	return consumed, nil
}

// ---- second-factor sign-in (AUTH-06) ----

func (s *Service) SignInWithTwofa(ctx context.Context, challengeToken, code string) (core.SessionTokens, error) {
	if !validOpaqueToken(challengeToken) {
		return core.SessionTokens{}, core.E(core.CodeUnauthenticated, "second-factor challenge is invalid or expired", nil)
	}
	digest := s.tokens.Digest(challengeToken)
	challenge, err := s.store.TwofaChallengeByDigest(ctx, digest, s.clock.Now())
	if err != nil {
		return core.SessionTokens{}, normalizeStoreError("read twofa challenge", err)
	}
	record, err := s.store.TwofaSecret(ctx, challenge.TenantID, challenge.AccountID)
	if err != nil || !record.Confirmed {
		return core.SessionTokens{}, core.E(core.CodeUnauthenticated, "second-factor challenge is invalid or expired", nil)
	}
	matched, err := s.verifySecondFactor(ctx, challenge.TenantID, challenge.AccountID, record.Ciphertext, code)
	if err != nil {
		return core.SessionTokens{}, err
	}
	if !matched {
		if failErr := s.store.FailTwofaChallenge(ctx, digest, s.clock.Now()); failErr != nil {
			return core.SessionTokens{}, normalizeStoreError("record failed second factor", failErr)
		}
		return core.SessionTokens{}, core.E(core.CodeUnauthenticated, "second-factor code is incorrect", nil)
	}
	if err := s.store.ConsumeTwofaChallenge(ctx, digest, s.clock.Now()); err != nil {
		return core.SessionTokens{}, normalizeStoreError("consume twofa challenge", err)
	}
	return s.newSession(ctx, challenge.AccountID, challenge.TenantID, core.SessionClientInfo{
		DeviceLabel: challenge.DeviceLabel,
		Platform:    challenge.Platform,
	})
}

func (s *Service) issueTwofaChallenge(ctx context.Context, tenantID, accountID string, client core.SessionClientInfo) (core.SignInOutcome, error) {
	challengeID, err := security.NewID("twofa")
	if err != nil {
		return core.SignInOutcome{}, core.E(core.CodeInternal, "could not create second-factor challenge", err)
	}
	rawToken, err := s.tokens.NewRawToken()
	if err != nil {
		return core.SignInOutcome{}, core.E(core.CodeInternal, "could not create second-factor challenge", err)
	}
	now := s.clock.Now()
	expiresAt := now.Add(s.twofaChallengeTTL)
	if err := s.store.CreateTwofaChallenge(ctx, core.CreateTwofaChallengeCommand{
		ChallengeID: challengeID,
		TenantID:    tenantID,
		AccountID:   accountID,
		TokenDigest: s.tokens.Digest(rawToken),
		DeviceLabel: client.DeviceLabel,
		Platform:    client.Platform,
		ExpiresAt:   expiresAt,
		Now:         now,
	}); err != nil {
		return core.SignInOutcome{}, normalizeStoreError("create twofa challenge", err)
	}
	return core.SignInOutcome{TwofaChallenge: rawToken, TwofaExpiresAt: &expiresAt}, nil
}

// ---- multi-step activation (AUTH-01..05, AUTH-10) ----

type ActivationTwofaStart struct {
	Secret          string `json:"secret"`
	ProvisioningURI string `json:"provisioningUri"`
}

func (s *Service) activationDigest(rawToken string) ([]byte, error) {
	if !validOpaqueToken(rawToken) {
		return nil, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	return s.tokens.Digest(rawToken), nil
}

func (s *Service) ActivationProgress(ctx context.Context, rawToken string) (core.ActivationProgressView, error) {
	digest, err := s.activationDigest(rawToken)
	if err != nil {
		return core.ActivationProgressView{}, err
	}
	view, err := s.store.ActivationProgress(ctx, digest, s.clock.Now())
	if err != nil {
		return core.ActivationProgressView{}, normalizeActivationStepError(err)
	}
	return view, nil
}

func (s *Service) SetActivationPassword(ctx context.Context, rawToken, phone, password string) error {
	digest, err := s.activationDigest(rawToken)
	if err != nil {
		return err
	}
	normalizedPhone, err := security.NormalizePhone(phone)
	if err != nil {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	normalizedPassword, err := s.passwords.NormalizeAndValidate(password)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	passwordHash, err := s.passwords.Hash(normalizedPassword)
	if err != nil {
		return core.E(core.CodeInternal, "could not process the password", err)
	}
	if err := s.store.SetActivationPassword(ctx, core.SetActivationPasswordCommand{
		TokenDigest:  digest,
		Phone:        normalizedPhone,
		PasswordHash: passwordHash,
		Now:          s.clock.Now(),
	}); err != nil {
		return normalizeActivationStepError(err)
	}
	return nil
}

func (s *Service) StartActivationContact(ctx context.Context, rawToken string, kind core.ContactKind, value string) error {
	digest, err := s.activationDigest(rawToken)
	if err != nil {
		return err
	}
	normalized, err := normalizeContact(kind, value)
	if err != nil {
		return err
	}
	verificationID, err := security.NewID("verify")
	if err != nil {
		return core.E(core.CodeInternal, "could not start contact confirmation", err)
	}
	code := s.tokens.ContactVerificationCode(verificationID)
	now := s.clock.Now()
	if err := s.store.StartActivationContact(ctx, core.StartActivationContactCommand{
		TokenDigest:    digest,
		VerificationID: verificationID,
		Kind:           kind,
		Value:          normalized,
		CodeDigest:     s.tokens.Digest(code),
		ExpiresAt:      now.Add(s.contactCodeTTL),
		Now:            now,
	}); err != nil {
		return normalizeActivationStepError(err)
	}
	return nil
}

func (s *Service) VerifyActivationContact(ctx context.Context, rawToken, code string) error {
	digest, err := s.activationDigest(rawToken)
	if err != nil {
		return err
	}
	normalized, err := normalizeConfirmationCode(code)
	if err != nil {
		return err
	}
	if err := s.store.VerifyActivationContact(ctx, core.VerifyActivationContactCommand{
		TokenDigest: digest,
		CodeDigest:  s.tokens.Digest(normalized),
		Now:         s.clock.Now(),
	}); err != nil {
		return normalizeActivationStepError(err)
	}
	return nil
}

func (s *Service) StartActivationTwofa(ctx context.Context, rawToken string) (ActivationTwofaStart, error) {
	box, err := s.secretBoxOrError()
	if err != nil {
		return ActivationTwofaStart{}, err
	}
	digest, err := s.activationDigest(rawToken)
	if err != nil {
		return ActivationTwofaStart{}, err
	}
	preview, err := s.store.ActivationProgress(ctx, digest, s.clock.Now())
	if err != nil {
		return ActivationTwofaStart{}, normalizeActivationStepError(err)
	}
	secret, err := security.NewTOTPSecret()
	if err != nil {
		return ActivationTwofaStart{}, core.E(core.CodeInternal, "could not create enrollment", err)
	}
	ciphertext, err := box.Seal([]byte(secret))
	if err != nil {
		return ActivationTwofaStart{}, core.E(core.CodeInternal, "could not protect enrollment", err)
	}
	if err := s.store.SetActivationTwofa(ctx, core.SetActivationTwofaCommand{
		TokenDigest:      digest,
		SecretCiphertext: ciphertext,
		Now:              s.clock.Now(),
	}); err != nil {
		return ActivationTwofaStart{}, normalizeActivationStepError(err)
	}
	return ActivationTwofaStart{
		Secret:          secret,
		ProvisioningURI: security.TOTPProvisioningURI(twofaIssuer, preview.DisplayName, secret),
	}, nil
}

func (s *Service) ConfirmActivationTwofa(ctx context.Context, rawToken, code string) ([]string, error) {
	digest, err := s.activationDigest(rawToken)
	if err != nil {
		return nil, err
	}
	ciphertext, err := s.store.ActivationTwofaSecret(ctx, digest, s.clock.Now())
	if err != nil {
		return nil, normalizeActivationStepError(err)
	}
	matched, err := s.verifyTotpAgainstSecret(ciphertext, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, core.E(core.CodeInvalidInput, "authenticator code is incorrect", nil)
	}
	codes, err := security.NewRecoveryCodes(recoveryCodeCount)
	if err != nil {
		return nil, core.E(core.CodeInternal, "could not create recovery codes", err)
	}
	digests := make([][]byte, len(codes))
	for index, recovery := range codes {
		digests[index] = s.tokens.Digest(recovery)
	}
	if err := s.store.ConfirmActivationTwofa(ctx, core.ConfirmActivationTwofaCommand{
		TokenDigest:     digest,
		RecoveryDigests: digests,
		Now:             s.clock.Now(),
	}); err != nil {
		return nil, normalizeActivationStepError(err)
	}
	return codes, nil
}

func (s *Service) FinishActivation(ctx context.Context, rawToken, phone, idempotencyKey string) error {
	digest, err := s.activationDigest(rawToken)
	if err != nil {
		return err
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedPhone, err := security.NormalizePhone(phone)
	if err != nil {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		TokenDigest []byte
		Phone       string
	}{TokenDigest: digest, Phone: normalizedPhone})
	if err != nil {
		return core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	if err := s.store.FinishActivation(ctx, core.FinishActivationCommand{
		TokenDigest:        digest,
		Phone:              normalizedPhone,
		IdempotencyKey:     normalizedKey,
		PayloadFingerprint: fingerprint,
		Now:                s.clock.Now(),
	}); err != nil {
		return normalizeActivationStepError(err)
	}
	return nil
}

// normalizeActivationStepError keeps step-order and code errors typed while
// collapsing everything else to the neutral activation failure.
func normalizeActivationStepError(err error) error {
	if core.IsCode(err, core.CodeInvalidState) || core.IsCode(err, core.CodeInvalidInput) ||
		core.IsCode(err, core.CodeConflict) || core.IsCode(err, core.CodeInvalidActivation) {
		return err
	}
	return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
}
