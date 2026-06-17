-- name: ChildLeavingRecordGetByChild :one
SELECT id, tenant_id, branch_id, child_id, left_at, reason_code, reason_note, created_at
FROM child_leaving_records
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildLeavingRecordInsert :one
INSERT INTO child_leaving_records (
    id, tenant_id, branch_id, child_id, left_at, reason_code, reason_note
)
VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
ON CONFLICT (child_id) DO NOTHING
RETURNING id, tenant_id, branch_id, child_id, left_at, reason_code, reason_note, created_at;
