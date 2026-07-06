-- name: ChildBookingPatternsListByChild :many
SELECT id, tenant_id, branch_id, child_id, effective_from, effective_to, created_at, updated_at, term_time_only
FROM child_booking_patterns
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY effective_from DESC, created_at DESC;

-- name: ChildBookingPatternsGetByID :one
SELECT id, tenant_id, branch_id, child_id, effective_from, effective_to, created_at, updated_at, term_time_only
FROM child_booking_patterns
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ChildBookingPatternsGetActiveForDate :one
SELECT id, tenant_id, branch_id, child_id, effective_from, effective_to, created_at, updated_at, term_time_only
FROM child_booking_patterns
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND effective_from <= $4
  AND (effective_to IS NULL OR effective_to >= $4)
ORDER BY effective_from DESC
LIMIT 1;

-- name: ChildBookingPatternsGetCurrentOpenByChild :one
SELECT id, tenant_id, branch_id, child_id, effective_from, effective_to, created_at, updated_at, term_time_only
FROM child_booking_patterns
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND is_current
LIMIT 1;

-- name: ChildBookingPatternsGetPreviousClosedByChild :one
SELECT id, tenant_id, branch_id, child_id, effective_from, effective_to, created_at, updated_at, term_time_only
FROM child_booking_patterns
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND effective_to IS NOT NULL
ORDER BY effective_to DESC
LIMIT 1;

-- name: ChildBookingPatternsInsert :one
INSERT INTO child_booking_patterns (id, tenant_id, branch_id, child_id, effective_from, effective_to, term_time_only)
VALUES ($1, $2, $3, $4, $5, sqlc.narg('effective_to'), $6)
RETURNING id, tenant_id, branch_id, child_id, effective_from, effective_to, created_at, updated_at, term_time_only;

-- name: ChildBookingPatternsCloseCurrent :exec
UPDATE child_booking_patterns
SET effective_to = $4, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND is_current;

-- name: ChildBookingPatternsCloseByID :exec
UPDATE child_booking_patterns
SET effective_to = $4, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3 AND is_current;

-- name: ChildBookingPatternsUpdateEffectiveFrom :exec
UPDATE child_booking_patterns
SET effective_from = $4, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3 AND is_current;

-- name: ChildBookingPatternsUpdateTermTimeOnly :exec
UPDATE child_booking_patterns
SET term_time_only = $4, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ChildBookingPatternsListByChildPaginated :many
SELECT id, tenant_id, branch_id, child_id, effective_from, effective_to, created_at, updated_at, term_time_only
FROM child_booking_patterns
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY effective_from DESC, created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: ChildBookingPatternsCountByChild :one
SELECT COUNT(*)
FROM child_booking_patterns
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;
