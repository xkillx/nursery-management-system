UPDATE children SET core_hourly_rate_minor = 0 WHERE core_hourly_rate_minor IS NULL;

ALTER TABLE children
  DROP CONSTRAINT IF EXISTS core_hourly_rate_minor_check;

ALTER TABLE children
  ALTER COLUMN core_hourly_rate_minor SET NOT NULL;

ALTER TABLE children
  ADD CONSTRAINT children_core_hourly_rate_minor_check
    CHECK (core_hourly_rate_minor >= 0);
