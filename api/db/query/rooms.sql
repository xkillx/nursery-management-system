-- name: RoomsListByBranch :many
SELECT id, tenant_id, branch_id, name, description, age_group, capacity, is_active, created_at, updated_at
FROM rooms
WHERE tenant_id = $1
  AND branch_id = $2
  AND (NOT $3::bool OR is_active = true)
ORDER BY name ASC;

-- name: RoomsGetByID :one
SELECT id, tenant_id, branch_id, name, description, age_group, capacity, is_active, created_at, updated_at
FROM rooms
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: RoomsCreate :exec
INSERT INTO rooms (id, tenant_id, branch_id, name, description, age_group, capacity, is_active)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, true);

-- name: RoomsUpdate :execrows
UPDATE rooms
SET
    name = CASE WHEN @set_name = 1 THEN @name ELSE name END,
    age_group = CASE WHEN @set_age_group = 1 THEN @age_group ELSE age_group END,
    capacity = CASE WHEN @set_capacity = 1 THEN @capacity ELSE capacity END,
    description = CASE WHEN @set_description = 1 THEN NULLIF(@description, '') ELSE description END,
    updated_at = now()
WHERE tenant_id = @tenant_id AND branch_id = @branch_id AND id = @id;

-- name: RoomsArchive :exec
UPDATE rooms
SET is_active = false, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: RoomsReactivate :exec
UPDATE rooms
SET is_active = true, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: RoomsCheckActiveNameExists :one
SELECT EXISTS (
    SELECT 1 FROM rooms
    WHERE tenant_id = $1 AND branch_id = $2 AND name = $3 AND is_active = true
      AND ($4::uuid IS NULL OR id != $4)
);

-- name: RoomsCountActiveChildren :one
SELECT 0::bigint AS count;

-- name: RoomsCountAssignedChildrenByBranch :many
SELECT primary_room_id AS room_id, COUNT(*)::bigint AS assigned_count
FROM children
WHERE tenant_id = $1
  AND branch_id = $2
  AND primary_room_id IS NOT NULL
  AND is_active = true
GROUP BY primary_room_id;

-- name: RoomsGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, name, description, age_group, capacity, is_active, created_at, updated_at
FROM rooms
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: RoomsExists :one
SELECT EXISTS (
    SELECT 1 FROM rooms
    WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
);
