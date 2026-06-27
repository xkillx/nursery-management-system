DROP INDEX IF EXISTS child_booking_pattern_entries_unique_day;

CREATE UNIQUE INDEX child_booking_pattern_entries_unique_day_session
  ON child_booking_pattern_entries (tenant_id, branch_id, pattern_id, day_of_week, session_type_id);
