package migrations

import (
	"strings"
	"testing"
)

func TestInvitationSchemaStoresDigestOnly(t *testing.T) {
	lower := strings.ToLower(initialUp)
	for _, forbidden := range []string{"token_ciphertext", "raw_token", "encrypted_token", "token_key_version"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("migration contains recoverable invitation-token field %q", forbidden)
		}
	}
	if !strings.Contains(lower, "token_digest") {
		t.Fatal("migration does not persist invitation token digest")
	}
}

func TestTenantScopedIdentityAndReplacementInvariantsAreDeclared(t *testing.T) {
	compact := strings.Join(strings.Fields(strings.ToLower(initialUp)), " ")
	required := []string{
		"normalized_value text not null check (normalized_value ~ '^\\+[1-9][0-9]{7,14}$')",
		"(status = 'pending_activation' and activated_at is null)",
		"(role_type = 'student' and scope_type = 'student')",
		"create unique index if not exists role_grants_one_active_owner_per_tenant on role_grants (tenant_id) where role_type = 'owner' and status = 'active'",
		"(status = 'revoked' and revoked_at is not null and revoked_by_account_id is not null and revocation_reason is not null)",
		"check ((kind = 'student_activation') = (student_id is not null))",
		"token_digest bytea not null unique check (octet_length(token_digest) = 32)",
		"access_digest bytea not null unique check (octet_length(access_digest) = 32)",
		"refresh_digest bytea not null unique check (octet_length(refresh_digest) = 32)",
		"payload_fingerprint bytea not null check (octet_length(payload_fingerprint) = 32)",
		"check (expires_at > issued_at)",
		"check ((consumed_idempotency_key is null) = (consumed_payload_fingerprint is null))",
		"check ((status = 'consumed') = (consumed_at is not null))",
		"check (refresh_expires_at > access_expires_at)",
		"(status = 'processing' and response_json is null and completed_at is null)",
		"foreign key (tenant_id, student_id, account_id) references students(tenant_id, id, account_id)",
		"foreign key (tenant_id, granted_by) references accounts(tenant_id, id)",
		"foreign key (tenant_id, revoked_by_account_id) references accounts(tenant_id, id)",
		"foreign key (tenant_id, issued_by_account_id) references accounts(tenant_id, id)",
		"foreign key (tenant_id, account_id, student_id, kind, superseded_by_id) references activation_invitations(tenant_id, account_id, student_id, kind, id)",
		"foreign key (tenant_id, account_id, family_id, replaced_by_id) references sessions(tenant_id, account_id, family_id, id)",
	}
	for _, declaration := range required {
		if !strings.Contains(compact, declaration) {
			t.Fatalf("migration is missing invariant %q", declaration)
		}
	}
}
