ALTER TABLE child_registration_consent_records ADD COLUMN signed_date DATE;
ALTER TABLE child_registration_consent_records ADD COLUMN signer_name TEXT;

CREATE TYPE registration_office_check_status AS ENUM ('unknown', 'complete', 'missing', 'not_applicable');

CREATE TABLE child_registration_office_checklists (
    id UUID NOT NULL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    deposit_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    deposit_paid_date DATE,
    application_date_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    application_date DATE,
    start_date_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    date_left DATE,
    sessions_days_requested_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    sessions_days_requested TEXT,
    term_time_only_space_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    contract_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    contract_date DATE,
    handbook_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    handbook_date DATE,
    red_book_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    red_book_checked_date DATE,
    birth_certificate_passport_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    birth_certificate_passport_checked_date DATE,
    proof_of_address_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    proof_of_address_checked_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE child_registration_completion_attestations ADD COLUMN office_checklist_updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
