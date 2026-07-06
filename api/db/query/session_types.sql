-- name: SessionTypesListByBranch :many
SELECT id, tenant_id, branch_id, name, start_time, end_time, is_active, created_at, updated_at, kind, flat_fee_minor
FROM session_types
WHERE tenant_id = $1
  AND branch_id = $2
  AND (NOT $3::bool OR is_active = true)
ORDER BY name ASC;

-- name: SessionTypesGetByID :one
SELECT id, tenant_id, branch_id, name, start_time, end_time, is_active, created_at, updated_at, kind, flat_fee_minor
FROM session_types
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: SessionTypesGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, name, start_time, end_time, is_active, created_at, updated_at, kind, flat_fee_minor
FROM session_types
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: SessionTypesCreate :exec
INSERT INTO session_types (id, tenant_id, branch_id, name, start_time, end_time, kind, flat_fee_minor)
VALUES ($1, $2, $3, $4, $5, $6, $7, sqlc.narg('flat_fee_minor'));

-- name: SessionTypesUpdate :execrows
UPDATE session_types
SET
    name = CASE WHEN @set_name = 1 THEN @name ELSE name END,
    start_time = CASE WHEN @set_start_time = 1 THEN @start_time ELSE start_time END,
    end_time = CASE WHEN @set_end_time = 1 THEN @end_time ELSE end_time END,
    kind = CASE WHEN @set_kind = 1 THEN sqlc.narg('new_kind') ELSE kind END,
    flat_fee_minor = CASE WHEN @set_flat_fee_minor::bool THEN sqlc.narg('new_flat_fee_minor') ELSE flat_fee_minor END,
    updated_at = now()
WHERE tenant_id = @tenant_id AND branch_id = @branch_id AND id = @id;

-- name: SessionTypesArchive :exec
UPDATE session_types
SET is_active = false, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: SessionTypesReactivate :exec
UPDATE session_types
SET is_active = true, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: SessionTypesCheckActiveNameExists :one
SELECT EXISTS (
    SELECT 1 FROM session_types
    WHERE tenant_id = $1 AND branch_id = $2 AND name = $3 AND is_active = true
      AND ($4::uuid IS NULL OR id != $4)
);

-- name: SessionTypesExists :one
SELECT EXISTS (
    SELECT 1 FROM session_types
    WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
);

-- name: SessionTypesListByBranchPaginated :many
SELECT id, tenant_id, branch_id, name, start_time, end_time, is_active, created_at, updated_at, kind, flat_fee_minor
FROM session_types
WHERE tenant_id = $1
  AND branch_id = $2
  AND (NOT $3::bool OR is_active = true)
ORDER BY name ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: SessionTypesCountByBranch :one
SELECT COUNT(*)
FROM session_types
WHERE tenant_id = $1
  AND branch_id = $2
  AND (NOT $3::bool OR is_active = true);
