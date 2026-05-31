DROP INDEX IF EXISTS idx_reconciliation_attempt_created;
DROP INDEX IF EXISTS idx_reconciliation_invoice_created;
DROP INDEX IF EXISTS uq_reconciliation_stripe_event_id;
DROP TABLE IF EXISTS payment_reconciliation_records;
DROP INDEX IF EXISTS idx_stripe_webhook_events_processing_status;
DROP INDEX IF EXISTS idx_stripe_webhook_events_event_type;
DROP TABLE IF EXISTS stripe_webhook_events;
