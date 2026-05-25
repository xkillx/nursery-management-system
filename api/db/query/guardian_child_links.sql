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

-- name: GuardianLinksEnd :exec
UPDATE guardian_child_links
SET ended_at = now(), ended_reason_code = $1, ended_reason_note = NULLIF($2, ''), updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5;
