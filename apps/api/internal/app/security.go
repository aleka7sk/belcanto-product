package app

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// P.1 session security (Figma Page 32: ACC-05/08/09, AUTH-06..08, flow B).

const defaultSecurityEventsLimit = 20
const maxSecurityEventsLimit = 50

func validateSessionClientInfo(client core.SessionClientInfo) (core.SessionClientInfo, error) {
	normalized := core.SessionClientInfo{}
	if client.DeviceLabel != "" {
		label, err := security.ValidateText("deviceLabel", client.DeviceLabel, 1, 120)
		if err != nil {
			return core.SessionClientInfo{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
		normalized.DeviceLabel = label
	}
	switch client.Platform {
	case "", "ios", "android", "web":
		normalized.Platform = client.Platform
	default:
		return core.SessionClientInfo{}, core.E(core.CodeInvalidInput, "platform must be ios, android or web", nil)
	}
	return normalized, nil
}

// reauthenticate verifies the caller's current password before a sensitive
// session mutation (HOF-12: recent authentication for session revoke).
func (s *Service) reauthenticate(ctx context.Context, principal core.Principal, currentPassword string) error {
	if currentPassword == "" {
		return core.E(core.CodeInvalidInput, "currentPassword is required", nil)
	}
	record, err := s.store.CredentialByAccount(ctx, principal.AccountID)
	encoded := s.passwords.DummyHash()
	if err == nil {
		encoded = record.PasswordHash
	}
	verified, verifyErr := s.passwords.VerifyCredential(currentPassword, encoded)
	if verifyErr != nil {
		return core.E(core.CodeInternal, "stored credential is invalid", verifyErr)
	}
	if err != nil && !core.IsCode(err, core.CodeNotFound) {
		return internalStoreError("read credential", err)
	}
	if err != nil || record.TenantID != principal.TenantID || record.Status != "active" || !verified {
		return core.E(core.CodeUnauthenticated, "fresh authentication is required", nil)
	}
	return nil
}

func (s *Service) ListSessions(ctx context.Context, principal core.Principal) ([]core.SessionDevice, error) {
	devices, err := s.store.ListSessions(ctx, principal, s.clock.Now())
	if err != nil {
		return nil, normalizeStoreError("list sessions", err)
	}
	return devices, nil
}

func (s *Service) RevokeSessionByID(ctx context.Context, principal core.Principal, sessionID, currentPassword string) error {
	normalizedID, err := security.ValidateIdentifier("sessionId", sessionID, 128)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if err := s.reauthenticate(ctx, principal, currentPassword); err != nil {
		return err
	}
	if err := s.store.RevokeSessionByID(ctx, core.RevokeSessionByIDCommand{
		Principal: principal,
		SessionID: normalizedID,
		Now:       s.clock.Now(),
	}); err != nil {
		return normalizeStoreError("revoke session", err)
	}
	return nil
}

func (s *Service) RevokeOtherSessions(ctx context.Context, principal core.Principal, currentPassword string) (int, error) {
	if err := s.reauthenticate(ctx, principal, currentPassword); err != nil {
		return 0, err
	}
	revoked, err := s.store.RevokeOtherSessions(ctx, core.RevokeOtherSessionsCommand{
		Principal: principal,
		Now:       s.clock.Now(),
	})
	if err != nil {
		return 0, normalizeStoreError("revoke other sessions", err)
	}
	return revoked, nil
}

type SecurityEventsPage struct {
	Events     []core.SecurityEvent `json:"events"`
	NextCursor string               `json:"nextCursor,omitempty"`
}

func (s *Service) ListSecurityEvents(ctx context.Context, principal core.Principal, cursor string, limit int) (SecurityEventsPage, error) {
	beforeID, err := decodeSecurityEventsCursor(cursor)
	if err != nil {
		return SecurityEventsPage{}, err
	}
	if limit <= 0 {
		limit = defaultSecurityEventsLimit
	}
	if limit > maxSecurityEventsLimit {
		return SecurityEventsPage{}, core.E(core.CodeInvalidInput, "limit must be at most 50", nil)
	}
	events, err := s.store.ListSecurityEvents(ctx, principal, core.SecurityEventsQuery{
		BeforeID: beforeID,
		Limit:    limit,
	})
	if err != nil {
		return SecurityEventsPage{}, normalizeStoreError("list security events", err)
	}
	page := SecurityEventsPage{Events: events}
	if len(events) == limit {
		page.NextCursor = encodeSecurityEventsCursor(events[len(events)-1].ID)
	}
	return page, nil
}

func encodeSecurityEventsCursor(beforeID int64) string {
	return base64.RawURLEncoding.EncodeToString([]byte("v1:" + strconv.FormatInt(beforeID, 10)))
}

func decodeSecurityEventsCursor(cursor string) (int64, error) {
	if cursor == "" {
		return 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, core.E(core.CodeInvalidInput, "cursor is invalid", nil)
	}
	value, hasPrefix := strings.CutPrefix(string(decoded), "v1:")
	if !hasPrefix {
		return 0, core.E(core.CodeInvalidInput, "cursor is invalid", nil)
	}
	beforeID, parseErr := strconv.ParseInt(value, 10, 64)
	if parseErr != nil || beforeID <= 0 {
		return 0, core.E(core.CodeInvalidInput, "cursor is invalid", nil)
	}
	return beforeID, nil
}

// RequestPasswordReset never reveals whether the phone belongs to an
// account (AUTH-07): unknown subjects succeed silently, the token digest is
// the only stored material, and delivery happens through the outbox worker
// which re-derives the link from the reset identifier.
func (s *Service) RequestPasswordReset(ctx context.Context, phone string) error {
	normalizedPhone, err := security.NormalizePhone(phone)
	if err != nil {
		return core.E(core.CodeInvalidInput, "phone must be in E.164 format", nil)
	}
	resetID, err := security.NewID("reset")
	if err != nil {
		return core.E(core.CodeInternal, "could not create password reset", err)
	}
	rawToken := s.tokens.PasswordResetToken(resetID)
	now := s.clock.Now()
	err = s.store.CreatePasswordReset(ctx, core.CreatePasswordResetCommand{
		ResetID:     resetID,
		Phone:       normalizedPhone,
		TokenDigest: s.tokens.Digest(rawToken),
		ExpiresAt:   now.Add(s.passwordResetTTL),
		Now:         now,
	})
	if err != nil && !core.IsCode(err, core.CodeNotFound) {
		return normalizeStoreError("create password reset", err)
	}
	return nil
}

// CompletePasswordReset consumes the one-time recovery token, rotates the
// credential and revokes every session family (AUTH-08: the old password
// and existing sessions stop working together).
func (s *Service) CompletePasswordReset(ctx context.Context, rawToken, newPassword string) error {
	if !validOpaqueToken(rawToken) {
		return core.E(core.CodeUnauthenticated, "recovery link is invalid or expired", nil)
	}
	normalizedPassword, err := s.passwords.NormalizeAndValidate(newPassword)
	if err != nil {
		return core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	passwordHash, err := s.passwords.Hash(normalizedPassword)
	if err != nil {
		return core.E(core.CodeInternal, "could not process the new password", err)
	}
	if err := s.store.CompletePasswordReset(ctx, core.CompletePasswordResetCommand{
		TokenDigest:  s.tokens.Digest(rawToken),
		PasswordHash: passwordHash,
		Now:          s.clock.Now(),
	}); err != nil {
		return normalizeStoreError("complete password reset", err)
	}
	return nil
}
