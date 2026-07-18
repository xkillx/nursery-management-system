ALTER TABLE bookings DROP CONSTRAINT bookings_session_source_check;

ALTER TABLE bookings DROP COLUMN session_entries;

UPDATE bookings SET session_template_id = '00000000-0000-0000-0000-000000000000' WHERE session_template_id IS NULL;

ALTER TABLE bookings ALTER COLUMN session_template_id SET NOT NULL;
