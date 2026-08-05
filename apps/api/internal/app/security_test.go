package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

func ownerSessionTokens(t *testing.T, fixture *fixture, client core.SessionClientInfo) core.SessionTokens {
	t.Helper()
	outcome, err := fixture.service.SignIn(context.Background(), "+77000000001", ownerPassword, client)
	if err != nil {
		t.Fatalf("sign in Owner session: %v", err)
	}
	if outcome.Tokens == nil {
		t.Fatal("Owner sign-in returned a second-factor challenge; tokens expected")
	}
	return *outcome.Tokens
}

func TestSessionInventoryAndTargetedRevocation(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	phoneTokens := ownerSessionTokens(t, fixture, core.SessionClientInfo{DeviceLabel: "iPhone 17", Platform: "ios"})
	laptopTokens := ownerSessionTokens(t, fixture, core.SessionClientInfo{DeviceLabel: "MacBook Pro", Platform: "web"})
	phonePrincipal, err := fixture.service.Authenticate(ctx, phoneTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate phone session: %v", err)
	}

	devices, err := fixture.service.ListSessions(ctx, phonePrincipal)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("session inventory size = %d, want 3 (fixture + phone + laptop)", len(devices))
	}
	currentCount := 0
	var laptopSessionID string
	for _, device := range devices {
		if device.Current {
			currentCount++
			if device.SessionID != phonePrincipal.SessionID {
				t.Fatalf("current session = %s, want %s", device.SessionID, phonePrincipal.SessionID)
			}
			if device.DeviceLabel != "iPhone 17" || device.Platform != "ios" {
				t.Fatalf("current session metadata = %#v", device)
			}
		}
		if device.DeviceLabel == "MacBook Pro" {
			laptopSessionID = device.SessionID
		}
	}
	if currentCount != 1 || laptopSessionID == "" {
		t.Fatalf("inventory markers: current=%d laptop=%q", currentCount, laptopSessionID)
	}

	if err := fixture.service.RevokeSessionByID(ctx, phonePrincipal, laptopSessionID, "wrong-password"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("revoke with wrong password = %v", err)
	}
	if err := fixture.service.RevokeSessionByID(ctx, phonePrincipal, phonePrincipal.SessionID, ownerPassword); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("revoke current session = %v, want INVALID_STATE", err)
	}
	if err := fixture.service.RevokeSessionByID(ctx, phonePrincipal, laptopSessionID, ownerPassword); err != nil {
		t.Fatalf("revoke laptop session: %v", err)
	}
	if _, err := fixture.service.Authenticate(ctx, laptopTokens.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("revoked laptop token authenticate = %v", err)
	}

	revoked, err := fixture.service.RevokeOtherSessions(ctx, phonePrincipal, ownerPassword)
	if err != nil {
		t.Fatalf("revoke other sessions: %v", err)
	}
	if revoked != 1 {
		t.Fatalf("revoked families = %d, want 1 (the fixture session)", revoked)
	}
	if _, err := fixture.service.Authenticate(ctx, phoneTokens.AccessToken); err != nil {
		t.Fatalf("current session must survive revoke-others: %v", err)
	}
	devices, err = fixture.service.ListSessions(ctx, phonePrincipal)
	if err != nil {
		t.Fatalf("list sessions after revocations: %v", err)
	}
	if len(devices) != 1 || !devices[0].Current {
		t.Fatalf("inventory after revocations = %#v", devices)
	}
}

func TestRefreshCarriesDeviceMetadataForward(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	tokens := ownerSessionTokens(t, fixture, core.SessionClientInfo{DeviceLabel: "iPhone 17", Platform: "ios"})
	rotated, err := fixture.service.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh session: %v", err)
	}
	principal, err := fixture.service.Authenticate(ctx, rotated.AccessToken)
	if err != nil {
		t.Fatalf("authenticate rotated session: %v", err)
	}
	devices, err := fixture.service.ListSessions(ctx, principal)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	var found bool
	for _, device := range devices {
		if device.Current {
			found = true
			if device.DeviceLabel != "iPhone 17" || device.Platform != "ios" {
				t.Fatalf("rotated session lost device metadata: %#v", device)
			}
			if device.LastSeenAt == nil {
				t.Fatal("rotated session must record lastSeenAt")
			}
		}
	}
	if !found {
		t.Fatalf("rotated session missing from inventory: %#v", devices)
	}
}

func TestSessionClientInfoValidation(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	if _, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{Platform: "windows"}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("unsupported platform = %v", err)
	}
	if _, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{DeviceLabel: strings.Repeat("д", 121)}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("oversized device label = %v", err)
	}
}

func TestSecurityEventsFeedIsActorScopedAndPaged(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	phonePrincipal := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)
	if _, err := fixture.service.RevokeOtherSessions(ctx, phonePrincipal, ownerPassword); err != nil {
		t.Fatalf("revoke other sessions: %v", err)
	}

	page, err := fixture.service.ListSecurityEvents(ctx, phonePrincipal, "", 2)
	if err != nil {
		t.Fatalf("list security events: %v", err)
	}
	if len(page.Events) != 2 || page.NextCursor == "" {
		t.Fatalf("first page = %#v", page)
	}
	if page.Events[0].ID <= page.Events[1].ID {
		t.Fatalf("events must be newest-first: %#v", page.Events)
	}
	second, err := fixture.service.ListSecurityEvents(ctx, phonePrincipal, page.NextCursor, 50)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	seen := map[string]bool{}
	for _, event := range append(page.Events, second.Events...) {
		seen[event.Action] = true
		if event.ID >= page.Events[0].ID && event != page.Events[0] {
			t.Fatalf("second page leaked newer event: %#v", event)
		}
	}
	if !seen["OtherSessionsRevoked"] || !seen["SessionCreated"] || !seen["AccountActivated"] {
		t.Fatalf("security feed actions = %#v", seen)
	}

	teacher := fixture.teacher
	teacherPage, err := fixture.service.ListSecurityEvents(ctx, teacher, "", 50)
	if err != nil {
		t.Fatalf("teacher security events: %v", err)
	}
	for _, event := range teacherPage.Events {
		if event.Action == "OtherSessionsRevoked" {
			t.Fatalf("teacher feed leaked Owner event: %#v", event)
		}
	}

	if _, err := fixture.service.ListSecurityEvents(ctx, phonePrincipal, "not-base64!", 10); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("invalid cursor = %v", err)
	}
	if _, err := fixture.service.ListSecurityEvents(ctx, phonePrincipal, "", 51); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("oversized limit = %v", err)
	}
}

func TestPasswordResetLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	outboxBefore := len(fixture.store.OutboxRecords())
	if err := fixture.service.RequestPasswordReset(ctx, "+77009999999"); err != nil {
		t.Fatalf("reset for unknown phone must succeed silently: %v", err)
	}
	if len(fixture.store.OutboxRecords()) != outboxBefore {
		t.Fatal("unknown phone must not enqueue delivery")
	}

	staleTokens := ownerSessionTokens(t, fixture, core.SessionClientInfo{})
	if err := fixture.service.RequestPasswordReset(ctx, "+7 700 000-00-01"); err != nil {
		t.Fatalf("request password reset: %v", err)
	}
	var resetID string
	for _, record := range fixture.store.OutboxRecords() {
		if record.EventType == "PasswordResetRequested" {
			resetID = record.AggregateID
		}
	}
	if resetID == "" {
		t.Fatal("password reset delivery event missing")
	}
	rawToken := fixture.codec.PasswordResetToken(resetID)

	if err := fixture.service.CompletePasswordReset(ctx, rawToken, "short"); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("weak password = %v", err)
	}
	const rotatedPassword = "Rotated-owner-password-456!"
	if err := fixture.service.CompletePasswordReset(ctx, rawToken, rotatedPassword); err != nil {
		t.Fatalf("complete password reset: %v", err)
	}
	if _, err := fixture.service.Authenticate(ctx, staleTokens.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("sessions must be revoked after reset: %v", err)
	}
	if _, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{}); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("old password after reset = %v", err)
	}
	if _, err := fixture.service.SignIn(ctx, "+77000000001", rotatedPassword, core.SessionClientInfo{}); err != nil {
		t.Fatalf("new password after reset: %v", err)
	}
	if err := fixture.service.CompletePasswordReset(ctx, rawToken, "Another-password-789!"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("consumed token replay = %v", err)
	}
}

func TestPasswordResetSupersessionAndExpiry(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	if err := fixture.service.RequestPasswordReset(ctx, "+77000000001"); err != nil {
		t.Fatalf("first reset request: %v", err)
	}
	firstID := lastResetID(t, fixture)
	if err := fixture.service.RequestPasswordReset(ctx, "+77000000001"); err != nil {
		t.Fatalf("second reset request: %v", err)
	}
	secondID := lastResetID(t, fixture)
	if firstID == secondID {
		t.Fatal("second request must mint a new reset")
	}
	if err := fixture.service.CompletePasswordReset(ctx, fixture.codec.PasswordResetToken(firstID), "Superseded-password-123!"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("superseded token = %v", err)
	}

	fixture.clock.Advance(31 * time.Minute)
	if err := fixture.service.CompletePasswordReset(ctx, fixture.codec.PasswordResetToken(secondID), "Expired-window-password-1!"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("expired token = %v", err)
	}
}

func lastResetID(t *testing.T, fixture *fixture) string {
	t.Helper()
	records := fixture.store.OutboxRecords()
	for index := len(records) - 1; index >= 0; index-- {
		if records[index].EventType == "PasswordResetRequested" {
			return records[index].AggregateID
		}
	}
	t.Fatal("no password reset delivery event recorded")
	return ""
}
