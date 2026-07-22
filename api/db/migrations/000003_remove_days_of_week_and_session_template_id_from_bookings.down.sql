ALTER TABLE bookings ADD COLUMN session_template_id uuid;
ALTER TABLE bookings ADD COLUMN days_of_week integer[] NOT NULL DEFAULT '{}';
