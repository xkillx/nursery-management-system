-- name: InvoiceRunAdvanceInsert :one
INSERT INTO invoice_run_advance (
    id, tenant_id, branch_id, billing_month,
    generated_invoice_count, skipped_term_count, exception_count,
    triggered_by, request_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, NULLIF($9, '')
)
ON CONFLICT (tenant_id, branch_id, billing_month) DO NOTHING
RETURNING id, tenant_id, branch_id, billing_month,
          generated_at, generated_invoice_count, skipped_term_count, exception_count,
          triggered_by, request_id;

-- name: InvoiceRunAdvanceGetForMonth :one
SELECT id, tenant_id, branch_id, billing_month,
       generated_at, generated_invoice_count, skipped_term_count, exception_count,
       triggered_by, request_id
FROM invoice_run_advance
WHERE tenant_id = $1 AND branch_id = $2 AND billing_month = $3;

-- name: InvoiceRunAdvanceCountForMonth :one
SELECT COUNT(*)::bigint AS count
FROM invoice_run_advance
WHERE tenant_id = $1 AND branch_id = $2 AND billing_month = $3;
