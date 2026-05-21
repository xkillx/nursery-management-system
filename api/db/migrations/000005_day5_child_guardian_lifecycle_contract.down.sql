DROP TRIGGER IF EXISTS parent_membership_guardians_active_entity_check ON parent_membership_guardians;
DROP FUNCTION IF EXISTS enforce_parent_mapping_active_entities;

ALTER TABLE parent_membership_guardians
DROP CONSTRAINT IF EXISTS parent_membership_guardians_end_reason_check;

ALTER TABLE parent_membership_guardians
ADD COLUMN ended_reason TEXT;

UPDATE parent_membership_guardians
SET ended_reason = CASE
    WHEN ended_reason_code IS NULL THEN NULL
    WHEN ended_reason_note IS NULL OR btrim(ended_reason_note) = '' THEN ended_reason_code::text
    ELSE ended_reason_code::text || ': ' || ended_reason_note
END
WHERE ended_at IS NOT NULL;

ALTER TABLE parent_membership_guardians
DROP COLUMN IF EXISTS ended_reason_note,
DROP COLUMN IF EXISTS ended_reason_code;

ALTER TABLE guardian_child_links
DROP CONSTRAINT IF EXISTS guardian_child_links_end_reason_check;

ALTER TABLE guardian_child_links
ADD COLUMN ended_reason TEXT;

UPDATE guardian_child_links
SET ended_reason = CASE
    WHEN ended_reason_code IS NULL THEN NULL
    WHEN ended_reason_note IS NULL OR btrim(ended_reason_note) = '' THEN ended_reason_code::text
    ELSE ended_reason_code::text || ': ' || ended_reason_note
END
WHERE ended_at IS NOT NULL;

ALTER TABLE guardian_child_links
DROP COLUMN IF EXISTS ended_reason_note,
DROP COLUMN IF EXISTS ended_reason_code;

ALTER TABLE guardians
DROP CONSTRAINT IF EXISTS guardians_active_consistency_check;

ALTER TABLE guardians
ADD CONSTRAINT guardians_active_consistency_check
CHECK ((is_active = true AND deactivated_at IS NULL) OR (is_active = false AND deactivated_at IS NOT NULL));

ALTER TABLE guardians
DROP COLUMN IF EXISTS deactivation_reason_note,
DROP COLUMN IF EXISTS deactivation_reason_code;

ALTER TABLE guardians
RENAME COLUMN deactivated_at TO ended_at;

ALTER TABLE children
DROP CONSTRAINT IF EXISTS children_active_consistency_check;

ALTER TABLE children
ADD CONSTRAINT children_active_consistency_check
CHECK ((is_active = true AND left_at IS NULL) OR (is_active = false AND left_at IS NOT NULL));

ALTER TABLE children
DROP COLUMN IF EXISTS left_reason_note,
DROP COLUMN IF EXISTS left_reason_code;

DROP INDEX IF EXISTS idx_audit_logs_actor_membership;

ALTER TABLE audit_logs
DROP CONSTRAINT IF EXISTS audit_logs_reason_other_note_check;

ALTER TABLE audit_logs
DROP CONSTRAINT IF EXISTS audit_logs_reason_shape_check;

ALTER TABLE audit_logs
DROP CONSTRAINT IF EXISTS audit_logs_actor_membership_fkey;

ALTER TABLE audit_logs
DROP COLUMN IF EXISTS reason_note,
DROP COLUMN IF EXISTS reason_code,
DROP COLUMN IF EXISTS actor_membership_id;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lifecycle_reason_code') THEN
        DROP TYPE lifecycle_reason_code;
    END IF;
END $$;
