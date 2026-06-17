-- 1. Drop the new child_* sub-record tables.
DROP TABLE IF EXISTS child_leaving_records CASCADE;
DROP TABLE IF EXISTS child_billing_profiles CASCADE;
DROP TABLE IF EXISTS child_room_assignments CASCADE;
DROP TABLE IF EXISTS child_collection_settings CASCADE;
DROP TABLE IF EXISTS child_funding_records CASCADE;
DROP TABLE IF EXISTS child_consent_records CASCADE;
DROP TABLE IF EXISTS child_safeguarding_profiles CASCADE;
DROP TABLE IF EXISTS child_health_profiles CASCADE;
DROP TABLE IF EXISTS child_contacts CASCADE;
DROP TYPE  IF EXISTS child_contact_type CASCADE;
DROP TABLE IF EXISTS child_profiles CASCADE;

-- 2. Re-add the dropped `children` columns.
CREATE TYPE child_left_reason_code AS ENUM (
    'duplicate_record', 'entered_in_error', 'left_nursery',
    'safeguarding_direction', 'contact_update', 'access_revoked', 'other'
);

ALTER TABLE children
    ADD COLUMN IF NOT EXISTS primary_room_id UUID REFERENCES rooms(id),
    ADD COLUMN IF NOT EXISTS core_hourly_rate_minor INTEGER,
    ADD COLUMN IF NOT EXISTS left_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS left_reason_code child_left_reason_code,
    ADD COLUMN IF NOT EXISTS left_reason_note TEXT;

DROP TYPE IF EXISTS child_left_reason_code;

-- 3. Re-create the four `child_registration_*` tables and the three enum types
-- (matches the schema in 000016/000020/000021/000023/000024/000025/000026).

CREATE TYPE registration_yes_no_unknown AS ENUM ('unknown', 'no', 'yes');
CREATE TYPE registration_immunisation_status AS ENUM ('unknown', 'up_to_date', 'refused', 'partial', 'not_recorded');
CREATE TYPE registration_contact_type AS ENUM ('parent_carer', 'emergency_contact', 'authorised_collector');

CREATE TABLE child_registration_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,

    sex TEXT,
    religion TEXT,
    ethnic_origin TEXT,
    first_language TEXT,
    other_languages TEXT[] NOT NULL DEFAULT '{}',
    home_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    home_postcode TEXT,
    home_telephone TEXT,

    disability_status registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    disability_notes TEXT,
    access_requirements TEXT,

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

    doctor_name TEXT,
    doctor_address TEXT,
    doctor_phone TEXT,
    health_visitor_name TEXT,
    health_visitor_address TEXT,
    health_visitor_phone TEXT,

    social_services_status registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    social_services_notes TEXT,
    social_worker_name TEXT,
    social_worker_phone TEXT,
    social_worker_email TEXT,
    concern_walking registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_speech_language registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_hearing registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_sight registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_emotional_wellbeing registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    concern_behaviour registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    professional_referrals JSONB NOT NULL DEFAULT '[]'::jsonb,

    over_18_collection_acknowledged BOOLEAN NOT NULL DEFAULT false,
    collection_password_hash TEXT,
    collection_password_updated_at TIMESTAMPTZ,
    collection_password_updated_by_user_id UUID,
    collection_password_updated_by_membership_id UUID,

    benefits_contribute_to_fees registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    working_tax_credit         registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_parent registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_nursery registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    funding_3yo_term_time      registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    funding_2yo_term_time      registration_yes_no_unknown NOT NULL DEFAULT 'unknown',
    funding_support_notes      TEXT,
    funding_support_reviewed   BOOLEAN NOT NULL DEFAULT false,

    routine_care_notes TEXT,
    routine_care_reviewed BOOLEAN NOT NULL DEFAULT false,

    registration_date DATE,

    gdpr_declared_by_name  TEXT,
    gdpr_declared_at       TIMESTAMPTZ,
    gdpr_declaration_date  DATE,

    demographics_home_reviewed     BOOLEAN NOT NULL DEFAULT false,
    medical_dietary_reviewed       BOOLEAN NOT NULL DEFAULT false,
    health_contacts_reviewed       BOOLEAN NOT NULL DEFAULT false,
    social_development_reviewed    BOOLEAN NOT NULL DEFAULT false,
    parent_responsibility_reviewed BOOLEAN NOT NULL DEFAULT false,
    emergency_collection_reviewed  BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE child_registration_contacts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    profile_id UUID NOT NULL,
    child_id UUID NOT NULL,
    contact_type registration_contact_type NOT NULL,
    sort_order INTEGER NOT NULL,
    full_name TEXT NOT NULL,
    relationship_to_child TEXT,
    address JSONB NOT NULL DEFAULT '{}'::jsonb,
    telephone TEXT,
    email TEXT,
    work_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    has_parental_responsibility BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE child_registration_consent_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    version INTEGER NOT NULL,
    source TEXT NOT NULL DEFAULT 'paper_form',
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
    signer_name TEXT,
    signed_date DATE,
    paper_form_on_file BOOLEAN NOT NULL DEFAULT true,
    entered_by_user_id UUID NOT NULL,
    entered_by_membership_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE child_registration_completion_attestations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    consent_record_id UUID,
    profile_updated_at TIMESTAMPTZ NOT NULL,
    attested_by_user_id UUID NOT NULL,
    attested_by_membership_id UUID NOT NULL,
    attested_at TIMESTAMPTZ NOT NULL,
    request_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
