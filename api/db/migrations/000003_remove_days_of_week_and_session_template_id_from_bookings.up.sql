ALTER TABLE bookings DROP COLUMN IF EXISTS days_of_week;
ALTER TABLE bookings DROP COLUMN IF EXISTS session_template_id;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_session_source_check;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_days_of_week_check;
