-- name: ParentChildMappingsFindActiveByPair :one
SELECT id, membership_id, child_id, ended_at,
       COALESCE(ended_reason_code::text, '') AS ended_reason_code,
       ended_reason_note, created_at, updated_at
FROM parent_membership_children
WHERE tenant_id = $1
  AND branch_id = $2
  AND membership_id = $3
  AND child_id = $4
  AND ended_at IS NULL
LIMIT 1;

-- name: ParentChildMappingsListActiveByMembership :many
SELECT id, membership_id, child_id, ended_at,
       COALESCE(ended_reason_code::text, '') AS ended_reason_code,
       ended_reason_note, created_at, updated_at
FROM parent_membership_children
WHERE tenant_id = $1
  AND branch_id = $2
  AND membership_id = $3
  AND ended_at IS NULL
ORDER BY created_at ASC;

-- name: ParentChildMappingsCreate :exec
INSERT INTO parent_membership_children (id, tenant_id, branch_id, membership_id, child_id)
VALUES ($1, $2, $3, $4, $5);

-- name: ParentChildMappingsGetByIDForUpdate :one
SELECT id, membership_id, child_id, ended_at,
       COALESCE(ended_reason_code::text, '') AS ended_reason_code,
       ended_reason_note, created_at, updated_at
FROM parent_membership_children
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: ParentChildMappingsEnd :exec
UPDATE parent_membership_children
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3
  AND branch_id = $4
  AND id = $5
  AND ended_at IS NULL;

-- name: ParentChildMappingsGetMembershipForScope :one
SELECT id, role, is_active
FROM memberships
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;
