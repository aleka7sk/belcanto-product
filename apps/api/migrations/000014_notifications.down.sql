-- Roll back notifications and activity. Refuses while any activity
-- entry exists: what a person was told is delivery history.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM activity_entries) THEN
        RAISE EXCEPTION 'cannot roll back notifications while activity history exists';
    END IF;
END
$$;

DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS activity_entries;
DROP INDEX IF EXISTS outbox_events_pending_idx;
ALTER TABLE outbox_events
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS next_attempt_at,
    DROP COLUMN IF EXISTS status;
