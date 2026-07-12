-- name: CreatePaymentLink :exec
INSERT INTO payment_links (
    id,
    tenant_id,
    branch_id,
    invoice_id,
    stripe_payment_link_id,
    stripe_payment_link_url,
    amount_minor,
    currency_code,
    created_by_user_id,
    created_by_membership_id,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
);

-- name: GetActivePaymentLinkForInvoice :one
SELECT
    id,
    tenant_id,
    branch_id,
    invoice_id,
    stripe_payment_link_id,
    stripe_payment_link_url,
    amount_minor,
    currency_code,
    created_by_user_id,
    created_by_membership_id,
    status,
    created_at,
    updated_at
FROM payment_links
WHERE tenant_id = $1
  AND branch_id = $2
  AND invoice_id = $3
  AND status = 'active'
LIMIT 1;
