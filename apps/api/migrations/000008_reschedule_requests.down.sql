-- Roll back Belcanto L.2 reschedule requests. Refuses while any decided
-- request exists: decisions about real people's lessons are history.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM lesson_reschedule_requests WHERE status IN ('approved', 'declined')
    ) THEN
        RAISE EXCEPTION 'cannot roll back reschedule requests while decision history exists';
    END IF;
END
$$;

DROP TABLE IF EXISTS lesson_reschedule_requests;
