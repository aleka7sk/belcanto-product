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

// TestPostgreSQLPrivacyAndDataRights proves the P.1 privacy slice against
// the real schema: policy acceptance idempotency under the unique
// constraint, optimistic privacy-settings concurrency, the single-open-
// export rule, and the DEC-104-safe deletion request lifecycle (reviewable
// request → cancel → new request; never a scheduled erasure).
func TestPostgreSQLPrivacyAndDataRights(t *testing.T) {
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
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x79}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgpriv", TenantName: "Belcanto PG Privacy",
		FullName: "PG Privacy Owner", Phone: "+77004000001",
		Operator: "pg-privacy-operator", Reason: "privacy integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	const ownerPassword = "Pg-privacy-password-123!"
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77004000001",
		Password: ownerPassword, IdempotencyKey: "pgpriv-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	outcome, err := service.SignIn(ctx, "+77004000001", ownerPassword, core.SessionClientInfo{})
	if err != nil || outcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, outcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO policy_versions (id, tenant_id, kind, version, title, body_ref, effective_from, created_at)
		VALUES ('pgpol_privacy_v1', 'tenant_pgpriv', 'privacy', '1.0',
		        'Политика конфиденциальности', 'policies/privacy/1.0', now(), now())
	`); err != nil {
		t.Fatalf("seed policy version: %v", err)
	}

	policies, err := service.ListPolicies(ctx, owner)
	if err != nil || len(policies) != 1 || policies[0].AcceptedAt != nil {
		t.Fatalf("policy catalog = %#v, %v", policies, err)
	}
	if err := service.AcceptPolicy(ctx, owner, "pgpol_privacy_v1"); err != nil {
		t.Fatalf("accept policy: %v", err)
	}
	if err := service.AcceptPolicy(ctx, owner, "pgpol_privacy_v1"); err != nil {
		t.Fatalf("repeated acceptance must be idempotent: %v", err)
	}
	if err := service.AcceptPolicy(ctx, owner, "pgpol_missing"); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("accept unknown policy = %v, want NOT_FOUND", err)
	}
	policies, err = service.ListPolicies(ctx, owner)
	if err != nil || len(policies) != 1 || policies[0].AcceptedAt == nil {
		t.Fatalf("policy catalog after acceptance = %#v, %v", policies, err)
	}

	settings, err := service.PrivacySettings(ctx, owner)
	if err != nil || settings.Version != 0 || settings.PushPreview != "hidden" {
		t.Fatalf("default privacy settings = %#v, %v", settings, err)
	}
	settings.AchievementsVisible = false
	settings.PushPreview = "full"
	updated, err := service.UpdatePrivacySettings(ctx, owner, settings)
	if err != nil || updated.Version != 1 || updated.AchievementsVisible {
		t.Fatalf("updated privacy settings = %#v, %v", updated, err)
	}
	if _, err := service.UpdatePrivacySettings(ctx, owner, settings); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale privacy update = %v, want CONFLICT", err)
	}
	reread, err := service.PrivacySettings(ctx, owner)
	if err != nil || reread != updated {
		t.Fatalf("persisted privacy settings = %#v, %v", reread, err)
	}

	export, err := service.CreateDataExport(ctx, owner, ownerPassword)
	if err != nil || export.Status != "requested" {
		t.Fatalf("create data export = %#v, %v", export, err)
	}
	if _, err := service.CreateDataExport(ctx, owner, ownerPassword); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("second open export = %v, want CONFLICT", err)
	}
	exports, err := service.ListDataExports(ctx, owner)
	if err != nil || len(exports) != 1 || exports[0].ID != export.ID {
		t.Fatalf("export inventory = %#v, %v", exports, err)
	}

	if _, err := service.DeletionRequest(ctx, owner); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("read absent deletion request = %v, want NOT_FOUND", err)
	}
	created, err := service.CreateDeletionRequest(ctx, owner, ownerPassword)
	if err != nil || created.Status != "pending_review" {
		t.Fatalf("create deletion request = %#v, %v", created, err)
	}
	if _, err := service.CreateDeletionRequest(ctx, owner, ownerPassword); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("second open deletion request = %v, want CONFLICT", err)
	}
	cancelled, err := service.CancelDeletionRequest(ctx, owner)
	if err != nil || cancelled.Status != "cancelled" || cancelled.CancelledAt == nil {
		t.Fatalf("cancel deletion request = %#v, %v", cancelled, err)
	}
	reopened, err := service.CreateDeletionRequest(ctx, owner, ownerPassword)
	if err != nil || reopened.ID == created.ID {
		t.Fatalf("re-create after cancel = %#v, %v", reopened, err)
	}

	// DEC-104 structural guard: the schema cannot express scheduled erasure.
	var hasScheduledFor bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'account_deletion_requests' AND column_name = 'scheduled_for'
		)
	`).Scan(&hasScheduledFor); err != nil {
		t.Fatalf("inspect deletion schema: %v", err)
	}
	if hasScheduledFor {
		t.Fatal("account_deletion_requests must not have scheduled_for while DEC-104 is open")
	}

	page, err := service.ListSecurityEvents(ctx, owner, "", 50)
	if err != nil {
		t.Fatalf("list security events: %v", err)
	}
	seen := map[string]bool{}
	for _, event := range page.Events {
		seen[event.Action] = true
	}
	for _, action := range []string{"PolicyAccepted", "PrivacySettingsUpdated", "DataExportRequested", "DeletionRequested", "DeletionRequestCancelled"} {
		if !seen[action] {
			t.Fatalf("security feed misses %s: %#v", action, seen)
		}
	}
}
