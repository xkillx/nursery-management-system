-- Office-Use Checklist: one current checklist per child for manager
-- registration/enrolment office-use tracking, separate from the
-- registration profile and from operational enrollment dates.

-- 1. Enum types for structured checklist fields
CREATE TYPE registration_office_check_status AS ENUM ('unknown', 'complete', 'missing', 'not_applicable');
CREATE TYPE registration_term_time_only_status AS ENUM ('unknown', 'yes', 'no', 'not_applicable');

-- 2. Main checklist table
CREATE TABLE child_registration_office_checklists (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,

    -- Deposit
    deposit_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    deposit_paid_date DATE,

    -- Application date
    application_date_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    application_date DATE,

    -- Start date check
    start_date_status registration_office_check_status NOT NULL DEFAULT 'unknown',

    -- Date left (optional, does not affect completeness)
    date_left DATE,

    -- Sessions/days requested
    sessions_days_requested_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    sessions_days_requested TEXT,

    -- Term-time-only space
    term_time_only_space_status registration_term_time_only_status NOT NULL DEFAULT 'unknown',

    -- Contract/signature
    contract_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    contract_date DATE,

    -- Handbook
    handbook_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    handbook_date DATE,

    -- Red Book
    red_book_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    red_book_checked_date DATE,

    -- Birth certificate / passport
    birth_certificate_passport_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    birth_certificate_passport_checked_date DATE,

    -- Proof of address
    proof_of_address_status registration_office_check_status NOT NULL DEFAULT 'unknown',
    proof_of_address_checked_date DATE,

    -- Free-text notes
    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Tenant/branch scope
    CONSTRAINT croc_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT croc_child_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),

    -- One checklist per child
    CONSTRAINT croc_child_unique UNIQUE (tenant_id, branch_id, child_id),

    -- Structural validation
    CONSTRAINT croc_application_date_required_when_complete
        CHECK (application_date_status != 'complete' OR application_date IS NOT NULL),
    CONSTRAINT croc_sessions_required_when_complete
        CHECK (sessions_days_requested_status != 'complete' OR (sessions_days_requested IS NOT NULL AND sessions_days_requested <> ''))
);
