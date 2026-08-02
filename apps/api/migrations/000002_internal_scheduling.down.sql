DROP TABLE IF EXISTS lesson_participants;
DROP TABLE IF EXISTS lessons;
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM teacher_assignments
        WHERE status = 'active'
        GROUP BY tenant_id, student_id
        HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot roll back internal scheduling after temporal Teacher reassignment data exists';
    END IF;
END;
$$;
DROP INDEX IF EXISTS teacher_assignments_effective_lookup_idx;
ALTER TABLE teacher_assignments DROP CONSTRAINT IF EXISTS teacher_assignments_no_active_overlap;
ALTER TABLE teacher_assignments DROP CONSTRAINT IF EXISTS teacher_assignments_effective_interval;
ALTER TABLE teacher_assignments
    DROP COLUMN IF EXISTS version,
    DROP COLUMN IF EXISTS effective_until,
    DROP COLUMN IF EXISTS effective_from;
CREATE UNIQUE INDEX IF NOT EXISTS teacher_assignments_one_active_primary
    ON teacher_assignments (tenant_id, student_id)
    WHERE status = 'active';
