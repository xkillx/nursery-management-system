-- Reverse the funding type unification.
-- Maps unified values back to the old booking-domain values.
UPDATE bookings
SET funding_type = CASE
    WHEN funding_type = 'universal_15'      THEN 'fifteen_hours'
    WHEN funding_type = 'working_parent'    THEN 'thirty_hours'
    WHEN funding_type = 'disadvantaged_2yo' THEN 'two_year_old'
    -- 'custom' was mapped to 'none'; reverse is lossy — leave as 'none'
    ELSE funding_type
END
WHERE funding_type IN ('universal_15', 'working_parent', 'disadvantaged_2yo');
