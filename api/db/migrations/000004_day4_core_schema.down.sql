DROP TRIGGER IF EXISTS memberships_role_guardian_mapping_check ON memberships;
DROP FUNCTION IF EXISTS prevent_non_parent_with_active_guardian_mapping;

DROP TRIGGER IF EXISTS parent_membership_guardians_role_check ON parent_membership_guardians;
DROP FUNCTION IF EXISTS enforce_parent_membership_guardian_role;

DROP INDEX IF EXISTS idx_parent_membership_guardians_guardian_active;
DROP INDEX IF EXISTS idx_parent_membership_guardians_active_pair;
DROP INDEX IF EXISTS idx_parent_membership_guardians_active_membership;

DROP TABLE IF EXISTS parent_membership_guardians;

DROP INDEX IF EXISTS idx_guardian_child_links_active_pair;
DROP INDEX IF EXISTS idx_guardian_child_links_guardian_active;
DROP INDEX IF EXISTS idx_guardian_child_links_child_active;

DROP TABLE IF EXISTS guardian_child_links;

DROP INDEX IF EXISTS idx_guardians_active;
DROP INDEX IF EXISTS idx_guardians_scope;

DROP TABLE IF EXISTS guardians;

DROP INDEX IF EXISTS idx_children_active;
DROP INDEX IF EXISTS idx_children_scope;

DROP TABLE IF EXISTS children;

DROP INDEX IF EXISTS idx_audit_logs_request_id;
ALTER TABLE audit_logs
DROP COLUMN IF EXISTS request_id;

DROP INDEX IF EXISTS idx_memberships_scope_active;
DROP INDEX IF EXISTS idx_memberships_user_active;
ALTER TABLE memberships
DROP CONSTRAINT IF EXISTS memberships_scope_id_unique;
ALTER TABLE memberships
DROP CONSTRAINT IF EXISTS memberships_active_consistency_check;
ALTER TABLE memberships
DROP COLUMN IF EXISTS ended_at;
ALTER TABLE memberships
DROP COLUMN IF EXISTS is_active;

ALTER TABLE branches
DROP CONSTRAINT IF EXISTS branches_tenant_id_id_unique;
