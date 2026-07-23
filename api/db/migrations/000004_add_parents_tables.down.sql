DROP TRIGGER IF EXISTS parent_user_scope_check ON parents;
DROP FUNCTION IF EXISTS enforce_parent_user_scope();
DROP TRIGGER IF EXISTS parent_children_scope_check ON parent_children;
DROP FUNCTION IF EXISTS enforce_parent_children_scope();
DROP TABLE IF EXISTS parent_children;
DROP TABLE IF EXISTS parents;
