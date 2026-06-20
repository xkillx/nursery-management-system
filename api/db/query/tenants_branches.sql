-- name: ListAllTenantBranches :many
SELECT tenant_id, id AS branch_id
FROM branches
WHERE is_active = true
ORDER BY tenant_id, id;
