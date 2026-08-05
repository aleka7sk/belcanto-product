-- Belcanto P.1 session security: session inventory metadata, password
-- recovery and the server configuration registry (Figma Page 32 contract,
-- HOF-12/HOF-22: TTLs and policy values live in server config, not code).
-- Additive only; existing rows keep working with NULL metadata.

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS device_label text
        CHECK (device_label IS NULL OR char_length(device_label) BETWEEN 1 AND 120);

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS platform text
        CHECK (platform IS NULL OR platform IN ('ios', 'android', 'web'));

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS last_seen_at timestamptz;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id           text PRIMARY KEY,
    tenant_id    text NOT NULL REFERENCES tenants(id),
    account_id   text NOT NULL,
    token_digest bytea NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    expires_at   timestamptz NOT NULL,
    consumed_at  timestamptz,
    superseded_at timestamptz,
    created_at   timestamptz NOT NULL,
    CHECK (expires_at > created_at),
    CHECK (consumed_at IS NULL OR superseded_at IS NULL),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX IF NOT EXISTS password_reset_tokens_account_open_idx
    ON password_reset_tokens (tenant_id, account_id)
    WHERE consumed_at IS NULL AND superseded_at IS NULL;

CREATE TABLE IF NOT EXISTS server_config (
    key        text PRIMARY KEY CHECK (key ~ '^[a-z][a-z0-9_.-]{2,100}$'),
    value      jsonb NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
