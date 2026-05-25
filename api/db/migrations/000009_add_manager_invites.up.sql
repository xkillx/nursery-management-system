CREATE TABLE manager_invites (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    email TEXT NOT NULL,
    email_normalized TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('practitioner', 'parent')),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    accepted_user_id UUID REFERENCES users(id),
    accepted_membership_id UUID REFERENCES memberships(id),
    revoked_at TIMESTAMPTZ,
    revoked_by_user_id UUID REFERENCES users(id),
    revoked_by_membership_id UUID REFERENCES memberships(id),
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    created_by_membership_id UUID NOT NULL REFERENCES memberships(id),
    resent_at TIMESTAMPTZ,
    resent_by_user_id UUID REFERENCES users(id),
    resent_by_membership_id UUID REFERENCES memberships(id),
    send_count INTEGER NOT NULL DEFAULT 1 CHECK (send_count >= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT manager_invites_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT manager_invites_accept_shape_check
        CHECK (
            (accepted_at IS NULL AND accepted_user_id IS NULL AND accepted_membership_id IS NULL)
            OR
            (accepted_at IS NOT NULL AND accepted_user_id IS NOT NULL AND accepted_membership_id IS NOT NULL)
        ),
    CONSTRAINT manager_invites_revoke_shape_check
        CHECK (
            (revoked_at IS NULL AND revoked_by_user_id IS NULL AND revoked_by_membership_id IS NULL)
            OR
            (revoked_at IS NOT NULL AND revoked_by_user_id IS NOT NULL AND revoked_by_membership_id IS NOT NULL)
        ),
    CONSTRAINT manager_invites_terminal_state_check
        CHECK (accepted_at IS NULL OR revoked_at IS NULL)
);

CREATE INDEX idx_manager_invites_scope_created ON manager_invites (tenant_id, branch_id, created_at DESC);
CREATE INDEX idx_manager_invites_scope_email ON manager_invites (tenant_id, branch_id, email_normalized);
CREATE INDEX idx_manager_invites_expires_at ON manager_invites (expires_at);
CREATE INDEX idx_manager_invites_pending_lookup ON manager_invites (tenant_id, branch_id, email_normalized, expires_at DESC) WHERE accepted_at IS NULL AND revoked_at IS NULL;
