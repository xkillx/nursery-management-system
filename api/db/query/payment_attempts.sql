-- name: GetParentInvoiceForCheckoutForUpdate :one
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status, i.currency_code,
    i.total_due_minor, i.amount_paid_minor, i.child_id
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
JOIN memberships m
  ON m.tenant_id = i.tenant_id
 AND m.branch_id = i.branch_id
 AND m.id = $3
 AND m.role = 'parent'
 AND m.is_active = true
 AND m.ended_at IS NULL
JOIN parent_membership_guardians pmg
  ON pmg.tenant_id = i.tenant_id
 AND pmg.branch_id = i.branch_id
 AND pmg.membership_id = m.id
 AND pmg.ended_at IS NULL
JOIN guardian_child_links gcl
  ON gcl.tenant_id = i.tenant_id
 AND gcl.branch_id = i.branch_id
 AND gcl.guardian_id = pmg.guardian_id
 AND gcl.child_id = i.child_id
 AND gcl.ended_at IS NULL
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.id = $4
  AND i.status IN ('issued', 'payment_failed', 'overdue')
FOR UPDATE OF i;

-- name: CreatePaymentAttempt :exec
INSERT INTO payment_attempts (
    id, tenant_id, branch_id, invoice_id,
    initiated_by_user_id, initiated_by_membership_id, request_id,
    status, amount_minor, currency_code
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetInvoicePaymentState :one
SELECT invoice_kind, status, currency_code, total_due_minor, amount_paid_minor
FROM invoices
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: MarkPaymentAttemptCheckoutCreated :execrows
UPDATE payment_attempts
SET status = 'checkout_created',
    stripe_checkout_session_id = $4,
    stripe_checkout_url = $5,
    stripe_payment_intent_id = $6,
    stripe_expires_at = $7,
    provider_error_code = NULL,
    provider_error_message = NULL,
    failure_reason = NULL,
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND status = 'checkout_creation_started';

-- name: MarkPaymentAttemptCheckoutCreationFailed :execrows
UPDATE payment_attempts
SET status = 'checkout_creation_failed',
    failure_reason = $4,
    provider_error_code = $5,
    provider_error_message = $6,
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND status = 'checkout_creation_started';
