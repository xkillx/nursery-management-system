-- name: TermScheduleChangeInsert :one
INSERT INTO term_schedule_change (
    id, tenant_id, branch_id, term_id,
    previous_booking_pattern_id, new_booking_pattern_id, change_kind,
    effective_from, request_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9
)
RETURNING id, tenant_id, branch_id, term_id,
          previous_booking_pattern_id, new_booking_pattern_id, change_kind,
          requested_at, effective_from,
          approved_by_membership_id, approval_decision, rejected_at, request_id;

-- name: TermScheduleChangeGetByID :one
SELECT id, tenant_id, branch_id, term_id,
       previous_booking_pattern_id, new_booking_pattern_id, change_kind,
       requested_at, effective_from,
       approved_by_membership_id, approval_decision, rejected_at, request_id
FROM term_schedule_change
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: TermScheduleChangeListByTerm :many
SELECT id, tenant_id, branch_id, term_id,
       previous_booking_pattern_id, new_booking_pattern_id, change_kind,
       requested_at, effective_from,
       approved_by_membership_id, approval_decision, rejected_at, request_id
FROM term_schedule_change
WHERE tenant_id = $1 AND branch_id = $2 AND term_id = $3
ORDER BY requested_at DESC, created_at IS NULL DESC, id DESC;

-- name: TermScheduleChangeApprove :execrows
UPDATE term_schedule_change
SET approved_by_membership_id = $4,
    approval_decision = 'approved',
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
  AND approval_decision IS NULL;

-- name: TermScheduleChangeReject :execrows
UPDATE term_schedule_change
SET approved_by_membership_id = $4,
    approval_decision = 'rejected',
    rejected_at = now(),
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
  AND approval_decision IS NULL;
