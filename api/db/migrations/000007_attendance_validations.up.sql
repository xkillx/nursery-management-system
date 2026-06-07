-- Day 9: Attendance validations

-- Replace Day 8 reason constraints with a unified shape constraint
ALTER TABLE attendance_events
DROP CONSTRAINT IF EXISTS attendance_events_correction_reason_check,
DROP CONSTRAINT IF EXISTS attendance_events_other_reason_note_check;

ALTER TABLE attendance_events
ADD CONSTRAINT attendance_events_reason_shape_check CHECK (
    (
        event_type IN ('check_in', 'check_out')
        AND reason_code IS NULL
        AND reason_note IS NULL
    )
    OR
    (
        event_type = 'correction'
        AND reason_code IN ('missed_check_in', 'missed_check_out', 'incorrect_time', 'duplicate_entry', 'other')
        AND (
            reason_code <> 'other'
            OR NULLIF(btrim(reason_note), '') IS NOT NULL
        )
    )
);
