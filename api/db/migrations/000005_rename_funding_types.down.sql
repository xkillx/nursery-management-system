-- Remove EYPP and DAF columns
ALTER TABLE child_funding_records DROP COLUMN IF EXISTS eypp_eligible;
ALTER TABLE child_funding_records DROP COLUMN IF EXISTS daf_eligible;

-- Restore CHECK constraints to original values
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_funding_type_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_funding_type_check
    CHECK (funding_type IN ('none', 'fifteen_hours', 'thirty_hours', 'two_year_old', 'custom', 'unknown'));

ALTER TABLE child_funding_history DROP CONSTRAINT IF EXISTS child_funding_history_funding_type_check;
ALTER TABLE child_funding_history ADD CONSTRAINT child_funding_history_funding_type_check
    CHECK (funding_type IN ('none', 'fifteen_hours', 'thirty_hours', 'two_year_old', 'custom', 'unknown'));

ALTER TABLE child_funding_records DROP CONSTRAINT IF EXISTS child_funding_records_funding_type_check;
ALTER TABLE child_funding_records ADD CONSTRAINT child_funding_records_funding_type_check
    CHECK (funding_type IN ('none', 'fifteen_hours', 'thirty_hours', 'two_year_old', 'custom', 'unknown'));

-- Reverse data migration
UPDATE child_funding_history SET funding_type = CASE
    WHEN funding_type = 'universal_15' THEN 'fifteen_hours'
    WHEN funding_type = 'working_parent' THEN 'thirty_hours'
    WHEN funding_type = 'disadvantaged_2yo' THEN 'two_year_old'
    ELSE funding_type
END;

UPDATE child_funding_records SET funding_type = CASE
    WHEN funding_type = 'universal_15' THEN 'fifteen_hours'
    WHEN funding_type = 'working_parent' THEN 'thirty_hours'
    WHEN funding_type = 'disadvantaged_2yo' THEN 'two_year_old'
    ELSE funding_type
END;
