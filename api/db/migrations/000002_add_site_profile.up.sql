CREATE TABLE site_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    nursery_name varchar(120) NOT NULL,
    description varchar(2000) NOT NULL DEFAULT '',
    phone varchar(32) NOT NULL,
    email varchar(254) NOT NULL,
    website varchar(2048) NOT NULL,
    address_street varchar(200) NOT NULL,
    address_city varchar(100) NOT NULL,
    address_postcode varchar(16) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT site_profiles_pkey PRIMARY KEY (id),
    CONSTRAINT site_profiles_branch_id_unique UNIQUE (branch_id),
    CONSTRAINT site_profiles_branch_id_fkey FOREIGN KEY (branch_id)
        REFERENCES branches(id) ON DELETE CASCADE
);

CREATE INDEX idx_site_profiles_tenant_branch ON site_profiles(tenant_id, branch_id);
