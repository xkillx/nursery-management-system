ALTER TABLE session_types DROP CONSTRAINT IF EXISTS session_types_flat_fee_nonneg;
ALTER TABLE session_types DROP COLUMN IF EXISTS flat_fee_minor;
ALTER TABLE session_types DROP CONSTRAINT IF EXISTS session_types_kind_check;
ALTER TABLE session_types DROP COLUMN IF EXISTS kind;
