CREATE TABLE bookings (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    session_template_id uuid NOT NULL,
    room_id uuid NOT NULL,
    days_of_week integer[] NOT NULL,
    effective_start_date date NOT NULL,
    effective_end_date date,
    funding_type text,
    funding_hours_per_week numeric(5,2),
    la_reference text,
    status text NOT NULL DEFAULT 'active',
    booked_by_membership_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT bookings_pkey PRIMARY KEY (id),
    CONSTRAINT bookings_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT bookings_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id),
    CONSTRAINT bookings_child_id_fkey FOREIGN KEY (child_id) REFERENCES children(id),
    CONSTRAINT bookings_session_template_id_fkey FOREIGN KEY (session_template_id) REFERENCES session_templates(id),
    CONSTRAINT bookings_room_id_fkey FOREIGN KEY (room_id) REFERENCES rooms(id),
    CONSTRAINT bookings_status_check CHECK (status IN ('active', 'paused', 'cancelled')),
    CONSTRAINT bookings_funding_type_check CHECK (funding_type IS NULL OR funding_type IN ('none', 'fifteen_hours', 'thirty_hours', 'two_year_old', 'custom')),
    CONSTRAINT bookings_days_of_week_check CHECK (array_length(days_of_week, 1) > 0),
    CONSTRAINT bookings_effective_dates_check CHECK (effective_end_date IS NULL OR effective_end_date >= effective_start_date)
);

CREATE INDEX idx_bookings_tenant_branch_child ON bookings USING btree (tenant_id, branch_id, child_id);
CREATE INDEX idx_bookings_tenant_branch_room_dates ON bookings USING btree (tenant_id, branch_id, room_id, effective_start_date);
CREATE INDEX idx_bookings_tenant_branch_status ON bookings USING btree (tenant_id, branch_id, status);
