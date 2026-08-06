-- Roll back Belcanto L.3 journals and progress. Refuses while any
-- published version or evidence exists: what a student was told about
-- their own learning is history and never disappears (DEC-006/007).

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM lesson_journal_versions)
       OR EXISTS (SELECT 1 FROM progress_evidence) THEN
        RAISE EXCEPTION 'cannot roll back journals while published learning history exists';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS progress_evidence_immutable ON progress_evidence;
DROP TABLE IF EXISTS progress_evidence;
DROP TRIGGER IF EXISTS lesson_journal_versions_immutable ON lesson_journal_versions;
DROP FUNCTION IF EXISTS reject_journal_version_mutation();
DROP TABLE IF EXISTS lesson_journal_versions;
DROP TABLE IF EXISTS lesson_journals;
