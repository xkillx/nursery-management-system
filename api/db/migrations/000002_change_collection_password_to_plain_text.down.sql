ALTER TABLE child_collection_settings DROP CONSTRAINT child_collection_settings_password_consistency;
ALTER TABLE child_collection_settings RENAME COLUMN collection_password TO collection_password_hash;
ALTER TABLE child_collection_settings ADD CONSTRAINT child_collection_settings_password_consistency CHECK (
    ((collection_password_hash IS NULL) AND (collection_password_updated_at IS NULL) AND (collection_password_updated_by_user_id IS NULL) AND (collection_password_updated_by_membership_id IS NULL)) OR
    ((collection_password_hash IS NOT NULL) AND (collection_password_updated_at IS NOT NULL) AND (collection_password_updated_by_user_id IS NOT NULL) AND (collection_password_updated_by_membership_id IS NOT NULL))
);
