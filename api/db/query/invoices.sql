-- name: InvoiceRunGet :one
SELECT id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at,
       requested_by_user_id, requested_by_membership_id, request_id,
       eligible_count, success_count, blocked_count, failed_count, details,
       created_at, updated_at
FROM invoice_runs
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: InvoiceGet :one
SELECT id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
       invoice_number, issued_sequence, generated_run_id, issued_run_id,
       issued_at, issued_by_user_id, issued_by_membership_id, locked_at,
       due_at, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor,
       amount_paid_minor, paid_at, payment_failed_at, payment_status_updated_at,
       adjusts_invoice_id, adjustment_reason_code, adjustment_reason_note,
       period_start_date, period_end_date, calculation_details,
       created_at, updated_at
FROM invoices
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: InvoiceLineListByInvoice :many
SELECT id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order,
       quantity_minutes, unit_amount_minor, line_amount_minor,
       raw_attended_minutes, rounded_attended_minutes, funded_allowance_minutes,
       funded_deduction_minutes, core_billable_minutes, session_count, details,
       created_at, updated_at
FROM invoice_lines
WHERE tenant_id = $1 AND branch_id = $2 AND invoice_id = $3
ORDER BY sort_order;

-- name: InvoiceNumberSequenceGetForUpdate :one
SELECT tenant_id, branch_id, billing_year, billing_month, next_sequence, created_at, updated_at
FROM invoice_number_sequences
WHERE tenant_id = $1 AND branch_id = $2 AND billing_year = $3 AND billing_month = $4
FOR UPDATE;
