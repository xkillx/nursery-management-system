CREATE TABLE funding_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    billing_month DATE NOT NULL,
    funded_allowance_minutes INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT funding_profiles_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT funding_profiles_child_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT funding_profiles_billing_month_first_day_check
        CHECK (billing_month = date_trunc('month', billing_month)::date),
    CONSTRAINT funding_profiles_allowance_bounds_check
        CHECK (funded_allowance_minutes >= 0 AND funded_allowance_minutes <= 44640),
    CONSTRAINT funding_profiles_scope_child_month_unique
        UNIQUE (tenant_id, branch_id, child_id, billing_month)
);

CREATE INDEX idx_funding_profiles_scope_month
    ON funding_profiles (tenant_id, branch_id, billing_month);

CREATE INDEX idx_funding_profiles_child_month
    ON funding_profiles (tenant_id, branch_id, child_id, billing_month);
