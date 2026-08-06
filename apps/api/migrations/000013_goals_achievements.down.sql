-- Roll back goals and achievements. Refuses while any goal or award
-- exists: what a student aimed for and earned is education history.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM student_goals)
       OR EXISTS (SELECT 1 FROM achievement_awards) THEN
        RAISE EXCEPTION 'cannot roll back goals while learning history exists';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS achievement_awards_guard ON achievement_awards;
DROP FUNCTION IF EXISTS guard_achievement_award_mutation();
DROP TABLE IF EXISTS achievement_awards;
DROP TRIGGER IF EXISTS achievement_definitions_guard ON achievement_definitions;
DROP FUNCTION IF EXISTS guard_achievement_definition_mutation();
DROP TABLE IF EXISTS achievement_definitions;
DROP TRIGGER IF EXISTS student_goals_guard ON student_goals;
DROP FUNCTION IF EXISTS guard_student_goal_mutation();
DROP TABLE IF EXISTS student_goals;
