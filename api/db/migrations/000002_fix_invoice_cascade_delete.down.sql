-- Revert invoice delete cascade changes.

-- 1. Remove ON DELETE CASCADE from invoice_lines foreign key
ALTER TABLE invoice_lines
    DROP CONSTRAINT IF EXISTS invoice_lines_invoice_scope_fkey;

ALTER TABLE invoice_lines
    ADD CONSTRAINT invoice_lines_invoice_scope_fkey
    FOREIGN KEY (tenant_id, branch_id, invoice_id)
    REFERENCES invoices(tenant_id, branch_id, id);

-- 2. Revert trigger to only allow 'draft' status
CREATE OR REPLACE FUNCTION protect_issued_invoice_lines() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    inv_status TEXT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        SELECT status INTO inv_status FROM invoices WHERE id = OLD.invoice_id AND tenant_id = OLD.tenant_id AND branch_id = OLD.branch_id;
    ELSE
        SELECT status INTO inv_status FROM invoices WHERE id = NEW.invoice_id AND tenant_id = NEW.tenant_id AND branch_id = NEW.branch_id;
    END IF;

    IF inv_status IS NOT NULL AND inv_status <> 'draft' THEN
        CASE TG_OP
            WHEN 'INSERT' THEN
                RAISE EXCEPTION 'cannot insert lines for invoice % with status %', NEW.invoice_id, inv_status;
            WHEN 'UPDATE' THEN
                RAISE EXCEPTION 'cannot update lines for invoice % with status %', NEW.invoice_id, inv_status;
            WHEN 'DELETE' THEN
                RAISE EXCEPTION 'cannot delete lines for invoice % with status %', OLD.invoice_id, inv_status;
        END CASE;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;
