-- Belcanto L.3 homework and practice (domain/homework.md Approved 1.0.0;
-- Figma Page 23 STU-PRACTICE-01..16; flows E/P). Homework belongs to a
-- Lesson, has an author and one Student, and moves through
-- draft → assigned → in_progress → submitted → reviewed → completed with
-- cancelled (reason preserved) and expired (deadline passed, stays in
-- history) on the side. Completed is final and homework is never
-- deleted: history survives in full. Practice attempts are versioned
-- immutable submissions; teacher feedback is immutable; an accepted
-- review may append named-area progress evidence (DEC-006 — no score).
-- Media rows carry only adapter storage keys — never URLs (upload
-- lifecycle: pending → uploading → ready | failed, resumable by offset).

CREATE TABLE media_objects (
    id               text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id        text NOT NULL REFERENCES tenants(id),
    owner_account_id text NOT NULL,
    kind             text NOT NULL CHECK (kind IN ('audio', 'video', 'image', 'pdf')),
    content_type     text NOT NULL CHECK (char_length(content_type) BETWEEN 3 AND 120),
    byte_size        bigint NOT NULL CHECK (byte_size > 0),
    uploaded_bytes   bigint NOT NULL DEFAULT 0 CHECK (uploaded_bytes >= 0 AND uploaded_bytes <= byte_size),
    status           text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'uploading', 'ready', 'failed')),
    storage_key      text NOT NULL CHECK (char_length(storage_key) BETWEEN 1 AND 200),
    created_at       timestamptz NOT NULL,
    updated_at       timestamptz NOT NULL,
    CHECK (status <> 'ready' OR uploaded_bytes = byte_size),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, owner_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX media_objects_owner_idx
    ON media_objects (tenant_id, owner_account_id, created_at DESC);

-- Ready media is a referenced artifact; its metadata never changes and
-- no deletion path exists while the retention decision (DEC-104) is open.
CREATE FUNCTION guard_media_object_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'media objects are immutable history (DEC-104 open)';
    END IF;
    IF OLD.status = 'ready' THEN
        RAISE EXCEPTION 'ready media is immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER media_objects_guard
    BEFORE UPDATE OR DELETE ON media_objects
    FOR EACH ROW EXECUTE FUNCTION guard_media_object_mutation();

CREATE TABLE homework_assignments (
    id                 text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id          text NOT NULL REFERENCES tenants(id),
    occurrence_id      text NOT NULL,
    student_id         text NOT NULL,
    teacher_account_id text NOT NULL,
    status             text NOT NULL DEFAULT 'draft' CHECK (status IN (
                           'draft', 'assigned', 'in_progress', 'submitted',
                           'reviewed', 'completed', 'cancelled', 'expired')),
    goal               text NOT NULL CHECK (char_length(goal) BETWEEN 1 AND 2000),
    readiness_criteria text CHECK (readiness_criteria IS NULL OR char_length(readiness_criteria) BETWEEN 1 AND 1000),
    due_at             timestamptz,
    cancel_reason      text CHECK (cancel_reason IS NULL OR char_length(cancel_reason) BETWEEN 1 AND 500),
    version            integer NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL,
    CHECK (status <> 'cancelled' OR cancel_reason IS NOT NULL),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES core_lesson_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, teacher_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX homework_assignments_student_idx
    ON homework_assignments (tenant_id, student_id, updated_at DESC);
CREATE INDEX homework_assignments_occurrence_idx
    ON homework_assignments (tenant_id, occurrence_id);

-- Completed is final and homework never disappears from history.
CREATE FUNCTION guard_homework_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'homework history is immutable';
    END IF;
    IF OLD.status = 'completed' THEN
        RAISE EXCEPTION 'completed homework is immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER homework_assignments_guard
    BEFORE UPDATE OR DELETE ON homework_assignments
    FOR EACH ROW EXECUTE FUNCTION guard_homework_mutation();

CREATE TABLE homework_tasks (
    tenant_id           text NOT NULL,
    homework_id         text NOT NULL,
    id                  text NOT NULL CHECK (char_length(id) BETWEEN 1 AND 128),
    position            integer NOT NULL CHECK (position > 0),
    title               text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    description         text CHECK (description IS NULL OR char_length(description) BETWEEN 1 AND 1000),
    recommended_minutes integer CHECK (recommended_minutes IS NULL OR recommended_minutes BETWEEN 1 AND 600),
    skill_area          text CHECK (skill_area IS NULL OR char_length(btrim(skill_area)) BETWEEN 1 AND 100),
    song_title          text CHECK (song_title IS NULL OR char_length(song_title) BETWEEN 1 AND 200),
    status              text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'done')),
    PRIMARY KEY (tenant_id, homework_id, id),
    UNIQUE (tenant_id, homework_id, position),
    FOREIGN KEY (tenant_id, homework_id) REFERENCES homework_assignments(tenant_id, id)
);

-- Tasks are editable only while the homework is a draft; afterwards the
-- only change a row accepts is its own completion status.
CREATE FUNCTION guard_homework_task_mutation() RETURNS trigger AS $$
DECLARE
    parent_status text;
BEGIN
    SELECT status INTO parent_status
        FROM homework_assignments
        WHERE tenant_id = OLD.tenant_id AND id = OLD.homework_id;
    IF TG_OP = 'DELETE' THEN
        IF parent_status = 'draft' THEN
            RETURN OLD;
        END IF;
        RAISE EXCEPTION 'homework tasks are removable only in draft';
    END IF;
    IF parent_status = 'draft' THEN
        RETURN NEW;
    END IF;
    IF NEW.position = OLD.position
        AND NEW.title = OLD.title
        AND NEW.description IS NOT DISTINCT FROM OLD.description
        AND NEW.recommended_minutes IS NOT DISTINCT FROM OLD.recommended_minutes
        AND NEW.skill_area IS NOT DISTINCT FROM OLD.skill_area
        AND NEW.song_title IS NOT DISTINCT FROM OLD.song_title
    THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'assigned homework tasks accept only status changes';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER homework_tasks_guard
    BEFORE UPDATE OR DELETE ON homework_tasks
    FOR EACH ROW EXECUTE FUNCTION guard_homework_task_mutation();

CREATE TABLE homework_attachments (
    tenant_id   text NOT NULL,
    homework_id text NOT NULL,
    media_id    text NOT NULL,
    position    integer NOT NULL CHECK (position > 0),
    PRIMARY KEY (tenant_id, homework_id, media_id),
    UNIQUE (tenant_id, homework_id, position),
    FOREIGN KEY (tenant_id, homework_id) REFERENCES homework_assignments(tenant_id, id),
    FOREIGN KEY (tenant_id, media_id) REFERENCES media_objects(tenant_id, id)
);

CREATE TABLE practice_submissions (
    id           text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id    text NOT NULL REFERENCES tenants(id),
    homework_id  text NOT NULL,
    student_id   text NOT NULL,
    attempt      integer NOT NULL CHECK (attempt > 0),
    note         text CHECK (note IS NULL OR char_length(note) BETWEEN 1 AND 1000),
    submitted_at timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, homework_id, attempt),
    FOREIGN KEY (tenant_id, homework_id) REFERENCES homework_assignments(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id)
);

CREATE FUNCTION reject_practice_history_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'practice history is immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER practice_submissions_immutable
    BEFORE UPDATE OR DELETE ON practice_submissions
    FOR EACH ROW EXECUTE FUNCTION reject_practice_history_mutation();

CREATE TABLE practice_submission_media (
    tenant_id     text NOT NULL,
    submission_id text NOT NULL,
    media_id      text NOT NULL,
    position      integer NOT NULL CHECK (position > 0),
    PRIMARY KEY (tenant_id, submission_id, media_id),
    UNIQUE (tenant_id, submission_id, position),
    FOREIGN KEY (tenant_id, submission_id) REFERENCES practice_submissions(tenant_id, id),
    FOREIGN KEY (tenant_id, media_id) REFERENCES media_objects(tenant_id, id)
);

CREATE TRIGGER practice_submission_media_immutable
    BEFORE UPDATE OR DELETE ON practice_submission_media
    FOR EACH ROW EXECUTE FUNCTION reject_practice_history_mutation();

CREATE TABLE practice_feedback (
    id                 text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id          text NOT NULL REFERENCES tenants(id),
    homework_id        text NOT NULL,
    submission_id      text NOT NULL,
    teacher_account_id text NOT NULL,
    decision           text NOT NULL CHECK (decision IN ('needs_revision', 'accepted')),
    body               text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
    next_step          text CHECK (next_step IS NULL OR char_length(next_step) BETWEEN 1 AND 1000),
    evidence_area      text CHECK (evidence_area IS NULL OR char_length(btrim(evidence_area)) BETWEEN 1 AND 100),
    evidence_note      text CHECK (evidence_note IS NULL OR char_length(evidence_note) BETWEEN 1 AND 1000),
    created_at         timestamptz NOT NULL,
    CHECK (decision = 'accepted' OR (evidence_area IS NULL AND evidence_note IS NULL)),
    CHECK ((evidence_area IS NULL) = (evidence_note IS NULL)),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, submission_id),
    FOREIGN KEY (tenant_id, homework_id) REFERENCES homework_assignments(tenant_id, id),
    FOREIGN KEY (tenant_id, submission_id) REFERENCES practice_submissions(tenant_id, id),
    FOREIGN KEY (tenant_id, teacher_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TRIGGER practice_feedback_immutable
    BEFORE UPDATE OR DELETE ON practice_feedback
    FOR EACH ROW EXECUTE FUNCTION reject_practice_history_mutation();
