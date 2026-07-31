ALTER TABLE email_outbox ADD COLUMN IF NOT EXISTS attachments JSONB;
ALTER TABLE email_outbox ADD COLUMN IF NOT EXISTS entity_id TEXT;
CREATE INDEX IF NOT EXISTS idx_email_outbox_entity ON email_outbox (event_type, entity_id) WHERE entity_id IS NOT NULL;
