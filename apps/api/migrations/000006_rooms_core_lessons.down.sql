-- Roll back Belcanto L.2 rooms and core lessons. Refuses while attendance
-- has been recorded or occurrences have moved beyond 'scheduled': those
-- rows are teaching history and dropping them would silently erase it.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM attendance_records) THEN
        RAISE EXCEPTION 'cannot roll back core lessons while attendance history exists';
    END IF;
    IF EXISTS (
        SELECT 1 FROM core_lesson_occurrences WHERE status <> 'scheduled'
    ) THEN
        RAISE EXCEPTION 'cannot roll back core lessons while occurrence history exists';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS lesson_participants_read_only ON lesson_participants;
DROP TRIGGER IF EXISTS lessons_read_only ON lessons;
DROP FUNCTION IF EXISTS reject_legacy_lesson_writes();
DROP TRIGGER IF EXISTS core_lesson_occurrence_capacity_guard ON core_lesson_occurrence_participants;
DROP FUNCTION IF EXISTS enforce_core_lesson_occurrence_capacity();
DROP TRIGGER IF EXISTS core_lesson_series_capacity_guard ON core_lesson_series_enrollments;
DROP FUNCTION IF EXISTS enforce_core_lesson_series_capacity();
DROP TABLE IF EXISTS attendance_records;
DROP TABLE IF EXISTS core_lesson_occurrence_participants;
DROP TABLE IF EXISTS core_lesson_occurrences;
DROP TABLE IF EXISTS core_lesson_series_enrollments;
DROP TABLE IF EXISTS core_lesson_series;
DROP TABLE IF EXISTS rooms;
