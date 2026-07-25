-- name: HolidayPeriodsCreate :exec
INSERT INTO holiday_periods (id, tenant_id, branch_id, name, start_date, end_date, type)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: HolidayPeriodsListByBranch :many
SELECT id, tenant_id, branch_id, name, start_date, end_date, type, created_at, updated_at
FROM holiday_periods
WHERE tenant_id = $1
  AND branch_id = $2
ORDER BY start_date DESC, name ASC;

-- name: HolidayPeriodsGetByID :one
SELECT id, tenant_id, branch_id, name, start_date, end_date, type, created_at, updated_at
FROM holiday_periods
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: HolidayPeriodsUpdate :execrows
UPDATE holiday_periods
SET
    name = CASE WHEN @set_name::bool THEN @name ELSE name END,
    start_date = CASE WHEN @set_start_date::bool THEN @start_date ELSE start_date END,
    end_date = CASE WHEN @set_end_date::bool THEN @end_date ELSE end_date END,
    type = CASE WHEN @set_type::bool THEN @type ELSE type END,
    updated_at = now()
WHERE tenant_id = @tenant_id AND branch_id = @branch_id AND id = @id;

-- name: HolidayPeriodsDelete :execrows
DELETE FROM holiday_periods
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: HolidayPeriodsListForBranchAndMonth :many
SELECT id, tenant_id, branch_id, name, start_date, end_date, type, created_at, updated_at
FROM holiday_periods
WHERE tenant_id = $1
  AND branch_id = $2
  AND end_date >= $3
  AND start_date <= $4
ORDER BY start_date ASC;
