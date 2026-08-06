-- Belcanto L.3 student repertoire (Figma STU-GROWTH-07, STU-PRACTICE-10/11;
-- aggregate StudentSong per the domain aggregate catalog). A song walks
-- the design's explicit journey «Знакомство → разучиваю → технически
-- устойчиво → интерпретация → готово к сцене»; the stage is a named
-- journey step set by the Teacher — never a computed readiness and never
-- a score (DEC-006). SongReadiness (concert eligibility, policy-computed)
-- is a separate future aggregate and deliberately absent here. Stage
-- history is append-only: the path of a piece never rewrites itself.

CREATE TABLE student_songs (
    id                     text PRIMARY KEY CHECK (char_length(id) BETWEEN 1 AND 128),
    tenant_id              text NOT NULL REFERENCES tenants(id),
    student_id             text NOT NULL,
    title                  text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    artist                 text CHECK (artist IS NULL OR char_length(artist) BETWEEN 1 AND 200),
    stage                  text NOT NULL CHECK (stage IN (
                               'acquaintance', 'learning', 'technically_stable',
                               'interpretation', 'stage_ready')),
    stage_note             text CHECK (stage_note IS NULL OR char_length(stage_note) BETWEEN 1 AND 500),
    assigned_by_account_id text NOT NULL,
    version                integer NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at             timestamptz NOT NULL,
    updated_at             timestamptz NOT NULL,
    UNIQUE (tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, assigned_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX student_songs_student_idx
    ON student_songs (tenant_id, student_id, updated_at DESC);

-- Repertoire is education history: items are never deleted.
CREATE FUNCTION reject_student_song_delete() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'repertoire history is immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER student_songs_no_delete
    BEFORE DELETE ON student_songs
    FOR EACH ROW EXECUTE FUNCTION reject_student_song_delete();

CREATE TABLE student_song_stage_history (
    tenant_id             text NOT NULL,
    song_id               text NOT NULL,
    seq                   integer NOT NULL CHECK (seq > 0),
    changed_at            timestamptz NOT NULL,
    from_stage            text CHECK (from_stage IS NULL OR from_stage IN (
                              'acquaintance', 'learning', 'technically_stable',
                              'interpretation', 'stage_ready')),
    to_stage              text NOT NULL CHECK (to_stage IN (
                              'acquaintance', 'learning', 'technically_stable',
                              'interpretation', 'stage_ready')),
    note                  text CHECK (note IS NULL OR char_length(note) BETWEEN 1 AND 500),
    changed_by_account_id text NOT NULL,
    PRIMARY KEY (tenant_id, song_id, seq),
    FOREIGN KEY (tenant_id, song_id) REFERENCES student_songs(tenant_id, id),
    FOREIGN KEY (tenant_id, changed_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE TRIGGER student_song_stage_history_immutable
    BEFORE UPDATE OR DELETE ON student_song_stage_history
    FOR EACH ROW EXECUTE FUNCTION reject_student_song_delete();
