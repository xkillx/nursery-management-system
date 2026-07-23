-- name: ParentsList :many
SELECT id, tenant_id, branch_id, first_name, last_name, email, phone,
       address_line1, address_line2, address_city, address_postcode,
       relationship_to_child, has_parental_responsibility, can_pick_up,
       is_emergency_contact, notes, user_id, is_active, created_at, updated_at
FROM parents
WHERE tenant_id = $1 AND branch_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ParentsListFiltered :many
SELECT id, tenant_id, branch_id, first_name, last_name, email, phone,
       address_line1, address_line2, address_city, address_postcode,
       relationship_to_child, has_parental_responsibility, can_pick_up,
       is_emergency_contact, notes, user_id, is_active, created_at, updated_at
FROM parents
WHERE tenant_id = $1
  AND branch_id = $2
  AND ($3::boolean IS NULL OR is_active = $3)
  AND ($4::text IS NULL OR (
       first_name ILIKE '%' || $4 || '%'
    OR COALESCE(last_name, '') ILIKE '%' || $4 || '%'
    OR COALESCE(email, '') ILIKE '%' || $4 || '%'
    OR COALESCE(phone, '') ILIKE '%' || $4 || '%'
  ))
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ParentsCount :one
SELECT COUNT(*) FROM parents
WHERE tenant_id = $1 AND branch_id = $2
  AND ($3::boolean IS NULL OR is_active = $3)
  AND ($4::text IS NULL OR (
       first_name ILIKE '%' || $4 || '%'
    OR COALESCE(last_name, '') ILIKE '%' || $4 || '%'
    OR COALESCE(email, '') ILIKE '%' || $4 || '%'
    OR COALESCE(phone, '') ILIKE '%' || $4 || '%'
  ));

-- name: ParentsGetByID :one
SELECT id, tenant_id, branch_id, first_name, last_name, email, phone,
       address_line1, address_line2, address_city, address_postcode,
       relationship_to_child, has_parental_responsibility, can_pick_up,
       is_emergency_contact, notes, user_id, is_active, created_at, updated_at
FROM parents
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ParentsGetByUserID :one
SELECT id, tenant_id, branch_id, first_name, last_name, email, phone,
       address_line1, address_line2, address_city, address_postcode,
       relationship_to_child, has_parental_responsibility, can_pick_up,
       is_emergency_contact, notes, user_id, is_active, created_at, updated_at
FROM parents
WHERE tenant_id = $1 AND user_id = $2 AND is_active = true;

-- name: ParentsCreate :exec
INSERT INTO parents (id, tenant_id, branch_id, first_name, last_name, email, phone,
                     address_line1, address_line2, address_city, address_postcode,
                     relationship_to_child, has_parental_responsibility, can_pick_up,
                     is_emergency_contact, notes, user_id, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);

-- name: ParentsUpdate :exec
UPDATE parents
SET first_name = $4,
    last_name = $5,
    email = $6,
    phone = $7,
    address_line1 = $8,
    address_line2 = $9,
    address_city = $10,
    address_postcode = $11,
    relationship_to_child = $12,
    has_parental_responsibility = $13,
    can_pick_up = $14,
    is_emergency_contact = $15,
    notes = $16,
    user_id = $17,
    is_active = $18,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ParentsSoftDelete :exec
UPDATE parents
SET is_active = false,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ParentsSetUserID :exec
UPDATE parents
SET user_id = $4,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;
