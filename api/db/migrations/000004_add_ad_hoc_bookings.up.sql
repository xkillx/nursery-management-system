CREATE TABLE ad_hoc_bookings (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    calendar_date date NOT NULL,
    session_type_id uuid NOT NULL,
    booked_by_membership_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ad_hoc_bookings_pkey PRIMARY KEY (id),
    CONSTRAINT ad_hoc_bookings_status_check CHECK (status IN ('active', 'cancelled')),
    CONSTRAINT ad_hoc_bookings_child_id_fkey FOREIGN KEY (child_id) REFERENCES children(id),
    CONSTRAINT ad_hoc_bookings_session_type_id_fkey FOREIGN KEY (session_type_id) REFERENCES session_types(id),
    CONSTRAINT ad_hoc_bookings_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE INDEX idx_ad_hoc_bookings_tenant_branch_child_date
    ON ad_hoc_bookings(tenant_id, branch_id, child_id, calendar_date);
CREATE INDEX idx_ad_hoc_bookings_tenant_branch_status
    ON ad_hoc_bookings(tenant_id, branch_id, status);

ALTER TABLE branches ADD COLUMN ad_hoc_rate_multiplier numeric(4,2) NOT NULL DEFAULT 1.50;
