-- Belcanto L.3 goals and achievements (Figma STU-GROWTH-04/08; Goal,
-- AchievementDefinition and AchievementAward aggregates per the domain
-- aggregate catalog). A goal has an explicit criterion and is reframed,
-- never «failed»: cancellation carries a reason and may link the
-- replacement goal. Completion carries an explicit decision note.
-- Achievements split the versioned definition from the evidence-backed
-- award; there are no ratings, points or streak penalties anywhere
-- (DEC-006), revocation preserves the original award, and nothing here
-- is ever deleted.

CREATE TABLE student_goals (
    id                     text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id              text NOT NULL REFERENCES tenants(id),
    student_id             text NOT NULL,
    criterion              text NOT NULL CHECK (char_length(criterion) BETWEEN 1 AND 500),
    description            text CHECK (description IS NULL OR char_length(description) BETWEEN 1 AND 1000),
    related_song_id        text,
    related_skill_area     text CHECK (related_skill_area IS NULL OR char_length(btrim(related_skill_area)) BETWEEN 1 AND 100),
    status                 text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'cancelled')),
    completion_note        text CHECK (completion_note IS NULL OR char_length(completion_note) BETWEEN 1 AND 500),
    cancel_reason          text CHECK (cancel_reason IS NULL OR char_length(cancel_reason) BETWEEN 1 AND 500),
    replaced_by_goal_id    text,
    created_by_account_id  text NOT NULL,
    version                integer NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at             timestamptz NOT NULL,
    updated_at             timestamptz NOT NULL,
    CHECK (status <> 'completed' OR completion_note IS NOT NULL),
    CHECK (status <> 'cancelled' OR cancel_reason IS NOT NULL),
    CHECK (replaced_by_goal_id IS NULL OR status = 'cancelled'),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, related_song_id) REFERENCES student_songs(tenant_id, id),
    FOREIGN KEY (tenant_id, replaced_by_goal_id) REFERENCES student_goals(tenant_id, id),
    FOREIGN KEY (tenant_id, created_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX student_goals_student_idx
    ON student_goals (tenant_id, student_id, updated_at DESC);

-- The criterion never rewrites (a material change is a replacement
-- goal); completed goals are immutable; a cancelled goal accepts only
-- the one-time reframe link; goals are never deleted.
CREATE FUNCTION guard_student_goal_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'goal history is immutable';
    END IF;
    IF NEW.criterion <> OLD.criterion
        OR NEW.description IS DISTINCT FROM OLD.description
        OR NEW.related_song_id IS DISTINCT FROM OLD.related_song_id
        OR NEW.related_skill_area IS DISTINCT FROM OLD.related_skill_area
    THEN
        RAISE EXCEPTION 'a goal criterion never rewrites; a material change is a replacement goal';
    END IF;
    IF OLD.status = 'completed' THEN
        RAISE EXCEPTION 'a completed goal is immutable';
    END IF;
    IF OLD.status = 'cancelled' THEN
        IF NEW.status = 'cancelled'
            AND NEW.cancel_reason = OLD.cancel_reason
            AND NEW.completion_note IS NOT DISTINCT FROM OLD.completion_note
            AND OLD.replaced_by_goal_id IS NULL
        THEN
            RETURN NEW;
        END IF;
        RAISE EXCEPTION 'a cancelled goal is immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER student_goals_guard
    BEFORE UPDATE OR DELETE ON student_goals
    FOR EACH ROW EXECUTE FUNCTION guard_student_goal_mutation();

CREATE TABLE achievement_definitions (
    id                    text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id             text NOT NULL REFERENCES tenants(id),
    name                  text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
    description           text NOT NULL CHECK (char_length(description) BETWEEN 1 AND 1000),
    category              text NOT NULL CHECK (char_length(btrim(category)) BETWEEN 1 AND 100),
    evidence_requirement  text CHECK (evidence_requirement IS NULL OR char_length(evidence_requirement) BETWEEN 1 AND 500),
    status                text NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'retired')),
    definition_version    integer NOT NULL DEFAULT 1 CHECK (definition_version > 0),
    created_by_account_id text NOT NULL,
    created_at            timestamptz NOT NULL,
    retired_at            timestamptz,
    CHECK ((status = 'retired') = (retired_at IS NOT NULL)),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, name),
    FOREIGN KEY (tenant_id, created_by_account_id) REFERENCES accounts(tenant_id, id)
);

-- A published definition is immutable within its version; the only
-- change a row accepts is retirement.
CREATE FUNCTION guard_achievement_definition_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'achievement definitions are immutable history';
    END IF;
    IF NEW.name <> OLD.name
        OR NEW.description <> OLD.description
        OR NEW.category <> OLD.category
        OR NEW.evidence_requirement IS DISTINCT FROM OLD.evidence_requirement
        OR NEW.definition_version <> OLD.definition_version
    THEN
        RAISE EXCEPTION 'a published definition is immutable; a material change is a new version';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER achievement_definitions_guard
    BEFORE UPDATE OR DELETE ON achievement_definitions
    FOR EACH ROW EXECUTE FUNCTION guard_achievement_definition_mutation();

CREATE TABLE achievement_awards (
    id                    text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id             text NOT NULL REFERENCES tenants(id),
    definition_id         text NOT NULL,
    definition_version    integer NOT NULL CHECK (definition_version > 0),
    student_id            text NOT NULL,
    evidence_note         text NOT NULL CHECK (char_length(evidence_note) BETWEEN 1 AND 1000),
    status                text NOT NULL DEFAULT 'awarded' CHECK (status IN ('awarded', 'revoked')),
    revoke_reason         text CHECK (revoke_reason IS NULL OR char_length(revoke_reason) BETWEEN 1 AND 500),
    revoked_at            timestamptz,
    awarded_by_account_id text NOT NULL,
    awarded_at            timestamptz NOT NULL,
    CHECK ((status = 'revoked') = (revoke_reason IS NOT NULL AND revoked_at IS NOT NULL)),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, definition_id) REFERENCES achievement_definitions(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, awarded_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX achievement_awards_student_idx
    ON achievement_awards (tenant_id, student_id, awarded_at DESC);

-- Revocation preserves the original award; nothing else ever changes.
CREATE FUNCTION guard_achievement_award_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'achievement awards are immutable history';
    END IF;
    IF NEW.definition_id <> OLD.definition_id
        OR NEW.definition_version <> OLD.definition_version
        OR NEW.student_id <> OLD.student_id
        OR NEW.evidence_note <> OLD.evidence_note
        OR NEW.awarded_by_account_id <> OLD.awarded_by_account_id
        OR NEW.awarded_at <> OLD.awarded_at
    THEN
        RAISE EXCEPTION 'an award is immutable; only revocation may be recorded';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER achievement_awards_guard
    BEFORE UPDATE OR DELETE ON achievement_awards
    FOR EACH ROW EXECUTE FUNCTION guard_achievement_award_mutation();
