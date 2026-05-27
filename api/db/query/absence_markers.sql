-- name: AbsenceMarkersCreate :one
INSERT INTO absence_markers (
    id, tenant_id, branch_id, child_id, local_date,
    marked_at, marked_by_user_id, marked_by_membership_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: AbsenceMarkersFindActiveByChildDate :one
SELECT * FROM absence_markers
WHERE tenant_id = $1
  AND branch_id = $2
  AND child_id = $3
  AND local_date = $4
  AND cleared_at IS NULL;

-- name: AbsenceMarkersGetByID :one
SELECT * FROM absence_markers
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: AbsenceMarkersClear :one
UPDATE absence_markers
SET cleared_at = $4,
    cleared_by_user_id = $5,
    cleared_by_membership_id = $6,
    updated_at = now()
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
  AND cleared_at IS NULL
RETURNING *;

-- name: AbsenceMarkersHasAttendanceForChildDate :one
SELECT EXISTS (
    SELECT 1 FROM attendance_sessions
    WHERE tenant_id = $1
      AND branch_id = $2
      AND child_id = $3
      AND check_in_local_date = $4
);
