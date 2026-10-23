BEGIN;

DELETE TABLE IF EXISTS refresh_tokens;

DELETE TABLE IF EXISTS password_reset_tokens;

COMMIT;