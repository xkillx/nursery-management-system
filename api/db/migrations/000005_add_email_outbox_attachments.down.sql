DROP INDEX IF EXISTS idx_email_outbox_entity;
ALTER TABLE email_outbox DROP COLUMN IF EXISTS entity_id;
ALTER TABLE email_outbox DROP COLUMN IF EXISTS attachments;
