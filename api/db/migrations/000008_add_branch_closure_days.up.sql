CREATE TABLE branch_closure_days (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    date date NOT NULL,
    reason text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX branch_closure_days_tenant_branch_date_idx
    ON branch_closure_days (tenant_id, branch_id, date);

CREATE INDEX branch_closure_days_tenant_branch_month_idx
    ON branch_closure_days (tenant_id, branch_id, date);
