ALTER TABLE session_types ADD COLUMN flat_fee_minor integer;
ALTER TABLE session_types ADD CONSTRAINT session_types_flat_fee_nonneg CHECK (flat_fee_minor IS NULL OR flat_fee_minor >= 0);
