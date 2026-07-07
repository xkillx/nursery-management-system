CREATE TABLE hourly_bookings (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL REFERENCES children(id),
    calendar_date DATE NOT NULL,
    start_time_minutes INT NOT NULL CHECK (start_time_minutes >= 0 AND start_time_minutes <= 1439),
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    session_type_id UUID NULL REFERENCES session_types(id),
    booked_by_membership_id UUID NOT NULL REFERENCES memberships(id),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX hourly_bookings_unique_slot
    ON hourly_bookings (tenant_id, branch_id, child_id, calendar_date, start_time_minutes);

CREATE INDEX hourly_bookings_child_date_idx
    ON hourly_bookings (tenant_id, branch_id, child_id, calendar_date);
