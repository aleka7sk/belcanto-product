-- Roll back Belcanto P.1 contacts and two-factor authentication. Refuses
-- to run while any account has confirmed TOTP enrollment: dropping the
-- secret would silently disable a second factor people rely on.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM twofa_totp_secrets WHERE confirmed_at IS NOT NULL) THEN
        RAISE EXCEPTION 'cannot roll back two-factor authentication while confirmed enrollments exist';
    END IF;
END
$$;

DROP TABLE IF EXISTS activation_progress;
DROP TABLE IF EXISTS twofa_challenges;
DROP INDEX IF EXISTS twofa_recovery_codes_active_idx;
DROP TABLE IF EXISTS twofa_recovery_codes;
DROP TABLE IF EXISTS twofa_totp_secrets;
DROP INDEX IF EXISTS contact_verifications_open_idx;
DROP TABLE IF EXISTS contact_verifications;
DROP TABLE IF EXISTS verified_contacts;
