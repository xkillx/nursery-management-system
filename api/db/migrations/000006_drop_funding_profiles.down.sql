CREATE TABLE funding_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    billing_month date NOT NULL,
    funded_allowance_minutes integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    funding_type text,
    funding_model text,
    funded_hours_per_week numeric(5,2),
    CONSTRAINT funding_profiles_allowance_bounds_check CHECK (((funded_allowance_minutes >= 0) AND (funded_allowance_minutes <= 44640))),
    CONSTRAINT funding_profiles_billing_month_first_day_check CHECK ((billing_month = (date_trunc('month'::text, (billing_month)::timestamp with time zone))::date))
);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_scope_child_month_unique UNIQUE (tenant_id, branch_id, child_id, billing_month);

CREATE INDEX idx_funding_profiles_child_month ON funding_profiles USING btree (tenant_id, branch_id, child_id, billing_month);

CREATE INDEX idx_funding_profiles_scope_month ON funding_profiles USING btree (tenant_id, branch_id, billing_month);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_child_scope_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);
