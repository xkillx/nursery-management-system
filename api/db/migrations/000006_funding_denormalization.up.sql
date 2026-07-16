-- Add denormalized funding columns to funding_profiles
ALTER TABLE funding_profiles
  ADD COLUMN funding_type text,
  ADD COLUMN funding_model text,
  ADD COLUMN funded_hours_per_week numeric(5,2);

-- Backfill from child_funding_records
UPDATE funding_profiles fp
SET funding_type = fr.funding_type,
    funding_model = fr.funding_model,
    funded_hours_per_week = fr.funded_hours_per_week
FROM child_funding_records fr
WHERE fp.child_id = fr.child_id;

-- Create funding history table
CREATE TABLE child_funding_history (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  branch_id uuid NOT NULL,
  child_id uuid NOT NULL,
  funding_type text,
  funding_model text,
  funded_hours_per_week numeric(5,2),
  funding_start_date date,
  funding_end_date date,
  changed_at timestamptz NOT NULL DEFAULT now(),
  changed_by_user_id uuid NOT NULL
);

-- Add CHECK constraints matching child_funding_records
ALTER TABLE child_funding_history
  ADD CONSTRAINT chk_funding_type
    CHECK (funding_type IS NULL OR funding_type IN ('fifteen_hours', 'thirty_hours')),
  ADD CONSTRAINT chk_funding_model
    CHECK (funding_model IS NULL OR funding_model IN ('term_time', 'stretched'));

-- Add index for history queries
CREATE INDEX idx_funding_history_lookup
  ON child_funding_history (tenant_id, branch_id, child_id, changed_at DESC);
