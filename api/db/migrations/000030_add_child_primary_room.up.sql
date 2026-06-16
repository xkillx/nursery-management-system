ALTER TABLE children
    ADD COLUMN primary_room_id UUID NULL REFERENCES rooms(id) ON DELETE SET NULL;

CREATE INDEX idx_children_primary_room
    ON children (tenant_id, branch_id, primary_room_id)
    WHERE primary_room_id IS NOT NULL;
