DROP INDEX IF EXISTS idx_branches_core_hourly_rate;

ALTER TABLE branches
DROP CONSTRAINT IF EXISTS branches_core_hourly_rate_positive_check;

ALTER TABLE branches
DROP COLUMN IF EXISTS core_hourly_rate_minor;
