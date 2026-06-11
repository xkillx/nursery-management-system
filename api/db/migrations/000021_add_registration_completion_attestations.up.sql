CREATE TABLE child_registration_completion_attestations (
    id uuid NOT NULL PRIMARY KEY,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,

    consent_record_id uuid,
    profile_updated_at TIMESTAMPTZ NOT NULL,
    office_checklist_updated_at TIMESTAMPTZ NOT NULL,

    attested_by_user_id uuid NOT NULL,
    attested_by_membership_id uuid NOT NULL,
    attested_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    request_id TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_completion_attestations_child
    ON child_registration_completion_attestations (tenant_id, branch_id, child_id, attested_at DESC);
