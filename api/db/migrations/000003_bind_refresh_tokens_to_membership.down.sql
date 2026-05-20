DROP INDEX IF EXISTS idx_refresh_tokens_membership_id;

ALTER TABLE refresh_tokens
DROP CONSTRAINT IF EXISTS refresh_tokens_membership_id_fkey;

ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS membership_id;
