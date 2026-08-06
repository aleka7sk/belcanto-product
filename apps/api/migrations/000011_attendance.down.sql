-- Roll back per-participant attendance. Refuses while any mark exists:
-- who was at a Lesson is education history and never disappears
-- (domain/lesson.md invariants).

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM core_lesson_attendance) THEN
        RAISE EXCEPTION 'cannot roll back attendance while marks exist';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS core_lesson_attendance_no_delete ON core_lesson_attendance;
DROP FUNCTION IF EXISTS reject_attendance_delete();
DROP TABLE IF EXISTS core_lesson_attendance;
