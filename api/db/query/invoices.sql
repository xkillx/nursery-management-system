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

-- name: PreflightListChildren :many
SELECT
    c.id AS child_id,
    c.full_name,
    c.date_of_birth,
    c.start_date,
    c.end_date,
    c.core_hourly_rate_minor,
    EXISTS (
        SELECT 1
        FROM guardian_child_links gcl
        WHERE gcl.tenant_id = c.tenant_id
          AND gcl.branch_id = c.branch_id
          AND gcl.child_id = c.id
          AND gcl.ended_at IS NULL
    ) AS has_guardian_link,
    fp.id AS funding_profile_id,
    fp.funded_allowance_minutes,
    i.id AS existing_invoice_id,
    i.status AS existing_invoice_status
FROM children c
LEFT JOIN funding_profiles fp
    ON fp.tenant_id = c.tenant_id
    AND fp.branch_id = c.branch_id
    AND fp.child_id = c.id
    AND fp.billing_month = $3
LEFT JOIN invoices i
    ON i.tenant_id = c.tenant_id
    AND i.branch_id = c.branch_id
    AND i.child_id = c.id
    AND i.billing_month = $3
    AND i.invoice_kind = 'monthly'
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.start_date < $4
  AND (c.end_date IS NULL OR c.end_date >= $3)
ORDER BY c.full_name, c.id;

-- name: PreflightListAttendanceSessions :many
SELECT
    id,
    child_id,
    status,
    check_in_at,
    check_out_at,
    check_in_local_date,
    check_out_local_date
FROM attendance_sessions
WHERE tenant_id = $1
  AND branch_id = $2
  AND check_in_local_date >= $3
  AND check_in_local_date < $4
  AND status IN ('open', 'complete', 'corrected')
ORDER BY child_id, check_in_local_date, check_in_at, id;

-- name: ListCandidateChildrenForUpdate :many
SELECT
    c.id AS child_id,
    c.full_name,
    c.date_of_birth,
    c.start_date,
    c.end_date,
    c.core_hourly_rate_minor,
    EXISTS (
        SELECT 1
        FROM guardian_child_links gcl
        WHERE gcl.tenant_id = c.tenant_id
          AND gcl.branch_id = c.branch_id
          AND gcl.child_id = c.id
          AND gcl.ended_at IS NULL
    ) AS has_guardian_link,
    fp.id AS funding_profile_id,
    fp.funded_allowance_minutes,
    i.id AS existing_invoice_id,
    i.status AS existing_invoice_status
FROM children c
LEFT JOIN funding_profiles fp
    ON fp.tenant_id = c.tenant_id
    AND fp.branch_id = c.branch_id
    AND fp.child_id = c.id
    AND fp.billing_month = $3
LEFT JOIN invoices i
    ON i.tenant_id = c.tenant_id
    AND i.branch_id = c.branch_id
    AND i.child_id = c.id
    AND i.billing_month = $3
    AND i.invoice_kind = 'monthly'
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.start_date < $4
  AND (c.end_date IS NULL OR c.end_date >= $3)
ORDER BY c.full_name, c.id
FOR UPDATE OF c;

-- name: ListSelectedChildrenForUpdate :many
SELECT
    c.id AS child_id,
    c.full_name,
    c.date_of_birth,
    c.start_date,
    c.end_date,
    c.core_hourly_rate_minor,
    EXISTS (
        SELECT 1
        FROM guardian_child_links gcl
        WHERE gcl.tenant_id = c.tenant_id
          AND gcl.branch_id = c.branch_id
          AND gcl.child_id = c.id
          AND gcl.ended_at IS NULL
    ) AS has_guardian_link,
    fp.id AS funding_profile_id,
    fp.funded_allowance_minutes,
    i.id AS existing_invoice_id,
    i.status AS existing_invoice_status
FROM children c
LEFT JOIN funding_profiles fp
    ON fp.tenant_id = c.tenant_id
    AND fp.branch_id = c.branch_id
    AND fp.child_id = c.id
LEFT JOIN invoices i
    ON i.tenant_id = c.tenant_id
    AND i.branch_id = c.branch_id
    AND i.child_id = c.id
    AND i.invoice_kind = 'monthly'
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.id = ANY($3::uuid[])
ORDER BY c.full_name, c.id
FOR UPDATE OF c;

-- name: ListAttendanceSessionsForGeneration :many
SELECT
    id,
    child_id,
    status,
    check_in_at,
    check_out_at,
    check_in_local_date,
    check_out_local_date
FROM attendance_sessions
WHERE tenant_id = $1
  AND branch_id = $2
  AND check_in_local_date >= $3
  AND check_in_local_date < $4
  AND status IN ('open', 'complete', 'corrected')
ORDER BY child_id, check_in_local_date, check_in_at, id;

-- name: CreateInvoiceRun :exec
INSERT INTO invoice_runs (
    id, tenant_id, branch_id, billing_month, run_type, status,
    started_at, requested_by_user_id, requested_by_membership_id, request_id
) VALUES (
    $1, $2, $3, $4, $5, $6,
    now(), $7, $8, $9
);

-- name: CompleteInvoiceRun :exec
UPDATE invoice_runs
SET status = $4,
    eligible_count = $5,
    success_count = $6,
    blocked_count = $7,
    details = $8,
    completed_at = now()
WHERE id = $1 AND tenant_id = $2 AND branch_id = $3;

-- name: GetMonthlyInvoiceForUpdate :one
SELECT id, status, invoice_kind, subtotal_minor, funded_deduction_minor, total_due_minor, calculation_details
FROM invoices
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND billing_month = $4
  AND invoice_kind = 'monthly'
FOR UPDATE;

-- name: CreateDraftInvoice :exec
INSERT INTO invoices (
    id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
    currency_code, generated_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
    period_start_date, period_end_date, calculation_details
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12,
    $13, $14, $15
);

-- name: UpdateDraftInvoice :exec
UPDATE invoices
SET generated_run_id = $4,
    subtotal_minor = $5,
    funded_deduction_minor = $6,
    total_due_minor = $7,
    calculation_details = $8,
    updated_at = now()
WHERE id = $1 AND tenant_id = $2 AND branch_id = $3 AND status = 'draft';

-- name: DeleteDraftSystemInvoiceLines :execrows
DELETE FROM invoice_lines
WHERE tenant_id = $1
  AND branch_id = $2
  AND invoice_id = $3
  AND line_kind IN ('core_childcare', 'funded_deduction');

-- name: ListDraftExtraLines :many
SELECT id, line_kind, line_amount_minor, details
FROM invoice_lines
WHERE tenant_id = $1
  AND branch_id = $2
  AND invoice_id = $3
  AND line_kind = 'extra'
ORDER BY sort_order;

-- name: InsertInvoiceLine :exec
INSERT INTO invoice_lines (
    id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order,
    quantity_minutes, unit_amount_minor, line_amount_minor,
    raw_attended_minutes, rounded_attended_minutes, funded_allowance_minutes,
    funded_deduction_minutes, core_billable_minutes, session_count, details
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10,
    $11, $12, $13,
    $14, $15, $16, $17
);

-- name: InvoiceListForManagerReview :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id, c.full_name AS child_name,
    i.billing_month,
    i.period_start_date, i.period_end_date,
    i.currency_code,
    i.subtotal_minor, i.funded_deduction_minor, i.total_due_minor,
    i.amount_paid_minor,
    i.due_at, i.issued_at, i.locked_at,
    i.paid_at, i.payment_failed_at, i.payment_status_updated_at,
    i.adjusts_invoice_id, i.adjustment_reason_code, i.adjustment_reason_note,
    i.generated_run_id,
    gr.status AS generated_run_status,
    gr.started_at AS generated_run_started_at,
    gr.completed_at AS generated_run_completed_at,
    gr.details AS generated_run_details,
    i.calculation_details,
    i.created_at, i.updated_at
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
ORDER BY i.billing_month DESC, c.full_name ASC, i.created_at DESC, i.id ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceGetForManagerReview :one
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id, c.full_name AS child_name,
    i.billing_month,
    i.period_start_date, i.period_end_date,
    i.currency_code,
    i.subtotal_minor, i.funded_deduction_minor, i.total_due_minor,
    i.amount_paid_minor,
    i.due_at, i.issued_at, i.locked_at,
    i.paid_at, i.payment_failed_at, i.payment_status_updated_at,
    i.adjusts_invoice_id, i.adjustment_reason_code, i.adjustment_reason_note,
    i.generated_run_id,
    gr.status AS generated_run_status,
    gr.started_at AS generated_run_started_at,
    gr.completed_at AS generated_run_completed_at,
    gr.details AS generated_run_details,
    i.calculation_details,
    i.created_at, i.updated_at
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2 AND i.id = $3;

-- name: InvoiceLinesForManagerReview :many
SELECT
    id, line_kind, description, sort_order,
    quantity_minutes, unit_amount_minor, line_amount_minor,
    raw_attended_minutes, rounded_attended_minutes,
    funded_allowance_minutes, funded_deduction_minutes, core_billable_minutes,
    session_count
FROM invoice_lines
WHERE tenant_id = $1 AND branch_id = $2 AND invoice_id = $3
ORDER BY sort_order;

-- name: AllocateInvoiceNumberSequence :one
INSERT INTO invoice_number_sequences (
    tenant_id, branch_id, billing_year, billing_month, next_sequence
) VALUES (
    $1, $2, $3, $4, 2
)
ON CONFLICT (tenant_id, branch_id, billing_year, billing_month)
DO UPDATE SET
    next_sequence = invoice_number_sequences.next_sequence + 1,
    updated_at = now()
RETURNING next_sequence - 1 AS issued_sequence;

-- name: GetInvoiceForIssueForUpdate :one
SELECT i.id, i.child_id, c.full_name AS child_name, i.billing_month,
       i.invoice_kind, i.status, i.total_due_minor
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1 AND i.branch_id = $2 AND i.id = $3
FOR UPDATE OF i;

-- name: ListDraftInvoicesForIssueForUpdate :many
SELECT i.id, i.child_id, c.full_name AS child_name, i.billing_month,
       i.invoice_kind, i.status, i.total_due_minor
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.billing_month = $3
  AND i.invoice_kind = 'monthly'
  AND i.status = 'draft'
ORDER BY c.full_name ASC, i.id ASC
FOR UPDATE OF i;

-- name: ListSelectedInvoicesForIssueForUpdate :many
SELECT i.id, i.child_id, c.full_name AS child_name, i.billing_month,
       i.invoice_kind, i.status, i.total_due_minor
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.id = ANY($3::uuid[])
ORDER BY c.full_name ASC, i.id ASC
FOR UPDATE OF i;

-- name: MarkInvoiceIssued :execrows
UPDATE invoices
SET status = 'issued',
    invoice_number = $4,
    issued_sequence = $5,
    issued_run_id = $6,
    issued_at = $7,
    issued_by_user_id = $8,
    issued_by_membership_id = $9,
    locked_at = $7,
    due_at = $7,
    updated_at = now()
WHERE id = $1
  AND tenant_id = $2
  AND branch_id = $3
  AND status = 'draft';

-- name: TryAcquireOverdueTransitionJobLock :one
SELECT pg_try_advisory_xact_lock(200020) AS acquired;

-- name: MarkIssuedInvoicesOverdue :many
UPDATE invoices
SET status = 'overdue',
    payment_status_updated_at = now(),
    updated_at = now()
WHERE status = 'issued'
  AND amount_paid_minor < total_due_minor
  AND due_at < $1
RETURNING id, tenant_id, branch_id;
