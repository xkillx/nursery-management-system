-- Parents table: first-class entity for parent/guardian data.
CREATE TABLE parents (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    first_name text NOT NULL,
    last_name text,
    email text,
    phone text,
    address_line1 text,
    address_line2 text,
    address_city text,
    address_postcode text,
    relationship_to_child text,
    has_parental_responsibility boolean DEFAULT false NOT NULL,
    can_pick_up boolean DEFAULT false NOT NULL,
    is_emergency_contact boolean DEFAULT false NOT NULL,
    notes text,
    user_id uuid,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT parents_first_name_check CHECK ((btrim(first_name) <> ''::text))
);

-- Parent-children link table with soft-delete lifecycle.
CREATE TABLE parent_children (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    parent_id uuid NOT NULL,
    child_id uuid NOT NULL,
    ended_at timestamp with time zone,
    ended_reason_code lifecycle_reason_code,
    ended_reason_note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT parent_children_end_reason_check CHECK (
        (((ended_at IS NULL) AND (ended_reason_code IS NULL) AND (ended_reason_note IS NULL)) OR
         ((ended_at IS NOT NULL) AND (ended_reason_code IS NOT NULL) AND
          ((ended_reason_code <> 'other'::lifecycle_reason_code) OR
           ((ended_reason_note IS NOT NULL) AND (btrim(ended_reason_note) <> ''::text)))))
    )
);

-- Primary keys
ALTER TABLE ONLY parents
    ADD CONSTRAINT parents_pkey PRIMARY KEY (id);

ALTER TABLE ONLY parent_children
    ADD CONSTRAINT parent_children_pkey PRIMARY KEY (id);

-- Unique constraint on parents(tenant_id, branch_id, id) to support FK from parent_children.
CREATE UNIQUE INDEX idx_parents_tenant_branch_id ON parents USING btree (tenant_id, branch_id, id);

-- Foreign keys
ALTER TABLE ONLY parents
    ADD CONSTRAINT parents_tenant_scope_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

ALTER TABLE ONLY parents
    ADD CONSTRAINT parents_branch_scope_fkey
    FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id) ON DELETE CASCADE;

ALTER TABLE ONLY parents
    ADD CONSTRAINT parents_user_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE ONLY parent_children
    ADD CONSTRAINT parent_children_tenant_scope_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

ALTER TABLE ONLY parent_children
    ADD CONSTRAINT parent_children_branch_scope_fkey
    FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id) ON DELETE CASCADE;

ALTER TABLE ONLY parent_children
    ADD CONSTRAINT parent_children_parent_scope_fkey
    FOREIGN KEY (tenant_id, branch_id, parent_id) REFERENCES parents(tenant_id, branch_id, id) ON DELETE CASCADE;

ALTER TABLE ONLY parent_children
    ADD CONSTRAINT parent_children_child_scope_fkey
    FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id) ON DELETE CASCADE;

-- Unique constraint: one active link per parent-child pair per tenant+branch.
CREATE UNIQUE INDEX idx_parent_children_active_pair
    ON parent_children USING btree (tenant_id, branch_id, parent_id, child_id)
    WHERE (ended_at IS NULL);

-- Indexes for common lookups.
CREATE INDEX idx_parents_tenant_branch ON parents USING btree (tenant_id, branch_id);
CREATE INDEX idx_parents_user ON parents USING btree (user_id) WHERE (user_id IS NOT NULL);
CREATE INDEX idx_parent_children_parent_active ON parent_children USING btree (parent_id) WHERE (ended_at IS NULL);
CREATE INDEX idx_parent_children_child_active ON parent_children USING btree (child_id) WHERE (ended_at IS NULL);

-- Trigger: enforce parent and child are in the same tenant+branch.
CREATE FUNCTION enforce_parent_children_scope() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    parent_exists BOOLEAN;
    child_exists BOOLEAN;
BEGIN
    IF NEW.ended_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM parents
        WHERE id = NEW.parent_id
          AND tenant_id = NEW.tenant_id
          AND branch_id = NEW.branch_id
    ) INTO parent_exists;

    IF NOT parent_exists THEN
        RAISE EXCEPTION 'parent_children requires parent in same tenant and branch';
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM children
        WHERE id = NEW.child_id
          AND tenant_id = NEW.tenant_id
          AND branch_id = NEW.branch_id
    ) INTO child_exists;

    IF NOT child_exists THEN
        RAISE EXCEPTION 'parent_children requires child in same tenant and branch';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER parent_children_scope_check
    BEFORE INSERT OR UPDATE OF parent_id, child_id, ended_at ON parent_children
    FOR EACH ROW EXECUTE FUNCTION enforce_parent_children_scope();

-- Trigger: enforce user_id scope (if set, user must belong to same tenant).
CREATE FUNCTION enforce_parent_user_scope() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    user_exists BOOLEAN;
BEGIN
    IF NEW.user_id IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM users
        WHERE id = NEW.user_id
          AND tenant_id = NEW.tenant_id
    ) INTO user_exists;

    IF NOT user_exists THEN
        RAISE EXCEPTION 'parents.user_id must belong to the same tenant';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER parent_user_scope_check
    BEFORE INSERT OR UPDATE OF user_id ON parents
    FOR EACH ROW EXECUTE FUNCTION enforce_parent_user_scope();
