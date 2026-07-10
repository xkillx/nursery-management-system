-- Add voided_at and void_reason columns to invoices table.
-- Update status CHECK to include 'void' and add void-shape constraint.
-- Update status transition trigger to allow draft -> void.

ALTER TABLE invoices ADD COLUMN voided_at timestamp with time zone;
ALTER TABLE invoices ADD COLUMN void_reason text;

ALTER TABLE invoices DROP CONSTRAINT invoices_status_valid;
ALTER TABLE invoices ADD CONSTRAINT invoices_status_valid CHECK (status = ANY (ARRAY['draft'::text, 'issued'::text, 'payment_failed'::text, 'paid'::text, 'overdue'::text, 'void'::text]));

ALTER TABLE invoices ADD CONSTRAINT invoices_void_shape CHECK (((status <> 'void'::text) OR ((voided_at IS NOT NULL) AND (void_reason IS NOT NULL) AND (btrim(void_reason) <> ''::text))));

CREATE OR REPLACE FUNCTION enforce_invoice_status_transition() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF OLD.status = NEW.status THEN
            RETURN NEW;
        END IF;

        IF OLD.status = 'paid' THEN
            RAISE EXCEPTION 'invoice % is paid and cannot transition', OLD.id;
        END IF;

        IF NEW.status = 'draft' THEN
            RAISE EXCEPTION 'invoice % cannot transition back to draft', OLD.id;
        END IF;

        IF OLD.status = 'payment_failed' AND NEW.status = 'overdue' THEN
            RAISE EXCEPTION 'invoice % is payment_failed and cannot transition to overdue', OLD.id;
        END IF;

        CASE OLD.status
            WHEN 'draft' THEN
                IF NEW.status NOT IN ('issued', 'void') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from draft to %', OLD.id, NEW.status;
                END IF;
            WHEN 'issued' THEN
                IF NEW.status NOT IN ('overdue', 'payment_failed', 'paid') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from issued to %', OLD.id, NEW.status;
                END IF;
            WHEN 'overdue' THEN
                IF NEW.status NOT IN ('payment_failed', 'paid') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from overdue to %', OLD.id, NEW.status;
                END IF;
            WHEN 'payment_failed' THEN
                IF NEW.status NOT IN ('paid') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from payment_failed to %', OLD.id, NEW.status;
                END IF;
        END CASE;
    END IF;

    RETURN NEW;
END;
$$;
