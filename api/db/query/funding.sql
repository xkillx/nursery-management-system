-- name: FundingChildEnrollmentGetForUpdate :one
SELECT id, start_date, end_date
FROM children
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
FOR UPDATE;

-- name: FundingOverviewList :many
SELECT
  c.id AS child_id,
  c.first_name AS child_first_name,
  c.middle_name AS child_middle_name,
  c.last_name AS child_last_name,
  c.is_active,
  c.start_date,
  c.end_date,
  fr.id AS funding_record_id,
  fr.funding_enabled,
  fr.funding_type,
  fr.funding_model,
  fr.funded_hours_per_week,
  fr.funding_start_date,
  fr.funding_end_date,
  fr.updated_at AS funding_updated_at,
  c.profile_photo_path
FROM children c
LEFT JOIN child_funding_records fr
  ON fr.tenant_id = c.tenant_id
  AND fr.branch_id = c.branch_id
  AND fr.child_id = c.id
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.start_date < ($3 + INTERVAL '1 month')::date
  AND (c.end_date IS NULL OR c.end_date >= $3)
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, c.id ASC;

-- name: FundingOverviewListPaginated :many
SELECT
  c.id AS child_id,
  c.first_name AS child_first_name,
  c.middle_name AS child_middle_name,
  c.last_name AS child_last_name,
  c.is_active,
  c.start_date,
  c.end_date,
  fr.id AS funding_record_id,
  fr.funding_enabled,
  fr.funding_type,
  fr.funding_model,
  fr.funded_hours_per_week,
  fr.funding_start_date,
  fr.funding_end_date,
  fr.updated_at AS funding_updated_at,
  c.profile_photo_path
FROM children c
LEFT JOIN child_funding_records fr
  ON fr.tenant_id = c.tenant_id
  AND fr.branch_id = c.branch_id
  AND fr.child_id = c.id
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.start_date < ($3 + INTERVAL '1 month')::date
  AND (c.end_date IS NULL OR c.end_date >= $3)
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, c.id ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: FundingOverviewCount :one
SELECT COUNT(*)
FROM children c
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.start_date < ($3 + INTERVAL '1 month')::date
  AND (c.end_date IS NULL OR c.end_date >= $3);

-- name: ChildFundingHistoryInsert :exec
INSERT INTO child_funding_history (
    id, tenant_id, branch_id, child_id,
    funding_type, funding_model, funded_hours_per_week,
    funding_start_date, funding_end_date,
    changed_at, changed_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: ChildFundingHistoryListByChild :many
SELECT id, tenant_id, branch_id, child_id,
       funding_type, funding_model, funded_hours_per_week,
       funding_start_date, funding_end_date,
       changed_at, changed_by_user_id
FROM child_funding_history
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY changed_at DESC;

-- name: FundingExpiringSoon :many
SELECT
  fr.id AS funding_record_id,
  fr.child_id,
  c.first_name AS child_first_name,
  c.middle_name AS child_middle_name,
  c.last_name AS child_last_name,
  fr.funding_type,
  fr.funded_hours_per_week,
  fr.funding_end_date
FROM child_funding_records fr
JOIN children c ON c.tenant_id = fr.tenant_id AND c.branch_id = fr.branch_id AND c.id = fr.child_id
WHERE fr.tenant_id = $1
  AND fr.branch_id = $2
  AND fr.funding_enabled = true
  AND fr.funding_end_date IS NOT NULL
  AND fr.funding_end_date >= CURRENT_DATE
  AND fr.funding_end_date <= CURRENT_DATE + make_interval(days => $3::int)
ORDER BY fr.funding_end_date ASC;

-- name: FundingFundedChildrenCount :one
SELECT
  COUNT(*) FILTER (WHERE fr.funding_type = 'universal_15') AS fifteen_hour_count,
  COUNT(*) FILTER (WHERE fr.funding_type IN ('working_parent', 'working_parent_under_3')) AS thirty_hour_count,
  COUNT(*) AS total_funded_count
FROM child_funding_records fr
WHERE fr.tenant_id = $1
  AND fr.branch_id = $2
  AND fr.funding_enabled = true
  AND fr.funding_type IS NOT NULL
  AND fr.funding_type != 'none';

-- name: FundingBookedHoursThisWeek :one
SELECT COALESCE(SUM(b.funding_hours_per_week), 0)::numeric AS total_booked_hours
FROM bookings b
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND b.status = 'active'
  AND b.funding_type IS NOT NULL
  AND b.funding_type != 'none'
  AND b.effective_start_date <= CURRENT_DATE
  AND (b.effective_end_date IS NULL OR b.effective_end_date >= CURRENT_DATE);

-- name: FundingExpiringSoonCount :one
SELECT COUNT(*) AS expiring_soon_count
FROM child_funding_records fr
WHERE fr.tenant_id = $1
  AND fr.branch_id = $2
  AND fr.funding_enabled = true
  AND fr.funding_end_date IS NOT NULL
  AND fr.funding_end_date >= CURRENT_DATE
  AND fr.funding_end_date <= CURRENT_DATE + make_interval(days => $3::int);

-- name: FundingChildAllocation :many
SELECT
  b.id AS booking_id,
  b.effective_start_date,
  b.effective_end_date,
  b.days_of_week,
  st.name AS session_type_name,
  EXTRACT(EPOCH FROM (st.end_time - st.start_time))::int / 60 AS session_duration_minutes
FROM bookings b
JOIN session_template_entries ste ON ste.template_id = b.session_template_id
    AND ste.tenant_id = b.tenant_id AND ste.branch_id = b.branch_id
JOIN session_types st ON st.id = ste.session_type_id
    AND st.tenant_id = b.tenant_id AND st.branch_id = b.branch_id
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND b.child_id = $3
  AND b.status = 'active'
  AND b.effective_start_date <= @billing_month_end
  AND (b.effective_end_date IS NULL OR b.effective_end_date >= @billing_month_start)
ORDER BY b.effective_start_date, ste.day_of_week;
