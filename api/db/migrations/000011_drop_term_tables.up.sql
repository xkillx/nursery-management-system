DROP TABLE IF EXISTS term_schedule_change CASCADE;
DROP TABLE IF EXISTS term CASCADE;
DROP TABLE IF EXISTS child_booking_pattern_entries CASCADE;
DROP TABLE IF EXISTS child_booking_patterns CASCADE;
ALTER TABLE children DROP COLUMN IF EXISTS current_term_id;
