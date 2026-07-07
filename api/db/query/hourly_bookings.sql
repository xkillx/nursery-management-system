-- name: HourlyBookingsCreate :exec
INSERT INTO hourly_bookings (id, tenant_id, branch_id, child_id, calendar_date, start_time_minutes, duration_minutes, session_type_id, booked_by_membership_id, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'active');

-- name: HourlyBookingsListByBranch :many
SELECT id, tenant_id, branch_id, child_id, calendar_date, start_time_minutes, duration_minutes, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM hourly_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4::date IS NULL OR calendar_date >= $4)
  AND ($5::date IS NULL OR calendar_date <= $5)
  AND (NOT $6::bool OR status = 'active')
ORDER BY calendar_date ASC, created_at ASC;

-- name: HourlyBookingsListByBranchPaginated :many
SELECT id, tenant_id, branch_id, child_id, calendar_date, start_time_minutes, duration_minutes, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM hourly_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4::date IS NULL OR calendar_date >= $4)
  AND ($5::date IS NULL OR calendar_date <= $5)
  AND (NOT $6::bool OR status = 'active')
ORDER BY calendar_date ASC, created_at ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: HourlyBookingsCountByBranch :one
SELECT COUNT(*)
FROM hourly_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4::date IS NULL OR calendar_date >= $4)
  AND ($5::date IS NULL OR calendar_date <= $5)
  AND (NOT $6::bool OR status = 'active');

-- name: HourlyBookingsListByChildAndDateRange :many
SELECT id, tenant_id, branch_id, child_id, calendar_date, start_time_minutes, duration_minutes, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM hourly_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND calendar_date >= @from_date
  AND calendar_date <= @to_date
  AND status = 'active'
ORDER BY calendar_date ASC;

-- name: HourlyBookingsGetByID :one
SELECT id, tenant_id, branch_id, child_id, calendar_date, start_time_minutes, duration_minutes, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM hourly_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: HourlyBookingsCancel :exec
UPDATE hourly_bookings
SET status = 'cancelled', updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3 AND status = 'active';
