CREATE TABLE absence_markers (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    local_date DATE NOT NULL,
    marked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    marked_by_user_id UUID NOT NULL REFERENCES users(id),
    marked_by_membership_id UUID NOT NULL,
    cleared_at TIMESTAMPTZ,
    cleared_by_user_id UUID,
    cleared_by_membership_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT absence_markers_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT absence_markers_child_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT absence_markers_marked_membership_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, marked_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id),
    CONSTRAINT absence_markers_cleared_membership_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, cleared_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id),
    CONSTRAINT absence_markers_cleared_at_null_implies_user_null
        CHECK ((cleared_at IS NULL) = (cleared_by_user_id IS NULL)),
    CONSTRAINT absence_markers_cleared_at_null_implies_membership_null
        CHECK ((cleared_at IS NULL) = (cleared_by_membership_id IS NULL)),
    CONSTRAINT absence_markers_cleared_at_gte_marked_at
        CHECK (cleared_at IS NULL OR cleared_at >= marked_at)
);

CREATE UNIQUE INDEX idx_absence_markers_active_unique
    ON absence_markers (tenant_id, branch_id, child_id, local_date)
    WHERE cleared_at IS NULL;

CREATE INDEX idx_absence_markers_active_scope_date
    ON absence_markers (tenant_id, branch_id, local_date)
    WHERE cleared_at IS NULL;

CREATE UNIQUE INDEX idx_absence_markers_scope_id
    ON absence_markers (tenant_id, branch_id, id);
