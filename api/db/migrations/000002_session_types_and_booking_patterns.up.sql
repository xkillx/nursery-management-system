CREATE TABLE session_types (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name text NOT NULL,
    start_time time NOT NULL,
    end_time time NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT session_types_time_check CHECK ((start_time < end_time))
);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_pkey PRIMARY KEY (id);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE UNIQUE INDEX session_types_active_name_unique ON session_types USING btree (tenant_id, branch_id, name) WHERE (is_active = true);

CREATE INDEX session_types_active_by_branch ON session_types USING btree (tenant_id, branch_id) WHERE (is_active = true);

CREATE INDEX session_types_branch_id ON session_types USING btree (branch_id);

CREATE INDEX session_types_tenant_id ON session_types USING btree (tenant_id);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_branch_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

CREATE TABLE child_booking_patterns (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    effective_from date NOT NULL,
    effective_to date,
    is_current boolean GENERATED ALWAYS AS ((effective_to IS NULL)) STORED NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_booking_patterns_dates_check CHECK (((effective_to IS NULL) OR (effective_to >= effective_from)))
);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE UNIQUE INDEX child_booking_patterns_one_open_per_child ON child_booking_patterns USING btree (tenant_id, branch_id, child_id) WHERE is_current;

CREATE INDEX child_booking_patterns_by_child ON child_booking_patterns USING btree (tenant_id, branch_id, child_id, effective_from DESC);

CREATE INDEX child_booking_patterns_branch_id ON child_booking_patterns USING btree (branch_id);

CREATE INDEX child_booking_patterns_tenant_id ON child_booking_patterns USING btree (tenant_id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

CREATE TABLE child_booking_pattern_entries (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    pattern_id uuid NOT NULL,
    day_of_week integer NOT NULL,
    session_type_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_booking_pattern_entries_dow_check CHECK (((day_of_week >= 1) AND (day_of_week <= 7)))
);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE UNIQUE INDEX child_booking_pattern_entries_unique_day_session ON child_booking_pattern_entries USING btree (tenant_id, branch_id, pattern_id, day_of_week, session_type_id);

CREATE INDEX child_booking_pattern_entries_by_pattern ON child_booking_pattern_entries USING btree (tenant_id, branch_id, pattern_id, day_of_week);

CREATE INDEX child_booking_pattern_entries_branch_id ON child_booking_pattern_entries USING btree (branch_id);

CREATE INDEX child_booking_pattern_entries_tenant_id ON child_booking_pattern_entries USING btree (tenant_id);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_pattern_fkey FOREIGN KEY (tenant_id, branch_id, pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_session_type_fkey FOREIGN KEY (tenant_id, branch_id, session_type_id) REFERENCES session_types(tenant_id, branch_id, id);
