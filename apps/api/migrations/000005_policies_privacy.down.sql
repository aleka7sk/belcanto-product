-- Roll back Belcanto P.1 policies, privacy and data rights. Refuses to run
-- while open deletion requests exist: dropping the register would silently
-- discard a person's explicit request.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM account_deletion_requests
        WHERE status IN ('requested', 'pending_review')
    ) THEN
        RAISE EXCEPTION 'cannot roll back data rights while open deletion requests exist';
    END IF;
END
$$;

DROP INDEX IF EXISTS account_deletion_requests_open_idx;
DROP TABLE IF EXISTS account_deletion_requests;
DROP INDEX IF EXISTS data_export_requests_open_idx;
DROP TABLE IF EXISTS data_export_requests;
DROP TABLE IF EXISTS privacy_settings;
DROP TABLE IF EXISTS policy_acceptances;
DROP TABLE IF EXISTS policy_versions;
