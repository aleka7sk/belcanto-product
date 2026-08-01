package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type Service struct {
	store             Store
	tokens            *security.TokenCodec
	passwords         PasswordService
	clock             Clock
	activationBaseURL string
	accessTTL         time.Duration
	refreshTTL        time.Duration
	invitationTTL     time.Duration
}

type PasswordService interface {
	NormalizeAndValidate(string) (string, error)
	Hash(string) (string, error)
	VerifyCredential(string, string) (bool, error)
	DummyHash() string
}

type Options struct {
	ActivationBaseURL string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	InvitationTTL     time.Duration
	Clock             Clock
}

func NewService(store Store, tokens *security.TokenCodec, passwords PasswordService, options Options) *Service {
	clock := options.Clock
	if clock == nil {
		clock = realClock{}
	}
	return &Service{
		store:             store,
		tokens:            tokens,
		passwords:         passwords,
		clock:             clock,
		activationBaseURL: strings.TrimRight(options.ActivationBaseURL, "#/?"),
		accessTTL:         options.AccessTTL,
		refreshTTL:        options.RefreshTTL,
		invitationTTL:     options.InvitationTTL,
	}
}

type BootstrapOwnerInput struct {
	TenantID   string
	TenantName string
	FullName   string
	Phone      string
	Operator   string
	Reason     string
}

type BootstrapAccessResult struct {
	AccountID      string
	ActivationLink string
	ExpiresAt      time.Time
}

func (s *Service) Ready(ctx context.Context) error {
	if err := s.store.Ready(ctx); err != nil {
		return core.E(core.CodeUnavailable, "service is not ready", err)
	}
	return nil
}

func (s *Service) BootstrapOwner(ctx context.Context, input BootstrapOwnerInput) (string, time.Time, error) {
	result, err := s.BootstrapOwnerWithAccount(ctx, input)
	return result.ActivationLink, result.ExpiresAt, err
}

func (s *Service) BootstrapOwnerWithAccount(ctx context.Context, input BootstrapOwnerInput) (BootstrapAccessResult, error) {
	phone, err := security.NormalizePhone(input.Phone)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, "invalid owner phone", err)
	}
	input.TenantID, err = security.ValidateIdentifier("tenant id", input.TenantID, 128)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.TenantName, err = security.ValidateText("tenant name", input.TenantName, 1, 200)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.FullName, err = security.ValidateText("Owner full name", input.FullName, 1, 200)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Operator, err = security.ValidateText("bootstrap operator", input.Operator, 1, 200)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Reason, err = security.ValidateText("bootstrap reason", input.Reason, 1, 500)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}

	ids, err := newIDs("inv", "acct", "person", "membership", "role")
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInternal, "could not generate bootstrap identifiers", err)
	}
	rawToken := s.tokens.InvitationToken(ids[0])
	now := s.clock.Now()
	expiresAt := now.Add(s.invitationTTL)
	err = s.store.BootstrapOwner(ctx, core.BootstrapOwnerCommand{
		TenantID:     input.TenantID,
		TenantName:   input.TenantName,
		FullName:     input.FullName,
		Phone:        phone,
		InvitationID: ids[0],
		AccountID:    ids[1],
		PersonID:     ids[2],
		MembershipID: ids[3],
		RoleGrantID:  ids[4],
		Operator:     input.Operator,
		Reason:       input.Reason,
		TokenDigest:  s.tokens.Digest(rawToken),
		Now:          now,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return BootstrapAccessResult{}, normalizeStoreError("bootstrap Owner", err)
	}
	return BootstrapAccessResult{AccountID: ids[1], ActivationLink: s.activationLink(rawToken), ExpiresAt: expiresAt}, nil
}

type BootstrapStaffInput struct {
	TenantID       string
	OwnerAccountID string
	FullName       string
	Phone          string
	Role           core.Role
	Operator       string
	Reason         string
}

func (s *Service) BootstrapStaff(ctx context.Context, input BootstrapStaffInput) (string, time.Time, error) {
	result, err := s.BootstrapStaffWithAccount(ctx, input)
	return result.ActivationLink, result.ExpiresAt, err
}

func (s *Service) BootstrapStaffWithAccount(ctx context.Context, input BootstrapStaffInput) (BootstrapAccessResult, error) {
	phone, err := security.NormalizePhone(input.Phone)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, "invalid staff phone", err)
	}
	input.TenantID, err = security.ValidateIdentifier("tenant id", input.TenantID, 128)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.OwnerAccountID, err = security.ValidateIdentifier("Owner account id", input.OwnerAccountID, 128)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.FullName, err = security.ValidateText("staff full name", input.FullName, 1, 200)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.Role != core.RoleAdministrator && input.Role != core.RoleTeacher {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, "staff bootstrap role must be Administrator or Teacher", nil)
	}
	input.Operator, err = security.ValidateText("bootstrap operator", input.Operator, 1, 200)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Reason, err = security.ValidateText("bootstrap reason", input.Reason, 1, 500)
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	ids, err := newIDs("inv", "acct", "person", "membership", "role")
	if err != nil {
		return BootstrapAccessResult{}, core.E(core.CodeInternal, "could not generate staff identifiers", err)
	}
	rawToken := s.tokens.InvitationToken(ids[0])
	now := s.clock.Now()
	expiresAt := now.Add(s.invitationTTL)
	err = s.store.BootstrapStaff(ctx, core.BootstrapStaffCommand{
		TenantID: input.TenantID, OwnerAccountID: input.OwnerAccountID,
		FullName: input.FullName, Phone: phone, Role: input.Role,
		InvitationID: ids[0], AccountID: ids[1], PersonID: ids[2],
		MembershipID: ids[3], RoleGrantID: ids[4],
		Operator: input.Operator, Reason: input.Reason,
		TokenDigest: s.tokens.Digest(rawToken), Now: now, ExpiresAt: expiresAt,
	})
	if err != nil {
		return BootstrapAccessResult{}, normalizeStoreError("bootstrap staff", err)
	}
	return BootstrapAccessResult{AccountID: ids[1], ActivationLink: s.activationLink(rawToken), ExpiresAt: expiresAt}, nil
}

type ReissueBootstrapInvitationInput struct {
	TenantID  string
	AccountID string
	Operator  string
	Reason    string
}

func (s *Service) ReissueBootstrapInvitation(ctx context.Context, input ReissueBootstrapInvitationInput) (core.BootstrapInvitationResult, string, error) {
	var err error
	input.TenantID, err = security.ValidateIdentifier("tenant id", input.TenantID, 128)
	if err != nil {
		return core.BootstrapInvitationResult{}, "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.AccountID, err = security.ValidateIdentifier("account id", input.AccountID, 128)
	if err != nil {
		return core.BootstrapInvitationResult{}, "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Operator, err = security.ValidateText("recovery operator", input.Operator, 1, 200)
	if err != nil {
		return core.BootstrapInvitationResult{}, "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Reason, err = security.ValidateText("recovery reason", input.Reason, 1, 500)
	if err != nil {
		return core.BootstrapInvitationResult{}, "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	invitationID, err := security.NewID("inv")
	if err != nil {
		return core.BootstrapInvitationResult{}, "", core.E(core.CodeInternal, "could not generate recovery invitation", err)
	}
	rawToken := s.tokens.InvitationToken(invitationID)
	now := s.clock.Now()
	result, err := s.store.ReissueBootstrapInvitation(ctx, core.ReissueBootstrapInvitationCommand{
		TenantID: input.TenantID, AccountID: input.AccountID,
		InvitationID: invitationID, Operator: input.Operator, Reason: input.Reason,
		TokenDigest: s.tokens.Digest(rawToken), Now: now, ExpiresAt: now.Add(s.invitationTTL),
	})
	if err != nil {
		return core.BootstrapInvitationResult{}, "", normalizeStoreError("reissue bootstrap invitation", err)
	}
	return result, s.activationLink(rawToken), nil
}

func (s *Service) PreviewActivation(ctx context.Context, rawToken string) (core.ActivationPreview, error) {
	if !validOpaqueToken(rawToken) {
		return core.ActivationPreview{}, core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	preview, err := s.store.PreviewActivation(ctx, s.tokens.Digest(rawToken), s.clock.Now())
	if err != nil {
		return core.ActivationPreview{}, normalizeActivationPreviewError(err)
	}
	return preview, nil
}

type CompleteActivationInput struct {
	Token          string
	Phone          string
	Password       string
	IdempotencyKey string
}

func (s *Service) CompleteActivation(ctx context.Context, input CompleteActivationInput) error {
	if !validOpaqueToken(input.Token) {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	phone, err := security.NormalizePhone(input.Phone)
	if err != nil {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	normalizedPassword, err := s.passwords.NormalizeAndValidate(input.Password)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		TokenDigest   string `json:"tokenDigest"`
		Phone         string `json:"phone"`
		PasswordProof string `json:"passwordProof"`
	}{
		TokenDigest:   fmt.Sprintf("%x", s.tokens.Digest(input.Token)),
		Phone:         phone,
		PasswordProof: fmt.Sprintf("%x", s.tokens.Digest("activation-password-v1:"+normalizedPassword)),
	})
	if err != nil {
		return core.E(core.CodeInternal, "could not fingerprint activation", err)
	}
	digest := s.tokens.Digest(input.Token)
	replayed, err := s.store.ValidateActivation(ctx, core.ActivationValidationCommand{
		TokenDigest: digest, Phone: phone, IdempotencyKey: idempotencyKey,
		PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return normalizeActivationCompletionError(err)
	}
	if replayed {
		return nil
	}
	passwordHash, err := s.passwords.Hash(normalizedPassword)
	if err != nil {
		// Validation already succeeded above; any later hash failure is an
		// infrastructure failure (for example crypto/rand), never client input.
		return core.E(core.CodeInternal, "could not hash activation password", err)
	}
	err = s.store.CompleteActivation(ctx, core.ActivationCompleteCommand{
		TokenDigest:        digest,
		Phone:              phone,
		PasswordHash:       passwordHash,
		IdempotencyKey:     idempotencyKey,
		PayloadFingerprint: fingerprint,
		Now:                s.clock.Now(),
	})
	if err != nil {
		return normalizeActivationCompletionError(err)
	}
	return nil
}

func (s *Service) SignIn(ctx context.Context, phone, password string) (core.SessionTokens, error) {
	normalizedPhone, normalizeErr := security.NormalizePhone(phone)
	encoded := s.passwords.DummyHash()
	if normalizeErr != nil {
		_, _ = s.passwords.VerifyCredential(password, encoded)
		return core.SessionTokens{}, core.E(core.CodeUnauthenticated, "phone or password is incorrect", nil)
	}
	record, lookupErr := s.store.CredentialByPhone(ctx, normalizedPhone)
	if lookupErr == nil {
		encoded = record.PasswordHash
	}
	verified, verifyErr := s.passwords.VerifyCredential(password, encoded)
	if verifyErr != nil {
		return core.SessionTokens{}, core.E(core.CodeInternal, "stored sign-in credential is invalid", verifyErr)
	}
	if lookupErr != nil && !core.IsCode(lookupErr, core.CodeNotFound) {
		return core.SessionTokens{}, internalStoreError("read sign-in credential", lookupErr)
	}
	if lookupErr != nil || !verified || record.Status != "active" {
		return core.SessionTokens{}, core.E(core.CodeUnauthenticated, "phone or password is incorrect", nil)
	}
	return s.newSession(ctx, record.AccountID, record.TenantID)
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (core.SessionTokens, error) {
	if !validOpaqueToken(rawRefreshToken) {
		return core.SessionTokens{}, core.E(core.CodeUnauthenticated, "refresh token is invalid", nil)
	}
	material, rawAccess, rawRefresh, err := s.newSessionMaterial("")
	if err != nil {
		return core.SessionTokens{}, core.E(core.CodeInternal, "could not generate session", err)
	}
	accountID, tenantID, err := s.store.RotateSession(ctx, s.tokens.Digest(rawRefreshToken), material, s.clock.Now())
	if err != nil {
		if core.IsCode(err, core.CodeUnauthenticated) {
			return core.SessionTokens{}, core.E(core.CodeUnauthenticated, "refresh token is invalid", nil)
		}
		return core.SessionTokens{}, internalStoreError("rotate refresh session", err)
	}
	_ = accountID
	_ = tenantID
	return core.SessionTokens{
		AccessToken:      rawAccess,
		RefreshToken:     rawRefresh,
		AccessExpiresAt:  material.AccessExpiresAt,
		RefreshExpiresAt: material.RefreshExpiresAt,
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, rawAccessToken string) (core.Principal, error) {
	if !validOpaqueToken(rawAccessToken) {
		return core.Principal{}, core.E(core.CodeUnauthenticated, "access token is invalid", nil)
	}
	principal, err := s.store.PrincipalByAccessDigest(ctx, s.tokens.Digest(rawAccessToken), s.clock.Now())
	if err != nil {
		if core.IsCode(err, core.CodeUnauthenticated) {
			return core.Principal{}, core.E(core.CodeUnauthenticated, "access token is invalid", nil)
		}
		return core.Principal{}, internalStoreError("authenticate access session", err)
	}
	return principal, nil
}

func (s *Service) SignOut(ctx context.Context, rawAccessToken string) error {
	if !validOpaqueToken(rawAccessToken) {
		return nil
	}
	if err := s.store.RevokeSession(ctx, s.tokens.Digest(rawAccessToken), s.clock.Now()); err != nil {
		return normalizeStoreError("revoke session family", err)
	}
	return nil
}

type GrantDelegationInput struct {
	AdministratorAccountID string
	Reason                 string
	ExpiresAt              *time.Time
	CurrentPassword        string
	IdempotencyKey         string
}

func (s *Service) GrantDelegation(ctx context.Context, principal core.Principal, input GrantDelegationInput) (core.DelegationResult, error) {
	if input.CurrentPassword == "" {
		return core.DelegationResult{}, core.E(core.CodeInvalidInput, "currentPassword is required", nil)
	}
	if err := s.reauthenticateOwner(ctx, principal, input.CurrentPassword); err != nil {
		return core.DelegationResult{}, err
	}
	var err error
	input.AdministratorAccountID, err = security.ValidateIdentifier("Administrator account id", input.AdministratorAccountID, 128)
	if err != nil {
		return core.DelegationResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Reason, err = security.ValidateText("delegation reason", input.Reason, 1, 500)
	if err != nil {
		return core.DelegationResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.IdempotencyKey, err = security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.DelegationResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(s.clock.Now()) {
		return core.DelegationResult{}, core.E(core.CodeInvalidInput, "delegation expiry must be in the future", nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AdministratorID string     `json:"administratorId"`
		Bundle          string     `json:"bundle"`
		Reason          string     `json:"reason"`
		ExpiresAt       *time.Time `json:"expiresAt"`
	}{input.AdministratorAccountID, core.StudentOnboardingManagerV1, input.Reason, input.ExpiresAt})
	if err != nil {
		return core.DelegationResult{}, core.E(core.CodeInternal, "could not fingerprint delegation", err)
	}
	id, err := security.NewID("delegation")
	if err != nil {
		return core.DelegationResult{}, core.E(core.CodeInternal, "could not generate delegation id", err)
	}
	result, err := s.store.GrantDelegation(ctx, core.GrantDelegationCommand{
		ID:                 id,
		TenantID:           principal.TenantID,
		OwnerAccountID:     principal.AccountID,
		AdministratorID:    input.AdministratorAccountID,
		Bundle:             core.StudentOnboardingManagerV1,
		Reason:             input.Reason,
		ExpiresAt:          input.ExpiresAt,
		IdempotencyKey:     input.IdempotencyKey,
		PayloadFingerprint: fingerprint,
		Now:                s.clock.Now(),
	})
	if err != nil {
		return core.DelegationResult{}, normalizeStoreError("grant delegation", err)
	}
	return result, nil
}

type RevokeDelegationInput struct {
	DelegationID    string
	Reason          string
	CurrentPassword string
	IdempotencyKey  string
}

func (s *Service) RevokeDelegation(ctx context.Context, principal core.Principal, input RevokeDelegationInput) error {
	if input.CurrentPassword == "" {
		return core.E(core.CodeInvalidInput, "currentPassword is required", nil)
	}
	if err := s.reauthenticateOwner(ctx, principal, input.CurrentPassword); err != nil {
		return err
	}
	var err error
	input.DelegationID, err = security.ValidateIdentifier("delegation id", input.DelegationID, 128)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Reason, err = security.ValidateText("revocation reason", input.Reason, 1, 500)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.IdempotencyKey, err = security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		DelegationID string `json:"delegationId"`
		Reason       string `json:"reason"`
	}{input.DelegationID, input.Reason})
	if err != nil {
		return core.E(core.CodeInternal, "could not fingerprint delegation revocation", err)
	}
	err = s.store.RevokeDelegation(ctx, core.RevokeDelegationCommand{
		TenantID:           principal.TenantID,
		OwnerAccountID:     principal.AccountID,
		DelegationID:       input.DelegationID,
		Reason:             input.Reason,
		IdempotencyKey:     input.IdempotencyKey,
		PayloadFingerprint: fingerprint,
		Now:                s.clock.Now(),
	})
	return normalizeStoreError("revoke delegation", err)
}

type CreateStudentInput struct {
	FullName            string
	Phone               string
	EnrollmentReference string
	TeacherAccountID    string
	Locale              string
	Timezone            string
	AdultConfirmed      bool
	IdempotencyKey      string
}

func (s *Service) CreateStudent(ctx context.Context, principal core.Principal, input CreateStudentInput) (core.StudentResult, error) {
	phone, err := security.NormalizePhone(input.Phone)
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, "student phone is invalid", err)
	}
	if !input.AdultConfirmed {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, "adult confirmation is required", nil)
	}
	input.FullName, err = security.ValidateText("student full name", input.FullName, 1, 200)
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.EnrollmentReference, err = security.ValidateText("enrollment reference", input.EnrollmentReference, 1, 100)
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.TeacherAccountID, err = security.ValidateIdentifier("Teacher account id", input.TeacherAccountID, 128)
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.IdempotencyKey, err = security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.Locale == "" {
		input.Locale = "ru-KZ"
	}
	if input.Timezone == "" {
		input.Timezone = "Asia/Almaty"
	}
	input.Locale, err = security.ValidateLocale(input.Locale)
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.Timezone, err = security.ValidateTimezone(input.Timezone)
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		FullName            string `json:"fullName"`
		Phone               string `json:"phone"`
		EnrollmentReference string `json:"enrollmentReference"`
		TeacherAccountID    string `json:"teacherAccountId"`
		Locale              string `json:"locale"`
		Timezone            string `json:"timezone"`
		AdultConfirmed      bool   `json:"adultConfirmed"`
	}{input.FullName, phone, input.EnrollmentReference, input.TeacherAccountID, input.Locale, input.Timezone, input.AdultConfirmed})
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInternal, "could not fingerprint student creation", err)
	}
	ids, err := newIDs("person", "membership", "student", "acct", "role", "assignment")
	if err != nil {
		return core.StudentResult{}, core.E(core.CodeInternal, "could not generate student identifiers", err)
	}
	result, err := s.store.CreateStudent(ctx, core.CreateStudentCommand{
		TenantID:            principal.TenantID,
		ActorAccountID:      principal.AccountID,
		PersonID:            ids[0],
		MembershipID:        ids[1],
		StudentID:           ids[2],
		AccountID:           ids[3],
		RoleGrantID:         ids[4],
		TeacherAssignmentID: ids[5],
		FullName:            input.FullName,
		Phone:               phone,
		EnrollmentReference: input.EnrollmentReference,
		TeacherAccountID:    input.TeacherAccountID,
		Locale:              input.Locale,
		Timezone:            input.Timezone,
		AdultConfirmed:      input.AdultConfirmed,
		IdempotencyKey:      input.IdempotencyKey,
		PayloadFingerprint:  fingerprint,
		Now:                 s.clock.Now(),
	})
	if err != nil {
		return core.StudentResult{}, normalizeStoreError("create Student", err)
	}
	return result, nil
}

type PublishFirstMinuteInput struct {
	StudentID       string
	WhatWorked      string
	CurrentFocus    string
	NextStep        string
	ExpectedVersion int64
	IdempotencyKey  string
}

func (s *Service) PublishFirstMinute(ctx context.Context, principal core.Principal, input PublishFirstMinuteInput) (core.FirstMinute, error) {
	var err error
	input.StudentID, err = security.ValidateIdentifier("student id", input.StudentID, 128)
	if err != nil {
		return core.FirstMinute{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.WhatWorked, err = security.ValidateText("whatWorked", input.WhatWorked, 1, 500)
	if err != nil {
		return core.FirstMinute{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.CurrentFocus, err = security.ValidateText("currentFocus", input.CurrentFocus, 1, 500)
	if err != nil {
		return core.FirstMinute{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	input.NextStep, err = security.ValidateText("nextStep", input.NextStep, 1, 500)
	if err != nil {
		return core.FirstMinute{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.ExpectedVersion < 0 {
		return core.FirstMinute{}, core.E(core.CodeInvalidInput, "expectedVersion must be non-negative", nil)
	}
	input.IdempotencyKey, err = security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.FirstMinute{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(input)
	if err != nil {
		return core.FirstMinute{}, core.E(core.CodeInternal, "could not fingerprint first minute", err)
	}
	revisionID, err := security.NewID("first_minute")
	if err != nil {
		return core.FirstMinute{}, core.E(core.CodeInternal, "could not generate first-minute id", err)
	}
	result, err := s.store.PublishFirstMinute(ctx, core.PublishFirstMinuteCommand{
		TenantID:           principal.TenantID,
		ActorAccountID:     principal.AccountID,
		StudentID:          input.StudentID,
		RevisionID:         revisionID,
		WhatWorked:         input.WhatWorked,
		CurrentFocus:       input.CurrentFocus,
		NextStep:           input.NextStep,
		ExpectedVersion:    input.ExpectedVersion,
		IdempotencyKey:     input.IdempotencyKey,
		PayloadFingerprint: fingerprint,
		Now:                s.clock.Now(),
	})
	if err != nil {
		return core.FirstMinute{}, normalizeStoreError("publish First Belcanto Minute", err)
	}
	return result, nil
}

func (s *Service) IssueInvitation(ctx context.Context, principal core.Principal, studentID, idempotencyKey string, mode core.InvitationMode) (core.InvitationResult, string, error) {
	studentID, err := security.ValidateIdentifier("student id", studentID, 128)
	if err != nil {
		return core.InvitationResult{}, "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	idempotencyKey, err = security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.InvitationResult{}, "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if mode != core.InvitationIssue && mode != core.InvitationReissue {
		return core.InvitationResult{}, "", core.E(core.CodeInvalidInput, "invitation mode is invalid", nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		StudentID string              `json:"studentId"`
		Mode      core.InvitationMode `json:"mode"`
	}{studentID, mode})
	if err != nil {
		return core.InvitationResult{}, "", core.E(core.CodeInternal, "could not fingerprint invitation", err)
	}
	invitationID, err := security.NewID("inv")
	if err != nil {
		return core.InvitationResult{}, "", core.E(core.CodeInternal, "could not generate invitation id", err)
	}
	rawToken := s.tokens.InvitationToken(invitationID)
	now := s.clock.Now()
	result, err := s.store.IssueInvitation(ctx, core.IssueInvitationCommand{
		TenantID:           principal.TenantID,
		ActorAccountID:     principal.AccountID,
		StudentID:          studentID,
		InvitationID:       invitationID,
		TokenDigest:        s.tokens.Digest(rawToken),
		Mode:               mode,
		ExpiresAt:          now.Add(s.invitationTTL),
		IdempotencyKey:     idempotencyKey,
		PayloadFingerprint: fingerprint,
		Now:                now,
	})
	if err != nil {
		return core.InvitationResult{}, "", normalizeStoreError("issue activation invitation", err)
	}
	return result, s.activationLink(s.tokens.InvitationToken(result.InvitationID)), nil
}

func (s *Service) RevokeInvitation(ctx context.Context, principal core.Principal, invitationID, idempotencyKey string) error {
	invitationID, err := security.ValidateIdentifier("invitation id", invitationID, 128)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	idempotencyKey, err = security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		InvitationID string `json:"invitationId"`
	}{invitationID})
	if err != nil {
		return core.E(core.CodeInternal, "could not fingerprint invitation revocation", err)
	}
	err = s.store.RevokeInvitation(ctx, core.RevokeInvitationCommand{
		TenantID:           principal.TenantID,
		ActorAccountID:     principal.AccountID,
		InvitationID:       invitationID,
		IdempotencyKey:     idempotencyKey,
		PayloadFingerprint: fingerprint,
		Now:                s.clock.Now(),
	})
	return normalizeStoreError("revoke activation invitation", err)
}

func (s *Service) BootstrapView(ctx context.Context, principal core.Principal) (core.BootstrapView, error) {
	view, err := s.store.BootstrapView(ctx, principal, s.clock.Now())
	if err != nil {
		return core.BootstrapView{}, normalizeStoreError("read bootstrap view", err)
	}
	return view, nil
}

func (s *Service) ListStaff(ctx context.Context, principal core.Principal, role core.Role) ([]core.StaffMember, error) {
	if role != core.RoleAdministrator && role != core.RoleTeacher {
		return nil, core.E(core.CodeInvalidInput, "role must be Administrator or Teacher", nil)
	}
	staff, err := s.store.ListStaff(ctx, principal, role, s.clock.Now())
	if err != nil {
		return nil, normalizeStoreError("list staff", err)
	}
	return staff, nil
}

func (s *Service) ListStudentOnboarding(ctx context.Context, principal core.Principal) ([]core.StudentOnboardingItem, error) {
	items, err := s.store.ListStudentOnboarding(ctx, principal, s.clock.Now())
	if err != nil {
		return nil, normalizeStoreError("list Student onboarding", err)
	}
	return items, nil
}

func (s *Service) reauthenticateOwner(ctx context.Context, principal core.Principal, currentPassword string) error {
	if !principal.HasRole(core.RoleOwner) {
		return core.E(core.CodeForbidden, "owner authorization is required", nil)
	}
	record, err := s.store.CredentialByAccount(ctx, principal.AccountID)
	encoded := s.passwords.DummyHash()
	if err == nil {
		encoded = record.PasswordHash
	}
	verified, verifyErr := s.passwords.VerifyCredential(currentPassword, encoded)
	if verifyErr != nil {
		return core.E(core.CodeInternal, "stored Owner credential is invalid", verifyErr)
	}
	if err != nil && !core.IsCode(err, core.CodeNotFound) {
		return internalStoreError("read Owner credential", err)
	}
	if err != nil || record.TenantID != principal.TenantID || record.Status != "active" || !verified {
		return core.E(core.CodeUnauthenticated, "fresh owner authentication is required", nil)
	}
	return nil
}

func (s *Service) newSession(ctx context.Context, accountID, tenantID string) (core.SessionTokens, error) {
	material, rawAccess, rawRefresh, err := s.newSessionMaterial("")
	if err != nil {
		return core.SessionTokens{}, core.E(core.CodeInternal, "could not generate session", err)
	}
	if err := s.store.CreateSession(ctx, accountID, tenantID, material); err != nil {
		if core.IsCode(err, core.CodeUnauthenticated) {
			return core.SessionTokens{}, err
		}
		return core.SessionTokens{}, normalizeStoreError("create session", err)
	}
	return core.SessionTokens{
		AccessToken:      rawAccess,
		RefreshToken:     rawRefresh,
		AccessExpiresAt:  material.AccessExpiresAt,
		RefreshExpiresAt: material.RefreshExpiresAt,
	}, nil
}

func (s *Service) newSessionMaterial(familyID string) (core.SessionMaterial, string, string, error) {
	sessionID, err := security.NewID("session")
	if err != nil {
		return core.SessionMaterial{}, "", "", err
	}
	if familyID == "" {
		familyID, err = security.NewID("session_family")
		if err != nil {
			return core.SessionMaterial{}, "", "", err
		}
	}
	access, err := s.tokens.NewRawToken()
	if err != nil {
		return core.SessionMaterial{}, "", "", err
	}
	refresh, err := s.tokens.NewRawToken()
	if err != nil {
		return core.SessionMaterial{}, "", "", err
	}
	now := s.clock.Now()
	return core.SessionMaterial{
		SessionID:        sessionID,
		FamilyID:         familyID,
		AccessDigest:     s.tokens.Digest(access),
		RefreshDigest:    s.tokens.Digest(refresh),
		AccessExpiresAt:  now.Add(s.accessTTL),
		RefreshExpiresAt: now.Add(s.refreshTTL),
		CreatedAt:        now,
	}, access, refresh, nil
}

func (s *Service) activationLink(rawToken string) string {
	return s.activationBaseURL + "#token=" + url.QueryEscape(rawToken)
}

func validOpaqueToken(raw string) bool {
	return opaqueTokenPattern.MatchString(raw)
}

var opaqueTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func normalizeActivationPreviewError(err error) error {
	if core.IsCode(err, core.CodeInvalidActivation) {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	return internalStoreError("preview activation", err)
}

func normalizeActivationCompletionError(err error) error {
	if core.IsCode(err, core.CodeInvalidActivation) {
		return core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil)
	}
	if core.IsCode(err, core.CodeConflict) || core.IsCode(err, core.CodeInvalidInput) {
		return err
	}
	return internalStoreError("complete activation", err)
}

func internalStoreError(operation string, err error) error {
	if core.IsCode(err, core.CodeInternal) {
		return err
	}
	return core.E(core.CodeInternal, operation, err)
}

func normalizeStoreError(operation string, err error) error {
	if err == nil {
		return nil
	}
	var appError *core.AppError
	if errors.As(err, &appError) {
		return err
	}
	return internalStoreError(operation, err)
}

func newIDs(prefixes ...string) ([]string, error) {
	result := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		id, err := security.NewID(prefix)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}
