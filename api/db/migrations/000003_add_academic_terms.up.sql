CREATE TABLE academic_terms (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name varchar(120) NOT NULL,
    kind varchar(20) NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT academic_terms_pkey PRIMARY KEY (id),
    CONSTRAINT academic_terms_start_before_end CHECK (start_date < end_date),
    CONSTRAINT academic_terms_kind_check CHECK (kind IN ('autumn', 'spring', 'summer')),
    CONSTRAINT academic_terms_branch_id_fkey FOREIGN KEY (branch_id)
        REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_academic_terms_tenant_branch_name_active
    ON academic_terms(tenant_id, branch_id, name) WHERE is_active = true;

CREATE INDEX idx_academic_terms_tenant_branch ON academic_terms(tenant_id, branch_id);
CREATE INDEX idx_academic_terms_dates ON academic_terms(tenant_id, branch_id, start_date, end_date);
