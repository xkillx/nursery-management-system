CREATE TABLE email_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    idempotency_key TEXT NOT NULL,
    event_type TEXT NOT NULL,
    recipient TEXT NOT NULL,
    recipient_name TEXT,
    subject TEXT NOT NULL,
    template_name TEXT NOT NULL,
    template_version INT NOT NULL DEFAULT 1,
    payload_json JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 8,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error TEXT,
    provider_message_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ,
    CONSTRAINT email_outbox_status_check CHECK (status IN ('pending','processing','sent','failed','dead_letter')),
    CONSTRAINT email_outbox_unique_idempotency UNIQUE (tenant_id, branch_id, idempotency_key)
);

CREATE TABLE email_delivery (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_message_id TEXT NOT NULL,
    email_outbox_id UUID NOT NULL REFERENCES email_outbox(id),
    status TEXT NOT NULL,
    response_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_outbox_pending ON email_outbox (status, next_retry_at) WHERE status = 'pending';
CREATE INDEX idx_email_outbox_tenant ON email_outbox (tenant_id, branch_id);
CREATE INDEX idx_email_delivery_provider_msg ON email_delivery (provider_message_id);
