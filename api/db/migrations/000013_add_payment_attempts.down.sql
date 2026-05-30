DROP INDEX IF EXISTS idx_payment_attempts_open_attempts;
DROP INDEX IF EXISTS uq_payment_attempts_stripe_session_id;
DROP INDEX IF EXISTS idx_payment_attempts_invoice_created;
DROP TABLE IF EXISTS payment_attempts;
