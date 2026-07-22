-- name: BookingsCreate :exec
INSERT INTO bookings (id, tenant_id, branch_id, child_id, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, term_time_only)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', $11, $12);

-- name: BookingsGetByID :one
SELECT id, tenant_id, branch_id, child_id, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, term_time_only, created_at, updated_at
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: BookingsGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, child_id, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, term_time_only, created_at, updated_at
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: BookingsListByBranchPaginated :many
SELECT id, tenant_id, branch_id, child_id, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, term_time_only, created_at, updated_at
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4 = '' OR status = $4)
  AND ($5 = '' OR funding_type = $5)
  AND ($6::date IS NULL OR effective_start_date >= $6)
  AND ($7::date IS NULL OR (effective_end_date IS NULL OR effective_end_date >= $7))
  AND (NOT $8::bool OR status = 'active')
ORDER BY effective_start_date DESC, created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: BookingsCountByBranch :one
SELECT COUNT(*)
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4 = '' OR status = $4)
  AND ($5 = '' OR funding_type = $5)
  AND ($6::date IS NULL OR effective_start_date >= $6)
  AND ($7::date IS NULL OR (effective_end_date IS NULL OR effective_end_date >= $7))
  AND (NOT $8::bool OR status = 'active');

-- name: BookingsUpdate :exec
UPDATE bookings
SET effective_start_date = $4,
    effective_end_date = $5,
    funding_type = $6,
    funding_hours_per_week = $7,
    la_reference = $8,
    term_time_only = $9,
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: BookingsCancel :exec
UPDATE bookings
SET status = 'cancelled', updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3 AND status = 'active';

-- name: BookingsPause :exec
UPDATE bookings
SET status = 'paused', updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3 AND status = 'active';

-- name: BookingsListByChildAndDateRange :many
SELECT id, tenant_id, branch_id, child_id, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, term_time_only, created_at, updated_at
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND effective_start_date <= @to_date
  AND (effective_end_date IS NULL OR effective_end_date >= @from_date)
  AND status = 'active'
ORDER BY effective_start_date ASC;

-- name: BookingsUnifiedListByBranch :many
SELECT
    'recurring' AS booking_type,
    b.id,
    b.tenant_id,
    b.branch_id,
    b.child_id,
    b.effective_start_date AS start_date,
    b.effective_end_date AS end_date,
    b.status,
    b.created_at,
    b.updated_at,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name
FROM bookings b
JOIN children c ON c.id = b.child_id AND c.tenant_id = b.tenant_id AND c.branch_id = b.branch_id
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND ($3::uuid IS NULL OR b.child_id = $3)
  AND ($4 = '' OR b.status = $4)
  AND ($5 = '' OR b.funding_type = $5)
  AND ($6::date IS NULL OR (b.effective_end_date IS NULL OR b.effective_end_date >= $6))
  AND ($7::date IS NULL OR b.effective_start_date <= $7)
  AND (NOT $8::bool OR b.status = 'active')

UNION ALL

SELECT
    'ad_hoc' AS booking_type,
    ah.id,
    ah.tenant_id,
    ah.branch_id,
    ah.child_id,
    ah.calendar_date AS start_date,
    ah.calendar_date AS end_date,
    ah.status,
    ah.created_at,
    ah.updated_at,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name
FROM ad_hoc_bookings ah
JOIN children c ON c.id = ah.child_id AND c.tenant_id = ah.tenant_id AND c.branch_id = ah.branch_id
WHERE ah.tenant_id = $1
  AND ah.branch_id = $2
  AND ($3::uuid IS NULL OR ah.child_id = $3)
  AND ($6::date IS NULL OR ah.calendar_date >= $6)
  AND ($7::date IS NULL OR ah.calendar_date <= $7)
  AND (NOT $8::bool OR ah.status = 'active')

UNION ALL

SELECT
    'hourly' AS booking_type,
    h.id,
    h.tenant_id,
    h.branch_id,
    h.child_id,
    h.calendar_date AS start_date,
    h.calendar_date AS end_date,
    h.status,
    h.created_at,
    h.updated_at,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name
FROM hourly_bookings h
JOIN children c ON c.id = h.child_id AND c.tenant_id = h.tenant_id AND c.branch_id = h.branch_id
WHERE h.tenant_id = $1
  AND h.branch_id = $2
  AND ($3::uuid IS NULL OR h.child_id = $3)
  AND ($6::date IS NULL OR h.calendar_date >= $6)
  AND ($7::date IS NULL OR h.calendar_date <= $7)
  AND (NOT $8::bool OR h.status = 'active')

ORDER BY start_date DESC, created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: BookingEntriesForChildInMonth :many
SELECT
    (entry->>'day_of_week')::int AS day_of_week,
    st.id AS session_type_id,
    st.name AS session_type_name,
    st.start_time AS session_type_start_time,
    st.end_time AS session_type_end_time
FROM bookings b
CROSS JOIN jsonb_array_elements(b.session_entries) AS entry
JOIN session_types st ON st.id = (entry->>'session_type_id')::uuid
    AND st.tenant_id = b.tenant_id
    AND st.branch_id = b.branch_id
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND b.child_id = $3
  AND b.status = 'active'
  AND b.effective_start_date <= @month_end
  AND (b.effective_end_date IS NULL OR b.effective_end_date >= @month_start)
ORDER BY day_of_week;
