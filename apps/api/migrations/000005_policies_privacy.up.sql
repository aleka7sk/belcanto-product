-- Belcanto P.1 policies, privacy and data rights (Figma Page 32:
-- ACC-10..12, ACC-14..18; HOF-12). DEC-104 is open: deletion requests
-- never schedule automatic erasure and export rows promise no SLA —
-- statuses only, with an explicit cancel window.

CREATE TABLE IF NOT EXISTS policy_versions (
    id             text PRIMARY KEY,
    tenant_id      text NOT NULL REFERENCES tenants(id),
    kind           text NOT NULL CHECK (kind IN ('privacy', 'terms', 'community', 'media_consent')),
    version        text NOT NULL CHECK (char_length(version) BETWEEN 1 AND 50),
    title          text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    body_ref       text NOT NULL CHECK (char_length(body_ref) BETWEEN 1 AND 500),
    effective_from timestamptz NOT NULL,
    created_at     timestamptz NOT NULL,
    UNIQUE (tenant_id, kind, version),
    UNIQUE (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS policy_acceptances (
    id                text PRIMARY KEY,
    tenant_id         text NOT NULL REFERENCES tenants(id),
    account_id        text NOT NULL,
    policy_version_id text NOT NULL,
    accepted_at       timestamptz NOT NULL,
    UNIQUE (tenant_id, account_id, policy_version_id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, policy_version_id) REFERENCES policy_versions(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS privacy_settings (
    tenant_id                 text NOT NULL REFERENCES tenants(id),
    account_id                text NOT NULL,
    community_profile_visible boolean NOT NULL DEFAULT true,
    achievements_visible      boolean NOT NULL DEFAULT true,
    staff_messages_allowed    boolean NOT NULL DEFAULT true,
    mentions_allowed          boolean NOT NULL DEFAULT true,
    push_preview              text NOT NULL DEFAULT 'hidden'
        CHECK (push_preview IN ('hidden', 'title', 'full')),
    version                   bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    updated_at                timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, account_id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS data_export_requests (
    id           text PRIMARY KEY,
    tenant_id    text NOT NULL REFERENCES tenants(id),
    account_id   text NOT NULL,
    status       text NOT NULL CHECK (status IN ('requested', 'processing', 'ready', 'expired', 'cancelled')),
    requested_at timestamptz NOT NULL,
    ready_at     timestamptz,
    expires_at   timestamptz,
    archive_ref  text CHECK (archive_ref IS NULL OR char_length(archive_ref) BETWEEN 1 AND 500),
    CHECK (
        (status IN ('requested', 'processing', 'cancelled') AND ready_at IS NULL)
        OR (status IN ('ready', 'expired') AND ready_at IS NOT NULL)
    ),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX IF NOT EXISTS data_export_requests_open_idx
    ON data_export_requests (tenant_id, account_id)
    WHERE status IN ('requested', 'processing');

-- DEC-104 guard: no scheduled_for column exists on purpose. Deletion stays
-- an explicit request in a reviewable state until the legal retention and
-- deletion policy is decided; nothing here can trigger erasure.
CREATE TABLE IF NOT EXISTS account_deletion_requests (
    id           text PRIMARY KEY,
    tenant_id    text NOT NULL REFERENCES tenants(id),
    account_id   text NOT NULL,
    status       text NOT NULL CHECK (status IN ('requested', 'pending_review', 'cancelled')),
    requested_at timestamptz NOT NULL,
    cancelled_at timestamptz,
    CHECK (
        (status = 'cancelled' AND cancelled_at IS NOT NULL)
        OR (status <> 'cancelled' AND cancelled_at IS NULL)
    ),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS account_deletion_requests_open_idx
    ON account_deletion_requests (tenant_id, account_id)
    WHERE status IN ('requested', 'pending_review');
