-- name: GuardianLinksFindActiveByPair :one
SELECT id, guardian_id, child_id, ended_at, COALESCE(ended_reason_code::text, '') AS ended_reason_code, ended_reason_note, created_at, updated_at
FROM guardian_child_links
WHERE tenant_id = $1 AND branch_id = $2 AND guardian_id = $3 AND child_id = $4 AND ended_at IS NULL
LIMIT 1;

-- name: GuardianLinksCreate :exec
INSERT INTO guardian_child_links (id, tenant_id, branch_id, guardian_id, child_id) VALUES ($1, $2, $3, $4, $5);

-- name: GuardianLinksGetByIDForUpdate :one
SELECT id, guardian_id, child_id, ended_at, COALESCE(ended_reason_code::text, '') AS ended_reason_code, ended_reason_note, created_at, updated_at
FROM guardian_child_links
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
FOR UPDATE;

-- name: GuardianLinksListActiveByChild :many
SELECT
    gcl.id,
    gcl.guardian_id,
    gcl.child_id,
    gcl.created_at,
    gcl.updated_at,
    g.id AS guardian_table_id,
    g.full_name AS guardian_full_name,
    g.email AS guardian_email,
    g.phone AS guardian_phone,
    g.is_active AS guardian_is_active
FROM guardian_child_links gcl
JOIN guardians g ON g.tenant_id = gcl.tenant_id AND g.branch_id = gcl.branch_id AND g.id = gcl.guardian_id AND g.is_active = true
WHERE gcl.tenant_id = $1 AND gcl.branch_id = $2 AND gcl.child_id = $3 AND gcl.ended_at IS NULL
ORDER BY g.full_name ASC, gcl.created_at ASC;

-- name: GuardianLinksEnd :exec
UPDATE guardian_child_links
SET ended_at = now(), ended_reason_code = $1, ended_reason_note = NULLIF($2, ''), updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5;
