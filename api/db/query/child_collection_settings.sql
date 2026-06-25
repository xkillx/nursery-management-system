-- name: ChildCollectionSettingGetByChild :one
SELECT id, tenant_id, branch_id, child_id,
       over_18_collection_acknowledged,
       collection_password, collection_password_updated_at,
       collection_password_updated_by_user_id, collection_password_updated_by_membership_id,
       created_at, updated_at, collection_password_hint
FROM child_collection_settings
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildCollectionSettingUpsert :one
INSERT INTO child_collection_settings (
    id, tenant_id, branch_id, child_id,
    over_18_collection_acknowledged,
    collection_password_hint
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (child_id) DO UPDATE SET
    over_18_collection_acknowledged = EXCLUDED.over_18_collection_acknowledged,
    collection_password_hint = EXCLUDED.collection_password_hint,
    updated_at = now()
RETURNING id, tenant_id, branch_id, child_id,
          over_18_collection_acknowledged,
          collection_password, collection_password_updated_at,
          collection_password_updated_by_user_id, collection_password_updated_by_membership_id,
          created_at, updated_at, collection_password_hint;

-- name: ChildCollectionSettingSetPassword :one
UPDATE child_collection_settings SET
    collection_password = $5,
    collection_password_hint = $6,
    collection_password_updated_at = $7,
    collection_password_updated_by_user_id = $8,
    collection_password_updated_by_membership_id = $9,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND id = $4
RETURNING id, tenant_id, branch_id, child_id,
          over_18_collection_acknowledged,
          collection_password, collection_password_updated_at,
          collection_password_updated_by_user_id, collection_password_updated_by_membership_id,
          created_at, updated_at, collection_password_hint;
