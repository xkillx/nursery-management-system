-- name: InsertWebhookEvent :one
INSERT INTO stripe_webhook_events (
    id, stripe_event_id, event_type, livemode, api_version,
    provider_created_at, processing_status, processing_reason,
    request_id, raw_payload
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (stripe_event_id) DO NOTHING
RETURNING id;

-- name: UpdateWebhookEventStatus :execrows
UPDATE stripe_webhook_events
SET processing_status = $2,
    processing_reason = $3,
    processed_at = now(),
    error_message = $4,
    updated_at = now()
WHERE id = $1;

-- name: GetWebhookEventByStripeEventID :one
SELECT id, stripe_event_id, event_type, processing_status, processing_reason
FROM stripe_webhook_events
WHERE stripe_event_id = $1;

-- name: GetPaymentAttemptAndInvoiceForWebhook :one
SELECT
    pa.id AS attempt_id,
    pa.status AS attempt_status,
    pa.amount_minor AS attempt_amount_minor,
    pa.currency_code AS attempt_currency_code,
    pa.stripe_checkout_session_id AS attempt_session_id,
    i.id AS invoice_id,
    i.status AS invoice_status,
    i.total_due_minor AS invoice_total_due_minor,
    i.amount_paid_minor AS invoice_amount_paid_minor,
    i.currency_code AS invoice_currency_code,
    i.paid_at AS invoice_paid_at,
    i.payment_failed_at AS invoice_payment_failed_at
FROM payment_attempts pa
JOIN invoices i ON i.tenant_id = pa.tenant_id AND i.branch_id = pa.branch_id AND i.id = pa.invoice_id
WHERE pa.tenant_id = $1
  AND pa.branch_id = $2
  AND pa.invoice_id = $3
  AND pa.id = $4
  AND pa.stripe_checkout_session_id = $5
FOR UPDATE OF pa, i;

-- name: MarkPaymentAttemptPaid :execrows
UPDATE payment_attempts
SET status = 'paid',
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND status IN ('checkout_created', 'payment_failed');

-- name: MarkPaymentAttemptFailed :execrows
UPDATE payment_attempts
SET status = 'payment_failed',
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND status IN ('checkout_created');

-- name: MarkPaymentAttemptExpired :execrows
UPDATE payment_attempts
SET status = 'expired',
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND status IN ('checkout_created');

-- name: MarkInvoicePaid :execrows
UPDATE invoices
SET status = 'paid',
    amount_paid_minor = total_due_minor,
    paid_at = COALESCE(paid_at, now()),
    payment_status_updated_at = now(),
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND status IN ('issued', 'overdue', 'payment_failed');

-- name: MarkInvoicePaymentFailed :execrows
UPDATE invoices
SET status = 'payment_failed',
    payment_failed_at = COALESCE(payment_failed_at, now()),
    payment_status_updated_at = now(),
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND status IN ('issued', 'overdue');

-- name: InsertReconciliationRecord :exec
INSERT INTO payment_reconciliation_records (
    id, tenant_id, branch_id, invoice_id, payment_attempt_id,
    stripe_webhook_event_id, stripe_event_id, stripe_event_type,
    stripe_checkout_session_id, stripe_payment_intent_id,
    outcome, reason_code,
    previous_invoice_status, new_invoice_status,
    attempt_previous_status, attempt_new_status,
    amount_minor, currency_code, details
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19);
