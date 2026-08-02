CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;

ALTER TABLE teacher_assignments
    ADD COLUMN effective_from timestamptz,
    ADD COLUMN effective_until timestamptz,
    ADD COLUMN version bigint NOT NULL DEFAULT 0 CHECK (version >= 0);

UPDATE teacher_assignments
SET effective_from = assigned_at,
    effective_until = ended_at;

ALTER TABLE teacher_assignments
    ALTER COLUMN effective_from SET NOT NULL,
    ALTER COLUMN effective_from SET DEFAULT now(),
    ADD CONSTRAINT teacher_assignments_effective_interval
        CHECK (effective_until IS NULL OR effective_until >= effective_from);

DROP INDEX teacher_assignments_one_active_primary;

ALTER TABLE teacher_assignments
    ADD CONSTRAINT teacher_assignments_no_active_overlap
    EXCLUDE USING gist (
        tenant_id public.gist_text_ops WITH =,
        student_id public.gist_text_ops WITH =,
        tstzrange(effective_from, effective_until, '[)') WITH &&
    ) WHERE (status = 'active');

CREATE INDEX teacher_assignments_effective_lookup_idx
    ON teacher_assignments (tenant_id, student_id, effective_from DESC)
    WHERE status = 'active';

CREATE TABLE lessons (
    id                      text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id               text NOT NULL REFERENCES tenants(id),
    title                   text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    starts_at               timestamptz NOT NULL,
    duration_minutes        integer NOT NULL CHECK (duration_minutes BETWEEN 1 AND 1440),
    location                text CHECK (location IS NULL OR char_length(btrim(location)) BETWEEN 1 AND 200),
    teacher_account_id      text NOT NULL,
    status                  text NOT NULL CHECK (status = 'scheduled'),
    version                 bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_by_account_id   text NOT NULL,
    created_at              timestamptz NOT NULL,
    updated_at              timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, teacher_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, created_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TABLE lesson_participants (
    tenant_id       text NOT NULL,
    lesson_id       text NOT NULL,
    student_id      text NOT NULL,
    added_at        timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, lesson_id, student_id),
    FOREIGN KEY (tenant_id, lesson_id) REFERENCES lessons(tenant_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id) ON DELETE RESTRICT
);

CREATE INDEX lessons_tenant_starts_at_idx
    ON lessons (tenant_id, starts_at, id);

CREATE INDEX lessons_teacher_starts_at_idx
    ON lessons (tenant_id, teacher_account_id, starts_at, id);

CREATE INDEX lesson_participants_student_idx
    ON lesson_participants (tenant_id, student_id, lesson_id);
