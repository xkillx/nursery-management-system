-- name: BranchClosureDaysCreate :exec
INSERT INTO branch_closure_days (id, tenant_id, branch_id, date, reason)
VALUES ($1, $2, $3, $4, $5);

-- name: BranchClosureDaysListByBranchAndDateRange :many
SELECT id, tenant_id, branch_id, date, reason, created_at
FROM branch_closure_days
WHERE tenant_id = $1
  AND branch_id = $2
  AND date >= $3
  AND date <= $4
ORDER BY date ASC;

-- name: BranchClosureDaysDelete :execrows
DELETE FROM branch_closure_days
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: BranchClosureDaysDateExists :one
SELECT EXISTS (
    SELECT 1 FROM branch_closure_days
    WHERE tenant_id = $1
      AND branch_id = $2
      AND date = $3
);

-- name: BranchClosureDaysListClosureDatesForMonth :many
SELECT date
FROM branch_closure_days
WHERE tenant_id = $1
  AND branch_id = $2
  AND date >= $3
  AND date <= $4
ORDER BY date ASC;
