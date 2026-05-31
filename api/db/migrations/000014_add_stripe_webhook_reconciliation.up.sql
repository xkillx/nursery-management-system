-- Stripe Webhook Events: raw inbox of verified provider events
CREATE TABLE stripe_webhook_events (
    id UUID PRIMARY KEY,
    stripe_event_id TEXT NOT NULL UNIQUE,
    event_type TEXT NOT NULL,
    livemode BOOLEAN NOT NULL,
    api_version TEXT,
    provider_created_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    processing_status TEXT NOT NULL,
    processing_reason TEXT,
    request_id TEXT,
    raw_payload JSONB NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_webhook_events_processing_status
        CHECK (processing_status IN ('received', 'processed', 'ignored', 'rejected'))
);

CREATE INDEX idx_stripe_webhook_events_event_type
    ON stripe_webhook_events (event_type, received_at DESC);

CREATE INDEX idx_stripe_webhook_events_processing_status
    ON stripe_webhook_events (processing_status, received_at DESC);


-- Payment Reconciliation Records: manager-facing payment timeline
CREATE TABLE payment_reconciliation_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    invoice_id UUID NOT NULL,
    payment_attempt_id UUID NOT NULL,
    stripe_webhook_event_id UUID NOT NULL,
    stripe_event_id TEXT NOT NULL,
    stripe_event_type TEXT NOT NULL,
    stripe_checkout_session_id TEXT NOT NULL,
    stripe_payment_intent_id TEXT,
    outcome TEXT NOT NULL,
    reason_code TEXT NOT NULL,
    previous_invoice_status TEXT,
    new_invoice_status TEXT,
    attempt_previous_status TEXT,
    attempt_new_status TEXT,
    amount_minor INTEGER,
    currency_code CHAR(3),
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_reconciliation_branch
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT fk_reconciliation_invoice
        FOREIGN KEY (tenant_id, branch_id, invoice_id) REFERENCES invoices(tenant_id, branch_id, id),
    CONSTRAINT fk_reconciliation_attempt
        FOREIGN KEY (tenant_id, branch_id, payment_attempt_id) REFERENCES payment_attempts(tenant_id, branch_id, id),
    CONSTRAINT fk_reconciliation_webhook_event
        FOREIGN KEY (stripe_webhook_event_id) REFERENCES stripe_webhook_events(id),
    CONSTRAINT chk_reconciliation_outcome
        CHECK (outcome IN ('paid', 'payment_failed', 'expired', 'ignored', 'rejected'))
);

CREATE UNIQUE INDEX uq_reconciliation_stripe_event_id
    ON payment_reconciliation_records (stripe_event_id);

CREATE INDEX idx_reconciliation_invoice_created
    ON payment_reconciliation_records (tenant_id, branch_id, invoice_id, created_at DESC);

CREATE INDEX idx_reconciliation_attempt_created
    ON payment_reconciliation_records (tenant_id, branch_id, payment_attempt_id, created_at DESC);
