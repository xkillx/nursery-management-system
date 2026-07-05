-- name: AdHocBookingsListByBranch :many
SELECT id, tenant_id, branch_id, child_id, calendar_date, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM ad_hoc_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::uuid IS NULL OR child_id = $3)
  AND ($4::date IS NULL OR calendar_date >= $4)
  AND ($5::date IS NULL OR calendar_date <= $5)
  AND (NOT $6::bool OR status = 'active')
ORDER BY calendar_date ASC, created_at ASC;

-- name: AdHocBookingsListByChildAndDateRange :many
SELECT id, tenant_id, branch_id, child_id, calendar_date, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM ad_hoc_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND calendar_date >= @from_date
  AND calendar_date <= @to_date
  AND status = 'active'
ORDER BY calendar_date ASC;

-- name: AdHocBookingsGetByID :one
SELECT id, tenant_id, branch_id, child_id, calendar_date, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM ad_hoc_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: AdHocBookingsGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, child_id, calendar_date, session_type_id, booked_by_membership_id, status, created_at, updated_at
FROM ad_hoc_bookings
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: AdHocBookingsCreate :exec
INSERT INTO ad_hoc_bookings (id, tenant_id, branch_id, child_id, calendar_date, session_type_id, booked_by_membership_id, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'active');

-- name: AdHocBookingsCancel :exec
UPDATE ad_hoc_bookings
SET status = 'cancelled', updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3 AND status = 'active';
