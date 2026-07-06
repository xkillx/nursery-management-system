-- name: ChildRoomAssignmentsListByChild :many
SELECT id, tenant_id, branch_id, child_id, room_id, start_date, end_date, created_at
FROM child_room_assignments
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY start_date DESC, created_at DESC;

-- name: ChildRoomAssignmentsGetCurrentByChild :one
SELECT id, tenant_id, branch_id, child_id, room_id, start_date, end_date, created_at
FROM child_room_assignments
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND is_current
LIMIT 1;

-- name: ChildRoomAssignmentsCloseCurrent :exec
UPDATE child_room_assignments
SET end_date = $4
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND is_current;

-- name: ChildRoomAssignmentsInsert :one
INSERT INTO child_room_assignments (id, tenant_id, branch_id, child_id, room_id, start_date)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, branch_id, child_id, room_id, start_date, end_date, created_at;

-- name: ChildRoomAssignmentsGetByID :one
SELECT id, tenant_id, branch_id, child_id, room_id, start_date, end_date, created_at
FROM child_room_assignments
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ChildRoomAssignmentsCloseByID :exec
UPDATE child_room_assignments
SET end_date = $4
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3 AND is_current;

-- name: ChildRoomAssignmentsListByChildPaginated :many
SELECT id, tenant_id, branch_id, child_id, room_id, start_date, end_date, created_at
FROM child_room_assignments
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY start_date DESC, created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: ChildRoomAssignmentsCountByChild :one
SELECT COUNT(*)
FROM child_room_assignments
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;
