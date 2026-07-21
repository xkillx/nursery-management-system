-- Rename funding type enum values in child_funding_records
UPDATE child_funding_records SET funding_type = CASE
    WHEN funding_type = 'fifteen_hours' THEN 'universal_15'
    WHEN funding_type = 'thirty_hours' THEN 'working_parent'
    WHEN funding_type = 'two_year_old' THEN 'disadvantaged_2yo'
    ELSE funding_type
END;

-- Rename funding type enum values in child_funding_history
UPDATE child_funding_history SET funding_type = CASE
    WHEN funding_type = 'fifteen_hours' THEN 'universal_15'
    WHEN funding_type = 'thirty_hours' THEN 'working_parent'
    WHEN funding_type = 'two_year_old' THEN 'disadvantaged_2yo'
    ELSE funding_type
END;

-- Update CHECK constraint on child_funding_records
ALTER TABLE child_funding_records DROP CONSTRAINT IF EXISTS child_funding_records_funding_type_check;
ALTER TABLE child_funding_records ADD CONSTRAINT child_funding_records_funding_type_check
    CHECK (funding_type IN ('none', 'universal_15', 'working_parent', 'working_parent_under_3', 'disadvantaged_2yo', 'unknown'));

-- Update CHECK constraint on child_funding_history
ALTER TABLE child_funding_history DROP CONSTRAINT IF EXISTS child_funding_history_funding_type_check;
ALTER TABLE child_funding_history ADD CONSTRAINT child_funding_history_funding_type_check
    CHECK (funding_type IN ('none', 'universal_15', 'working_parent', 'working_parent_under_3', 'disadvantaged_2yo', 'unknown'));

-- Update CHECK constraint on bookings if it exists
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_funding_type_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_funding_type_check
    CHECK (funding_type IN ('none', 'universal_15', 'working_parent', 'working_parent_under_3', 'disadvantaged_2yo', 'unknown'));

-- Add EYPP and DAF boolean columns (schema-ready for future)
ALTER TABLE child_funding_records ADD COLUMN eypp_eligible BOOL NOT NULL DEFAULT false;
ALTER TABLE child_funding_records ADD COLUMN daf_eligible BOOL NOT NULL DEFAULT false;
