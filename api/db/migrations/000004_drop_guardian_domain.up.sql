-- 000004_drop_guardian_domain.up.sql
--
-- Replace the two-hop parent-access model with a one-hop model.
--
-- Drops:
--   - guardian_child_links       (active links between guardians and children)
--   - parent_membership_guardians (intermediate table linking parent memberships to guardians)
--   - guardians                  (the guardian entity; no remaining product role)
--
-- Creates:
--   - parent_membership_children (direct mapping from parent memberships to children)
--   - new triggers and indexes mirroring the dropped table's structure
--
-- Cascade model: when a parent membership is ended, all active parent_membership_children
-- rows for that membership are ended with ended_reason_code = 'system_cascade'.

-- Step 1: drop triggers bound to the old tables
DROP TRIGGER IF EXISTS parent_membership_guardians_active_entity_check ON parent_membership_guardians;
DROP TRIGGER IF EXISTS parent_membership_guardians_role_check ON parent_membership_guardians;
DROP TRIGGER IF EXISTS memberships_role_guardian_mapping_check ON memberships;

-- Step 2: drop the old tables (CASCADE so dependent FKs and indices fall with the tables)
DROP TABLE IF EXISTS parent_membership_guardians CASCADE;
DROP TABLE IF EXISTS guardian_child_links CASCADE;
DROP TABLE IF EXISTS guardians CASCADE;

-- Step 3: drop the trigger functions that are no longer needed
DROP FUNCTION IF EXISTS enforce_parent_mapping_active_entities() CASCADE;
DROP FUNCTION IF EXISTS enforce_parent_membership_guardian_role() CASCADE;
DROP FUNCTION IF EXISTS prevent_non_parent_with_active_guardian_mapping() CASCADE;

-- Step 4: create the new parent_membership_children table
CREATE TABLE parent_membership_children (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    membership_id uuid NOT NULL,
    child_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    ended_at timestamp with time zone,
    ended_reason_code lifecycle_reason_code,
    ended_reason_note text,
    CONSTRAINT parent_membership_children_end_reason_check CHECK (
        (((ended_at IS NULL) AND (ended_reason_code IS NULL) AND (ended_reason_note IS NULL)) OR
         ((ended_at IS NOT NULL) AND (ended_reason_code IS NOT NULL) AND
          ((ended_reason_code <> 'other'::lifecycle_reason_code) OR
           ((ended_reason_note IS NOT NULL) AND (btrim(ended_reason_note) <> ''::text)))))
    )
);

-- Step 5: indexes
-- One active mapping per (tenant, branch, membership, child) pair.
CREATE UNIQUE INDEX idx_parent_membership_children_active_pair
    ON parent_membership_children USING btree (tenant_id, branch_id, membership_id, child_id)
    WHERE (ended_at IS NULL);
CREATE INDEX idx_parent_membership_children_child_active
    ON parent_membership_children USING btree (child_id) WHERE (ended_at IS NULL);
CREATE INDEX idx_parent_membership_children_membership_active
    ON parent_membership_children USING btree (membership_id) WHERE (ended_at IS NULL);

-- Step 6: foreign keys
ALTER TABLE ONLY parent_membership_children
    ADD CONSTRAINT parent_membership_children_branch_scope_fkey
    FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);
ALTER TABLE ONLY parent_membership_children
    ADD CONSTRAINT parent_membership_children_membership_scope_fkey
    FOREIGN KEY (tenant_id, branch_id, membership_id) REFERENCES memberships(tenant_id, branch_id, id);
ALTER TABLE ONLY parent_membership_children
    ADD CONSTRAINT parent_membership_children_child_scope_fkey
    FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

-- Step 7: trigger functions
-- Enforce role=parent and is_active=true on the parent membership for a new active mapping.
CREATE FUNCTION enforce_parent_membership_child_role() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    membership_role TEXT;
    membership_is_active BOOLEAN;
BEGIN
    IF NEW.ended_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT role, is_active INTO membership_role, membership_is_active
    FROM memberships
    WHERE id = NEW.membership_id;

    IF membership_role IS NULL THEN
        RAISE EXCEPTION 'parent_membership_children requires valid membership';
    END IF;
    IF membership_role <> 'parent' THEN
        RAISE EXCEPTION 'parent_membership_children requires parent role membership';
    END IF;
    IF membership_is_active IS DISTINCT FROM true THEN
        RAISE EXCEPTION 'parent_membership_children requires active membership';
    END IF;
    RETURN NEW;
END;
$$;

-- Enforce child exists in the same tenant+branch on insert.
CREATE FUNCTION enforce_parent_membership_child_scope() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    child_exists BOOLEAN;
BEGIN
    IF NEW.ended_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM children
        WHERE id = NEW.child_id
          AND tenant_id = NEW.tenant_id
          AND branch_id = NEW.branch_id
    ) INTO child_exists;

    IF NOT child_exists THEN
        RAISE EXCEPTION 'parent_membership_children requires child in same tenant and branch';
    END IF;

    RETURN NEW;
END;
$$;

-- When a parent membership is ended, end all of its active child mappings.
CREATE FUNCTION cascade_parent_membership_child_end() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.ended_at IS NOT NULL
       AND (OLD.ended_at IS NULL OR OLD.ended_at IS DISTINCT FROM NEW.ended_at) THEN
        UPDATE parent_membership_children
        SET ended_at = NEW.ended_at,
            updated_at = now(),
            ended_reason_code = 'system_cascade',
            ended_reason_note = NULL
        WHERE membership_id = NEW.id
          AND ended_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;

-- Step 8: triggers
CREATE TRIGGER parent_membership_children_role_check
    BEFORE INSERT OR UPDATE ON parent_membership_children
    FOR EACH ROW EXECUTE FUNCTION enforce_parent_membership_child_role();

CREATE TRIGGER parent_membership_children_scope_check
    BEFORE INSERT OR UPDATE OF membership_id, child_id, ended_at ON parent_membership_children
    FOR EACH ROW EXECUTE FUNCTION enforce_parent_membership_child_scope();

CREATE TRIGGER memberships_end_cascade_children
    AFTER UPDATE OF ended_at ON memberships
    FOR EACH ROW EXECUTE FUNCTION cascade_parent_membership_child_end();
