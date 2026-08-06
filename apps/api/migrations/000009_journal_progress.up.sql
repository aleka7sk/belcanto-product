-- Belcanto L.3 lesson journals and evidence-based progress (Figma Pages
-- 22/26/27; flows C/D/H/I; HOF DEC-006, DEC-007). The journal keeps the
-- school's language of the First Belcanto Minute: what worked, current
-- focus, next step. A draft is teacher-private; publishing snapshots it
-- as an immutable version the student can see; a correction publishes
-- the next version with an explicit note — published history never
-- changes (DEC-007). Progress is evidence, never a score (DEC-006):
-- publishing may attach observations tied to named areas.

CREATE TABLE lesson_journals (
    id                    text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id             text NOT NULL REFERENCES tenants(id),
    occurrence_id         text NOT NULL,
    student_id            text NOT NULL,
    teacher_account_id    text NOT NULL,
    status                text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    current_version       integer NOT NULL DEFAULT 0 CHECK (current_version >= 0),
    draft_what_worked     text CHECK (draft_what_worked IS NULL OR char_length(draft_what_worked) BETWEEN 1 AND 2000),
    draft_current_focus   text CHECK (draft_current_focus IS NULL OR char_length(draft_current_focus) BETWEEN 1 AND 2000),
    draft_next_step       text CHECK (draft_next_step IS NULL OR char_length(draft_next_step) BETWEEN 1 AND 2000),
    created_at            timestamptz NOT NULL,
    updated_at            timestamptz NOT NULL,
    CHECK (status = 'draft' OR current_version > 0),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, occurrence_id, student_id),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES core_lesson_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, teacher_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX lesson_journals_student_idx
    ON lesson_journals (tenant_id, student_id, updated_at DESC);

CREATE TABLE lesson_journal_versions (
    tenant_id               text NOT NULL,
    journal_id              text NOT NULL,
    version                 integer NOT NULL CHECK (version > 0),
    what_worked             text NOT NULL CHECK (char_length(what_worked) BETWEEN 1 AND 2000),
    current_focus           text NOT NULL CHECK (char_length(current_focus) BETWEEN 1 AND 2000),
    next_step               text NOT NULL CHECK (char_length(next_step) BETWEEN 1 AND 2000),
    correction_note         text CHECK (correction_note IS NULL OR char_length(correction_note) BETWEEN 1 AND 500),
    published_by_account_id text NOT NULL,
    published_at            timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, journal_id, version),
    CHECK (version = 1 OR correction_note IS NOT NULL),
    FOREIGN KEY (tenant_id, journal_id) REFERENCES lesson_journals(tenant_id, id),
    FOREIGN KEY (tenant_id, published_by_account_id) REFERENCES accounts(tenant_id, id)
);

-- DEC-007 guard: a published version is a record of what the student
-- was told; it never changes and never disappears.
CREATE FUNCTION reject_journal_version_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'published journal versions are immutable (DEC-007)';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER lesson_journal_versions_immutable
    BEFORE UPDATE OR DELETE ON lesson_journal_versions
    FOR EACH ROW EXECUTE FUNCTION reject_journal_version_mutation();

-- DEC-006: progress is observed evidence tied to a named area — there
-- is no numeric score anywhere in this model, on purpose.
CREATE TABLE progress_evidence (
    id                     text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id              text NOT NULL REFERENCES tenants(id),
    student_id             text NOT NULL,
    source_kind            text NOT NULL CHECK (source_kind IN ('lesson_journal', 'practice', 'review')),
    source_id              text NOT NULL CHECK (char_length(source_id) BETWEEN 1 AND 160),
    area                   text NOT NULL CHECK (char_length(btrim(area)) BETWEEN 1 AND 100),
    note                   text NOT NULL CHECK (char_length(note) BETWEEN 1 AND 1000),
    recorded_by_account_id text NOT NULL,
    recorded_at            timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, recorded_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX progress_evidence_student_idx
    ON progress_evidence (tenant_id, student_id, recorded_at DESC);

CREATE TRIGGER progress_evidence_immutable
    BEFORE UPDATE OR DELETE ON progress_evidence
    FOR EACH ROW EXECUTE FUNCTION reject_journal_version_mutation();
