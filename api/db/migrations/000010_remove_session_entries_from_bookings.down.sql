-- Add session_entries JSONB column back to bookings
ALTER TABLE bookings ADD COLUMN session_entries jsonb;
