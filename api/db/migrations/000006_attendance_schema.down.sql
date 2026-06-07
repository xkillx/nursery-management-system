-- Day 8: Attendance schema rollback

ALTER TABLE attendance_sessions
DROP CONSTRAINT IF EXISTS attendance_sessions_corrected_by_event_fkey,
DROP CONSTRAINT IF EXISTS attendance_sessions_check_out_event_fkey,
DROP CONSTRAINT IF EXISTS attendance_sessions_check_in_event_fkey;

DROP INDEX IF EXISTS idx_attendance_events_child_date;
DROP INDEX IF EXISTS idx_attendance_events_session_time;
DROP INDEX IF EXISTS attendance_events_scope_id_unique;
DROP INDEX IF EXISTS idx_attendance_sessions_open_scope;
DROP INDEX IF EXISTS idx_attendance_sessions_child_date;
DROP INDEX IF EXISTS idx_attendance_sessions_one_open_child;

DROP TABLE IF EXISTS attendance_events;
DROP TABLE IF EXISTS attendance_sessions;
