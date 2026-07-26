-- Add unique constraint on bookings for (tenant_id, branch_id, id)
-- This is required for foreign key references from booking_session_entries
ALTER TABLE bookings ADD CONSTRAINT bookings_scope_id_unique UNIQUE (tenant_id, branch_id, id);

-- Create booking_session_entries table (normalized replacement for bookings.session_entries JSONB)
CREATE TABLE booking_session_entries (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    branch_id       uuid NOT NULL,
    booking_id      uuid NOT NULL,
    day_of_week     integer NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 4),
    session_type_id uuid NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT booking_session_entries_booking_fkey
        FOREIGN KEY (tenant_id, branch_id, booking_id)
        REFERENCES bookings(tenant_id, branch_id, id),
    CONSTRAINT booking_session_entries_session_type_fkey
        FOREIGN KEY (tenant_id, branch_id, session_type_id)
        REFERENCES session_types(tenant_id, branch_id, id),
    CONSTRAINT booking_session_entries_unique_day
        UNIQUE (tenant_id, branch_id, booking_id, day_of_week)
);

CREATE INDEX booking_session_entries_tenant_id ON booking_session_entries(tenant_id);
CREATE INDEX booking_session_entries_branch_id ON booking_session_entries(branch_id);
CREATE INDEX booking_session_entries_by_booking ON booking_session_entries(tenant_id, branch_id, booking_id);

-- Migrate existing JSONB data
INSERT INTO booking_session_entries (id, tenant_id, branch_id, booking_id, day_of_week, session_type_id)
SELECT
    gen_random_uuid(),
    b.tenant_id,
    b.branch_id,
    b.id,
    (entry->>'day_of_week')::int,
    (entry->>'session_type_id')::uuid
FROM bookings b
CROSS JOIN jsonb_array_elements(b.session_entries) AS entry
WHERE b.session_entries IS NOT NULL;
