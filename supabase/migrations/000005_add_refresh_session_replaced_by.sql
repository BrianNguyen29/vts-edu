-- 000005_add_refresh_session_replaced_by.sql
-- Adds replaced_by_token_hash to support refresh-token rotation and reuse detection.

ALTER TABLE refresh_sessions
ADD COLUMN IF NOT EXISTS replaced_by_token_hash text;

CREATE INDEX IF NOT EXISTS idx_refresh_sessions_replaced_by
    ON refresh_sessions (replaced_by_token_hash)
    WHERE replaced_by_token_hash IS NOT NULL;
