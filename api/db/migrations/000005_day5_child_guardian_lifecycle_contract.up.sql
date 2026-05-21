DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lifecycle_reason_code') THEN
        CREATE TYPE lifecycle_reason_code AS ENUM (
            'duplicate_record',
            'entered_in_error',
            'left_nursery',
            'safeguarding_direction',
            'contact_update',
            'access_revoked',
            'other'
        );
    END IF;
END $$;

ALTER TABLE audit_logs
ADD COLUMN actor_membership_id UUID,
ADD COLUMN reason_code lifecycle_reason_code,
ADD COLUMN reason_note TEXT;

ALTER TABLE audit_logs
ADD CONSTRAINT audit_logs_actor_membership_fkey
FOREIGN KEY (actor_membership_id) REFERENCES memberships(id);

ALTER TABLE audit_logs
ADD CONSTRAINT audit_logs_reason_shape_check
CHECK (reason_code IS NOT NULL OR reason_note IS NULL);

ALTER TABLE audit_logs
ADD CONSTRAINT audit_logs_reason_other_note_check
CHECK (
    reason_code IS DISTINCT FROM 'other'::lifecycle_reason_code
    OR (reason_note IS NOT NULL AND btrim(reason_note) <> '')
);

CREATE INDEX idx_audit_logs_actor_membership ON audit_logs (actor_membership_id) WHERE actor_membership_id IS NOT NULL;

ALTER TABLE children
ADD COLUMN left_reason_code lifecycle_reason_code,
ADD COLUMN left_reason_note TEXT;

UPDATE children
SET left_reason_code = 'left_nursery'::lifecycle_reason_code
WHERE is_active = false AND left_reason_code IS NULL;

ALTER TABLE children
DROP CONSTRAINT IF EXISTS children_active_consistency_check;

ALTER TABLE children
ADD CONSTRAINT children_active_consistency_check
CHECK (
    (
        is_active = true
        AND left_at IS NULL
        AND left_reason_code IS NULL
        AND left_reason_note IS NULL
    )
    OR
    (
        is_active = false
        AND left_at IS NOT NULL
        AND left_reason_code IS NOT NULL
        AND (
            left_reason_code <> 'other'::lifecycle_reason_code
            OR (left_reason_note IS NOT NULL AND btrim(left_reason_note) <> '')
        )
    )
);

ALTER TABLE guardians
RENAME COLUMN ended_at TO deactivated_at;

ALTER TABLE guardians
ADD COLUMN deactivation_reason_code lifecycle_reason_code,
ADD COLUMN deactivation_reason_note TEXT;

UPDATE guardians
SET deactivation_reason_code = 'access_revoked'::lifecycle_reason_code
WHERE is_active = false AND deactivation_reason_code IS NULL;

ALTER TABLE guardians
DROP CONSTRAINT IF EXISTS guardians_active_consistency_check;

ALTER TABLE guardians
ADD CONSTRAINT guardians_active_consistency_check
CHECK (
    (
        is_active = true
        AND deactivated_at IS NULL
        AND deactivation_reason_code IS NULL
        AND deactivation_reason_note IS NULL
    )
    OR
    (
        is_active = false
        AND deactivated_at IS NOT NULL
        AND deactivation_reason_code IS NOT NULL
        AND (
            deactivation_reason_code <> 'other'::lifecycle_reason_code
            OR (deactivation_reason_note IS NOT NULL AND btrim(deactivation_reason_note) <> '')
        )
    )
);

ALTER TABLE guardian_child_links
ADD COLUMN ended_reason_code lifecycle_reason_code,
ADD COLUMN ended_reason_note TEXT;

UPDATE guardian_child_links
SET ended_reason_code = 'other'::lifecycle_reason_code,
    ended_reason_note = ended_reason
WHERE ended_at IS NOT NULL
  AND ended_reason IS NOT NULL
  AND ended_reason_code IS NULL;

UPDATE guardian_child_links
SET ended_reason_code = 'access_revoked'::lifecycle_reason_code
WHERE ended_at IS NOT NULL
  AND ended_reason IS NULL
  AND ended_reason_code IS NULL;

ALTER TABLE guardian_child_links
DROP COLUMN ended_reason;

ALTER TABLE guardian_child_links
ADD CONSTRAINT guardian_child_links_end_reason_check
CHECK (
    (
        ended_at IS NULL
        AND ended_reason_code IS NULL
        AND ended_reason_note IS NULL
    )
    OR
    (
        ended_at IS NOT NULL
        AND ended_reason_code IS NOT NULL
        AND (
            ended_reason_code <> 'other'::lifecycle_reason_code
            OR (ended_reason_note IS NOT NULL AND btrim(ended_reason_note) <> '')
        )
    )
);

ALTER TABLE parent_membership_guardians
ADD COLUMN ended_reason_code lifecycle_reason_code,
ADD COLUMN ended_reason_note TEXT;

UPDATE parent_membership_guardians
SET ended_reason_code = 'other'::lifecycle_reason_code,
    ended_reason_note = ended_reason
WHERE ended_at IS NOT NULL
  AND ended_reason IS NOT NULL
  AND ended_reason_code IS NULL;

UPDATE parent_membership_guardians
SET ended_reason_code = 'access_revoked'::lifecycle_reason_code
WHERE ended_at IS NOT NULL
  AND ended_reason IS NULL
  AND ended_reason_code IS NULL;

ALTER TABLE parent_membership_guardians
DROP COLUMN ended_reason;

ALTER TABLE parent_membership_guardians
ADD CONSTRAINT parent_membership_guardians_end_reason_check
CHECK (
    (
        ended_at IS NULL
        AND ended_reason_code IS NULL
        AND ended_reason_note IS NULL
    )
    OR
    (
        ended_at IS NOT NULL
        AND ended_reason_code IS NOT NULL
        AND (
            ended_reason_code <> 'other'::lifecycle_reason_code
            OR (ended_reason_note IS NOT NULL AND btrim(ended_reason_note) <> '')
        )
    )
);

CREATE OR REPLACE FUNCTION enforce_parent_mapping_active_entities()
RETURNS trigger AS $$
DECLARE
    membership_active BOOLEAN;
    guardian_active BOOLEAN;
BEGIN
    IF NEW.ended_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT is_active INTO membership_active
    FROM memberships
    WHERE id = NEW.membership_id
      AND tenant_id = NEW.tenant_id
      AND branch_id = NEW.branch_id;

    IF membership_active IS DISTINCT FROM true THEN
        RAISE EXCEPTION 'parent_membership_guardians requires active membership';
    END IF;

    SELECT is_active INTO guardian_active
    FROM guardians
    WHERE id = NEW.guardian_id
      AND tenant_id = NEW.tenant_id
      AND branch_id = NEW.branch_id;

    IF guardian_active IS DISTINCT FROM true THEN
        RAISE EXCEPTION 'parent_membership_guardians requires active guardian';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER parent_membership_guardians_active_entity_check
BEFORE INSERT OR UPDATE OF membership_id, guardian_id, ended_at ON parent_membership_guardians
FOR EACH ROW
EXECUTE FUNCTION enforce_parent_mapping_active_entities();
