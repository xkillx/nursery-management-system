-- Day 9: Attendance validations rollback

ALTER TABLE attendance_events
DROP CONSTRAINT IF EXISTS attendance_events_reason_shape_check;

-- Restore Day 8 constraints
ALTER TABLE attendance_events
ADD CONSTRAINT attendance_events_correction_reason_check
    CHECK (event_type <> 'correction' OR reason_code IN ('missed_check_in', 'missed_check_out', 'incorrect_time', 'duplicate_entry', 'other'));

ALTER TABLE attendance_events
ADD CONSTRAINT attendance_events_other_reason_note_check
    CHECK (reason_code <> 'other' OR NULLIF(reason_note, '') IS NOT NULL);
