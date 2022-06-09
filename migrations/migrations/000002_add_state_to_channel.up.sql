BEGIN;
ALTER TABLE channels
ADD COLUMN IF NOT EXISTS channel_state TEXT NOT NULL DEFAULT 'no_state_provided';
COMMIT;
