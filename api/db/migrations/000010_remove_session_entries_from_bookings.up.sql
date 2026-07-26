-- Remove session_entries JSONB column from bookings
ALTER TABLE bookings DROP COLUMN IF EXISTS session_entries;
