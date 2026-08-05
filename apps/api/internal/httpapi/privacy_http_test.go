package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestHTTPPrivacyAndDataRightsJourney walks ACC-10..18 over the wire:
// policy catalog and acceptance, optimistic privacy settings, re-auth-gated
// export and deletion requests, and the DEC-104-safe cancel window.
func TestHTTPPrivacyAndDataRightsJourney(t *testing.T) {
	fixture := newHTTPFixture(t)

	fixture.store.SeedPolicyVersionForTest("polv_http_privacy", fixture.owner.TenantID,
		"privacy", "1.0", "Политика конфиденциальности", "policies/privacy/1.0",
		time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC))

	response := fixture.do(t, http.MethodGet, "/v1/policies", nil, "", "")
	assertHTTPError(t, response, http.StatusUnauthorized, core.CodeUnauthenticated)

	response = fixture.do(t, http.MethodGet, "/v1/policies", nil, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list policies status = %d", response.StatusCode)
	}
	var policies []core.PolicyVersion
	if err := json.NewDecoder(response.Body).Decode(&policies); err != nil {
		t.Fatalf("decode policies: %v", err)
	}
	response.Body.Close()
	if len(policies) != 1 || policies[0].AcceptedAt != nil {
		t.Fatalf("policy catalog = %#v", policies)
	}

	response = fixture.do(t, http.MethodPost, "/v1/me/policy-acceptances",
		map[string]any{"policyVersionId": "polv_http_privacy"}, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("accept policy status = %d", response.StatusCode)
	}
	response.Body.Close()

	response = fixture.do(t, http.MethodGet, "/v1/me/privacy", nil, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("privacy settings status = %d", response.StatusCode)
	}
	var settings core.PrivacySettings
	if err := json.NewDecoder(response.Body).Decode(&settings); err != nil {
		t.Fatalf("decode privacy settings: %v", err)
	}
	response.Body.Close()
	if settings.Version != 0 || settings.PushPreview != "hidden" {
		t.Fatalf("default privacy settings = %#v", settings)
	}

	settings.StaffMessagesAllowed = false
	response = fixture.do(t, http.MethodPut, "/v1/me/privacy", settings, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("update privacy settings status = %d", response.StatusCode)
	}
	var updated core.PrivacySettings
	if err := json.NewDecoder(response.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated settings: %v", err)
	}
	response.Body.Close()
	if updated.Version != 1 || updated.StaffMessagesAllowed {
		t.Fatalf("updated privacy settings = %#v", updated)
	}

	response = fixture.do(t, http.MethodPut, "/v1/me/privacy", settings, fixture.ownerAccess, "")
	assertHTTPError(t, response, http.StatusConflict, core.CodeConflict)

	response = fixture.do(t, http.MethodPost, "/v1/me/data-exports",
		map[string]any{"currentPassword": "Wrong-password-1!"}, fixture.ownerAccess, "")
	assertHTTPError(t, response, http.StatusUnauthorized, core.CodeUnauthenticated)

	response = fixture.do(t, http.MethodPost, "/v1/me/data-exports",
		map[string]any{"currentPassword": httpOwnerPassword}, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create data export status = %d", response.StatusCode)
	}
	var export core.DataExportRequest
	if err := json.NewDecoder(response.Body).Decode(&export); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	response.Body.Close()
	if export.Status != "requested" {
		t.Fatalf("created export = %#v", export)
	}

	response = fixture.do(t, http.MethodPost, "/v1/me/data-exports",
		map[string]any{"currentPassword": httpOwnerPassword}, fixture.ownerAccess, "")
	assertHTTPError(t, response, http.StatusConflict, core.CodeConflict)

	response = fixture.do(t, http.MethodGet, "/v1/me/data-exports", nil, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list exports status = %d", response.StatusCode)
	}
	var exports []core.DataExportRequest
	if err := json.NewDecoder(response.Body).Decode(&exports); err != nil {
		t.Fatalf("decode exports: %v", err)
	}
	response.Body.Close()
	if len(exports) != 1 || exports[0].ID != export.ID {
		t.Fatalf("export inventory = %#v", exports)
	}

	response = fixture.do(t, http.MethodGet, "/v1/me/deletion-request", nil, fixture.ownerAccess, "")
	assertHTTPError(t, response, http.StatusNotFound, core.CodeNotFound)

	response = fixture.do(t, http.MethodPost, "/v1/me/deletion-request",
		map[string]any{"currentPassword": httpOwnerPassword}, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create deletion request status = %d", response.StatusCode)
	}
	var deletion core.DeletionRequest
	if err := json.NewDecoder(response.Body).Decode(&deletion); err != nil {
		t.Fatalf("decode deletion request: %v", err)
	}
	response.Body.Close()
	if deletion.Status != "pending_review" {
		t.Fatalf("created deletion request = %#v (DEC-104: reviewable only)", deletion)
	}

	response = fixture.do(t, http.MethodPost, "/v1/me/deletion-request/cancel", nil, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("cancel deletion request status = %d", response.StatusCode)
	}
	var cancelled core.DeletionRequest
	if err := json.NewDecoder(response.Body).Decode(&cancelled); err != nil {
		t.Fatalf("decode cancelled request: %v", err)
	}
	response.Body.Close()
	if cancelled.Status != "cancelled" || cancelled.CancelledAt == nil {
		t.Fatalf("cancelled deletion request = %#v", cancelled)
	}

	response = fixture.do(t, http.MethodPost, "/v1/me/deletion-request/cancel", nil, fixture.ownerAccess, "")
	assertHTTPError(t, response, http.StatusConflict, core.CodeInvalidState)
}
