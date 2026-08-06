-- Belcanto L.2 reschedule and cancellation requests (Figma Page 24
-- 378:803, Page 29; flows J/K/L). A participant student or the lesson's
-- teacher asks; Owner/Administrator decides. DEC-102 is open: nothing
-- here computes or applies any late-cancellation consequence — the
-- request records who asked, when and why, and the decision applies the
-- change itself. Approval moves the occurrence in place (references from
-- attendance and future journals stay intact); the request row and the
-- audit trail keep the history.

CREATE TABLE lesson_reschedule_requests (
    id                      text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id               text NOT NULL REFERENCES tenants(id),
    occurrence_id           text NOT NULL,
    requested_by_account_id text NOT NULL,
    kind                    text NOT NULL CHECK (kind IN ('reschedule', 'cancellation')),
    proposed_starts_at      timestamptz,
    reason                  text NOT NULL CHECK (char_length(btrim(reason)) BETWEEN 1 AND 500),
    status                  text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'declined', 'withdrawn')),
    decided_by_account_id   text,
    decision_note           text CHECK (decision_note IS NULL OR char_length(decision_note) BETWEEN 1 AND 500),
    decided_at              timestamptz,
    version                 bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_at              timestamptz NOT NULL,
    updated_at              timestamptz NOT NULL,
    CHECK (
        (kind = 'reschedule' AND proposed_starts_at IS NOT NULL)
        OR (kind = 'cancellation' AND proposed_starts_at IS NULL)
    ),
    CHECK (
        (status IN ('approved', 'declined') AND decided_by_account_id IS NOT NULL AND decided_at IS NOT NULL)
        OR (status IN ('pending', 'withdrawn') AND decided_by_account_id IS NULL AND decided_at IS NULL)
    ),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES core_lesson_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, requested_by_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, decided_by_account_id) REFERENCES accounts(tenant_id, id)
);

-- One open request per occurrence and requester: repeated asks amend the
-- open one through withdraw + re-create, never pile up.
CREATE UNIQUE INDEX lesson_reschedule_requests_open_idx
    ON lesson_reschedule_requests (tenant_id, occurrence_id, requested_by_account_id)
    WHERE status = 'pending';

CREATE INDEX lesson_reschedule_requests_pending_idx
    ON lesson_reschedule_requests (tenant_id, created_at)
    WHERE status = 'pending';
