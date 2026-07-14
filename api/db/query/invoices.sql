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
    c.first_name,
    c.middle_name,
    c.last_name,
    c.date_of_birth,
    c.start_date,
    c.end_date,
    b.core_hourly_rate_minor,
    EXISTS (
        SELECT 1
        FROM child_contacts cc
        WHERE cc.tenant_id = c.tenant_id
          AND cc.branch_id = c.branch_id
          AND cc.child_id = c.id
          AND cc.contact_type = 'parent_carer'
    ) AS has_parent_carer_contact,
    fp.id AS funding_profile_id,
    fp.funded_allowance_minutes,
    i.id AS existing_invoice_id,
    i.status AS existing_invoice_status
FROM children c
JOIN branches b ON b.tenant_id = c.tenant_id AND b.id = c.branch_id
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
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, c.id ASC;

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
    c.first_name,
    c.middle_name,
    c.last_name,
    c.date_of_birth,
    c.start_date,
    c.end_date,
    b.core_hourly_rate_minor,
    EXISTS (
        SELECT 1
        FROM child_contacts cc
        WHERE cc.tenant_id = c.tenant_id
          AND cc.branch_id = c.branch_id
          AND cc.child_id = c.id
          AND cc.contact_type = 'parent_carer'
    ) AS has_parent_carer_contact,
    fp.id AS funding_profile_id,
    fp.funded_allowance_minutes,
    i.id AS existing_invoice_id,
    i.status AS existing_invoice_status
FROM children c
JOIN branches b ON b.tenant_id = c.tenant_id AND b.id = c.branch_id
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
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, c.id ASC
FOR UPDATE OF c;

-- name: ListSelectedChildrenForUpdate :many
SELECT
    c.id AS child_id,
    c.first_name,
    c.middle_name,
    c.last_name,
    c.date_of_birth,
    c.start_date,
    c.end_date,
    b.core_hourly_rate_minor,
    EXISTS (
        SELECT 1
        FROM child_contacts cc
        WHERE cc.tenant_id = c.tenant_id
          AND cc.branch_id = c.branch_id
          AND cc.child_id = c.id
          AND cc.contact_type = 'parent_carer'
    ) AS has_parent_carer_contact,
    fp.id AS funding_profile_id,
    fp.funded_allowance_minutes,
    i.id AS existing_invoice_id,
    i.status AS existing_invoice_status
FROM children c
JOIN branches b ON b.tenant_id = c.tenant_id AND b.id = c.branch_id
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
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, c.id ASC
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
  AND line_kind IN ('core_childcare', 'funded_deduction', 'hourly');

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

-- name: InvoiceLineGet :one
SELECT id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order,
       quantity_minutes, unit_amount_minor, line_amount_minor,
       raw_attended_minutes, rounded_attended_minutes, funded_allowance_minutes,
       funded_deduction_minutes, core_billable_minutes, session_count, details,
       created_at, updated_at
FROM invoice_lines
WHERE tenant_id = $1 AND branch_id = $2 AND invoice_id = $3 AND id = $4;

-- name: InvoiceLineUpdate :execrows
UPDATE invoice_lines
SET description = $4,
    quantity_minutes = $5,
    unit_amount_minor = $6,
    line_amount_minor = $7,
    updated_at = now()
WHERE id = $1
  AND tenant_id = $2
  AND branch_id = $3
  AND line_kind IN ('extra', 'ad_hoc');

-- name: InvoiceLineDelete :execrows
DELETE FROM invoice_lines
WHERE id = $1
  AND tenant_id = $2
  AND branch_id = $3
  AND line_kind IN ('extra', 'ad_hoc');

-- name: InvoiceListForManagerReview :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.billing_month DESC, c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, i.created_at DESC, i.id ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceCountForManagerReview :one
SELECT COUNT(*)
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%');

-- name: InvoiceListForManagerReviewSortByBillingMonthAsc :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.billing_month ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceListForManagerReviewSortByDueAtAsc :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.due_at ASC NULLS LAST
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceListForManagerReviewSortByDueAtDesc :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.due_at DESC NULLS LAST
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceListForManagerReviewSortByTotalAmountAsc :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.total_due_minor ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceListForManagerReviewSortByTotalAmountDesc :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.total_due_minor DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceGetForManagerReview :one
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    r.name AS room_name,
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
LEFT JOIN child_room_assignments cra ON cra.child_id = i.child_id AND cra.tenant_id = i.tenant_id AND cra.branch_id = i.branch_id AND cra.is_current
LEFT JOIN rooms r ON r.id = cra.room_id AND r.tenant_id = i.tenant_id AND r.branch_id = i.branch_id
WHERE i.tenant_id = $1 AND i.branch_id = $2 AND i.id = $3;

-- name: InvoiceLinesForManagerReview :many
SELECT
    id, line_kind, description, sort_order,
    quantity_minutes, unit_amount_minor, line_amount_minor,
    raw_attended_minutes, rounded_attended_minutes,
    funded_allowance_minutes, funded_deduction_minutes, core_billable_minutes,
    session_count, details
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
SELECT i.id, i.child_id,
       c.first_name AS child_first_name,
       c.middle_name AS child_middle_name,
       c.last_name AS child_last_name,
       i.billing_month,
       i.invoice_kind, i.status, i.total_due_minor
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1 AND i.branch_id = $2 AND i.id = $3
FOR UPDATE OF i;

-- name: ListDraftInvoicesForIssueForUpdate :many
SELECT i.id, i.child_id,
       c.first_name AS child_first_name,
       c.middle_name AS child_middle_name,
       c.last_name AS child_last_name,
       i.billing_month,
       i.invoice_kind, i.status, i.total_due_minor
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.billing_month = $3
  AND i.invoice_kind = 'monthly'
  AND i.status = 'draft'
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, i.id ASC
FOR UPDATE OF i;

-- name: ListSelectedInvoicesForIssueForUpdate :many
SELECT i.id, i.child_id,
       c.first_name AS child_first_name,
       c.middle_name AS child_middle_name,
       c.last_name AS child_last_name,
       i.billing_month,
       i.invoice_kind, i.status, i.total_due_minor
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.id = ANY($3::uuid[])
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, i.id ASC
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
    due_at = $10,
    updated_at = now()
WHERE id = $1
  AND tenant_id = $2
  AND branch_id = $3
  AND status = 'draft';

-- name: MarkInvoiceVoid :execrows
UPDATE invoices
SET status = 'void',
    voided_at = $4,
    void_reason = $5,
    updated_at = now()
WHERE id = $1
  AND tenant_id = $2
  AND branch_id = $3
  AND status = 'draft';

-- name: TryAcquireOverdueTransitionJobLock :one
SELECT pg_try_advisory_xact_lock(200020) AS acquired;

-- name: TryAcquireReminderJobLock :one
SELECT pg_try_advisory_xact_lock(200021) AS acquired;

-- name: MarkIssuedInvoicesOverdue :many
UPDATE invoices
SET status = 'overdue',
    payment_status_updated_at = now(),
    updated_at = now()
FROM branches b
WHERE invoices.branch_id = b.id
  AND invoices.tenant_id = b.tenant_id
  AND invoices.status = 'issued'
  AND invoices.amount_paid_minor < invoices.total_due_minor
  AND invoices.due_at + (b.overdue_grace_days || ' days')::interval < $1
RETURNING invoices.id, invoices.tenant_id, invoices.branch_id;

-- name: ListInvoicesDueSoon :many
SELECT
    i.id,
    i.tenant_id,
    i.branch_id,
    i.due_at
FROM invoices i
JOIN branches b ON b.id = i.branch_id AND b.tenant_id = i.tenant_id
WHERE i.status = 'issued'
  AND i.amount_paid_minor < i.total_due_minor
  AND i.due_at::date = (now() + (b.reminder_days_before || ' days')::interval)::date
  AND NOT EXISTS (
      SELECT 1 FROM invoice_reminder_log l
      WHERE l.invoice_id = i.id
        AND l.reminder_type = 'due_soon'
        AND l.sent_at_date = CURRENT_DATE
  );

-- name: ListInvoicesDueToday :many
SELECT
    i.id,
    i.tenant_id,
    i.branch_id,
    i.due_at
FROM invoices i
WHERE i.status = 'issued'
  AND i.amount_paid_minor < i.total_due_minor
  AND i.due_at::date = CURRENT_DATE
  AND NOT EXISTS (
      SELECT 1 FROM invoice_reminder_log l
      WHERE l.invoice_id = i.id
        AND l.reminder_type = 'due_today'
        AND l.sent_at_date = CURRENT_DATE
  );

-- name: InsertInvoiceReminderLog :exec
INSERT INTO invoice_reminder_log (tenant_id, branch_id, invoice_id, reminder_type, sent_at_date)
VALUES ($1, $2, $3, $4, CURRENT_DATE);

-- name: BillingListAdHocBookingsForMonth :many
SELECT
    ab.id,
    ab.child_id,
    ab.calendar_date,
    ab.session_type_id,
    st.name AS session_type_name,
    st.start_time AS session_type_start_time,
    st.end_time AS session_type_end_time,
    st.flat_fee_minor AS session_type_flat_fee_minor
FROM ad_hoc_bookings ab
JOIN session_types st
  ON st.tenant_id = ab.tenant_id
 AND st.branch_id = ab.branch_id
 AND st.id = ab.session_type_id
WHERE ab.tenant_id = $1
  AND ab.branch_id = $2
  AND ab.child_id = $3
  AND ab.calendar_date >= $4
  AND ab.calendar_date <= $5
  AND ab.status = 'active'
ORDER BY ab.calendar_date ASC, ab.created_at ASC;

-- Advance-pay billing: list active terms covering the billing month, joined with child
-- and branch data so the application layer can drive invoice generation off a single query.
-- name: BillingListActiveTermsForGeneration :many
SELECT
    t.id            AS term_id,
    t.tenant_id,
    t.branch_id,
    t.child_id,
    t.term_start_date,
    t.term_end_date,
    t.booking_pattern_id,
    t.site_hourly_rate_minor,
    t.status,
    c.first_name,
    c.middle_name,
    c.last_name,
    c.date_of_birth,
    c.start_date,
    c.end_date,
    EXISTS (
        SELECT 1
        FROM child_contacts cc
        WHERE cc.tenant_id = t.tenant_id
          AND cc.branch_id = t.branch_id
          AND cc.child_id = t.child_id
          AND cc.contact_type = 'parent_carer'
    ) AS has_parent_carer_contact,
    fp.id AS funding_profile_id,
    fp.funded_allowance_minutes,
    bp.term_time_only,
    COALESCE(fr.funding_model, 'unknown') AS funding_model,
    fr.funded_hours_per_week,
    b.ad_hoc_rate_multiplier
FROM term t
JOIN children c
  ON c.tenant_id = t.tenant_id
 AND c.branch_id = t.branch_id
 AND c.id = t.child_id
JOIN child_booking_patterns bp
  ON bp.tenant_id = t.tenant_id
 AND bp.branch_id = t.branch_id
 AND bp.id = t.booking_pattern_id
LEFT JOIN funding_profiles fp
  ON fp.tenant_id = t.tenant_id
 AND fp.branch_id = t.branch_id
 AND fp.child_id = t.child_id
 AND fp.billing_month = $3
LEFT JOIN child_funding_records fr
  ON fr.tenant_id = t.tenant_id
 AND fr.branch_id = t.branch_id
 AND fr.child_id = t.child_id
JOIN branches b
  ON b.tenant_id = t.tenant_id
 AND b.id = t.branch_id
WHERE t.tenant_id = $1
  AND t.branch_id = $2
  AND t.status = ANY (ARRAY['active', 'pending_renewal']::text[])
  AND t.term_start_date <= $4
  AND t.term_end_date >= $3
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, t.id ASC
FOR UPDATE OF t;

-- name: InvoiceListForParent :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
    i.billing_month,
    i.period_start_date, i.period_end_date,
    i.currency_code,
    i.subtotal_minor, i.funded_deduction_minor, i.total_due_minor,
    i.amount_paid_minor,
    i.due_at, i.issued_at,
    i.paid_at, i.payment_failed_at, i.payment_status_updated_at,
    i.calculation_details
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
JOIN memberships m
  ON m.tenant_id = i.tenant_id
 AND m.branch_id = i.branch_id
 AND m.id = $3
 AND m.role = 'parent'
 AND m.is_active = true
 AND m.ended_at IS NULL
JOIN parent_membership_children pmc
  ON pmc.tenant_id = i.tenant_id
 AND pmc.branch_id = i.branch_id
 AND pmc.membership_id = m.id
 AND pmc.child_id = i.child_id
 AND pmc.ended_at IS NULL
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.status IN ('issued', 'payment_failed', 'paid', 'overdue')
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
ORDER BY
  CASE i.status
    WHEN 'overdue' THEN 1
    WHEN 'payment_failed' THEN 2
    WHEN 'issued' THEN 3
    WHEN 'paid' THEN 4
    ELSE 5
  END,
  i.due_at ASC NULLS LAST,
  i.billing_month DESC,
  c.first_name ASC,
  c.middle_name ASC NULLS FIRST,
  c.last_name ASC NULLS FIRST,
  i.id ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InvoiceCountForParent :one
SELECT COUNT(*)
FROM invoices i
JOIN memberships m
  ON m.tenant_id = i.tenant_id
 AND m.branch_id = i.branch_id
 AND m.id = $3
 AND m.role = 'parent'
 AND m.is_active = true
 AND m.ended_at IS NULL
JOIN parent_membership_children pmc
  ON pmc.tenant_id = i.tenant_id
 AND pmc.branch_id = i.branch_id
 AND pmc.membership_id = m.id
 AND pmc.child_id = i.child_id
 AND pmc.ended_at IS NULL
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.status IN ('issued', 'payment_failed', 'paid', 'overdue')
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid);

-- name: InvoiceGetForParent :one
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
    i.billing_month,
    i.period_start_date, i.period_end_date,
    i.currency_code,
    i.subtotal_minor, i.funded_deduction_minor, i.total_due_minor,
    i.amount_paid_minor,
    i.due_at, i.issued_at,
    i.paid_at, i.payment_failed_at, i.payment_status_updated_at,
    i.calculation_details
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
JOIN memberships m
  ON m.tenant_id = i.tenant_id
 AND m.branch_id = i.branch_id
 AND m.id = $3
 AND m.role = 'parent'
 AND m.is_active = true
 AND m.ended_at IS NULL
JOIN parent_membership_children pmc
  ON pmc.tenant_id = i.tenant_id
 AND pmc.branch_id = i.branch_id
 AND pmc.membership_id = m.id
 AND pmc.child_id = i.child_id
 AND pmc.ended_at IS NULL
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.id = $4
  AND i.status IN ('issued', 'payment_failed', 'paid', 'overdue');

-- name: InvoiceLinesForParent :many
SELECT
    il.line_kind, il.description, il.sort_order,
    il.quantity_minutes, il.unit_amount_minor, il.line_amount_minor, il.details
FROM invoice_lines il
JOIN invoices i ON i.tenant_id = il.tenant_id AND i.branch_id = il.branch_id AND i.id = il.invoice_id
JOIN memberships m
  ON m.tenant_id = i.tenant_id
 AND m.branch_id = i.branch_id
 AND m.id = $3
 AND m.role = 'parent'
 AND m.is_active = true
 AND m.ended_at IS NULL
JOIN parent_membership_children pmc
  ON pmc.tenant_id = i.tenant_id
 AND pmc.branch_id = i.branch_id
 AND pmc.membership_id = m.id
 AND pmc.child_id = i.child_id
 AND pmc.ended_at IS NULL
WHERE il.tenant_id = $1
  AND il.branch_id = $2
  AND il.invoice_id = $4
  AND i.status IN ('issued', 'payment_failed', 'paid', 'overdue')
ORDER BY il.sort_order;

-- name: InvoiceExportForManagerReview :many
SELECT
    i.id, i.invoice_kind, i.invoice_number, i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.middle_name AS child_middle_name,
    c.last_name AS child_last_name,
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
    i.created_at, i.updated_at,
    c.profile_photo_path AS child_profile_photo_path
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
LEFT JOIN invoice_runs gr ON gr.tenant_id = i.tenant_id AND gr.branch_id = i.branch_id AND gr.id = i.generated_run_id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.billing_month DESC, c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, i.created_at DESC, i.id ASC;

-- name: InvoiceExportDetailForManagerReview :many
SELECT
    i.invoice_number,
    i.status,
    i.child_id,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name,
    i.billing_month,
    il.line_kind,
    il.description,
    il.quantity_minutes,
    il.unit_amount_minor,
    il.line_amount_minor
FROM invoices i
JOIN children c ON c.tenant_id = i.tenant_id AND c.branch_id = i.branch_id AND c.id = i.child_id
JOIN invoice_lines il ON il.tenant_id = i.tenant_id AND il.branch_id = i.branch_id AND il.invoice_id = i.id
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month')::date IS NULL OR i.billing_month = sqlc.narg('billing_month')::date)
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
  AND (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('child_id')::uuid IS NULL OR i.child_id = sqlc.narg('child_id')::uuid)
  AND (sqlc.narg('search')::text IS NULL OR i.invoice_number ILIKE '%' || sqlc.narg('search')::text || '%' OR (c.first_name || ' ' || c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY i.billing_month DESC, c.first_name ASC, c.last_name ASC, il.sort_order ASC;

-- name: InvoiceSummaryByMonth :many
SELECT
    i.billing_month,
    COALESCE(SUM(i.total_due_minor), 0)::integer AS total_invoiced_minor,
    COALESCE(SUM(CASE WHEN i.status = 'paid' THEN i.amount_paid_minor ELSE 0 END), 0)::integer AS total_collected_minor,
    COALESCE(SUM(CASE WHEN i.status IN ('issued', 'overdue', 'payment_failed') THEN i.total_due_minor - i.amount_paid_minor ELSE 0 END), 0)::integer AS total_outstanding_minor,
    COALESCE(SUM(CASE WHEN i.status = 'overdue' THEN i.total_due_minor - i.amount_paid_minor ELSE 0 END), 0)::integer AS total_overdue_minor,
    COUNT(*)::integer AS invoice_count
FROM invoices i
WHERE i.tenant_id = $1 AND i.branch_id = $2
  AND (sqlc.narg('billing_month_from')::date IS NULL OR i.billing_month >= sqlc.narg('billing_month_from')::date)
  AND (sqlc.narg('billing_month_to')::date IS NULL OR i.billing_month <= sqlc.narg('billing_month_to')::date)
GROUP BY i.billing_month
ORDER BY i.billing_month DESC;
