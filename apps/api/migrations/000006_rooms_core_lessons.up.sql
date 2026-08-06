-- Belcanto L.2 rooms and core lessons (Figma Pages 24/26/29; HOF DEC-001,
-- DEC-002, DEC-004). Core lessons are the paid learning obligation and
-- never mix with events. Format is structural: individual holds exactly
-- one student, group holds at most three (DEC-002) — enforced by trigger
-- on both series enrollments and occurrence participants. Recurrence has
-- exactly two scopes (DEC-004): the weekly series and the single
-- occurrence; occurrence rows may override teacher/room/time.

CREATE TABLE rooms (
    id          text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id   text NOT NULL REFERENCES tenants(id),
    name        text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    capacity    integer CHECK (capacity IS NULL OR capacity BETWEEN 1 AND 500),
    status      text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    version     bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, name)
);

-- server_config already exists (migration 000003) and carries the open
-- P0/P1 decision knobs (HOF-22): DEC-101 waitlist TTL, DEC-102 late
-- cancellation (default: no automatic consequences), DEC-106 quiet-hours
-- bypass, DEC-107 moderation SLA.

CREATE TABLE core_lesson_series (
    id                    text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id             text NOT NULL REFERENCES tenants(id),
    format                text NOT NULL CHECK (format IN ('individual', 'group')),
    title                 text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    teacher_account_id    text NOT NULL,
    room_id               text,
    weekday               smallint NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    start_minutes         smallint NOT NULL CHECK (start_minutes BETWEEN 0 AND 1439),
    duration_minutes      integer NOT NULL CHECK (duration_minutes BETWEEN 1 AND 1440),
    effective_from        date NOT NULL,
    effective_until       date CHECK (effective_until IS NULL OR effective_until >= effective_from),
    status                text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'ended')),
    version               bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_by_account_id text NOT NULL,
    created_at            timestamptz NOT NULL,
    updated_at            timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, teacher_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, room_id) REFERENCES rooms(tenant_id, id),
    FOREIGN KEY (tenant_id, created_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX core_lesson_series_teacher_idx
    ON core_lesson_series (tenant_id, teacher_account_id, status);

CREATE TABLE core_lesson_series_enrollments (
    tenant_id   text NOT NULL,
    series_id   text NOT NULL,
    student_id  text NOT NULL,
    added_at    timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, series_id, student_id),
    FOREIGN KEY (tenant_id, series_id) REFERENCES core_lesson_series(tenant_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id) ON DELETE RESTRICT
);

CREATE INDEX core_lesson_series_enrollments_student_idx
    ON core_lesson_series_enrollments (tenant_id, student_id, series_id);

CREATE TABLE core_lesson_occurrences (
    id                    text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id             text NOT NULL REFERENCES tenants(id),
    series_id             text,
    format                text NOT NULL CHECK (format IN ('individual', 'group')),
    title                 text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    starts_at             timestamptz NOT NULL,
    duration_minutes      integer NOT NULL CHECK (duration_minutes BETWEEN 1 AND 1440),
    teacher_account_id    text NOT NULL,
    room_id               text,
    location_note         text CHECK (location_note IS NULL OR char_length(btrim(location_note)) BETWEEN 1 AND 200),
    status                text NOT NULL DEFAULT 'scheduled' CHECK (status IN (
        'scheduled', 'completed', 'cancelled_school', 'cancelled_student', 'rescheduled', 'no_show'
    )),
    version               bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_by_account_id text NOT NULL,
    created_at            timestamptz NOT NULL,
    updated_at            timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, series_id) REFERENCES core_lesson_series(tenant_id, id),
    FOREIGN KEY (tenant_id, teacher_account_id) REFERENCES accounts(tenant_id, id),
    FOREIGN KEY (tenant_id, room_id) REFERENCES rooms(tenant_id, id),
    FOREIGN KEY (tenant_id, created_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX core_lesson_occurrences_tenant_starts_idx
    ON core_lesson_occurrences (tenant_id, starts_at, id);

CREATE INDEX core_lesson_occurrences_teacher_starts_idx
    ON core_lesson_occurrences (tenant_id, teacher_account_id, starts_at, id);

CREATE INDEX core_lesson_occurrences_series_idx
    ON core_lesson_occurrences (tenant_id, series_id, starts_at)
    WHERE series_id IS NOT NULL;

CREATE TABLE core_lesson_occurrence_participants (
    tenant_id      text NOT NULL,
    occurrence_id  text NOT NULL,
    student_id     text NOT NULL,
    added_at       timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, occurrence_id, student_id),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES core_lesson_occurrences(tenant_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id) ON DELETE RESTRICT
);

CREATE INDEX core_lesson_occurrence_participants_student_idx
    ON core_lesson_occurrence_participants (tenant_id, student_id, occurrence_id);

CREATE TABLE attendance_records (
    tenant_id            text NOT NULL,
    occurrence_id        text NOT NULL,
    student_id           text NOT NULL,
    state                text NOT NULL CHECK (state IN ('present', 'absent_excused', 'absent_unexcused', 'late')),
    note                 text CHECK (note IS NULL OR char_length(note) BETWEEN 1 AND 500),
    marked_by_account_id text NOT NULL,
    marked_at            timestamptz NOT NULL,
    version              bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    PRIMARY KEY (tenant_id, occurrence_id, student_id),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES core_lesson_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, marked_by_account_id) REFERENCES accounts(tenant_id, id)
);

-- DEC-002 structural guard: individual = exactly one student, group = at
-- most three. The trigger closes the race a plain CHECK cannot see.
CREATE FUNCTION enforce_core_lesson_series_capacity() RETURNS trigger AS $$
DECLARE
    series_format text;
    member_count  integer;
BEGIN
    SELECT format INTO series_format
    FROM core_lesson_series
    WHERE tenant_id = NEW.tenant_id AND id = NEW.series_id
    FOR UPDATE;
    IF series_format IS NULL THEN
        RAISE EXCEPTION 'core lesson series % was not found', NEW.series_id;
    END IF;
    SELECT count(*) INTO member_count
    FROM core_lesson_series_enrollments
    WHERE tenant_id = NEW.tenant_id AND series_id = NEW.series_id;
    IF series_format = 'individual' AND member_count > 1 THEN
        RAISE EXCEPTION 'individual core lesson holds exactly one student (DEC-002)';
    END IF;
    IF series_format = 'group' AND member_count > 3 THEN
        RAISE EXCEPTION 'group core lesson holds at most three students (DEC-002)';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER core_lesson_series_capacity_guard
    AFTER INSERT ON core_lesson_series_enrollments
    FOR EACH ROW EXECUTE FUNCTION enforce_core_lesson_series_capacity();

CREATE FUNCTION enforce_core_lesson_occurrence_capacity() RETURNS trigger AS $$
DECLARE
    occurrence_format text;
    member_count      integer;
BEGIN
    SELECT format INTO occurrence_format
    FROM core_lesson_occurrences
    WHERE tenant_id = NEW.tenant_id AND id = NEW.occurrence_id
    FOR UPDATE;
    IF occurrence_format IS NULL THEN
        RAISE EXCEPTION 'core lesson occurrence % was not found', NEW.occurrence_id;
    END IF;
    SELECT count(*) INTO member_count
    FROM core_lesson_occurrence_participants
    WHERE tenant_id = NEW.tenant_id AND occurrence_id = NEW.occurrence_id;
    IF occurrence_format = 'individual' AND member_count > 1 THEN
        RAISE EXCEPTION 'individual core lesson holds exactly one student (DEC-002)';
    END IF;
    IF occurrence_format = 'group' AND member_count > 3 THEN
        RAISE EXCEPTION 'group core lesson holds at most three students (DEC-002)';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER core_lesson_occurrence_capacity_guard
    AFTER INSERT ON core_lesson_occurrence_participants
    FOR EACH ROW EXECUTE FUNCTION enforce_core_lesson_occurrence_capacity();

-- Migrate legacy generic lessons into ad-hoc core occurrences. A legacy
-- lesson with more than three participants cannot be classified under
-- DEC-002 — the migration refuses deterministically instead of guessing.
DO $$
DECLARE
    offending text;
BEGIN
    SELECT lesson_id INTO offending
    FROM lesson_participants
    GROUP BY tenant_id, lesson_id
    HAVING count(*) > 3
    LIMIT 1;
    IF offending IS NOT NULL THEN
        RAISE EXCEPTION 'legacy lesson % has more than three participants and cannot migrate under DEC-002', offending;
    END IF;
END
$$;

INSERT INTO core_lesson_occurrences (
    id, tenant_id, series_id, format, title, starts_at, duration_minutes,
    teacher_account_id, room_id, location_note, status, version,
    created_by_account_id, created_at, updated_at
)
SELECT
    l.id, l.tenant_id, NULL,
    CASE WHEN COALESCE(p.participant_count, 0) > 1 THEN 'group' ELSE 'individual' END,
    l.title, l.starts_at, l.duration_minutes,
    l.teacher_account_id, NULL, l.location, l.status, l.version,
    l.created_by_account_id, l.created_at, l.updated_at
FROM lessons l
LEFT JOIN (
    SELECT tenant_id, lesson_id, count(*) AS participant_count
    FROM lesson_participants
    GROUP BY tenant_id, lesson_id
) p ON p.tenant_id = l.tenant_id AND p.lesson_id = l.id;

INSERT INTO core_lesson_occurrence_participants (tenant_id, occurrence_id, student_id, added_at)
SELECT tenant_id, lesson_id, student_id, added_at
FROM lesson_participants;

-- Product history is immutable: the legacy tables stay readable forever
-- but accept no further writes — the core model is the only write path.
CREATE FUNCTION reject_legacy_lesson_writes() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'legacy lesson tables are read-only after the core lesson migration';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER lessons_read_only
    BEFORE INSERT OR UPDATE OR DELETE ON lessons
    FOR EACH ROW EXECUTE FUNCTION reject_legacy_lesson_writes();

CREATE TRIGGER lesson_participants_read_only
    BEFORE INSERT OR UPDATE OR DELETE ON lesson_participants
    FOR EACH ROW EXECUTE FUNCTION reject_legacy_lesson_writes();
