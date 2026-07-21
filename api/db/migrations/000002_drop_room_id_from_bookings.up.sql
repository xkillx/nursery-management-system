ALTER TABLE ONLY bookings
    DROP CONSTRAINT bookings_room_id_fkey;

DROP INDEX idx_bookings_tenant_branch_room_dates;

ALTER TABLE bookings
    DROP COLUMN room_id;