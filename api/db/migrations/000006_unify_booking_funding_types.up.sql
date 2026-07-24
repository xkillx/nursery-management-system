-- Unify booking funding type values to match the funding module's vocabulary.
-- Maps old booking-domain values to the canonical funding-module values.
UPDATE bookings
SET funding_type = CASE
    WHEN funding_type = 'fifteen_hours' THEN 'universal_15'
    WHEN funding_type = 'thirty_hours'  THEN 'working_parent'
    WHEN funding_type = 'two_year_old'  THEN 'disadvantaged_2yo'
    WHEN funding_type = 'custom'        THEN 'none'
    ELSE funding_type
END
WHERE funding_type IN ('fifteen_hours', 'thirty_hours', 'two_year_old', 'custom');
