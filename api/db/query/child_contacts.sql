-- name: ChildContactsListByChild :many
SELECT id, tenant_id, branch_id, child_id, contact_type, sort_order,
       full_name, relationship_to_child,
       address, telephone, email, work_address, has_parental_responsibility,
       created_at, updated_at
FROM child_contacts
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY contact_type ASC, sort_order ASC;

-- name: ChildContactsDeleteByTypes :exec
DELETE FROM child_contacts
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
  AND contact_type::text = ANY($4::text[]);

-- name: ChildContactsInsert :one
INSERT INTO child_contacts (
    id, tenant_id, branch_id, child_id, contact_type, sort_order,
    full_name, relationship_to_child,
    address, telephone, email, work_address, has_parental_responsibility
)
VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, NULLIF($8, ''),
    $9, NULLIF($10, ''), NULLIF($11, ''), $12, $13
)
RETURNING id, tenant_id, branch_id, child_id, contact_type, sort_order,
          full_name, relationship_to_child,
          address, telephone, email, work_address, has_parental_responsibility,
          created_at, updated_at;
