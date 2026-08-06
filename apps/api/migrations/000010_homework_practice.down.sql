-- Roll back Belcanto L.3 homework and practice. Refuses while any
-- homework, submission or feedback exists: the execution history of a
-- student's work is preserved in full and never disappears
-- (domain/homework.md invariants).

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM homework_assignments)
       OR EXISTS (SELECT 1 FROM practice_submissions)
       OR EXISTS (SELECT 1 FROM practice_feedback) THEN
        RAISE EXCEPTION 'cannot roll back homework while practice history exists';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS practice_feedback_immutable ON practice_feedback;
DROP TABLE IF EXISTS practice_feedback;
DROP TRIGGER IF EXISTS practice_submission_media_immutable ON practice_submission_media;
DROP TABLE IF EXISTS practice_submission_media;
DROP TRIGGER IF EXISTS practice_submissions_immutable ON practice_submissions;
DROP FUNCTION IF EXISTS reject_practice_history_mutation();
DROP TABLE IF EXISTS practice_submissions;
DROP TABLE IF EXISTS homework_attachments;
DROP TRIGGER IF EXISTS homework_tasks_guard ON homework_tasks;
DROP FUNCTION IF EXISTS guard_homework_task_mutation();
DROP TABLE IF EXISTS homework_tasks;
DROP TRIGGER IF EXISTS homework_assignments_guard ON homework_assignments;
DROP FUNCTION IF EXISTS guard_homework_mutation();
DROP TABLE IF EXISTS homework_assignments;
DROP TRIGGER IF EXISTS media_objects_guard ON media_objects;
DROP FUNCTION IF EXISTS guard_media_object_mutation();
DROP TABLE IF EXISTS media_objects;
