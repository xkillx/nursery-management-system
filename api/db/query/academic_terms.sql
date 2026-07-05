-- name: AcademicTermsListByBranch :many
SELECT id, tenant_id, branch_id, name, kind, start_date, end_date, is_active, created_at, updated_at
FROM academic_terms
WHERE tenant_id = $1
  AND branch_id = $2
  AND (NOT $3::bool OR is_active = true)
ORDER BY start_date DESC, name ASC;

-- name: AcademicTermsGetByID :one
SELECT id, tenant_id, branch_id, name, kind, start_date, end_date, is_active, created_at, updated_at
FROM academic_terms
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3;

-- name: AcademicTermsGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, name, kind, start_date, end_date, is_active, created_at, updated_at
FROM academic_terms
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE;

-- name: AcademicTermsCreate :exec
INSERT INTO academic_terms (id, tenant_id, branch_id, name, kind, start_date, end_date, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, true);

-- name: AcademicTermsUpdate :execrows
UPDATE academic_terms
SET
    name = CASE WHEN @set_name::bool THEN @name ELSE name END,
    kind = CASE WHEN @set_kind::bool THEN @kind ELSE kind END,
    start_date = CASE WHEN @set_start_date::bool THEN @start_date ELSE start_date END,
    end_date = CASE WHEN @set_end_date::bool THEN @end_date ELSE end_date END,
    updated_at = now()
WHERE tenant_id = @tenant_id AND branch_id = @branch_id AND id = @id;

-- name: AcademicTermsArchive :exec
UPDATE academic_terms
SET is_active = false, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: AcademicTermsCheckActiveNameExists :one
SELECT EXISTS (
    SELECT 1 FROM academic_terms
    WHERE tenant_id = $1 AND branch_id = $2 AND name = $3 AND is_active = true
      AND ($4::uuid IS NULL OR id != $4)
);

-- name: AcademicTermsListActiveDateRanges :many
SELECT start_date, end_date
FROM academic_terms
WHERE tenant_id = $1
  AND branch_id = $2
  AND is_active = true
  AND end_date >= $3
  AND start_date <= $4
ORDER BY start_date ASC;
