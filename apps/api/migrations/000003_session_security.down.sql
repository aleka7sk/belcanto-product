-- Roll back Belcanto P.1 session security. Refuses to run while unconsumed
-- password reset tokens exist: dropping them would silently strand users
-- who hold a live recovery link.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM password_reset_tokens
        WHERE consumed_at IS NULL AND superseded_at IS NULL AND expires_at > now()
    ) THEN
        RAISE EXCEPTION 'cannot roll back session security while live password reset tokens exist';
    END IF;
END
$$;

DROP TABLE IF EXISTS server_config;
DROP INDEX IF EXISTS password_reset_tokens_account_open_idx;
DROP TABLE IF EXISTS password_reset_tokens;

ALTER TABLE sessions DROP COLUMN IF EXISTS last_seen_at;
ALTER TABLE sessions DROP COLUMN IF EXISTS platform;
ALTER TABLE sessions DROP COLUMN IF EXISTS device_label;
