-- Owner access model: tenant-wide owner memberships, active-site filtering,
-- and manager-role invite support for owner administration paths.

-- 1. branches.is_active for active-site filtering
ALTER TABLE branches ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- 2. Extend memberships.role to include 'owner'
ALTER TABLE memberships DROP CONSTRAINT memberships_role_check;
ALTER TABLE memberships ADD CONSTRAINT memberships_role_check
  CHECK (role IN ('owner', 'manager', 'practitioner', 'parent'));

-- 3. Make memberships.branch_id nullable for owner memberships
ALTER TABLE memberships ALTER COLUMN branch_id DROP NOT NULL;

-- 4. Owner memberships must have NULL branch; non-owner must have non-NULL branch
ALTER TABLE memberships ADD CONSTRAINT memberships_owner_branch_check
  CHECK (
    (role = 'owner' AND branch_id IS NULL)
    OR
    (role IN ('manager', 'practitioner', 'parent') AND branch_id IS NOT NULL)
  );

-- 5. Replace (tenant_id, branch_id, user_id) unique with partial indexes
--    Original cannot enforce uniqueness for NULL branch_id (owner memberships).
ALTER TABLE memberships DROP CONSTRAINT memberships_tenant_id_branch_id_user_id_key;

--    One owner membership per (tenant, user)
CREATE UNIQUE INDEX idx_memberships_owner_tenant_user
  ON memberships (tenant_id, user_id) WHERE role = 'owner';

--    One non-owner membership per (tenant, branch, user)
CREATE UNIQUE INDEX idx_memberships_branch_user
  ON memberships (tenant_id, branch_id, user_id) WHERE role IN ('manager', 'practitioner', 'parent');

-- 6. Extend manager_invites.role to include 'manager' for owner-created invites
ALTER TABLE manager_invites DROP CONSTRAINT manager_invites_role_check;
ALTER TABLE manager_invites ADD CONSTRAINT manager_invites_role_check
  CHECK (role IN ('manager', 'practitioner', 'parent'));

-- 7. Indexes for owner query paths
CREATE INDEX idx_branches_active_tenant ON branches (tenant_id) WHERE is_active = true;
CREATE INDEX idx_memberships_active_managers ON memberships (tenant_id, branch_id)
  WHERE role = 'manager' AND is_active = true AND ended_at IS NULL;
CREATE INDEX idx_manager_invites_pending_manager ON manager_invites (tenant_id, branch_id)
  WHERE role = 'manager' AND accepted_at IS NULL AND revoked_at IS NULL;
