-- name: FundingProfileGet :one
SELECT id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes, created_at, updated_at
FROM funding_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND billing_month = $4;

-- name: FundingProfileGetForUpdate :one
SELECT id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes, created_at, updated_at
FROM funding_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND billing_month = $4
FOR UPDATE;

-- name: FundingProfileCreate :one
INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: FundingProfileUpdateAllowance :one
UPDATE funding_profiles
SET funded_allowance_minutes = $1, updated_at = now()
WHERE tenant_id = $2 AND branch_id = $3 AND child_id = $4 AND billing_month = $5
RETURNING *;

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
  fp.id AS funding_profile_id,
  fp.funded_allowance_minutes,
  fp.updated_at AS funding_updated_at,
  c.profile_photo_path
FROM children c
LEFT JOIN funding_profiles fp
  ON fp.tenant_id = c.tenant_id
  AND fp.branch_id = c.branch_id
  AND fp.child_id = c.id
  AND fp.billing_month = $3
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
  fp.id AS funding_profile_id,
  fp.funded_allowance_minutes,
  fp.updated_at AS funding_updated_at,
  c.profile_photo_path
FROM children c
LEFT JOIN funding_profiles fp
  ON fp.tenant_id = c.tenant_id
  AND fp.branch_id = c.branch_id
  AND fp.child_id = c.id
  AND fp.billing_month = $3
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
