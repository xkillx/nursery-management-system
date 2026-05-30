CREATE TABLE payment_attempts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    invoice_id UUID NOT NULL,
    initiated_by_user_id UUID NOT NULL REFERENCES users(id),
    initiated_by_membership_id UUID NOT NULL,
    request_id TEXT,
    status TEXT NOT NULL,
    amount_minor INTEGER NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'GBP',
    stripe_checkout_session_id TEXT,
    stripe_checkout_url TEXT,
    stripe_payment_intent_id TEXT,
    stripe_expires_at TIMESTAMPTZ,
    provider_error_code TEXT,
    provider_error_message TEXT,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_payment_attempts_branch FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT fk_payment_attempts_invoice FOREIGN KEY (tenant_id, branch_id, invoice_id) REFERENCES invoices(tenant_id, branch_id, id),
    CONSTRAINT fk_payment_attempts_membership FOREIGN KEY (tenant_id, branch_id, initiated_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id),
    CONSTRAINT chk_payment_attempts_status CHECK (status IN ('checkout_creation_started', 'checkout_created', 'checkout_creation_failed', 'paid', 'payment_failed', 'cancelled', 'expired')),
    CONSTRAINT chk_payment_attempts_amount_positive CHECK (amount_minor > 0),
    CONSTRAINT chk_payment_attempts_currency_gbp CHECK (currency_code = 'GBP'),
    CONSTRAINT chk_payment_attempts_created_has_session CHECK (status != 'checkout_created' OR (stripe_checkout_session_id IS NOT NULL AND stripe_checkout_session_id != '' AND stripe_checkout_url IS NOT NULL AND stripe_checkout_url != '')),
    CONSTRAINT chk_payment_attempts_failed_has_reason CHECK (status != 'checkout_creation_failed' OR (failure_reason IS NOT NULL AND failure_reason != '') OR (provider_error_code IS NOT NULL AND provider_error_code != '') OR (provider_error_message IS NOT NULL AND provider_error_message != '')),
    CONSTRAINT uq_payment_attempts_scoped_id UNIQUE (tenant_id, branch_id, id)
);

CREATE INDEX idx_payment_attempts_invoice_created ON payment_attempts (tenant_id, branch_id, invoice_id, created_at DESC);
CREATE UNIQUE INDEX uq_payment_attempts_stripe_session_id ON payment_attempts (stripe_checkout_session_id) WHERE stripe_checkout_session_id IS NOT NULL;
CREATE INDEX idx_payment_attempts_open_attempts ON payment_attempts (tenant_id, branch_id, invoice_id, status) WHERE status IN ('checkout_creation_started', 'checkout_created');
