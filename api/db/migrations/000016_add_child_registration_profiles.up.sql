-- Registration/Enrolment Profile: one current manager-editable profile per child,
-- repeated profile contacts, and collection password metadata.

-- 1. Enum types for structured status fields
CREATE TYPE registration_yes_no_unknown AS ENUM ('unknown', 'no', 'yes');
CREATE TYPE registration_immunisation_status AS ENUM ('unknown', 'up_to_date', 'refused', 'partial', 'not_recorded');
CREATE TYPE registration_contact_type AS ENUM ('parent_carer', 'emergency_contact', 'authorised_collector');

-- 2. Main profile table
CREATE TABLE child_registration_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,

    -- Demographics / home
    sex TEXT,
    religion TEXT,
    ethnic_origin TEXT,
    first_language TEXT,
    other_languages TEXT[] NOT NULL DEFAULT '{}',
    home_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    home_postcode TEXT,
    home_telephone TEXT,

    -- Disability / access
    disability_status registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    disability_notes TEXT,
    access_requirements TEXT,

    -- Medical / dietary
    medical_conditions_status registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    medical_conditions_notes TEXT,
    prescribed_medication_status registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    medication_notes TEXT,
    immunisation_status registration_immunisation_status NOT NULL DEFAULT 'unknown',
    immunisation_country TEXT,
    illness_diagnosis_history TEXT,
    dietary_requirements_status registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    dietary_requirements_notes TEXT,
    dietary_side_effects TEXT,

    -- Health contacts
    doctor_name TEXT,
    doctor_address TEXT,
    doctor_phone TEXT,
    health_visitor_name TEXT,
    health_visitor_address TEXT,
    health_visitor_phone TEXT,

    -- Social / development
    social_services_status registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    social_services_notes TEXT,
    social_worker_contact_details TEXT,
    concern_walking registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_speech_language registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_hearing registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_sight registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_emotional_wellbeing registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_behaviour registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    professional_referrals JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Parental / family
    parental_responsibility_notes TEXT,

    -- Collection
    over18_collection_acknowledged BOOLEAN NOT NULL DEFAULT false,
    collection_password_hash TEXT,
    collection_password_updated_at TIMESTAMPTZ,
    collection_password_updated_by_user_id UUID REFERENCES users(id),
    collection_password_updated_by_membership_id UUID REFERENCES memberships(id),

    -- Funding support
    benefits_contribute_to_fees registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    working_tax_credit registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_parent registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_nursery registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    funding_3yo_term_time registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    funding_2yo_term_time registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    funding_support_notes TEXT,

    -- Routine / GDPR declaration
    routine_care_notes TEXT,
    gdpr_declared_by_name TEXT,
    gdpr_declared_at TIMESTAMPTZ,
    gdpr_declaration_date DATE,

    -- Section review booleans
    demographics_home_reviewed BOOLEAN NOT NULL DEFAULT false,
    medical_dietary_reviewed BOOLEAN NOT NULL DEFAULT false,
    health_contacts_reviewed BOOLEAN NOT NULL DEFAULT false,
    social_development_reviewed BOOLEAN NOT NULL DEFAULT false,
    parent_responsibility_reviewed BOOLEAN NOT NULL DEFAULT false,
    emergency_collection_reviewed BOOLEAN NOT NULL DEFAULT false,
    funding_support_reviewed BOOLEAN NOT NULL DEFAULT false,
    routine_care_reviewed BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Foreign keys
    CONSTRAINT child_registration_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_registration_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_registration_profiles_scope_unique
        UNIQUE (tenant_id, branch_id, child_id),

    -- JSON type checks
    CONSTRAINT child_registration_profiles_home_address_is_object
        CHECK (jsonb_typeof(home_address) = 'object'),
    CONSTRAINT child_registration_profiles_referrals_is_array
        CHECK (jsonb_typeof(professional_referrals) = 'array'),

    -- Collection password consistency: if hash is set, metadata must be present
    CONSTRAINT child_registration_profiles_password_consistency
        CHECK (
            (collection_password_hash IS NULL AND collection_password_updated_at IS NULL AND collection_password_updated_by_user_id IS NULL AND collection_password_updated_by_membership_id IS NULL)
            OR
            (collection_password_hash IS NOT NULL AND collection_password_updated_at IS NOT NULL AND collection_password_updated_by_user_id IS NOT NULL AND collection_password_updated_by_membership_id IS NOT NULL)
        )
);

CREATE UNIQUE INDEX idx_child_registration_profiles_scope_child
    ON child_registration_profiles (tenant_id, branch_id, child_id);

-- Unique constraint for contact FK reference
ALTER TABLE child_registration_profiles ADD CONSTRAINT child_registration_profiles_scope_id_unique
    UNIQUE (tenant_id, branch_id, id);

-- 3. Repeated registration contacts
CREATE TABLE child_registration_contacts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    profile_id UUID NOT NULL,
    child_id UUID NOT NULL,

    contact_type registration_contact_type NOT NULL,
    sort_order INTEGER NOT NULL,
    full_name TEXT NOT NULL CHECK (btrim(full_name) <> ''),
    relationship_to_child TEXT,
    address JSONB NOT NULL DEFAULT '{}'::jsonb,
    telephone TEXT,
    email TEXT,
    work_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    has_parental_responsibility BOOLEAN,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Foreign keys
    CONSTRAINT child_registration_contacts_profile_fkey
        FOREIGN KEY (tenant_id, branch_id, profile_id)
        REFERENCES child_registration_profiles(tenant_id, branch_id, id)
        ON DELETE CASCADE,
    CONSTRAINT child_registration_contacts_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),

    -- JSON checks
    CONSTRAINT child_registration_contacts_address_is_object
        CHECK (jsonb_typeof(address) = 'object'),
    CONSTRAINT child_registration_contacts_work_address_is_object
        CHECK (jsonb_typeof(work_address) = 'object'),

    -- Sort order must be non-negative
    CONSTRAINT child_registration_contacts_sort_order_check
        CHECK (sort_order >= 0),

    -- One unique sort order per contact type per profile
    CONSTRAINT child_registration_contacts_type_sort_unique
        UNIQUE (tenant_id, branch_id, profile_id, contact_type, sort_order)
);

CREATE INDEX idx_child_registration_contacts_profile_type
    ON child_registration_contacts (tenant_id, branch_id, profile_id, contact_type, sort_order);
