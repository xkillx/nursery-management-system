-- 1. Drop the registration tables. No data preservation.
DROP TABLE IF EXISTS child_registration_contacts CASCADE;
DROP TABLE IF EXISTS child_registration_consent_records CASCADE;
DROP TABLE IF EXISTS child_registration_profiles CASCADE;
DROP TABLE IF EXISTS child_registration_completion_attestations CASCADE;
DROP TYPE  IF EXISTS registration_yes_no_unknown CASCADE;
DROP TYPE  IF EXISTS registration_immunisation_status CASCADE;
DROP TYPE  IF EXISTS registration_contact_type CASCADE;

-- 2. Drop the operational columns we are moving off `children`.
ALTER TABLE children
    DROP COLUMN IF EXISTS primary_room_id,
    DROP COLUMN IF EXISTS core_hourly_rate_minor,
    DROP COLUMN IF EXISTS left_at,
    DROP COLUMN IF EXISTS left_reason_code,
    DROP COLUMN IF EXISTS left_reason_note;

-- 3. New child_profiles table (1:1 with children).
CREATE TABLE child_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    -- Demographics / home
    sex TEXT,
    religion TEXT,
    ethnic_origin TEXT,
    first_language TEXT,
    other_languages TEXT,
    home_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    home_postcode TEXT,
    home_telephone TEXT,

    -- Disability / access
    disability_status TEXT NOT NULL DEFAULT 'unknown',  -- 'unknown' | 'no' | 'yes'
    disability_notes TEXT,
    access_requirements TEXT,

    -- Routine care
    routine_care_notes TEXT,

    -- GDPR declaration (paper form metadata)
    gdpr_declared_by_name TEXT,
    gdpr_declared_at TIMESTAMPTZ,
    gdpr_declaration_date DATE,

    -- Registration date captured at intake (paper form "today" the family registered)
    registration_date DATE,

    -- Section review booleans
    demographics_home_reviewed BOOLEAN NOT NULL DEFAULT false,
    medical_dietary_reviewed  BOOLEAN NOT NULL DEFAULT false,
    health_contacts_reviewed  BOOLEAN NOT NULL DEFAULT false,
    social_development_reviewed BOOLEAN NOT NULL DEFAULT false,
    parent_responsibility_reviewed BOOLEAN NOT NULL DEFAULT false,
    emergency_collection_reviewed BOOLEAN NOT NULL DEFAULT false,
    routine_care_reviewed BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_profiles_address_is_object
        CHECK (jsonb_typeof(home_address) = 'object')
);

CREATE INDEX idx_child_profiles_child ON child_profiles (tenant_id, branch_id, child_id);

-- 4. child_contacts: parent/carer, emergency, authorised collector. 1:many.
CREATE TYPE child_contact_type AS ENUM ('parent_carer', 'emergency_contact', 'authorised_collector');

CREATE TABLE child_contacts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    contact_type child_contact_type NOT NULL,
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

    CONSTRAINT child_contacts_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_contacts_address_is_object
        CHECK (jsonb_typeof(address) = 'object'),
    CONSTRAINT child_contacts_work_address_is_object
        CHECK (jsonb_typeof(work_address) = 'object'),
    CONSTRAINT child_contacts_sort_order_nonneg
        CHECK (sort_order >= 0),
    CONSTRAINT child_contacts_type_sort_unique
        UNIQUE (tenant_id, branch_id, child_id, contact_type, sort_order)
);

CREATE INDEX idx_child_contacts_child_type
    ON child_contacts (tenant_id, branch_id, child_id, contact_type, sort_order);

-- 5. child_health_profiles: 1:1 with children. Medical, allergy, dietary, GP, health visitor.
CREATE TABLE child_health_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    medical_conditions_status TEXT NOT NULL DEFAULT 'unknown',  -- unknown | no | yes
    medical_conditions_notes TEXT,
    prescribed_medication_status TEXT NOT NULL DEFAULT 'unknown',
    medication_notes TEXT,
    immunisation_status TEXT NOT NULL DEFAULT 'unknown',  -- unknown | up_to_date | refused | partial | not_recorded
    immunisation_country TEXT,
    illness_diagnosis_history TEXT,
    dietary_requirements_status TEXT NOT NULL DEFAULT 'unknown',
    dietary_requirements_notes TEXT,
    dietary_side_effects TEXT,

    doctor_name TEXT,
    doctor_address TEXT,
    doctor_phone TEXT,
    health_visitor_name TEXT,
    health_visitor_address TEXT,
    health_visitor_phone TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_health_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_health_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id)
);

CREATE INDEX idx_child_health_profiles_child ON child_health_profiles (tenant_id, branch_id, child_id);

-- 6. child_safeguarding_profiles: 1:1 with children. Social services, concerns, professional referrals, restricted notes.
CREATE TABLE child_safeguarding_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    social_services_status TEXT NOT NULL DEFAULT 'unknown',  -- unknown | no | yes
    social_services_notes TEXT,
    social_worker_name TEXT,
    social_worker_phone TEXT,
    social_worker_email TEXT,
    concern_walking TEXT NOT NULL DEFAULT 'unknown',
    concern_speech_language TEXT NOT NULL DEFAULT 'unknown',
    concern_hearing TEXT NOT NULL DEFAULT 'unknown',
    concern_sight TEXT NOT NULL DEFAULT 'unknown',
    concern_emotional_wellbeing TEXT NOT NULL DEFAULT 'unknown',
    concern_behaviour TEXT NOT NULL DEFAULT 'unknown',
    professional_referrals JSONB NOT NULL DEFAULT '[]'::jsonb,
    restricted_notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_safeguarding_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_safeguarding_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_safeguarding_profiles_referrals_is_array
        CHECK (jsonb_typeof(professional_referrals) = 'array')
);

CREATE INDEX idx_child_safeguarding_profiles_child ON child_safeguarding_profiles (tenant_id, branch_id, child_id);

-- 7. child_consent_records: 1:1 with children. Single current row (no version).
CREATE TABLE child_consent_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    urgent_medical_treatment BOOLEAN NOT NULL,
    urgent_medical_treatment_exceptions TEXT,
    plasters BOOLEAN NOT NULL,
    safeguarding_reporting_acknowledgement BOOLEAN NOT NULL,
    information_sharing_consent BOOLEAN NOT NULL,
    gdpr_data_processing_consent BOOLEAN NOT NULL,
    area_senco_liaison BOOLEAN NOT NULL,
    health_visitor_liaison BOOLEAN NOT NULL,
    transition_documents BOOLEAN NOT NULL,
    local_outings BOOLEAN NOT NULL,
    face_painting BOOLEAN NOT NULL,
    parent_supplied_sun_cream BOOLEAN NOT NULL,
    parent_supplied_nappy_cream BOOLEAN NOT NULL,
    development_profile_photos BOOLEAN NOT NULL,
    nursery_display_boards BOOLEAN NOT NULL,
    promotional_literature BOOLEAN NOT NULL,
    nursery_website BOOLEAN NOT NULL,
    staff_student_coursework BOOLEAN NOT NULL,
    social_media BOOLEAN NOT NULL,
    social_media_channel_notes TEXT,

    notes_exceptions TEXT,

    -- Paper-form signer metadata (still useful, not stored in profile)
    signer_name TEXT NOT NULL,
    signed_date DATE NOT NULL,
    paper_form_on_file BOOLEAN NOT NULL,

    -- Last write actor
    entered_by_user_id UUID NOT NULL REFERENCES users(id),
    entered_by_membership_id UUID NOT NULL REFERENCES memberships(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_consent_records_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_consent_records_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id)
);

CREATE INDEX idx_child_consent_records_child ON child_consent_records (tenant_id, branch_id, child_id);

-- 8. child_funding_records: 1:1 with children. Eligibility/intent (15/30h, 2yo, tax-free, benefits).
CREATE TABLE child_funding_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    benefits_contribute_to_fees TEXT NOT NULL DEFAULT 'unknown',  -- unknown | no | yes
    working_tax_credit         TEXT NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_parent TEXT NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_nursery TEXT NOT NULL DEFAULT 'unknown',
    funding_3yo_term_time      TEXT NOT NULL DEFAULT 'unknown',
    funding_2yo_term_time      TEXT NOT NULL DEFAULT 'unknown',
    funding_support_notes      TEXT,
    funding_support_reviewed   BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_funding_records_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_funding_records_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id)
);

CREATE INDEX idx_child_funding_records_child ON child_funding_records (tenant_id, branch_id, child_id);

-- 9. child_collection_settings: 1:1 with children. The collection password hash and metadata.
CREATE TABLE child_collection_settings (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    over_18_collection_acknowledged BOOLEAN NOT NULL DEFAULT false,
    collection_password_hash TEXT,
    collection_password_updated_at TIMESTAMPTZ,
    collection_password_updated_by_user_id UUID REFERENCES users(id),
    collection_password_updated_by_membership_id UUID REFERENCES memberships(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_collection_settings_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_collection_settings_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_collection_settings_password_consistency
        CHECK (
            (collection_password_hash IS NULL
             AND collection_password_updated_at IS NULL
             AND collection_password_updated_by_user_id IS NULL
             AND collection_password_updated_by_membership_id IS NULL)
            OR
            (collection_password_hash IS NOT NULL
             AND collection_password_updated_at IS NOT NULL
             AND collection_password_updated_by_user_id IS NOT NULL
             AND collection_password_updated_by_membership_id IS NOT NULL)
        )
);

CREATE INDEX idx_child_collection_settings_child ON child_collection_settings (tenant_id, branch_id, child_id);

-- 10. child_room_assignments: 1:many per child. History of room placements.
CREATE TABLE child_room_assignments (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(id),
    start_date DATE NOT NULL,
    end_date DATE,  -- NULL = current
    is_current BOOLEAN NOT NULL GENERATED ALWAYS AS (end_date IS NULL) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_room_assignments_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_room_assignments_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_room_assignments_dates_check
        CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_child_room_assignments_child_current
    ON child_room_assignments (tenant_id, branch_id, child_id) WHERE is_current;
CREATE INDEX idx_child_room_assignments_room_current
    ON child_room_assignments (tenant_id, branch_id, room_id) WHERE is_current;

-- 11. child_billing_profiles: 1:1 with children. Per-child billing basis.
CREATE TABLE child_billing_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    billing_basis TEXT NOT NULL DEFAULT 'site_rate',  -- 'site_rate' | 'custom'
    custom_rate_minor INTEGER,                        -- nullable, only when billing_basis='custom'
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_billing_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_billing_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_billing_profiles_basis_check
        CHECK (billing_basis IN ('site_rate', 'custom')),
    CONSTRAINT child_billing_profiles_custom_rate_consistency
        CHECK (
            (billing_basis = 'site_rate' AND custom_rate_minor IS NULL)
            OR
            (billing_basis = 'custom' AND custom_rate_minor IS NOT NULL AND custom_rate_minor > 0)
        )
);

CREATE INDEX idx_child_billing_profiles_child ON child_billing_profiles (tenant_id, branch_id, child_id);

-- 12. child_leaving_records: 1:1 with children. Set when child is marked inactive.
CREATE TABLE child_leaving_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    left_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reason_code TEXT NOT NULL,  -- duplicate_record | entered_in_error | left_nursery | safeguarding_direction | contact_update | access_revoked | other
    reason_note TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_leaving_records_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_leaving_records_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_leaving_records_reason_check
        CHECK (reason_code IN ('duplicate_record','entered_in_error','left_nursery','safeguarding_direction','contact_update','access_revoked','other'))
);

CREATE INDEX idx_child_leaving_records_child ON child_leaving_records (tenant_id, branch_id, child_id);
