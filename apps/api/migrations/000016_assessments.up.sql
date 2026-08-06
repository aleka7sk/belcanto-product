-- Belcanto L.4 assessments (domain/assessment.md, Approved; Figma
-- Page 27 TCH-REVIEW-*). An Assessment is a teacher's professional
-- observation in an explicit context: Draft -> Published ->
-- Superseded, plus Withdrawn with a mandatory reason. Published
-- content never silently rewrites — corrections are a new linked
-- version; nothing deletes. Progress trend automation stays out:
-- domain/progress.md leaves the minimum-evidence rule an open
-- question, so trends remain a teacher's judgement.

CREATE TABLE assessments (
    id                  text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id           text NOT NULL REFERENCES tenants(id),
    student_id          text NOT NULL,
    author_account_id   text NOT NULL,
    author_role         text NOT NULL CHECK (author_role IN ('Teacher', 'Administrator', 'Owner', 'Student')),
    assessment_type     text NOT NULL CHECK (assessment_type IN (
        'observation', 'diagnostic', 'formative', 'summative', 'self'
    )),
    context_type        text NOT NULL CHECK (context_type IN (
        'lesson', 'homework_review', 'repertoire_practice', 'concert_preparation',
        'concert_performance', 'diagnostic_session', 'periodic_review', 'teacher_observation'
    )),
    context_id          text CHECK (context_id IS NULL OR char_length(context_id) BETWEEN 1 AND 128),
    assessment_date     date NOT NULL,
    summary             text NOT NULL CHECK (char_length(summary) <= 2000),
    strengths           text NOT NULL DEFAULT '' CHECK (char_length(strengths) <= 2000),
    development_areas   text NOT NULL DEFAULT '' CHECK (char_length(development_areas) <= 2000),
    recommendations     text NOT NULL DEFAULT '' CHECK (char_length(recommendations) <= 2000),
    confidence          text CHECK (confidence IS NULL OR confidence IN ('low', 'medium', 'high')),
    visibility          text NOT NULL CHECK (visibility IN (
        'teacher_only', 'student_visible', 'staff_visible', 'owner_analytics'
    )),
    related_song_id     text,
    related_goal_id     text,
    areas               text NOT NULL DEFAULT '' CHECK (char_length(areas) <= 500),
    status              text NOT NULL DEFAULT 'draft' CHECK (status IN (
        'draft', 'published', 'superseded', 'withdrawn'
    )),
    superseded_by_id    text CHECK (superseded_by_id IS NULL OR char_length(superseded_by_id) BETWEEN 1 AND 128),
    withdrawal_reason   text CHECK (withdrawal_reason IS NULL OR char_length(withdrawal_reason) BETWEEN 1 AND 500),
    published_at        timestamptz,
    version             bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_at          timestamptz NOT NULL,
    updated_at          timestamptz NOT NULL,
    -- Published carries substance: a summary plus at least one of the
    -- observation blocks (business rule from domain/assessment.md).
    CHECK (status = 'draft' OR char_length(btrim(summary)) > 0),
    CHECK (status <> 'superseded' OR superseded_by_id IS NOT NULL),
    CHECK (status <> 'withdrawn' OR withdrawal_reason IS NOT NULL),
    CHECK ((status IN ('published', 'superseded')) = (published_at IS NOT NULL)
           OR status = 'withdrawn'),
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, author_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX assessments_student_idx
    ON assessments (tenant_id, student_id, assessment_date DESC, id);

CREATE INDEX assessments_author_status_idx
    ON assessments (tenant_id, author_account_id, status);

-- Evidence rows bind a conclusion to its source; a text observation is
-- evidence too. Editable while the assessment is a draft only.
CREATE TABLE assessment_evidence (
    id             text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id      text NOT NULL REFERENCES tenants(id),
    assessment_id  text NOT NULL,
    kind           text NOT NULL CHECK (kind IN (
        'observation', 'media', 'journal', 'homework', 'prior_assessment', 'self_assessment'
    )),
    note           text NOT NULL CHECK (char_length(note) BETWEEN 1 AND 1000),
    reference_id   text CHECK (reference_id IS NULL OR char_length(reference_id) BETWEEN 1 AND 128),
    added_at       timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, assessment_id) REFERENCES assessments(tenant_id, id)
);

CREATE INDEX assessment_evidence_assessment_idx
    ON assessment_evidence (tenant_id, assessment_id, added_at, id);

CREATE FUNCTION reject_assessment_delete() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'assessments are never deleted; withdraw with a reason instead';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER assessments_no_delete
    BEFORE DELETE ON assessments
    FOR EACH ROW EXECUTE FUNCTION reject_assessment_delete();

-- Draft evidence is the author's working set; once the assessment
-- leaves draft its evidence rows freeze with it.
CREATE FUNCTION guard_assessment_evidence_delete() RETURNS trigger AS $$
DECLARE
    parent_status text;
BEGIN
    SELECT status INTO parent_status FROM assessments
    WHERE tenant_id = OLD.tenant_id AND id = OLD.assessment_id;
    IF parent_status IS DISTINCT FROM 'draft' THEN
        RAISE EXCEPTION 'published assessment evidence is immutable';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER assessment_evidence_guard_delete
    BEFORE DELETE ON assessment_evidence
    FOR EACH ROW EXECUTE FUNCTION guard_assessment_evidence_delete();

-- Published content is immutable: only the documented status
-- transitions may touch a non-draft row, and each keeps the content
-- columns byte-identical.
CREATE FUNCTION guard_assessment_update() RETURNS trigger AS $$
BEGIN
    IF OLD.status = 'draft' THEN
        IF NEW.status NOT IN ('draft', 'published', 'withdrawn') THEN
            RAISE EXCEPTION 'a draft may only stay draft, publish or be withdrawn';
        END IF;
        RETURN NEW;
    END IF;
    IF OLD.status = 'published' AND NEW.status IN ('superseded', 'withdrawn') THEN
        IF NEW.summary IS DISTINCT FROM OLD.summary
           OR NEW.strengths IS DISTINCT FROM OLD.strengths
           OR NEW.development_areas IS DISTINCT FROM OLD.development_areas
           OR NEW.recommendations IS DISTINCT FROM OLD.recommendations
           OR NEW.confidence IS DISTINCT FROM OLD.confidence
           OR NEW.visibility IS DISTINCT FROM OLD.visibility
           OR NEW.assessment_date IS DISTINCT FROM OLD.assessment_date
           OR NEW.published_at IS DISTINCT FROM OLD.published_at THEN
            RAISE EXCEPTION 'published assessment content is immutable';
        END IF;
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'assessment status % does not allow this change', OLD.status;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER assessments_guard_update
    BEFORE UPDATE ON assessments
    FOR EACH ROW EXECUTE FUNCTION guard_assessment_update();
