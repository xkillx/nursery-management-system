ALTER TABLE branches
ADD CONSTRAINT branches_tenant_id_id_unique UNIQUE (tenant_id, id);

ALTER TABLE memberships
ADD COLUMN is_active BOOLEAN,
ADD COLUMN ended_at TIMESTAMPTZ;

UPDATE memberships
SET is_active = true
WHERE is_active IS NULL;

UPDATE memberships
SET ended_at = now()
WHERE is_active = false AND ended_at IS NULL;

ALTER TABLE memberships
ALTER COLUMN is_active SET NOT NULL,
ALTER COLUMN is_active SET DEFAULT true;

ALTER TABLE memberships
ADD CONSTRAINT memberships_active_consistency_check
CHECK ((is_active = true AND ended_at IS NULL) OR (is_active = false AND ended_at IS NOT NULL));

ALTER TABLE memberships
ADD CONSTRAINT memberships_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE INDEX idx_memberships_user_active ON memberships (user_id, is_active, created_at);
CREATE INDEX idx_memberships_scope_active ON memberships (tenant_id, branch_id, role, is_active);

ALTER TABLE audit_logs
ADD COLUMN request_id TEXT;

CREATE INDEX idx_audit_logs_request_id ON audit_logs (request_id) WHERE request_id IS NOT NULL;

CREATE TABLE children (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    branch_id UUID NOT NULL,
    full_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    core_hourly_rate_minor INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    left_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT children_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT children_enrollment_dates_check
        CHECK (end_date IS NULL OR start_date <= end_date),
    CONSTRAINT children_active_consistency_check
        CHECK ((is_active = true AND left_at IS NULL) OR (is_active = false AND left_at IS NOT NULL)),
    CONSTRAINT children_core_rate_nonnegative_check
        CHECK (core_hourly_rate_minor >= 0),
    CONSTRAINT children_scope_id_unique
        UNIQUE (tenant_id, branch_id, id)
);

CREATE INDEX idx_children_scope ON children (tenant_id, branch_id);
CREATE INDEX idx_children_active ON children (tenant_id, branch_id, is_active);

CREATE TABLE guardians (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    branch_id UUID NOT NULL,
    full_name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT guardians_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT guardians_active_consistency_check
        CHECK ((is_active = true AND ended_at IS NULL) OR (is_active = false AND ended_at IS NOT NULL)),
    CONSTRAINT guardians_scope_id_unique
        UNIQUE (tenant_id, branch_id, id)
);

CREATE INDEX idx_guardians_scope ON guardians (tenant_id, branch_id);
CREATE INDEX idx_guardians_active ON guardians (tenant_id, branch_id, is_active);

CREATE TABLE guardian_child_links (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    guardian_id UUID NOT NULL,
    child_id UUID NOT NULL,
    ended_at TIMESTAMPTZ,
    ended_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT guardian_child_links_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT guardian_child_links_guardian_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, guardian_id) REFERENCES guardians(tenant_id, branch_id, id),
    CONSTRAINT guardian_child_links_child_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id)
);

CREATE INDEX idx_guardian_child_links_child_active ON guardian_child_links (child_id) WHERE ended_at IS NULL;
CREATE INDEX idx_guardian_child_links_guardian_active ON guardian_child_links (guardian_id) WHERE ended_at IS NULL;
CREATE UNIQUE INDEX idx_guardian_child_links_active_pair
    ON guardian_child_links (tenant_id, branch_id, guardian_id, child_id)
    WHERE ended_at IS NULL;

CREATE TABLE parent_membership_guardians (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    membership_id UUID NOT NULL,
    guardian_id UUID NOT NULL,
    ended_at TIMESTAMPTZ,
    ended_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT parent_membership_guardians_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT parent_membership_guardians_membership_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, membership_id) REFERENCES memberships(tenant_id, branch_id, id),
    CONSTRAINT parent_membership_guardians_guardian_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, guardian_id) REFERENCES guardians(tenant_id, branch_id, id)
);

CREATE UNIQUE INDEX idx_parent_membership_guardians_active_membership
    ON parent_membership_guardians (membership_id)
    WHERE ended_at IS NULL;
CREATE UNIQUE INDEX idx_parent_membership_guardians_active_pair
    ON parent_membership_guardians (membership_id, guardian_id)
    WHERE ended_at IS NULL;
CREATE INDEX idx_parent_membership_guardians_guardian_active
    ON parent_membership_guardians (guardian_id)
    WHERE ended_at IS NULL;

CREATE OR REPLACE FUNCTION enforce_parent_membership_guardian_role()
RETURNS trigger AS $$
DECLARE
    membership_role TEXT;
BEGIN
    SELECT role INTO membership_role FROM memberships WHERE id = NEW.membership_id;
    IF membership_role IS NULL THEN
        RAISE EXCEPTION 'parent_membership_guardians requires valid membership';
    END IF;
    IF membership_role <> 'parent' THEN
        RAISE EXCEPTION 'parent_membership_guardians requires parent role membership';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER parent_membership_guardians_role_check
BEFORE INSERT OR UPDATE ON parent_membership_guardians
FOR EACH ROW
EXECUTE FUNCTION enforce_parent_membership_guardian_role();

CREATE OR REPLACE FUNCTION prevent_non_parent_with_active_guardian_mapping()
RETURNS trigger AS $$
BEGIN
    IF NEW.role <> 'parent' THEN
        IF EXISTS (
            SELECT 1
            FROM parent_membership_guardians
            WHERE membership_id = NEW.id
              AND ended_at IS NULL
        ) THEN
            RAISE EXCEPTION 'membership role must remain parent while active guardian mapping exists';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER memberships_role_guardian_mapping_check
BEFORE UPDATE OF role ON memberships
FOR EACH ROW
EXECUTE FUNCTION prevent_non_parent_with_active_guardian_mapping();
