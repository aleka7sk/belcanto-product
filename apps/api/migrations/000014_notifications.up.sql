-- Belcanto L.5 notifications and activity (Figma Page 31 ACT-01..03;
-- production prompt: an outbox row without a worker, retry policy and
-- delivery status is not a notification flow). The outbox gains an
-- explicit delivery lifecycle: pending → delivered | dead_letter with
-- retry scheduling. The worker materializes curated events into
-- per-recipient activity entries — the in-app channel is always on
-- (ACT-04 «Всегда включено»); per-category preferences are stored for
-- the push channel that ships with the device-permission modules.

ALTER TABLE outbox_events
    ADD COLUMN status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'delivered', 'dead_letter')),
    ADD COLUMN next_attempt_at timestamptz,
    ADD COLUMN last_error text CHECK (last_error IS NULL OR char_length(last_error) BETWEEN 1 AND 500);

-- Rows written before this migration stay pending and will be picked up
-- by the worker; already-published rows (none in practice) align.
UPDATE outbox_events SET status = 'delivered' WHERE published_at IS NOT NULL;

CREATE INDEX outbox_events_pending_idx
    ON outbox_events (status, next_attempt_at, id)
    WHERE status = 'pending';

CREATE TABLE activity_entries (
    id                   text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id            text NOT NULL REFERENCES tenants(id),
    recipient_account_id text NOT NULL,
    source_outbox_id     bigint NOT NULL,
    category             text NOT NULL CHECK (category IN ('important', 'learning', 'messages', 'community')),
    kind                 text NOT NULL CHECK (char_length(kind) BETWEEN 1 AND 100),
    target_type          text NOT NULL CHECK (char_length(target_type) BETWEEN 1 AND 100),
    target_id            text NOT NULL CHECK (char_length(target_id) BETWEEN 1 AND 160),
    payload              jsonb NOT NULL,
    occurred_at          timestamptz NOT NULL,
    read_at              timestamptz,
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, source_outbox_id, recipient_account_id),
    FOREIGN KEY (tenant_id, recipient_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX activity_entries_recipient_idx
    ON activity_entries (tenant_id, recipient_account_id, occurred_at DESC);

CREATE TABLE notification_preferences (
    tenant_id    text NOT NULL REFERENCES tenants(id),
    account_id   text NOT NULL,
    category     text NOT NULL CHECK (category IN ('important', 'learning', 'messages', 'community')),
    push_enabled boolean NOT NULL,
    updated_at   timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, account_id, category),
    FOREIGN KEY (tenant_id, account_id) REFERENCES accounts(tenant_id, id)
);
