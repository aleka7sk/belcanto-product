-- Belcanto L.5 community and safety (Figma Page 28; production prompt:
-- role-aware audience, report/block, moderation queue with reasons and
-- audit; no unrestricted student discovery and no Student↔Student DM —
-- chat ships as its own slice). Posts are text-only while DEC-103
-- (guardian consent for community media) is open. Nothing is deleted:
-- an author removal or a moderation hide is a status with a preserved
-- tombstone (COM-SAFE-05), and reporter identity is never shown to
-- non-moderators.

CREATE TABLE community_posts (
    id                text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id         text NOT NULL REFERENCES tenants(id),
    author_account_id text NOT NULL,
    kind              text NOT NULL CHECK (kind IN ('post', 'announcement')),
    title             text CHECK (title IS NULL OR char_length(title) BETWEEN 1 AND 200),
    body              text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
    audience          text NOT NULL CHECK (audience IN ('school', 'staff')),
    comments_enabled  boolean NOT NULL,
    pinned            boolean NOT NULL DEFAULT false,
    status            text NOT NULL DEFAULT 'published'
        CHECK (status IN ('published', 'hidden', 'removed')),
    status_reason     text CHECK (status_reason IS NULL OR char_length(status_reason) BETWEEN 1 AND 500),
    created_at        timestamptz NOT NULL,
    updated_at        timestamptz NOT NULL,
    CHECK (kind <> 'announcement' OR title IS NOT NULL),
    CHECK (status = 'published' OR status_reason IS NOT NULL),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, author_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX community_posts_feed_idx
    ON community_posts (tenant_id, pinned DESC, created_at DESC);

CREATE FUNCTION reject_community_delete() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'community history is immutable; removal is a status';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER community_posts_no_delete
    BEFORE DELETE ON community_posts
    FOR EACH ROW EXECUTE FUNCTION reject_community_delete();

CREATE TABLE community_comments (
    id                text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id         text NOT NULL REFERENCES tenants(id),
    post_id           text NOT NULL,
    author_account_id text NOT NULL,
    body              text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 1000),
    status            text NOT NULL DEFAULT 'published'
        CHECK (status IN ('published', 'hidden', 'removed')),
    status_reason     text CHECK (status_reason IS NULL OR char_length(status_reason) BETWEEN 1 AND 500),
    created_at        timestamptz NOT NULL,
    CHECK (status = 'published' OR status_reason IS NOT NULL),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, post_id) REFERENCES community_posts(tenant_id, id),
    FOREIGN KEY (tenant_id, author_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX community_comments_post_idx
    ON community_comments (tenant_id, post_id, created_at);

CREATE TRIGGER community_comments_no_delete
    BEFORE DELETE ON community_comments
    FOR EACH ROW EXECUTE FUNCTION reject_community_delete();

CREATE TABLE community_reports (
    id                  text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id           text NOT NULL REFERENCES tenants(id),
    target_type         text NOT NULL CHECK (target_type IN ('post', 'comment')),
    target_id           text NOT NULL CHECK (char_length(target_id) BETWEEN 1 AND 128),
    reason              text NOT NULL CHECK (reason IN ('abuse', 'personal_data', 'spam', 'other')),
    note                text CHECK (note IS NULL OR char_length(note) BETWEEN 1 AND 1000),
    reporter_account_id text NOT NULL,
    status              text NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'reviewed')),
    decision            text CHECK (decision IS NULL OR decision IN ('hidden', 'kept')),
    decision_reason     text CHECK (decision_reason IS NULL OR char_length(decision_reason) BETWEEN 1 AND 500),
    decided_by_account_id text,
    decided_at          timestamptz,
    created_at          timestamptz NOT NULL,
    CHECK (reason <> 'other' OR note IS NOT NULL),
    CHECK ((status = 'reviewed') = (decision IS NOT NULL AND decision_reason IS NOT NULL
                                    AND decided_by_account_id IS NOT NULL AND decided_at IS NOT NULL)),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, reporter_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, decided_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX community_reports_queue_idx
    ON community_reports (tenant_id, status, created_at DESC);

CREATE TRIGGER community_reports_no_delete
    BEFORE DELETE ON community_reports
    FOR EACH ROW EXECUTE FUNCTION reject_community_delete();

CREATE TABLE community_blocks (
    tenant_id          text NOT NULL REFERENCES tenants(id),
    blocker_account_id text NOT NULL,
    blocked_account_id text NOT NULL,
    created_at         timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, blocker_account_id, blocked_account_id),
    CHECK (blocker_account_id <> blocked_account_id),
    FOREIGN KEY (tenant_id, blocker_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, blocked_account_id) REFERENCES accounts(tenant_id, id)
);
