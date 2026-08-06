-- Roll back student repertoire. Refuses while any song exists: the path
-- of a piece is education history and never disappears.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM student_songs) THEN
        RAISE EXCEPTION 'cannot roll back repertoire while songs exist';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS student_song_stage_history_immutable ON student_song_stage_history;
DROP TABLE IF EXISTS student_song_stage_history;
DROP TRIGGER IF EXISTS student_songs_no_delete ON student_songs;
DROP FUNCTION IF EXISTS reject_student_song_delete();
DROP TABLE IF EXISTS student_songs;
