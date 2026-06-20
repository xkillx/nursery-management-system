-- 000004_drop_guardian_domain.down.sql
--
-- This migration is IRREVERSIBLE.
--
-- The three dropped guardian tables (guardians, guardian_child_links,
-- parent_membership_guardians) are not recreated on down. The pilot
-- is not live, and the data in those tables has been intentionally
-- discarded. The new parent_membership_children table is the source
-- of truth for parent-to-child access.
--
-- Restore from a backup taken before this migration was applied if
-- the old schema is required.

DO $$
BEGIN
    RAISE EXCEPTION '000004_drop_guardian_domain is irreversible: the dropped guardian tables are not recreated. Restore from a backup taken before this migration was applied.';
END
$$;
