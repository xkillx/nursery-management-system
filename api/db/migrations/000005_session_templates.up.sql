CREATE TABLE session_templates (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE UNIQUE INDEX session_templates_active_name_unique
    ON session_templates USING btree (tenant_id, branch_id, name)
    WHERE (is_active = true);

CREATE INDEX session_templates_active_by_branch
    ON session_templates USING btree (tenant_id, branch_id)
    WHERE (is_active = true);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_branch_fkey
    FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

CREATE TABLE session_template_entries (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    template_id uuid NOT NULL,
    day_of_week integer NOT NULL,
    session_type_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT session_template_entries_dow_check CHECK (((day_of_week >= 1) AND (day_of_week <= 7)))
);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE UNIQUE INDEX session_template_entries_unique_day_session
    ON session_template_entries USING btree (tenant_id, branch_id, template_id, day_of_week, session_type_id);

CREATE INDEX session_template_entries_by_template
    ON session_template_entries USING btree (tenant_id, branch_id, template_id, day_of_week);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_template_fkey
    FOREIGN KEY (tenant_id, branch_id, template_id) REFERENCES session_templates(tenant_id, branch_id, id);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_session_type_fkey
    FOREIGN KEY (tenant_id, branch_id, session_type_id) REFERENCES session_types(tenant_id, branch_id, id);
