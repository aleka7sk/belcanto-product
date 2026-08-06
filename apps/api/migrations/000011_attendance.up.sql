-- Belcanto L.2/L.4 per-participant attendance (domain/lesson.md Approved
-- 1.0.0: Missed state, audit with previous/new value and reason; Figma
-- TCH-JOURNAL-01/02). One row per Lesson participant who was marked —
-- an empty group seat has no row and is never an absence. A late mark
-- carries minutes; an absence carries a mandatory note. Rows may be
-- corrected after the fact, but every change is audited with its
-- reason at the application layer.

CREATE TABLE core_lesson_attendance (
    tenant_id              text NOT NULL REFERENCES tenants(id),
    occurrence_id          text NOT NULL,
    student_id             text NOT NULL,
    status                 text NOT NULL CHECK (status IN ('present', 'late', 'absent')),
    late_minutes           integer CHECK (late_minutes IS NULL OR late_minutes BETWEEN 1 AND 240),
    note                   text CHECK (note IS NULL OR char_length(note) BETWEEN 1 AND 500),
    recorded_by_account_id text NOT NULL,
    recorded_at            timestamptz NOT NULL,
    updated_at             timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, occurrence_id, student_id),
    CHECK (status <> 'late' OR late_minutes IS NOT NULL),
    CHECK (status = 'late' OR late_minutes IS NULL),
    CHECK (status <> 'absent' OR note IS NOT NULL),
    FOREIGN KEY (tenant_id, occurrence_id) REFERENCES core_lesson_occurrences(tenant_id, id),
    FOREIGN KEY (tenant_id, student_id) REFERENCES students(tenant_id, id),
    FOREIGN KEY (tenant_id, recorded_by_account_id) REFERENCES accounts(tenant_id, id)
);

CREATE INDEX core_lesson_attendance_student_idx
    ON core_lesson_attendance (tenant_id, student_id, recorded_at DESC);

-- Attendance is education history: rows are corrected, never removed.
CREATE FUNCTION reject_attendance_delete() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'attendance history is immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER core_lesson_attendance_no_delete
    BEFORE DELETE ON core_lesson_attendance
    FOR EACH ROW EXECUTE FUNCTION reject_attendance_delete();
