-- Drop history table
DROP TABLE IF EXISTS child_funding_history;

-- Remove denormalized columns from funding_profiles
ALTER TABLE funding_profiles
  DROP COLUMN IF EXISTS funding_type,
  DROP COLUMN IF EXISTS funding_model,
  DROP COLUMN IF EXISTS funded_hours_per_week;
