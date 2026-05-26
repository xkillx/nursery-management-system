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
