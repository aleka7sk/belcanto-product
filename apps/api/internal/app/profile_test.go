package app_test

import (
	"context"
	"testing"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

func TestProfileReadAndRename(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	view, err := fixture.service.MyProfile(ctx, owner)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if view.FullName != "Belcanto Owner" || view.TenantName != "Belcanto" {
		t.Fatalf("profile identity = %#v", view)
	}
	if view.Phone != "+77000000001" {
		t.Fatalf("profile phone = %q", view.Phone)
	}
	if len(view.Roles) == 0 || view.Roles[0] != core.RoleOwner {
		t.Fatalf("profile roles = %#v", view.Roles)
	}

	if _, err := fixture.service.UpdateMyProfile(ctx, owner, "   "); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("blank name update = %v, want INVALID_INPUT", err)
	}

	updated, err := fixture.service.UpdateMyProfile(ctx, owner, "  Belcanto Owner II  ")
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.FullName != "Belcanto Owner II" {
		t.Fatalf("updated name = %q, want trimmed rename", updated.FullName)
	}

	reread, err := fixture.service.MyProfile(ctx, owner)
	if err != nil || reread.FullName != "Belcanto Owner II" {
		t.Fatalf("reread profile = %#v, %v", reread, err)
	}

	page, err := fixture.service.ListSecurityEvents(ctx, owner, "", 50)
	if err != nil {
		t.Fatalf("list security events: %v", err)
	}
	found := false
	for _, event := range page.Events {
		if event.Action == "ProfileUpdated" {
			found = true
		}
	}
	if !found {
		t.Fatal("ProfileUpdated missing from the security feed")
	}
}
