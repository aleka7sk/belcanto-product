package memory

import (
	"context"
	"encoding/hex"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// P.1 session security parity with the PostgreSQL store.

type passwordReset struct {
	ID           string
	TenantID     string
	AccountID    string
	Digest       []byte
	ExpiresAt    time.Time
	ConsumedAt   *time.Time
	SupersededAt *time.Time
	CreatedAt    time.Time
}

var securityAuditActions = map[string]bool{
	"SessionCreated":            true,
	"SessionRefreshed":          true,
	"SessionRevoked":            true,
	"RefreshTokenReuseDetected": true,
	"AccountActivated":          true,
	"PasswordResetRequested":    true,
	"PasswordResetCompleted":    true,
	"OtherSessionsRevoked":      true,
	"ContactChangeStarted":      true,
	"ContactVerified":           true,
	"TwofaEnrolled":             true,
	"TwofaDisabled":             true,
	"TwofaChallengeFailed":      true,
	"ProfileUpdated":            true,
	"PolicyAccepted":            true,
	"PrivacySettingsUpdated":    true,
	"DataExportRequested":       true,
	"DeletionRequested":         true,
	"DeletionRequestCancelled":  true,
}

func (s *Store) ListSessions(_ context.Context, principal core.Principal, now time.Time) ([]core.SessionDevice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	devices := []core.SessionDevice{}
	for _, stored := range s.sessions {
		if stored.TenantID != principal.TenantID || stored.AccountID != principal.AccountID {
			continue
		}
		if stored.Status != "active" || !stored.Material.RefreshExpiresAt.After(now) {
			continue
		}
		device := core.SessionDevice{
			SessionID:   stored.Material.SessionID,
			DeviceLabel: stored.Material.DeviceLabel,
			Platform:    stored.Material.Platform,
			CreatedAt:   stored.Material.CreatedAt,
			Current:     stored.Material.SessionID == principal.SessionID,
		}
		if stored.LastSeenAt != nil {
			lastSeen := *stored.LastSeenAt
			device.LastSeenAt = &lastSeen
		}
		devices = append(devices, device)
	}
	sort.Slice(devices, func(left, right int) bool {
		if !devices[left].CreatedAt.Equal(devices[right].CreatedAt) {
			return devices[left].CreatedAt.After(devices[right].CreatedAt)
		}
		return devices[left].SessionID < devices[right].SessionID
	})
	return devices, nil
}

func (s *Store) RevokeSessionByID(_ context.Context, command core.RevokeSessionByIDCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	target := s.sessions[command.SessionID]
	if target == nil || target.TenantID != principal.TenantID || target.AccountID != principal.AccountID {
		return core.E(core.CodeNotFound, "session was not found", nil)
	}
	if current := s.sessions[principal.SessionID]; current != nil &&
		current.Material.FamilyID == target.Material.FamilyID {
		return core.E(core.CodeInvalidState, "sign out to end the current session", nil)
	}
	s.revokeFamily(target.TenantID, target.AccountID, target.Material.FamilyID, command.Now)
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "SessionRevoked",
		"session_family", target.Material.FamilyID, "allow", "", command.Now,
		map[string]any{"sessionId": command.SessionID})
	return nil
}

func (s *Store) RevokeOtherSessions(_ context.Context, command core.RevokeOtherSessionsCommand) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	current := s.sessions[principal.SessionID]
	if current == nil || current.TenantID != principal.TenantID || current.AccountID != principal.AccountID {
		return 0, core.E(core.CodeUnauthenticated, "session is inactive", nil)
	}
	families := map[string]bool{}
	for _, stored := range s.sessions {
		if stored.TenantID != principal.TenantID || stored.AccountID != principal.AccountID {
			continue
		}
		if stored.Status != "active" || stored.Material.FamilyID == current.Material.FamilyID {
			continue
		}
		families[stored.Material.FamilyID] = true
	}
	for familyID := range families {
		s.revokeFamily(principal.TenantID, principal.AccountID, familyID, command.Now)
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "OtherSessionsRevoked",
		"account", principal.AccountID, "allow", "", command.Now,
		map[string]any{"revokedFamilies": len(families)})
	return len(families), nil
}

func (s *Store) ListSecurityEvents(_ context.Context, principal core.Principal, query core.SecurityEventsQuery) ([]core.SecurityEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	events := []core.SecurityEvent{}
	for index := len(s.audit) - 1; index >= 0; index-- {
		record := s.audit[index]
		if record.TenantID != principal.TenantID || record.ActorID != principal.AccountID {
			continue
		}
		if !securityAuditActions[record.Action] {
			continue
		}
		if query.BeforeID != 0 && record.ID >= query.BeforeID {
			continue
		}
		events = append(events, core.SecurityEvent{
			ID:         record.ID,
			Action:     record.Action,
			Decision:   record.Decision,
			ReasonCode: record.Reason,
			TargetType: record.TargetType,
			TargetID:   record.TargetID,
			RecordedAt: record.RecordedAt,
		})
		if len(events) == query.Limit {
			break
		}
	}
	return events, nil
}

func (s *Store) CreatePasswordReset(_ context.Context, command core.CreatePasswordResetCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	accountID := s.accountPhone[command.Phone]
	acct := s.accounts[accountID]
	if accountID == "" || acct == nil || acct.Status != "active" {
		return core.E(core.CodeNotFound, "account was not found", nil)
	}
	for _, open := range s.resets {
		if open.AccountID == accountID && open.ConsumedAt == nil && open.SupersededAt == nil {
			superseded := command.Now
			open.SupersededAt = &superseded
		}
	}
	reset := &passwordReset{
		ID:        command.ResetID,
		TenantID:  acct.TenantID,
		AccountID: accountID,
		Digest:    cloneBytes(command.TokenDigest),
		ExpiresAt: command.ExpiresAt,
		CreatedAt: command.Now,
	}
	s.resets[reset.ID] = reset
	s.resetDigest[hex.EncodeToString(reset.Digest)] = reset.ID
	s.appendSecurityAudit(acct.TenantID, accountID, "PasswordResetRequested",
		"password_reset", reset.ID, "allow", "", command.Now, nil)
	s.appendOutbox(acct.TenantID, "PasswordResetRequested", reset.ID, command.Now)
	return nil
}

func (s *Store) CompletePasswordReset(_ context.Context, command core.CompletePasswordResetCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	resetID := s.resetDigest[hex.EncodeToString(command.TokenDigest)]
	reset := s.resets[resetID]
	if reset == nil || !security.EqualDigest(reset.Digest, command.TokenDigest) {
		return core.E(core.CodeUnauthenticated, "recovery link is invalid or expired", nil)
	}
	if reset.ConsumedAt != nil || reset.SupersededAt != nil || !reset.ExpiresAt.After(command.Now) {
		s.appendSecurityAudit(reset.TenantID, reset.AccountID, "PasswordResetCompleted",
			"password_reset", reset.ID, "deny", "inactive_or_expired_reset_token", command.Now, nil)
		return core.E(core.CodeUnauthenticated, "recovery link is invalid or expired", nil)
	}
	acct := s.accounts[reset.AccountID]
	if acct == nil || acct.Status != "active" {
		return core.E(core.CodeUnauthenticated, "recovery link is invalid or expired", nil)
	}
	consumed := command.Now
	reset.ConsumedAt = &consumed
	acct.PasswordHash = command.PasswordHash
	families := map[string]bool{}
	for _, stored := range s.sessions {
		if stored.TenantID == reset.TenantID && stored.AccountID == reset.AccountID {
			families[stored.Material.FamilyID] = true
		}
	}
	for familyID := range families {
		s.revokeFamily(reset.TenantID, reset.AccountID, familyID, command.Now)
	}
	s.appendSecurityAudit(reset.TenantID, reset.AccountID, "PasswordResetCompleted",
		"password_reset", reset.ID, "allow", "", command.Now, nil)
	return nil
}

func (s *Store) appendSecurityAudit(tenantID, actorID, action, targetType, targetID, decision, reason string, at time.Time, metadata map[string]any) {
	s.audit = append(s.audit, AuditRecord{
		ID:         int64(len(s.audit) + 1),
		TargetType: targetType,
		TenantID:   tenantID,
		ActorID:    actorID,
		Action:     action,
		TargetID:   targetID,
		Decision:   decision,
		Reason:     reason,
		Metadata:   metadata,
		RecordedAt: at,
	})
}
