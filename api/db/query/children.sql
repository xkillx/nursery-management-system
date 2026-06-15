-- name: ChildrenList :many
SELECT c.id,
       c.first_name,
       c.middle_name,
       c.last_name,
       c.date_of_birth,
       c.start_date,
       c.end_date,
       c.core_hourly_rate_minor,
       b.core_hourly_rate_minor AS site_core_hourly_rate_minor,
       c.notes,
       c.is_active,
       c.left_at,
       COALESCE(c.left_reason_code::text, '') AS left_reason_code,
       c.left_reason_note,
       EXISTS (
           SELECT 1
           FROM guardian_child_links gcl
           WHERE gcl.tenant_id = c.tenant_id
             AND gcl.branch_id = c.branch_id
             AND gcl.child_id = c.id
             AND gcl.ended_at IS NULL
       ) AS has_guardian_link,
       c.created_at,
       c.updated_at
FROM children c
JOIN branches b ON b.tenant_id = c.tenant_id AND b.id = c.branch_id
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND (
      sqlc.arg('status_filter') = 'all'
      OR (sqlc.arg('status_filter') = 'active' AND c.is_active = true)
      OR (sqlc.arg('status_filter') = 'inactive' AND c.is_active = false)
  )
ORDER BY c.updated_at DESC
LIMIT $3 OFFSET $4;

-- name: ChildrenGetByID :one
SELECT c.id,
       c.first_name,
       c.middle_name,
       c.last_name,
       c.date_of_birth,
       c.start_date,
       c.end_date,
       c.core_hourly_rate_minor,
       b.core_hourly_rate_minor AS site_core_hourly_rate_minor,
       c.notes,
       c.is_active,
       c.left_at,
       COALESCE(c.left_reason_code::text, '') AS left_reason_code,
       c.left_reason_note,
       EXISTS (
           SELECT 1
           FROM guardian_child_links gcl
           WHERE gcl.tenant_id = c.tenant_id
             AND gcl.branch_id = c.branch_id
             AND gcl.child_id = c.id
             AND gcl.ended_at IS NULL
       ) AS has_guardian_link,
       c.created_at,
       c.updated_at
FROM children c
JOIN branches b ON b.tenant_id = c.tenant_id AND b.id = c.branch_id
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.id = $3;

-- name: ChildrenCreate :exec
INSERT INTO children (
    id, tenant_id, branch_id, first_name, middle_name, last_name, date_of_birth, start_date, end_date,
    core_hourly_rate_minor, notes, is_active
)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9, $10, NULLIF($11, ''), true);

-- name: ChildrenUpdate :execrows
UPDATE children
SET
    first_name = CASE WHEN @set_first_name = 1 THEN @first_name ELSE first_name END,
    middle_name = CASE WHEN @set_middle_name = 1 THEN NULLIF(@middle_name, '') ELSE middle_name END,
    last_name = CASE WHEN @set_last_name = 1 THEN NULLIF(@last_name, '') ELSE last_name END,
    date_of_birth = CASE WHEN @set_date_of_birth = 1 THEN @date_of_birth ELSE date_of_birth END,
    start_date = CASE WHEN @set_start_date = 1 THEN @start_date ELSE start_date END,
    end_date = CASE WHEN @set_end_date = 1 THEN sqlc.narg('end_date') ELSE end_date END,
    core_hourly_rate_minor = CASE WHEN @set_core_hourly_rate_minor = 1 THEN @core_hourly_rate_minor ELSE core_hourly_rate_minor END,
    notes = CASE WHEN @set_notes = 1 THEN NULLIF(@notes, '') ELSE notes END,
    updated_at = now()
WHERE tenant_id = @tenant_id AND branch_id = @branch_id AND id = @id;

-- name: ChildrenMarkInactive :exec
UPDATE children
SET is_active = false,
    left_at = now(),
    left_reason_code = $1,
    left_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5;

-- name: ChildrenGetByIDForUpdate :one
SELECT c.id,
       c.first_name,
       c.middle_name,
       c.last_name,
       c.date_of_birth,
       c.start_date,
       c.end_date,
       c.core_hourly_rate_minor,
       b.core_hourly_rate_minor AS site_core_hourly_rate_minor,
       c.notes,
       c.is_active,
       c.left_at,
       COALESCE(c.left_reason_code::text, '') AS left_reason_code,
       c.left_reason_note,
       EXISTS (
           SELECT 1
           FROM guardian_child_links gcl
           WHERE gcl.tenant_id = c.tenant_id
             AND gcl.branch_id = c.branch_id
             AND gcl.child_id = c.id
             AND gcl.ended_at IS NULL
       ) AS has_guardian_link,
       c.created_at,
       c.updated_at
FROM children c
JOIN branches b ON b.tenant_id = c.tenant_id AND b.id = c.branch_id
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.id = $3
FOR UPDATE;

-- name: ChildrenExistsInScope :one
SELECT EXISTS (
  SELECT 1 FROM children WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
);

-- name: ChildrenListAttendance :many
SELECT c.id,
       c.first_name,
       c.middle_name,
       c.last_name,
       (c.first_name IS NOT NULL AND btrim(c.first_name) <> ''
        AND c.date_of_birth IS NOT NULL
        AND c.start_date IS NOT NULL
        AND EXISTS (
            SELECT 1
            FROM guardian_child_links gcl
            WHERE gcl.tenant_id = c.tenant_id
              AND gcl.branch_id = c.branch_id
              AND gcl.child_id = c.id
              AND gcl.ended_at IS NULL
        )) AS enrollment_complete,
       CASE
         WHEN s.id IS NOT NULL THEN 'checked_in'
         WHEN am.id IS NOT NULL THEN 'absent'
         ELSE 'not_checked_in'
       END AS attendance_state,
       s.id AS open_session_id,
       s.check_in_at AS checked_in_at,
       s.id IS NOT NULL AS has_incomplete_session,
       am.id AS absence_marker_id,
       am.marked_at AS absence_marked_at
FROM children c
LEFT JOIN attendance_sessions s
  ON s.tenant_id = c.tenant_id
 AND s.branch_id = c.branch_id
 AND s.child_id = c.id
 AND s.status = 'open'
LEFT JOIN absence_markers am
  ON am.tenant_id = c.tenant_id
 AND am.branch_id = c.branch_id
 AND am.child_id = c.id
 AND am.local_date = $3
 AND am.cleared_at IS NULL
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND (
      (c.is_active = true AND c.start_date <= $3 AND (c.end_date IS NULL OR c.end_date >= $3))
      OR s.id IS NOT NULL
  )
ORDER BY c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, c.id ASC;

-- name: ChildrenGetForCorrection :one
SELECT c.id, c.start_date, c.end_date
FROM children c
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.id = $3;
