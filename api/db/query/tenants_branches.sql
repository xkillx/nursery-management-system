-- name: ListAllTenantBranches :many
SELECT tenant_id, id AS branch_id
FROM branches
WHERE is_active = true
ORDER BY tenant_id, id;

-- name: GetOverdueGraceDays :one
SELECT overdue_grace_days
FROM branches
WHERE tenant_id = $1 AND id = $2;

-- name: UpdateOverdueGraceDays :exec
UPDATE branches
SET overdue_grace_days = $3, updated_at = now()
WHERE tenant_id = $1 AND id = $2;

-- name: GetReminderDaysBefore :one
SELECT reminder_days_before
FROM branches
WHERE tenant_id = $1 AND id = $2;

-- name: UpdateReminderDaysBefore :exec
UPDATE branches
SET reminder_days_before = $3, updated_at = now()
WHERE tenant_id = $1 AND id = $2;
