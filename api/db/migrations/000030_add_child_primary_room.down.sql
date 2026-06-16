DROP INDEX IF EXISTS idx_children_primary_room;

ALTER TABLE children
    DROP COLUMN IF EXISTS primary_room_id;
