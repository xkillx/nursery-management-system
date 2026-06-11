-- Reverse owner access model changes.
-- Removes owner memberships before restoring NOT NULL branch_id constraint.

-- 1. Remove owner-specific indexes
DROP INDEX IF EXISTS idx_manager_invites_pending_manager;
DROP INDEX IF EXISTS idx_memberships_active_managers;
DROP INDEX IF EXISTS idx_branches_active_tenant;

-- 2. Revert manager_invites.role check
ALTER TABLE manager_invites DROP CONSTRAINT manager_invites_role_check;
ALTER TABLE manager_invites ADD CONSTRAINT manager_invites_role_check
  CHECK (role IN ('practitioner', 'parent'));

-- 3. Remove partial unique indexes and restore original unique constraint
DROP INDEX IF EXISTS idx_memberships_branch_user;
DROP INDEX IF EXISTS idx_memberships_owner_tenant_user;

-- 4. Delete owner memberships before restoring NOT NULL
DELETE FROM refresh_tokens WHERE membership_id IN (
    SELECT id FROM memberships WHERE role = 'owner'
);
DELETE FROM memberships WHERE role = 'owner';

-- 5. Drop owner branch check
ALTER TABLE memberships DROP CONSTRAINT memberships_owner_branch_check;

-- 6. Restore NOT NULL on branch_id
ALTER TABLE memberships ALTER COLUMN branch_id SET NOT NULL;

-- 7. Restore original unique constraint
ALTER TABLE memberships ADD CONSTRAINT memberships_tenant_id_branch_id_user_id_key
  UNIQUE (tenant_id, branch_id, user_id);

-- 8. Revert memberships.role check
ALTER TABLE memberships DROP CONSTRAINT memberships_role_check;
ALTER TABLE memberships ADD CONSTRAINT memberships_role_check
  CHECK (role IN ('manager', 'practitioner', 'parent'));

-- 9. Remove branches.is_active
ALTER TABLE branches DROP COLUMN is_active;
