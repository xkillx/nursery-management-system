-- Remove booking_session_entries table
DROP TABLE IF EXISTS booking_session_entries;

-- Remove unique constraint from bookings
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_scope_id_unique;
