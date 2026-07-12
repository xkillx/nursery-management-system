-- U1: Add overdue_grace_days to branches
ALTER TABLE branches
    ADD COLUMN overdue_grace_days integer NOT NULL DEFAULT 3;

ALTER TABLE branches
    ADD CONSTRAINT branches_overdue_grace_days_range
    CHECK (overdue_grace_days >= 0 AND overdue_grace_days <= 30);

-- U7: Create invoice_reminder_log for idempotency
CREATE TABLE invoice_reminder_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    invoice_id uuid NOT NULL REFERENCES invoices(id),
    reminder_type text NOT NULL CHECK (reminder_type IN ('due_soon', 'due_today')),
    sent_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (invoice_id, reminder_type, sent_at::date)
);
