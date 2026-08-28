BEGIN;

ALTER TABLE user_sessions
    DROP CONSTRAINT IF EXISTS user_sessions_phone_key;

COMMIT;
