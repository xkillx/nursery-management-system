CREATE TABLE child_registration_consent_records (
    id uuid NOT NULL PRIMARY KEY,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    version INTEGER NOT NULL,
    source TEXT NOT NULL DEFAULT 'paper_form',

    -- Signer metadata
    signer_name TEXT NOT NULL,
    signed_date DATE NOT NULL,
    paper_form_on_file BOOLEAN NOT NULL,

    -- Consent decisions (BOOLEAN not nullable because they must be explicit)
    urgent_medical_treatment BOOLEAN NOT NULL,
    urgent_medical_treatment_exceptions TEXT,
    plasters BOOLEAN NOT NULL,
    safeguarding_reporting_acknowledgement BOOLEAN NOT NULL,
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

    -- Entered-by actor metadata
    entered_by_user_id uuid NOT NULL,
    entered_by_membership_id uuid NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, branch_id, child_id, version)
);

CREATE INDEX idx_consent_records_child ON child_registration_consent_records (tenant_id, branch_id, child_id, version DESC);
