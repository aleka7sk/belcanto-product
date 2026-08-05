-- Belcanto P.1 contacts and two-factor authentication (Figma Page 32:
-- AUTH-03/04/10, ACC-03/06; HOF-12). Digest-only storage for verification
-- codes, recovery codes and sign-in challenges; the TOTP secret is the only
-- reversibly encrypted material because RFC 6238 verification needs it.

CREATE TABLE IF NOT EXISTS verified_contacts (
    id               text PRIMARY KEY,
    tenant_id        text NOT NULL REFERENCES tenants(id),
    account_id       text NOT NULL,
    contact_kind     text NOT NULL CHECK (contact_kind IN ('email', 'phone')),
    normalized_value text NOT NULL CHECK (char_length(normalized_value) BETWEEN 3 AND 254),
    verified_at      timestamptz NOT NULL,
    UNIQUE (tenant_id, account_id, contact_kind),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS contact_verifications (
    id                 text PRIMARY KEY,
    tenant_id          text NOT NULL REFERENCES tenants(id),
    account_id         text NOT NULL,
    contact_kind       text NOT NULL CHECK (contact_kind IN ('email', 'phone')),
    normalized_value   text NOT NULL CHECK (char_length(normalized_value) BETWEEN 3 AND 254),
    purpose            text NOT NULL CHECK (purpose IN ('activation_contact', 'contact_change')),
    code_digest        bytea NOT NULL CHECK (octet_length(code_digest) = 32),
    expires_at         timestamptz NOT NULL,
    attempts_remaining integer NOT NULL CHECK (attempts_remaining BETWEEN 0 AND 10),
    consumed_at        timestamptz,
    superseded_at      timestamptz,
    created_at         timestamptz NOT NULL,
    CHECK (expires_at > created_at),
    CHECK (consumed_at IS NULL OR superseded_at IS NULL),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX IF NOT EXISTS contact_verifications_open_idx
    ON contact_verifications (tenant_id, account_id, purpose)
    WHERE consumed_at IS NULL AND superseded_at IS NULL;

CREATE TABLE IF NOT EXISTS twofa_totp_secrets (
    tenant_id        text NOT NULL REFERENCES tenants(id),
    account_id       text NOT NULL,
    secret_ciphertext bytea NOT NULL CHECK (octet_length(secret_ciphertext) BETWEEN 28 AND 128),
    confirmed_at     timestamptz,
    created_at       timestamptz NOT NULL,
    updated_at       timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, account_id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS twofa_recovery_codes (
    id           bigserial PRIMARY KEY,
    tenant_id    text NOT NULL REFERENCES tenants(id),
    account_id   text NOT NULL,
    code_digest  bytea NOT NULL UNIQUE CHECK (octet_length(code_digest) = 32),
    used_at      timestamptz,
    superseded_at timestamptz,
    created_at   timestamptz NOT NULL,
    CHECK (used_at IS NULL OR superseded_at IS NULL),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX IF NOT EXISTS twofa_recovery_codes_active_idx
    ON twofa_recovery_codes (tenant_id, account_id)
    WHERE used_at IS NULL AND superseded_at IS NULL;

CREATE TABLE IF NOT EXISTS twofa_challenges (
    id                 text PRIMARY KEY,
    tenant_id          text NOT NULL REFERENCES tenants(id),
    account_id         text NOT NULL,
    token_digest       bytea NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    device_label       text CHECK (device_label IS NULL OR char_length(device_label) BETWEEN 1 AND 120),
    platform           text CHECK (platform IS NULL OR platform IN ('ios', 'android', 'web')),
    expires_at         timestamptz NOT NULL,
    attempts_remaining integer NOT NULL CHECK (attempts_remaining BETWEEN 0 AND 10),
    consumed_at        timestamptz,
    created_at         timestamptz NOT NULL,
    CHECK (expires_at > created_at),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS activation_progress (
    invitation_id       text PRIMARY KEY REFERENCES activation_invitations(id),
    tenant_id           text NOT NULL REFERENCES tenants(id),
    account_id          text NOT NULL,
    password_set_at     timestamptz,
    contact_kind        text CHECK (contact_kind IS NULL OR contact_kind IN ('email', 'phone')),
    contact_value       text CHECK (contact_value IS NULL OR char_length(contact_value) BETWEEN 3 AND 254),
    contact_verified_at timestamptz,
    twofa_enrolled_at   timestamptz,
    completed_at        timestamptz,
    updated_at          timestamptz NOT NULL,
    CHECK (contact_verified_at IS NULL OR contact_kind IS NOT NULL),
    CHECK (completed_at IS NULL OR password_set_at IS NOT NULL),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);
