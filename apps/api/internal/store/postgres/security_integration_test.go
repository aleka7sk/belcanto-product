package postgres_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

// TestPostgreSQLSessionSecurityLifecycle proves the P.1 slice against the
// real schema: device-labeled sessions, targeted family revocation,
// revoke-others, the actor-scoped security feed, and one-time password
// recovery that rotates the credential and kills every session family.
func TestPostgreSQLSessionSecurityLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, _ := isolatedPool(t, ctx, databaseURL)
	if err := migrations.Up(ctx, pool); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	store := postgres.New(pool)
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x77}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgsec", TenantName: "Belcanto PG Security",
		FullName: "PG Security Owner", Phone: "+77003000001",
		Operator: "pg-security-operator", Reason: "session security integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	const ownerPassword = "Pg-security-password-123!"
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77003000001",
		Password: ownerPassword, IdempotencyKey: "pgsec-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}

	phoneOutcome, err := service.SignIn(ctx, "+77003000001", ownerPassword,
		core.SessionClientInfo{DeviceLabel: "iPhone 17", Platform: "ios"})
	if err != nil || phoneOutcome.Tokens == nil {
		t.Fatalf("sign in phone: %v (tokens=%v)", err, phoneOutcome.Tokens != nil)
	}
	phoneTokens := *phoneOutcome.Tokens
	laptopOutcome, err := service.SignIn(ctx, "+77003000001", ownerPassword,
		core.SessionClientInfo{DeviceLabel: "MacBook Pro", Platform: "web"})
	if err != nil || laptopOutcome.Tokens == nil {
		t.Fatalf("sign in laptop: %v (tokens=%v)", err, laptopOutcome.Tokens != nil)
	}
	laptopTokens := *laptopOutcome.Tokens
	phonePrincipal, err := service.Authenticate(ctx, phoneTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate phone: %v", err)
	}

	rotated, err := service.Refresh(ctx, phoneTokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh phone session: %v", err)
	}
	phonePrincipal, err = service.Authenticate(ctx, rotated.AccessToken)
	if err != nil {
		t.Fatalf("authenticate rotated phone: %v", err)
	}

	devices, err := service.ListSessions(ctx, phonePrincipal)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("active devices = %d, want 2", len(devices))
	}
	var laptopSessionID string
	for _, device := range devices {
		if device.Current {
			if device.DeviceLabel != "iPhone 17" || device.Platform != "ios" || device.LastSeenAt == nil {
				t.Fatalf("rotation dropped device metadata: %#v", device)
			}
		} else {
			laptopSessionID = device.SessionID
			if device.DeviceLabel != "MacBook Pro" {
				t.Fatalf("laptop metadata = %#v", device)
			}
		}
	}

	if err := service.RevokeSessionByID(ctx, phonePrincipal, phonePrincipal.SessionID, ownerPassword); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("revoke current session = %v", err)
	}
	if err := service.RevokeSessionByID(ctx, phonePrincipal, laptopSessionID, ownerPassword); err != nil {
		t.Fatalf("revoke laptop session: %v", err)
	}
	if _, err := service.Authenticate(ctx, laptopTokens.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("revoked laptop token = %v", err)
	}

	extraOutcome, err := service.SignIn(ctx, "+77003000001", ownerPassword, core.SessionClientInfo{})
	if err != nil || extraOutcome.Tokens == nil {
		t.Fatalf("sign in extra session: %v", err)
	}
	extraTokens := *extraOutcome.Tokens
	revoked, err := service.RevokeOtherSessions(ctx, phonePrincipal, ownerPassword)
	if err != nil {
		t.Fatalf("revoke other sessions: %v", err)
	}
	if revoked != 1 {
		t.Fatalf("revoked families = %d, want 1", revoked)
	}
	if _, err := service.Authenticate(ctx, extraTokens.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("extra session survived revoke-others: %v", err)
	}

	page, err := service.ListSecurityEvents(ctx, phonePrincipal, "", 3)
	if err != nil {
		t.Fatalf("list security events: %v", err)
	}
	if len(page.Events) != 3 || page.NextCursor == "" {
		t.Fatalf("security page = %#v", page)
	}
	rest, err := service.ListSecurityEvents(ctx, phonePrincipal, page.NextCursor, 50)
	if err != nil {
		t.Fatalf("list security events tail: %v", err)
	}
	seen := map[string]bool{}
	for _, event := range append(page.Events, rest.Events...) {
		seen[event.Action] = true
	}
	for _, action := range []string{"AccountActivated", "SessionCreated", "SessionRefreshed", "SessionRevoked", "OtherSessionsRevoked"} {
		if !seen[action] {
			t.Fatalf("security feed missing %s: %#v", action, seen)
		}
	}

	if err := service.RequestPasswordReset(ctx, "+77003000001"); err != nil {
		t.Fatalf("request password reset: %v", err)
	}
	var resetID string
	if err := pool.QueryRow(ctx, `
		SELECT aggregate_id FROM outbox_events
		WHERE event_type = 'PasswordResetRequested'
		ORDER BY id DESC LIMIT 1
	`).Scan(&resetID); err != nil {
		t.Fatalf("read reset delivery event: %v", err)
	}
	var storedRawToken string
	if err := pool.QueryRow(ctx, `
		SELECT COALESCE((SELECT payload::text FROM outbox_events
		WHERE event_type = 'PasswordResetRequested' AND payload::text LIKE '%' || $1 || '%'
		ORDER BY id DESC LIMIT 1), '')
	`, codec.PasswordResetToken(resetID)).Scan(&storedRawToken); err != nil {
		t.Fatalf("scan raw-token leak probe: %v", err)
	}
	if storedRawToken != "" {
		t.Fatal("raw recovery token leaked into the outbox payload")
	}

	const rotatedPassword = "Pg-rotated-password-456!"
	if err := service.CompletePasswordReset(ctx, codec.PasswordResetToken(resetID), rotatedPassword); err != nil {
		t.Fatalf("complete password reset: %v", err)
	}
	if _, err := service.Authenticate(ctx, rotated.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("session survived password reset: %v", err)
	}
	if _, err := service.SignIn(ctx, "+77003000001", ownerPassword, core.SessionClientInfo{}); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("old password after reset = %v", err)
	}
	if _, err := service.SignIn(ctx, "+77003000001", rotatedPassword, core.SessionClientInfo{}); err != nil {
		t.Fatalf("new password after reset: %v", err)
	}
	if err := service.CompletePasswordReset(ctx, codec.PasswordResetToken(resetID), "Replay-password-789!"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("consumed reset token replay = %v", err)
	}
}
