BEGIN;
ALTER TABLE channels DROP COLUMN channel_state;
COMMIT;
