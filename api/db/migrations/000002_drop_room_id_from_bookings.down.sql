ALTER TABLE bookings
    ADD COLUMN room_id uuid NOT NULL;

CREATE INDEX idx_bookings_tenant_branch_room_dates ON bookings USING btree (tenant_id, branch_id, room_id, effective_start_date);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT bookings_room_id_fkey FOREIGN KEY (room_id) REFERENCES rooms(id);