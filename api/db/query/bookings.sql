-- name: BookingsCreate :exec
INSERT INTO bookings (id, tenant_id, branch_id, child_id, session_template_id, room_id, days_of_week, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'active', $14);

-- name: BookingsGetByID :one
SELECT id, tenant_id, branch_id, child_id, session_template_id, room_id, days_of_week, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, created_at, updated_at
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: BookingsGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, child_id, session_template_id, room_id, days_of_week, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, created_at, updated_at
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: BookingsListByBranchPaginated :many
SELECT id, tenant_id, branch_id, child_id, session_template_id, room_id, days_of_week, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, created_at, updated_at
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4::uuid IS NULL OR room_id = $4)
  AND ($5 = '' OR status = $5)
  AND ($6 = '' OR funding_type = $6)
  AND ($7::date IS NULL OR effective_start_date >= $7)
  AND ($8::date IS NULL OR (effective_end_date IS NULL OR effective_end_date >= $8))
  AND (NOT $9::bool OR status = 'active')
ORDER BY effective_start_date DESC, created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: BookingsCountByBranch :one
SELECT COUNT(*)
FROM bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4::uuid IS NULL OR room_id = $4)
  AND ($5 = '' OR status = $5)
  AND ($6 = '' OR funding_type = $6)
  AND ($7::date IS NULL OR effective_start_date >= $7)
  AND ($8::date IS NULL OR (effective_end_date IS NULL OR effective_end_date >= $8))
  AND (NOT $9::bool OR status = 'active');

-- name: BookingsUpdate :exec
UPDATE bookings
SET room_id = $4,
    days_of_week = $5,
    effective_start_date = $6,
    effective_end_date = $7,
    funding_type = $8,
    funding_hours_per_week = $9,
    la_reference = $10,
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
SELECT id, tenant_id, branch_id, child_id, session_template_id, room_id, days_of_week, effective_start_date, effective_end_date, funding_type, funding_hours_per_week, la_reference, session_entries, status, booked_by_membership_id, created_at, updated_at
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
    b.room_id,
    b.session_template_id,
    b.status,
    b.created_at,
    b.updated_at,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name,
    r.name AS room_name
FROM bookings b
JOIN children c ON c.id = b.child_id AND c.tenant_id = b.tenant_id AND c.branch_id = b.branch_id
JOIN rooms r ON r.id = b.room_id AND r.tenant_id = b.tenant_id AND r.branch_id = b.branch_id
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND ($3::uuid IS NULL OR b.child_id = $3)
  AND ($4::uuid IS NULL OR b.room_id = $4)
  AND ($5 = '' OR b.status = $5)
  AND ($6 = '' OR b.funding_type = $6)
  AND ($7::date IS NULL OR (b.effective_end_date IS NULL OR b.effective_end_date >= $7))
  AND ($8::date IS NULL OR b.effective_start_date <= $8)
  AND (NOT $9::bool OR b.status = 'active')

UNION ALL

SELECT
    'ad_hoc' AS booking_type,
    ah.id,
    ah.tenant_id,
    ah.branch_id,
    ah.child_id,
    ah.calendar_date AS start_date,
    ah.calendar_date AS end_date,
    NULL::uuid AS room_id,
    ah.session_type_id AS session_template_id,
    ah.status,
    ah.created_at,
    ah.updated_at,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name,
    NULL::text AS room_name
FROM ad_hoc_bookings ah
JOIN children c ON c.id = ah.child_id AND c.tenant_id = ah.tenant_id AND c.branch_id = ah.branch_id
WHERE ah.tenant_id = $1
  AND ah.branch_id = $2
  AND ($3::uuid IS NULL OR ah.child_id = $3)
  AND ($7::date IS NULL OR ah.calendar_date >= $7)
  AND ($8::date IS NULL OR ah.calendar_date <= $8)
  AND (NOT $9::bool OR ah.status = 'active')

UNION ALL

SELECT
    'hourly' AS booking_type,
    h.id,
    h.tenant_id,
    h.branch_id,
    h.child_id,
    h.calendar_date AS start_date,
    h.calendar_date AS end_date,
    NULL::uuid AS room_id,
    h.session_type_id AS session_template_id,
    h.status,
    h.created_at,
    h.updated_at,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name,
    NULL::text AS room_name
FROM hourly_bookings h
JOIN children c ON c.id = h.child_id AND c.tenant_id = h.tenant_id AND c.branch_id = h.branch_id
WHERE h.tenant_id = $1
  AND h.branch_id = $2
  AND ($3::uuid IS NULL OR h.child_id = $3)
  AND ($7::date IS NULL OR h.calendar_date >= $7)
  AND ($8::date IS NULL OR h.calendar_date <= $8)
  AND (NOT $9::bool OR h.status = 'active')

ORDER BY start_date DESC, created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: BookingsRegisterForDate :many
SELECT
    b.child_id,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name,
    b.room_id,
    r.name AS room_name,
    b.session_template_id,
    st.name AS session_template_name,
    b.effective_start_date,
    b.effective_end_date,
    att.id AS attendance_id,
    att.status AS attendance_status,
    att.check_in_at,
    att.check_out_at
FROM bookings b
JOIN children c ON c.id = b.child_id AND c.tenant_id = b.tenant_id AND c.branch_id = b.branch_id
JOIN rooms r ON r.id = b.room_id AND r.tenant_id = b.tenant_id AND r.branch_id = b.branch_id
JOIN session_templates st ON st.id = b.session_template_id AND st.tenant_id = b.tenant_id AND st.branch_id = b.branch_id
LEFT JOIN attendance_sessions att ON att.child_id = b.child_id
    AND att.tenant_id = b.tenant_id
    AND att.branch_id = b.branch_id
    AND att.check_in_local_date = @register_date
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND b.status = 'active'
  AND b.effective_start_date <= @register_date
  AND (b.effective_end_date IS NULL OR b.effective_end_date >= @register_date)
  AND @register_date_dow = ANY(b.days_of_week)

UNION ALL

SELECT
    ah.child_id,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name,
    NULL::uuid AS room_id,
    NULL::text AS room_name,
    ah.session_type_id AS session_template_id,
    st.name AS session_template_name,
    ah.calendar_date AS effective_start_date,
    ah.calendar_date AS effective_end_date,
    att.id AS attendance_id,
    att.status AS attendance_status,
    att.check_in_at,
    att.check_out_at
FROM ad_hoc_bookings ah
JOIN children c ON c.id = ah.child_id AND c.tenant_id = ah.tenant_id AND c.branch_id = ah.branch_id
JOIN session_types st ON st.id = ah.session_type_id AND st.tenant_id = ah.tenant_id AND st.branch_id = ah.branch_id
LEFT JOIN attendance_sessions att ON att.child_id = ah.child_id
    AND att.tenant_id = ah.tenant_id
    AND att.branch_id = ah.branch_id
    AND att.check_in_local_date = @register_date
WHERE ah.tenant_id = $1
  AND ah.branch_id = $2
  AND ah.calendar_date = @register_date
  AND ah.status = 'active'

UNION ALL

SELECT
    h.child_id,
    c.first_name AS child_first_name,
    c.last_name AS child_last_name,
    NULL::uuid AS room_id,
    NULL::text AS room_name,
    h.session_type_id AS session_template_id,
    st.name AS session_template_name,
    h.calendar_date AS effective_start_date,
    h.calendar_date AS effective_end_date,
    att.id AS attendance_id,
    att.status AS attendance_status,
    att.check_in_at,
    att.check_out_at
FROM hourly_bookings h
JOIN children c ON c.id = h.child_id AND c.tenant_id = h.tenant_id AND c.branch_id = h.branch_id
JOIN session_types st ON st.id = h.session_type_id AND st.tenant_id = h.tenant_id AND st.branch_id = h.branch_id
LEFT JOIN attendance_sessions att ON att.child_id = h.child_id
    AND att.tenant_id = h.tenant_id
    AND att.branch_id = h.branch_id
    AND att.check_in_local_date = @register_date
WHERE h.tenant_id = $1
  AND h.branch_id = $2
  AND h.calendar_date = @register_date
  AND h.status = 'active'

ORDER BY child_last_name, child_first_name;
