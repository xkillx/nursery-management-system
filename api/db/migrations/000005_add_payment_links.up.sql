CREATE TABLE payment_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    invoice_id uuid NOT NULL REFERENCES invoices(id),
    stripe_payment_link_id text NOT NULL,
    stripe_payment_link_url text NOT NULL,
    amount_minor integer NOT NULL,
    currency_code text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_by_membership_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deactivated')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_payment_links_active_per_invoice
    ON payment_links (tenant_id, branch_id, invoice_id)
    WHERE status = 'active';

CREATE INDEX idx_payment_links_invoice_id
    ON payment_links (tenant_id, branch_id, invoice_id);
