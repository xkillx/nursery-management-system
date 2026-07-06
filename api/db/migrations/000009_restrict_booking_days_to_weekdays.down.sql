-- Restore full week support (Mon-Sun) for booking pattern and session template entries.

ALTER TABLE child_booking_pattern_entries
    DROP CONSTRAINT child_booking_pattern_entries_dow_check,
    ADD CONSTRAINT child_booking_pattern_entries_dow_check CHECK (((day_of_week >= 1) AND (day_of_week <= 7)));

ALTER TABLE session_template_entries
    DROP CONSTRAINT session_template_entries_dow_check,
    ADD CONSTRAINT session_template_entries_dow_check CHECK (((day_of_week >= 1) AND (day_of_week <= 7)));
