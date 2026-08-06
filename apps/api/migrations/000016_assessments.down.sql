-- Roll back assessments. Refuses while any assessment exists: the
-- educational history is immutable.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM assessments) THEN
        RAISE EXCEPTION 'cannot roll back assessments while history exists';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS assessments_guard_update ON assessments;
DROP FUNCTION IF EXISTS guard_assessment_update();
DROP TRIGGER IF EXISTS assessment_evidence_guard_delete ON assessment_evidence;
DROP FUNCTION IF EXISTS guard_assessment_evidence_delete();
DROP TABLE IF EXISTS assessment_evidence;
DROP TRIGGER IF EXISTS assessments_no_delete ON assessments;
DROP FUNCTION IF EXISTS reject_assessment_delete();
DROP TABLE IF EXISTS assessments;
