-- name: GetManagerInvoicePaymentStatus :one
SELECT
    i.id AS invoice_id,
    i.invoice_kind,
    i.invoice_number,
    COALESCE(i.invoice_number, '') AS invoice_number_display,
    i.child_id,
    c.full_name AS child_name,
    i.billing_month,
    i.status,
    i.currency_code,
    i.total_due_minor,
    i.amount_paid_minor,
    i.issued_at,
    i.due_at,
    i.paid_at,
    i.payment_failed_at,
    i.payment_status_updated_at,
    i.created_at,
    i.updated_at
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.id = $3;

-- name: GetLatestPaymentAttemptForInvoice :one
SELECT
    pa.id AS payment_attempt_id,
    pa.status,
    pa.amount_minor,
    pa.currency_code,
    pa.stripe_checkout_session_id,
    pa.stripe_payment_intent_id,
    pa.stripe_expires_at,
    pa.failure_reason,
    pa.provider_error_code,
    pa.provider_error_message,
    pa.created_at,
    pa.updated_at
FROM payment_attempts pa
WHERE pa.tenant_id = $1
  AND pa.branch_id = $2
  AND pa.invoice_id = $3
ORDER BY pa.created_at DESC, pa.id DESC;

-- name: GetLatestPaymentEventForInvoice :one
SELECT
    r.id AS payment_event_id,
    r.payment_attempt_id,
    r.stripe_event_id,
    r.stripe_event_type,
    r.stripe_checkout_session_id,
    r.stripe_payment_intent_id,
    r.outcome,
    r.reason_code,
    r.previous_invoice_status,
    r.new_invoice_status,
    r.attempt_previous_status,
    r.attempt_new_status,
    r.amount_minor,
    r.currency_code,
    e.processing_status AS webhook_processing_status,
    e.processing_reason AS webhook_processing_reason,
    e.received_at AS webhook_received_at,
    e.processed_at AS webhook_processed_at,
    r.created_at
FROM payment_reconciliation_records r
JOIN stripe_webhook_events e ON e.id = r.stripe_webhook_event_id
WHERE r.tenant_id = $1
  AND r.branch_id = $2
  AND r.invoice_id = $3
ORDER BY r.created_at DESC, r.id DESC;

-- name: ListPaymentEventsForInvoice :many
SELECT
    r.id AS payment_event_id,
    r.payment_attempt_id,
    r.stripe_event_id,
    r.stripe_event_type,
    r.stripe_checkout_session_id,
    r.stripe_payment_intent_id,
    r.outcome,
    r.reason_code,
    r.previous_invoice_status,
    r.new_invoice_status,
    r.attempt_previous_status,
    r.attempt_new_status,
    r.amount_minor,
    r.currency_code,
    e.processing_status AS webhook_processing_status,
    e.processing_reason AS webhook_processing_reason,
    e.received_at AS webhook_received_at,
    e.processed_at AS webhook_processed_at,
    r.created_at
FROM payment_reconciliation_records r
JOIN stripe_webhook_events e ON e.id = r.stripe_webhook_event_id
WHERE r.tenant_id = $1
  AND r.branch_id = $2
  AND r.invoice_id = $3
ORDER BY r.created_at DESC, r.id DESC
LIMIT $4 OFFSET $5;
