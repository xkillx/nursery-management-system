-- name: ParentChildrenListByParent :many
SELECT id, tenant_id, branch_id, parent_id, child_id, ended_at,
       COALESCE(ended_reason_code::text, '') AS ended_reason_code,
       ended_reason_note, created_at, updated_at
FROM parent_children
WHERE tenant_id = $1
  AND branch_id = $2
  AND parent_id = $3
  AND ended_at IS NULL
ORDER BY created_at ASC;

-- name: ParentChildrenListByChild :many
SELECT id, tenant_id, branch_id, parent_id, child_id, ended_at,
       COALESCE(ended_reason_code::text, '') AS ended_reason_code,
       ended_reason_note, created_at, updated_at
FROM parent_children
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND ended_at IS NULL
ORDER BY created_at ASC;

-- name: ParentChildrenFindActiveByPair :one
SELECT id, tenant_id, branch_id, parent_id, child_id, ended_at,
       COALESCE(ended_reason_code::text, '') AS ended_reason_code,
       ended_reason_note, created_at, updated_at
FROM parent_children
WHERE tenant_id = $1
  AND branch_id = $2
  AND parent_id = $3
  AND child_id = $4
  AND ended_at IS NULL
LIMIT 1;

-- name: ParentChildrenCreate :exec
INSERT INTO parent_children (id, tenant_id, branch_id, parent_id, child_id)
VALUES ($1, $2, $3, $4, $5);

-- name: ParentChildrenEnd :exec
UPDATE parent_children
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3
  AND branch_id = $4
  AND id = $5
  AND ended_at IS NULL;

-- name: ParentChildrenHasActiveForChild :one
SELECT EXISTS (
    SELECT 1 FROM parent_children
    WHERE tenant_id = $1
      AND branch_id = $2
      AND child_id = $3
      AND ended_at IS NULL
) AS has_active_parent;
