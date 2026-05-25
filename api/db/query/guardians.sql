-- name: GuardiansList :many
SELECT g.id,
       g.full_name,
       g.email,
       g.phone,
       g.notes,
       g.is_active,
       g.deactivated_at,
       COALESCE(g.deactivation_reason_code::text, '') AS deactivation_reason_code,
       g.deactivation_reason_note,
       g.created_at,
       g.updated_at
FROM guardians g
WHERE g.tenant_id = $1
  AND g.branch_id = $2
  AND (
      sqlc.arg('status_filter') = 'all'
      OR (sqlc.arg('status_filter') = 'active' AND g.is_active = true)
      OR (sqlc.arg('status_filter') = 'inactive' AND g.is_active = false)
  )
ORDER BY g.updated_at DESC
LIMIT $3 OFFSET $4;

-- name: GuardiansGetByID :one
SELECT g.id,
       g.full_name,
       g.email,
       g.phone,
       g.notes,
       g.is_active,
       g.deactivated_at,
       COALESCE(g.deactivation_reason_code::text, '') AS deactivation_reason_code,
       g.deactivation_reason_note,
       g.created_at,
       g.updated_at
FROM guardians g
WHERE g.tenant_id = $1
  AND g.branch_id = $2
  AND g.id = $3;

-- name: GuardiansCreate :exec
INSERT INTO guardians (id, tenant_id, branch_id, full_name, email, phone, notes, is_active)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), true);

-- name: GuardiansUpdate :execrows
UPDATE guardians
SET
    full_name = CASE WHEN @set_full_name = 1 THEN @full_name ELSE full_name END,
    email = CASE WHEN @set_email = 1 THEN NULLIF(@email, '') ELSE email END,
    phone = CASE WHEN @set_phone = 1 THEN NULLIF(@phone, '') ELSE phone END,
    notes = CASE WHEN @set_notes = 1 THEN NULLIF(@notes, '') ELSE notes END,
    updated_at = now()
WHERE tenant_id = @tenant_id AND branch_id = @branch_id AND id = @id;

-- name: GuardiansGetByIDForUpdate :one
SELECT g.id,
       g.full_name,
       g.email,
       g.phone,
       g.notes,
       g.is_active,
       g.deactivated_at,
       COALESCE(g.deactivation_reason_code::text, '') AS deactivation_reason_code,
       g.deactivation_reason_note,
       g.created_at,
       g.updated_at
FROM guardians g
WHERE g.tenant_id = $1
  AND g.branch_id = $2
  AND g.id = $3
FOR UPDATE;

-- name: GuardiansGetActive :one
SELECT is_active
FROM guardians
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: GuardiansDeactivate :exec
UPDATE guardians
SET is_active = false,
    deactivated_at = now(),
    deactivation_reason_code = $1,
    deactivation_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5;

-- name: GuardiansCascadeLinks :exec
UPDATE guardian_child_links
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = $2,
    updated_at = now()
WHERE tenant_id = $3
  AND branch_id = $4
  AND guardian_id = $5
  AND ended_at IS NULL;

-- name: GuardiansCascadeMappings :exec
UPDATE parent_membership_guardians
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = $2,
    updated_at = now()
WHERE tenant_id = $3
  AND branch_id = $4
  AND guardian_id = $5
  AND ended_at IS NULL;

-- name: GuardiansReactivate :exec
UPDATE guardians
SET is_active = true,
    deactivated_at = NULL,
    deactivation_reason_code = NULL,
    deactivation_reason_note = NULL,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;
