package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

func TestPolicyCatalogAndAcceptance(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	effective := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	fixture.store.SeedPolicyVersionForTest("policy_privacy_v2", "tenant_belcanto",
		"privacy", "2.0", "Политика конфиденциальности", "policies/privacy/2.0", effective)
	fixture.store.SeedPolicyVersionForTest("policy_terms_v1", "tenant_belcanto",
		"terms", "1.0", "Пользовательское соглашение", "policies/terms/1.0", effective)

	policies, err := fixture.service.ListPolicies(ctx, owner)
	if err != nil {
		t.Fatalf("list policies: %v", err)
	}
	if len(policies) != 2 {
		t.Fatalf("policy catalog size = %d, want 2", len(policies))
	}
	for _, policy := range policies {
		if policy.AcceptedAt != nil {
			t.Fatalf("policy %s starts accepted", policy.ID)
		}
	}

	if err := fixture.service.AcceptPolicy(ctx, owner, "policy_privacy_v2"); err != nil {
		t.Fatalf("accept policy: %v", err)
	}
	if err := fixture.service.AcceptPolicy(ctx, owner, "policy_privacy_v2"); err != nil {
		t.Fatalf("repeated acceptance must be idempotent: %v", err)
	}
	if err := fixture.service.AcceptPolicy(ctx, owner, "policy_missing"); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("accept unknown policy = %v, want NOT_FOUND", err)
	}

	policies, err = fixture.service.ListPolicies(ctx, owner)
	if err != nil {
		t.Fatalf("list policies after acceptance: %v", err)
	}
	accepted := 0
	for _, policy := range policies {
		if policy.ID == "policy_privacy_v2" {
			if policy.AcceptedAt == nil {
				t.Fatal("accepted policy lost its acceptance mark")
			}
			accepted++
		} else if policy.AcceptedAt != nil {
			t.Fatalf("policy %s unexpectedly accepted", policy.ID)
		}
	}
	if accepted != 1 {
		t.Fatalf("accepted policies = %d, want 1", accepted)
	}
}

func TestPrivacySettingsOptimisticConcurrency(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	settings, err := fixture.service.PrivacySettings(ctx, owner)
	if err != nil {
		t.Fatalf("read default privacy settings: %v", err)
	}
	if !settings.CommunityProfileVisible || settings.PushPreview != "hidden" || settings.Version != 0 {
		t.Fatalf("default privacy settings = %#v", settings)
	}

	settings.CommunityProfileVisible = false
	settings.PushPreview = "title"
	updated, err := fixture.service.UpdatePrivacySettings(ctx, owner, settings)
	if err != nil {
		t.Fatalf("update privacy settings: %v", err)
	}
	if updated.Version != 1 || updated.CommunityProfileVisible || updated.PushPreview != "title" {
		t.Fatalf("updated privacy settings = %#v", updated)
	}

	stale := settings
	stale.MentionsAllowed = false
	if _, err := fixture.service.UpdatePrivacySettings(ctx, owner, stale); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale version update = %v, want CONFLICT", err)
	}

	invalid := updated
	invalid.PushPreview = "everything"
	if _, err := fixture.service.UpdatePrivacySettings(ctx, owner, invalid); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("invalid pushPreview = %v, want INVALID_INPUT", err)
	}

	reread, err := fixture.service.PrivacySettings(ctx, owner)
	if err != nil {
		t.Fatalf("reread privacy settings: %v", err)
	}
	if reread != updated {
		t.Fatalf("persisted settings %#v, want %#v", reread, updated)
	}
}

func TestDataExportLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	if _, err := fixture.service.CreateDataExport(ctx, owner, ""); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("export without password = %v, want INVALID_INPUT", err)
	}
	if _, err := fixture.service.CreateDataExport(ctx, owner, "wrong-password"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("export with wrong password = %v, want UNAUTHENTICATED", err)
	}

	export, err := fixture.service.CreateDataExport(ctx, owner, ownerPassword)
	if err != nil {
		t.Fatalf("create data export: %v", err)
	}
	if export.Status != "requested" || export.ID == "" {
		t.Fatalf("created export = %#v", export)
	}
	if aggregate := lastOutboxAggregate(t, fixture, "DataExportRequested"); aggregate != export.ID {
		t.Fatalf("outbox aggregate = %s, want %s", aggregate, export.ID)
	}

	if _, err := fixture.service.CreateDataExport(ctx, owner, ownerPassword); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("second open export = %v, want CONFLICT", err)
	}

	exports, err := fixture.service.ListDataExports(ctx, owner)
	if err != nil {
		t.Fatalf("list data exports: %v", err)
	}
	if len(exports) != 1 || exports[0].ID != export.ID {
		t.Fatalf("export inventory = %#v", exports)
	}
}

func TestDeletionRequestLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	if _, err := fixture.service.DeletionRequest(ctx, owner); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("read absent deletion request = %v, want NOT_FOUND", err)
	}
	if _, err := fixture.service.CancelDeletionRequest(ctx, owner); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("cancel absent deletion request = %v, want INVALID_STATE", err)
	}
	if _, err := fixture.service.CreateDeletionRequest(ctx, owner, "wrong-password"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("deletion with wrong password = %v, want UNAUTHENTICATED", err)
	}

	created, err := fixture.service.CreateDeletionRequest(ctx, owner, ownerPassword)
	if err != nil {
		t.Fatalf("create deletion request: %v", err)
	}
	if created.Status != "pending_review" || created.CancelledAt != nil {
		t.Fatalf("created deletion request = %#v (DEC-104: must stay reviewable)", created)
	}
	if aggregate := lastOutboxAggregate(t, fixture, "DeletionRequested"); aggregate != created.ID {
		t.Fatalf("outbox aggregate = %s, want %s", aggregate, created.ID)
	}

	if _, err := fixture.service.CreateDeletionRequest(ctx, owner, ownerPassword); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("second open deletion request = %v, want CONFLICT", err)
	}

	open, err := fixture.service.DeletionRequest(ctx, owner)
	if err != nil || open.ID != created.ID {
		t.Fatalf("read open deletion request = %#v, %v", open, err)
	}

	cancelled, err := fixture.service.CancelDeletionRequest(ctx, owner)
	if err != nil {
		t.Fatalf("cancel deletion request: %v", err)
	}
	if cancelled.Status != "cancelled" || cancelled.CancelledAt == nil {
		t.Fatalf("cancelled deletion request = %#v", cancelled)
	}
	if _, err := fixture.service.DeletionRequest(ctx, owner); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("read after cancel = %v, want NOT_FOUND", err)
	}

	reopened, err := fixture.service.CreateDeletionRequest(ctx, owner, ownerPassword)
	if err != nil {
		t.Fatalf("re-create after cancel: %v", err)
	}
	if reopened.ID == created.ID {
		t.Fatal("re-created deletion request must be a new row")
	}
}

func TestPrivacyActionsAppearInSecurityFeed(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	settings, err := fixture.service.PrivacySettings(ctx, owner)
	if err != nil {
		t.Fatalf("read privacy settings: %v", err)
	}
	if _, err := fixture.service.UpdatePrivacySettings(ctx, owner, settings); err != nil {
		t.Fatalf("update privacy settings: %v", err)
	}
	if _, err := fixture.service.CreateDataExport(ctx, owner, ownerPassword); err != nil {
		t.Fatalf("create data export: %v", err)
	}

	page, err := fixture.service.ListSecurityEvents(ctx, owner, "", 50)
	if err != nil {
		t.Fatalf("list security events: %v", err)
	}
	seen := map[string]bool{}
	for _, event := range page.Events {
		seen[event.Action] = true
	}
	if !seen["PrivacySettingsUpdated"] || !seen["DataExportRequested"] {
		t.Fatalf("security feed misses privacy actions: %#v", seen)
	}
}
