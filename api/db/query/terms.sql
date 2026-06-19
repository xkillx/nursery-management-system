-- name: TermInsert :one
INSERT INTO term (
    id, tenant_id, branch_id, child_id,
    term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
    status, created_by_membership_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10
)
RETURNING id, tenant_id, branch_id, child_id,
          term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
          status, termination_reason_code, termination_reason_note, terminated_at,
          created_at, created_by_membership_id, updated_at;

-- name: TermGetByID :one
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: TermGetActiveForChild :one
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
  AND status = ANY (ARRAY['pre_term', 'active', 'pending_renewal']::text[])
LIMIT 1;

-- name: TermListByChild :many
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY term_start_date DESC, created_at DESC;

-- name: TermListActiveByBranch :many
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2
  AND status = ANY (ARRAY['pre_term', 'active', 'pending_renewal']::text[])
ORDER BY term_start_date ASC, created_at ASC;

-- name: TermListExpiringWithin :many
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2
  AND status = ANY (ARRAY['active', 'pending_renewal']::text[])
  AND term_end_date <= $3
ORDER BY term_end_date ASC, child_id ASC;

-- name: TermListEndingOnOrBefore :many
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2
  AND status = 'active'
  AND term_end_date <= $3
ORDER BY term_end_date ASC, child_id ASC;

-- name: TermListActiveInBillingMonth :many
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2
  AND status = ANY (ARRAY['active', 'pending_renewal']::text[])
  AND term_start_date <= $4
  AND term_end_date >= $3
ORDER BY child_id ASC;

-- name: TermUpdateStatus :execrows
UPDATE term
SET status = $4, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: TermTerminate :execrows
UPDATE term
SET status = 'terminated',
    terminated_at = $4,
    termination_reason_code = $5,
    termination_reason_note = $6,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
  AND status = ANY (ARRAY['pre_term', 'active', 'pending_renewal']::text[]);

-- name: TermListForChildUpdateLock :many
SELECT id, tenant_id, branch_id, child_id,
       term_start_date, term_end_date, booking_pattern_id, site_hourly_rate_minor,
       status, termination_reason_code, termination_reason_note, terminated_at,
       created_at, created_by_membership_id, updated_at
FROM term
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
  AND status = ANY (ARRAY['pre_term', 'active', 'pending_renewal']::text[])
FOR UPDATE;

-- name: ChildSetCurrentTermID :exec
UPDATE children
SET current_term_id = $4, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ChildClearCurrentTermID :exec
UPDATE children
SET current_term_id = NULL, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;
