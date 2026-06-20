-- name: SessionTemplatesListByBranch :many
SELECT id, tenant_id, branch_id, name, description, is_active, created_at, updated_at
FROM session_templates
WHERE tenant_id = $1
  AND branch_id = $2
  AND (NOT $3::bool OR is_active = true)
ORDER BY name ASC;

-- name: SessionTemplatesGetByID :one
SELECT id, tenant_id, branch_id, name, description, is_active, created_at, updated_at
FROM session_templates
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: SessionTemplatesGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, name, description, is_active, created_at, updated_at
FROM session_templates
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: SessionTemplatesCreate :exec
INSERT INTO session_templates (id, tenant_id, branch_id, name, description, is_active)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), true);

-- name: SessionTemplatesUpdate :execrows
UPDATE session_templates
SET
    name = CASE WHEN @set_name = 1 THEN @name ELSE name END,
    description = CASE WHEN @set_description = 1 THEN NULLIF(@description, '') ELSE description END,
    updated_at = now()
WHERE tenant_id = @tenant_id AND branch_id = @branch_id AND id = @id;

-- name: SessionTemplatesArchive :exec
UPDATE session_templates
SET is_active = false, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: SessionTemplatesReactivate :exec
UPDATE session_templates
SET is_active = true, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: SessionTemplatesCheckActiveNameExists :one
SELECT EXISTS (
    SELECT 1 FROM session_templates
    WHERE tenant_id = $1 AND branch_id = $2 AND name = $3 AND is_active = true
      AND ($4::uuid IS NULL OR id != $4)
);

-- name: SessionTemplatesExists :one
SELECT EXISTS (
    SELECT 1 FROM session_templates
    WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
);
