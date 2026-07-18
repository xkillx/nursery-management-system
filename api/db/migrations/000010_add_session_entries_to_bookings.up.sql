ALTER TABLE bookings ADD COLUMN session_entries jsonb;

ALTER TABLE bookings ALTER COLUMN session_template_id DROP NOT NULL;

ALTER TABLE bookings ADD CONSTRAINT bookings_session_source_check
  CHECK (session_template_id IS NOT NULL OR session_entries IS NOT NULL);
