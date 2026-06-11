ALTER TABLE children
  DROP CONSTRAINT IF EXISTS children_core_hourly_rate_minor_check;

ALTER TABLE children
  ALTER COLUMN core_hourly_rate_minor DROP NOT NULL;

ALTER TABLE children
  ADD CONSTRAINT core_hourly_rate_minor_check
    CHECK (core_hourly_rate_minor IS NULL OR core_hourly_rate_minor >= 0);
