-- Baseline schema (squashed — single migration for fresh builds).
-- No production data; rebuilt from scratch.

CREATE TYPE child_contact_type AS ENUM (
    'parent_carer',
    'emergency_contact',
    'authorised_collector'
);

CREATE TYPE lifecycle_reason_code AS ENUM (
    'duplicate_record',
    'entered_in_error',
    'left_nursery',
    'safeguarding_direction',
    'contact_update',
    'access_revoked',
    'other'
);

CREATE TYPE registration_term_time_only_status AS ENUM (
    'unknown',
    'yes',
    'no',
    'not_applicable'
);

CREATE FUNCTION enforce_invoice_status_transition() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF OLD.status = NEW.status THEN
            RETURN NEW;
        END IF;

        IF OLD.status = 'paid' THEN
            RAISE EXCEPTION 'invoice % is paid and cannot transition', OLD.id;
        END IF;

        IF NEW.status = 'draft' THEN
            RAISE EXCEPTION 'invoice % cannot transition back to draft', OLD.id;
        END IF;

        IF OLD.status = 'payment_failed' AND NEW.status = 'overdue' THEN
            RAISE EXCEPTION 'invoice % is payment_failed and cannot transition to overdue', OLD.id;
        END IF;

        CASE OLD.status
            WHEN 'draft' THEN
                IF NEW.status NOT IN ('issued', 'void') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from draft to %', OLD.id, NEW.status;
                END IF;
            WHEN 'issued' THEN
                IF NEW.status NOT IN ('overdue', 'payment_failed', 'paid') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from issued to %', OLD.id, NEW.status;
                END IF;
            WHEN 'overdue' THEN
                IF NEW.status NOT IN ('payment_failed', 'paid') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from overdue to %', OLD.id, NEW.status;
                END IF;
            WHEN 'payment_failed' THEN
                IF NEW.status NOT IN ('paid') THEN
                    RAISE EXCEPTION 'invoice % cannot transition from payment_failed to %', OLD.id, NEW.status;
                END IF;
        END CASE;
    END IF;

    RETURN NEW;
END;
$$;

CREATE FUNCTION protect_issued_invoice_immutability() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF OLD.status <> 'draft' THEN
        IF NEW.status IS DISTINCT FROM OLD.status
           OR NEW.amount_paid_minor IS DISTINCT FROM OLD.amount_paid_minor
           OR NEW.paid_at IS DISTINCT FROM OLD.paid_at
           OR NEW.payment_failed_at IS DISTINCT FROM OLD.payment_failed_at
           OR NEW.payment_status_updated_at IS DISTINCT FROM OLD.payment_status_updated_at
           OR NEW.updated_at IS DISTINCT FROM OLD.updated_at THEN
            IF NEW.id IS DISTINCT FROM OLD.id
               OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
               OR NEW.branch_id IS DISTINCT FROM OLD.branch_id
               OR NEW.child_id IS DISTINCT FROM OLD.child_id
               OR NEW.billing_month IS DISTINCT FROM OLD.billing_month
               OR NEW.invoice_kind IS DISTINCT FROM OLD.invoice_kind
               OR NEW.invoice_number IS DISTINCT FROM OLD.invoice_number
               OR NEW.issued_sequence IS DISTINCT FROM OLD.issued_sequence
               OR NEW.generated_run_id IS DISTINCT FROM OLD.generated_run_id
               OR NEW.issued_run_id IS DISTINCT FROM OLD.issued_run_id
               OR NEW.issued_at IS DISTINCT FROM OLD.issued_at
               OR NEW.issued_by_user_id IS DISTINCT FROM OLD.issued_by_user_id
               OR NEW.issued_by_membership_id IS DISTINCT FROM OLD.issued_by_membership_id
               OR NEW.locked_at IS DISTINCT FROM OLD.locked_at
               OR NEW.due_at IS DISTINCT FROM OLD.due_at
               OR NEW.currency_code IS DISTINCT FROM OLD.currency_code
               OR NEW.subtotal_minor IS DISTINCT FROM OLD.subtotal_minor
               OR NEW.funded_deduction_minor IS DISTINCT FROM OLD.funded_deduction_minor
               OR NEW.total_due_minor IS DISTINCT FROM OLD.total_due_minor
               OR NEW.adjusts_invoice_id IS DISTINCT FROM OLD.adjusts_invoice_id
               OR NEW.adjustment_reason_code IS DISTINCT FROM OLD.adjustment_reason_code
               OR NEW.adjustment_reason_note IS DISTINCT FROM OLD.adjustment_reason_note
               OR NEW.period_start_date IS DISTINCT FROM OLD.period_start_date
               OR NEW.period_end_date IS DISTINCT FROM OLD.period_end_date
               OR NEW.calculation_details IS DISTINCT FROM OLD.calculation_details
               OR NEW.created_at IS DISTINCT FROM OLD.created_at THEN
                RAISE EXCEPTION 'invoice % is not draft and cannot have header fields modified', OLD.id;
            END IF;
            RETURN NEW;
        ELSE
            RAISE EXCEPTION 'invoice % is not draft and cannot be modified', OLD.id;
        END IF;
    END IF;

    RETURN NEW;
END;
$$;

CREATE FUNCTION protect_issued_invoice_lines() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    inv_status TEXT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        SELECT status INTO inv_status FROM invoices WHERE id = OLD.invoice_id AND tenant_id = OLD.tenant_id AND branch_id = OLD.branch_id;
    ELSE
        SELECT status INTO inv_status FROM invoices WHERE id = NEW.invoice_id AND tenant_id = NEW.tenant_id AND branch_id = NEW.branch_id;
    END IF;

    IF inv_status IS NOT NULL AND inv_status <> 'draft' THEN
        CASE TG_OP
            WHEN 'INSERT' THEN
                RAISE EXCEPTION 'cannot insert lines for invoice % with status %', NEW.invoice_id, inv_status;
            WHEN 'UPDATE' THEN
                RAISE EXCEPTION 'cannot update lines for invoice % with status %', NEW.invoice_id, inv_status;
            WHEN 'DELETE' THEN
                RAISE EXCEPTION 'cannot delete lines for invoice % with status %', OLD.invoice_id, inv_status;
        END CASE;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION enforce_parent_membership_child_role() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    membership_role TEXT;
    membership_is_active BOOLEAN;
BEGIN
    IF NEW.ended_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT role, is_active INTO membership_role, membership_is_active
    FROM memberships
    WHERE id = NEW.membership_id;

    IF membership_role IS NULL THEN
        RAISE EXCEPTION 'parent_membership_children requires valid membership';
    END IF;
    IF membership_role <> 'parent' THEN
        RAISE EXCEPTION 'parent_membership_children requires parent role membership';
    END IF;
    IF membership_is_active IS DISTINCT FROM true THEN
        RAISE EXCEPTION 'parent_membership_children requires active membership';
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION enforce_parent_membership_child_scope() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    child_exists BOOLEAN;
BEGIN
    IF NEW.ended_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM children
        WHERE id = NEW.child_id
          AND tenant_id = NEW.tenant_id
          AND branch_id = NEW.branch_id
    ) INTO child_exists;

    IF NOT child_exists THEN
        RAISE EXCEPTION 'parent_membership_children requires child in same tenant and branch';
    END IF;

    RETURN NEW;
END;
$$;

CREATE FUNCTION cascade_parent_membership_child_end() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.ended_at IS NOT NULL
       AND (OLD.ended_at IS NULL OR OLD.ended_at IS DISTINCT FROM NEW.ended_at) THEN
        UPDATE parent_membership_children
        SET ended_at = NEW.ended_at,
            updated_at = now(),
            ended_reason_code = 'system_cascade',
            ended_reason_note = NULL
        WHERE membership_id = NEW.id
          AND ended_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TABLE tenants (
    id uuid NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE branches (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    core_hourly_rate_minor integer,
    ad_hoc_rate_multiplier numeric(4,2) NOT NULL DEFAULT 1.50,
    overdue_grace_days integer NOT NULL DEFAULT 3,
    reminder_days_before integer NOT NULL DEFAULT 3,
    CONSTRAINT branches_core_hourly_rate_positive_check CHECK (((core_hourly_rate_minor IS NULL) OR (core_hourly_rate_minor > 0))),
    CONSTRAINT branches_overdue_grace_days_range CHECK (overdue_grace_days >= 0 AND overdue_grace_days <= 30),
    CONSTRAINT branches_reminder_days_before_range CHECK (reminder_days_before >= 1 AND reminder_days_before <= 30)
);

CREATE TABLE site_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    nursery_name varchar(120) NOT NULL,
    description varchar(2000) NOT NULL DEFAULT '',
    phone varchar(32) NOT NULL,
    email varchar(254) NOT NULL,
    website varchar(2048) NOT NULL,
    address_street varchar(200) NOT NULL,
    address_city varchar(100) NOT NULL,
    address_postcode varchar(16) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE academic_terms (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name varchar(120) NOT NULL,
    kind varchar(20) NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT academic_terms_start_before_end CHECK (start_date < end_date),
    CONSTRAINT academic_terms_kind_check CHECK (kind IN ('autumn', 'spring', 'summer'))
);

CREATE TABLE users (
    id uuid NOT NULL,
    email text NOT NULL,
    email_normalized text NOT NULL,
    password_hash text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE memberships (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid,
    user_id uuid NOT NULL,
    role text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    ended_at timestamp with time zone,
    CONSTRAINT memberships_active_consistency_check CHECK ((((is_active = true) AND (ended_at IS NULL)) OR ((is_active = false) AND (ended_at IS NOT NULL)))),
    CONSTRAINT memberships_owner_branch_check CHECK ((((role = 'owner'::text) AND (branch_id IS NULL)) OR ((role = ANY (ARRAY['manager'::text, 'practitioner'::text, 'parent'::text])) AND (branch_id IS NOT NULL)))),
    CONSTRAINT memberships_role_check CHECK ((role = ANY (ARRAY['owner'::text, 'manager'::text, 'practitioner'::text, 'parent'::text])))
);

CREATE TABLE rooms (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    age_group text NOT NULL,
    capacity integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE children (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    date_of_birth date NOT NULL,
    start_date date NOT NULL,
    end_date date,
    is_active boolean DEFAULT true NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    first_name text NOT NULL,
    middle_name text,
    last_name text,
    current_term_id uuid,
    profile_photo_path text,
    CONSTRAINT children_enrollment_dates_check CHECK (((end_date IS NULL) OR (start_date <= end_date))),
    CONSTRAINT children_first_name_not_blank_check CHECK ((btrim(first_name) <> ''::text))
);

CREATE TABLE child_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    sex text,
    religion text,
    ethnic_origin text,
    first_language text,
    other_languages text,
    home_address jsonb DEFAULT '{}'::jsonb NOT NULL,
    home_postcode text,
    home_telephone text,
    disability_status text DEFAULT 'unknown'::text NOT NULL,
    disability_notes text,
    access_requirements text,
    routine_care_notes text,
    gdpr_declared_by_name text,
    gdpr_declared_at timestamp with time zone,
    gdpr_declaration_date date,
    registration_date date,
    demographics_home_reviewed boolean DEFAULT false NOT NULL,
    medical_dietary_reviewed boolean DEFAULT false NOT NULL,
    health_contacts_reviewed boolean DEFAULT false NOT NULL,
    social_development_reviewed boolean DEFAULT false NOT NULL,
    parent_responsibility_reviewed boolean DEFAULT false NOT NULL,
    emergency_collection_reviewed boolean DEFAULT false NOT NULL,
    routine_care_reviewed boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_profiles_address_is_object CHECK ((jsonb_typeof(home_address) = 'object'::text))
);

CREATE TABLE child_contacts (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    contact_type child_contact_type NOT NULL,
    sort_order integer NOT NULL,
    full_name text NOT NULL,
    relationship_to_child text,
    address jsonb DEFAULT '{}'::jsonb NOT NULL,
    telephone text,
    email text,
    work_address jsonb DEFAULT '{}'::jsonb NOT NULL,
    has_parental_responsibility boolean,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_contacts_address_is_object CHECK ((jsonb_typeof(address) = 'object'::text)),
    CONSTRAINT child_contacts_full_name_check CHECK ((btrim(full_name) <> ''::text)),
    CONSTRAINT child_contacts_sort_order_nonneg CHECK ((sort_order >= 0)),
    CONSTRAINT child_contacts_work_address_is_object CHECK ((jsonb_typeof(work_address) = 'object'::text))
);

CREATE TABLE child_health_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    medical_conditions_status text DEFAULT 'unknown'::text NOT NULL,
    medical_conditions_notes text,
    prescribed_medication_status text DEFAULT 'unknown'::text NOT NULL,
    medication_notes text,
    immunisation_status text DEFAULT 'unknown'::text NOT NULL,
    immunisation_country text,
    illness_diagnosis_history text,
    dietary_requirements_status text DEFAULT 'unknown'::text NOT NULL,
    dietary_requirements_notes text,
    dietary_side_effects text,
    doctor_name text,
    doctor_address text,
    doctor_phone text,
    health_visitor_name text,
    health_visitor_address text,
    health_visitor_phone text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE child_safeguarding_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    social_services_status text DEFAULT 'unknown'::text NOT NULL,
    social_services_notes text,
    social_worker_name text,
    social_worker_phone text,
    social_worker_email text,
    concern_walking text DEFAULT 'unknown'::text NOT NULL,
    concern_speech_language text DEFAULT 'unknown'::text NOT NULL,
    concern_hearing text DEFAULT 'unknown'::text NOT NULL,
    concern_sight text DEFAULT 'unknown'::text NOT NULL,
    concern_emotional_wellbeing text DEFAULT 'unknown'::text CONSTRAINT child_safeguarding_profiles_concern_emotional_wellbein_not_null NOT NULL,
    concern_behaviour text DEFAULT 'unknown'::text NOT NULL,
    professional_referrals jsonb DEFAULT '[]'::jsonb NOT NULL,
    restricted_notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_safeguarding_profiles_referrals_is_array CHECK ((jsonb_typeof(professional_referrals) = 'array'::text))
);

CREATE TABLE child_consent_records (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    urgent_medical_treatment boolean NOT NULL,
    urgent_medical_treatment_exceptions text,
    plasters boolean NOT NULL,
    safeguarding_reporting_acknowledgement boolean CONSTRAINT child_consent_records_safeguarding_reporting_acknowled_not_null NOT NULL,
    information_sharing_consent boolean NOT NULL,
    gdpr_data_processing_consent boolean NOT NULL,
    area_senco_liaison boolean NOT NULL,
    health_visitor_liaison boolean NOT NULL,
    transition_documents boolean NOT NULL,
    local_outings boolean NOT NULL,
    face_painting boolean NOT NULL,
    parent_supplied_sun_cream boolean NOT NULL,
    parent_supplied_nappy_cream boolean NOT NULL,
    development_profile_photos boolean NOT NULL,
    nursery_display_boards boolean NOT NULL,
    promotional_literature boolean NOT NULL,
    nursery_website boolean NOT NULL,
    staff_student_coursework boolean NOT NULL,
    social_media boolean NOT NULL,
    social_media_channel_notes text,
    notes_exceptions text,
    signer_name text NOT NULL,
    signed_date date NOT NULL,
    paper_form_on_file boolean NOT NULL,
    information_truthfulness_declaration boolean NOT NULL DEFAULT false,
    entered_by_user_id uuid NOT NULL,
    entered_by_membership_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE child_funding_records (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    funding_enabled boolean NOT NULL DEFAULT false,
    funding_type text NOT NULL DEFAULT 'unknown',
    funding_model text NOT NULL DEFAULT 'unknown',
    funded_hours_per_week numeric(5,2),
    funding_start_date date,
    funding_end_date date,
    eligibility_code text,
    eligibility_code_validated boolean NOT NULL DEFAULT false,
    evidence_received boolean NOT NULL DEFAULT false,
    benefits_status text NOT NULL DEFAULT 'unknown',
    benefit_notes text,
    manager_notes text,
    benefit_universal_credit boolean NOT NULL DEFAULT false,
    benefit_income_support boolean NOT NULL DEFAULT false,
    benefit_jobseekers_allowance boolean NOT NULL DEFAULT false,
    benefit_esa_income_related boolean NOT NULL DEFAULT false,
    benefit_child_tax_credit boolean NOT NULL DEFAULT false,
    benefit_other_support boolean NOT NULL DEFAULT false,
    other_benefit_name text,
    CONSTRAINT child_funding_records_funding_type_check CHECK (funding_type IN ('none','fifteen_hours','thirty_hours','two_year_old','custom','unknown')),
    CONSTRAINT child_funding_records_funding_model_check CHECK (funding_model IN ('term_time_only','stretched','unknown')),
    CONSTRAINT child_funding_records_benefits_status_check CHECK (benefits_status IN ('no','yes','unknown')),
    CONSTRAINT child_funding_records_end_after_start CHECK (funding_end_date IS NULL OR funding_start_date IS NULL OR funding_end_date > funding_start_date)
);

CREATE TABLE child_billing_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    billing_basis text DEFAULT 'site_rate'::text NOT NULL,
    custom_rate_minor integer,
    effective_from date DEFAULT CURRENT_DATE NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_billing_profiles_basis_check CHECK ((billing_basis = ANY (ARRAY['site_rate'::text, 'custom'::text]))),
    CONSTRAINT child_billing_profiles_custom_rate_consistency CHECK ((((billing_basis = 'site_rate'::text) AND (custom_rate_minor IS NULL)) OR ((billing_basis = 'custom'::text) AND (custom_rate_minor IS NOT NULL) AND (custom_rate_minor > 0))))
);

CREATE TABLE child_collection_settings (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    over_18_collection_acknowledged boolean DEFAULT false CONSTRAINT child_collection_settings_over_18_collection_acknowled_not_null NOT NULL,
    collection_password text,
    collection_password_hint text,
    collection_password_updated_at timestamp with time zone,
    collection_password_updated_by_user_id uuid,
    collection_password_updated_by_membership_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_collection_settings_password_consistency CHECK ((((collection_password IS NULL) AND (collection_password_updated_at IS NULL) AND (collection_password_updated_by_user_id IS NULL) AND (collection_password_updated_by_membership_id IS NULL)) OR ((collection_password IS NOT NULL) AND (collection_password_updated_at IS NOT NULL) AND (collection_password_updated_by_user_id IS NOT NULL) AND (collection_password_updated_by_membership_id IS NOT NULL))))
);

CREATE TABLE child_room_assignments (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    room_id uuid NOT NULL,
    start_date date NOT NULL,
    end_date date,
    is_current boolean GENERATED ALWAYS AS ((end_date IS NULL)) STORED NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_room_assignments_dates_check CHECK (((end_date IS NULL) OR (end_date >= start_date)))
);

CREATE TABLE child_leaving_records (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    left_at timestamp with time zone DEFAULT now() NOT NULL,
    reason_code text NOT NULL,
    reason_note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_leaving_records_reason_check CHECK ((reason_code = ANY (ARRAY['duplicate_record'::text, 'entered_in_error'::text, 'left_nursery'::text, 'safeguarding_direction'::text, 'contact_update'::text, 'access_revoked'::text, 'other'::text])))
);

CREATE TABLE attendance_sessions (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    status text NOT NULL,
    check_in_at timestamp with time zone NOT NULL,
    check_out_at timestamp with time zone,
    check_in_local_date date NOT NULL,
    check_out_local_date date,
    check_in_event_id uuid,
    check_out_event_id uuid,
    corrected_by_event_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT attendance_sessions_open_shape_check CHECK ((((status = 'open'::text) AND (check_out_at IS NULL) AND (check_out_local_date IS NULL)) OR ((status = ANY (ARRAY['complete'::text, 'corrected'::text])) AND (check_out_at IS NOT NULL) AND (check_out_local_date IS NOT NULL)))),
    CONSTRAINT attendance_sessions_status_check CHECK ((status = ANY (ARRAY['open'::text, 'complete'::text, 'corrected'::text]))),
    CONSTRAINT attendance_sessions_time_order_check CHECK (((check_out_at IS NULL) OR (check_out_at > check_in_at)))
);

CREATE TABLE attendance_events (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    session_id uuid NOT NULL,
    event_type text NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    local_date date NOT NULL,
    recorded_by_user_id uuid NOT NULL,
    recorded_by_membership_id uuid NOT NULL,
    request_id text,
    reason_code text,
    reason_note text,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT attendance_events_reason_shape_check CHECK ((((event_type = ANY (ARRAY['check_in'::text, 'check_out'::text])) AND (reason_code IS NULL) AND (reason_note IS NULL)) OR ((event_type = 'correction'::text) AND (reason_code = ANY (ARRAY['missed_check_in'::text, 'missed_check_out'::text, 'incorrect_time'::text, 'duplicate_entry'::text, 'other'::text])) AND ((reason_code <> 'other'::text) OR (NULLIF(btrim(reason_note), ''::text) IS NOT NULL))))),
    CONSTRAINT attendance_events_type_check CHECK ((event_type = ANY (ARRAY['check_in'::text, 'check_out'::text, 'correction'::text])))
);

CREATE TABLE absence_markers (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    local_date date NOT NULL,
    marked_at timestamp with time zone DEFAULT now() NOT NULL,
    marked_by_user_id uuid NOT NULL,
    marked_by_membership_id uuid NOT NULL,
    cleared_at timestamp with time zone,
    cleared_by_user_id uuid,
    cleared_by_membership_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT absence_markers_cleared_at_gte_marked_at CHECK (((cleared_at IS NULL) OR (cleared_at >= marked_at))),
    CONSTRAINT absence_markers_cleared_at_null_implies_membership_null CHECK (((cleared_at IS NULL) = (cleared_by_membership_id IS NULL))),
    CONSTRAINT absence_markers_cleared_at_null_implies_user_null CHECK (((cleared_at IS NULL) = (cleared_by_user_id IS NULL)))
);

CREATE TABLE funding_profiles (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    billing_month date NOT NULL,
    funded_allowance_minutes integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    funding_type text,
    funding_model text,
    funded_hours_per_week numeric(5,2),
    CONSTRAINT funding_profiles_allowance_bounds_check CHECK (((funded_allowance_minutes >= 0) AND (funded_allowance_minutes <= 44640))),
    CONSTRAINT funding_profiles_billing_month_first_day_check CHECK ((billing_month = (date_trunc('month'::text, (billing_month)::timestamp with time zone))::date))
);

CREATE TABLE child_funding_history (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    funding_type text,
    funding_model text,
    funded_hours_per_week numeric(5,2),
    funding_start_date date,
    funding_end_date date,
    changed_at timestamptz NOT NULL DEFAULT now(),
    changed_by_user_id uuid NOT NULL,
    CONSTRAINT chk_funding_type CHECK (funding_type IS NULL OR funding_type IN ('fifteen_hours', 'thirty_hours')),
    CONSTRAINT chk_funding_model CHECK (funding_model IS NULL OR funding_model IN ('term_time', 'stretched'))
);

CREATE TABLE session_types (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name text NOT NULL,
    start_time time NOT NULL,
    end_time time NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT session_types_time_check CHECK ((start_time < end_time))
);

CREATE TABLE child_booking_patterns (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    effective_from date NOT NULL,
    effective_to date,
    is_current boolean GENERATED ALWAYS AS ((effective_to IS NULL)) STORED NOT NULL,
    term_time_only boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_booking_patterns_dates_check CHECK (((effective_to IS NULL) OR (effective_to >= effective_from)))
);

CREATE TABLE child_booking_pattern_entries (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    pattern_id uuid NOT NULL,
    day_of_week integer NOT NULL,
    session_type_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT child_booking_pattern_entries_dow_check CHECK (((day_of_week >= 1) AND (day_of_week <= 5)))
);

CREATE TABLE session_templates (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE session_template_entries (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    template_id uuid NOT NULL,
    day_of_week integer NOT NULL,
    session_type_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT session_template_entries_dow_check CHECK (((day_of_week >= 1) AND (day_of_week <= 5)))
);

CREATE TABLE bookings (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    session_template_id uuid,
    room_id uuid NOT NULL,
    days_of_week integer[] NOT NULL,
    effective_start_date date NOT NULL,
    effective_end_date date,
    funding_type text,
    funding_hours_per_week numeric(5,2),
    la_reference text,
    status text NOT NULL DEFAULT 'active',
    booked_by_membership_id uuid NOT NULL,
    session_entries jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT bookings_status_check CHECK (status IN ('active', 'paused', 'cancelled')),
    CONSTRAINT bookings_funding_type_check CHECK (funding_type IS NULL OR funding_type IN ('none', 'fifteen_hours', 'thirty_hours', 'two_year_old', 'custom')),
    CONSTRAINT bookings_days_of_week_check CHECK (array_length(days_of_week, 1) > 0),
    CONSTRAINT bookings_effective_dates_check CHECK (effective_end_date IS NULL OR effective_end_date >= effective_start_date),
    CONSTRAINT bookings_session_source_check CHECK (session_template_id IS NOT NULL OR session_entries IS NOT NULL)
);

CREATE TABLE term (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    term_start_date date NOT NULL,
    term_end_date date NOT NULL,
    booking_pattern_id uuid NOT NULL,
    site_hourly_rate_minor integer NOT NULL,
    status text NOT NULL,
    termination_reason_code text,
    termination_reason_note text,
    terminated_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_membership_id uuid NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT term_dates_first_of_month CHECK ((term_start_date = date_trunc('month', term_start_date)::date)),
    CONSTRAINT term_end_after_start CHECK ((term_end_date >= term_start_date)),
    CONSTRAINT term_end_minus_start_is_12_months_minus_one_day CHECK ((term_end_date = ((term_start_date + interval '12 months') - interval '1 day')::date)),
    CONSTRAINT term_status_valid CHECK ((status = ANY (ARRAY['pre_term'::text, 'active'::text, 'pending_renewal'::text, 'ended'::text, 'terminated'::text]))),
    CONSTRAINT term_hourly_rate_nonneg CHECK ((site_hourly_rate_minor >= 0)),
    CONSTRAINT term_terminated_shape CHECK (
        ((status = 'terminated') AND (terminated_at IS NOT NULL) AND (termination_reason_code IS NOT NULL) AND (btrim(termination_reason_code) <> ''::text))
        OR (status <> 'terminated')
    )
);

CREATE TABLE term_schedule_change (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    term_id uuid NOT NULL,
    previous_booking_pattern_id uuid NOT NULL,
    new_booking_pattern_id uuid NOT NULL,
    change_kind text NOT NULL,
    requested_at timestamp with time zone DEFAULT now() NOT NULL,
    effective_from date NOT NULL,
    approved_by_membership_id uuid,
    approval_decision text,
    rejected_at timestamp with time zone,
    request_id text NOT NULL,
    CONSTRAINT term_schedule_change_kind_valid CHECK ((change_kind = ANY (ARRAY['decrease'::text, 'increase'::text]))),
    CONSTRAINT term_schedule_change_decision_valid CHECK (
        (approval_decision IS NULL)
        OR (approval_decision = ANY (ARRAY['approved'::text, 'rejected'::text]))
    ),
    CONSTRAINT term_schedule_change_first_of_month CHECK ((effective_from = date_trunc('month', effective_from)::date))
);

CREATE TABLE invoice_run_advance (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    billing_month date NOT NULL,
    generated_at timestamp with time zone DEFAULT now() NOT NULL,
    generated_invoice_count integer NOT NULL,
    skipped_term_count integer NOT NULL,
    exception_count integer NOT NULL,
    triggered_by text NOT NULL,
    request_id text,
    CONSTRAINT invoice_run_advance_triggered_by_valid CHECK ((triggered_by = ANY (ARRAY['scheduler'::text, 'manager_regenerate'::text]))),
    CONSTRAINT invoice_run_advance_first_of_month CHECK ((billing_month = date_trunc('month', billing_month)::date)),
    CONSTRAINT invoice_run_advance_counts_nonneg CHECK (
        (generated_invoice_count >= 0)
        AND (skipped_term_count >= 0)
        AND (exception_count >= 0)
    )
);

CREATE TABLE audit_logs (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    actor_user_id uuid,
    action_type text NOT NULL,
    action_entity_type text NOT NULL,
    action_entity_id uuid,
    reason text,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    request_id text,
    actor_membership_id uuid,
    reason_code lifecycle_reason_code,
    reason_note text,
    CONSTRAINT audit_logs_reason_other_note_check CHECK (((reason_code IS DISTINCT FROM 'other'::lifecycle_reason_code) OR ((reason_note IS NOT NULL) AND (btrim(reason_note) <> ''::text)))),
    CONSTRAINT audit_logs_reason_shape_check CHECK (((reason_code IS NOT NULL) OR (reason_note IS NULL)))
);

CREATE TABLE parent_membership_children (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    membership_id uuid NOT NULL,
    child_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    ended_at timestamp with time zone,
    ended_reason_code lifecycle_reason_code,
    ended_reason_note text,
    CONSTRAINT parent_membership_children_end_reason_check CHECK (
        (((ended_at IS NULL) AND (ended_reason_code IS NULL) AND (ended_reason_note IS NULL)) OR
         ((ended_at IS NOT NULL) AND (ended_reason_code IS NOT NULL) AND
          ((ended_reason_code <> 'other'::lifecycle_reason_code) OR
           ((ended_reason_note IS NOT NULL) AND (btrim(ended_reason_note) <> ''::text)))))
    )
);

CREATE TABLE invoice_number_sequences (
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    billing_year integer NOT NULL,
    billing_month integer NOT NULL,
    next_sequence integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT invoice_number_sequences_month_range CHECK (((billing_month >= 1) AND (billing_month <= 12))),
    CONSTRAINT invoice_number_sequences_next_seq_gte_1 CHECK ((next_sequence >= 1)),
    CONSTRAINT invoice_number_sequences_year_gte_2000 CHECK ((billing_year >= 2000))
);

CREATE TABLE invoice_runs (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    billing_month date NOT NULL,
    run_type text NOT NULL,
    status text DEFAULT 'started'::text NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    requested_by_user_id uuid NOT NULL,
    requested_by_membership_id uuid NOT NULL,
    request_id text,
    eligible_count integer DEFAULT 0 NOT NULL,
    success_count integer DEFAULT 0 NOT NULL,
    blocked_count integer DEFAULT 0 NOT NULL,
    failed_count integer DEFAULT 0 NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT invoice_runs_billing_month_first_day CHECK ((billing_month = (date_trunc('month'::text, (billing_month)::timestamp with time zone))::date)),
    CONSTRAINT invoice_runs_blocked_count_nonneg CHECK ((blocked_count >= 0)),
    CONSTRAINT invoice_runs_completed_at_consistent CHECK (((status = 'started'::text) = (completed_at IS NULL))),
    CONSTRAINT invoice_runs_eligible_count_nonneg CHECK ((eligible_count >= 0)),
    CONSTRAINT invoice_runs_failed_count_nonneg CHECK ((failed_count >= 0)),
    CONSTRAINT invoice_runs_run_type_valid CHECK ((run_type = ANY (ARRAY['draft_generation'::text, 'issue'::text]))),
    CONSTRAINT invoice_runs_status_valid CHECK ((status = ANY (ARRAY['started'::text, 'completed'::text, 'completed_with_exceptions'::text, 'failed'::text]))),
    CONSTRAINT invoice_runs_success_count_nonneg CHECK ((success_count >= 0))
);

CREATE TABLE invoices (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    billing_month date NOT NULL,
    invoice_kind text DEFAULT 'monthly'::text NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    invoice_number text,
    issued_sequence integer,
    generated_run_id uuid,
    issued_run_id uuid,
    issued_at timestamp with time zone,
    issued_by_user_id uuid,
    issued_by_membership_id uuid,
    locked_at timestamp with time zone,
    due_at timestamp with time zone,
    currency_code text DEFAULT 'GBP' NOT NULL,
    subtotal_minor integer DEFAULT 0 NOT NULL,
    funded_deduction_minor integer DEFAULT 0 NOT NULL,
    total_due_minor integer DEFAULT 0 NOT NULL,
    amount_paid_minor integer DEFAULT 0 NOT NULL,
    paid_at timestamp with time zone,
    payment_failed_at timestamp with time zone,
    payment_status_updated_at timestamp with time zone,
    adjusts_invoice_id uuid,
    adjustment_reason_code text,
    adjustment_reason_note text,
    period_start_date date NOT NULL,
    period_end_date date NOT NULL,
    calculation_details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    voided_at timestamp with time zone,
    void_reason text,
    CONSTRAINT invoices_adjustment_shape CHECK (((invoice_kind <> 'adjustment'::text) OR ((adjusts_invoice_id IS NOT NULL) AND (adjustment_reason_code IS NOT NULL) AND (btrim(adjustment_reason_code) <> ''::text) AND (adjustment_reason_note IS NOT NULL) AND (btrim(adjustment_reason_note) <> ''::text)))),
    CONSTRAINT invoices_amount_paid_lte_total CHECK ((amount_paid_minor <= total_due_minor)),
    CONSTRAINT invoices_amount_paid_nonneg CHECK ((amount_paid_minor >= 0)),
    CONSTRAINT invoices_billing_month_first_day CHECK ((billing_month = (date_trunc('month'::text, (billing_month)::timestamp with time zone))::date)),
    CONSTRAINT invoices_currency_gbp CHECK (currency_code = 'GBP'),
    CONSTRAINT invoices_draft_shape CHECK (((status <> 'draft'::text) OR ((issued_at IS NULL) AND (issued_by_user_id IS NULL) AND (issued_by_membership_id IS NULL) AND (locked_at IS NULL) AND (due_at IS NULL) AND (invoice_number IS NULL) AND (issued_sequence IS NULL) AND (issued_run_id IS NULL)))),
    CONSTRAINT invoices_funded_deduction_nonneg CHECK ((funded_deduction_minor >= 0)),
    CONSTRAINT invoices_invoice_kind_valid CHECK ((invoice_kind = ANY (ARRAY['monthly'::text, 'adjustment'::text]))),
    CONSTRAINT invoices_issued_shape CHECK (((status = 'draft'::text) OR ((issued_at IS NOT NULL) AND (issued_by_user_id IS NOT NULL) AND (issued_by_membership_id IS NOT NULL) AND (locked_at IS NOT NULL) AND (due_at IS NOT NULL) AND (invoice_number IS NOT NULL) AND (issued_sequence IS NOT NULL) AND (issued_run_id IS NOT NULL)))),
    CONSTRAINT invoices_monthly_shape CHECK (((invoice_kind <> 'monthly'::text) OR ((adjusts_invoice_id IS NULL) AND (adjustment_reason_code IS NULL) AND (adjustment_reason_note IS NULL)))),
    CONSTRAINT invoices_no_self_adjust CHECK (((adjusts_invoice_id IS NULL) OR (adjusts_invoice_id <> id))),
    CONSTRAINT invoices_paid_shape CHECK (((status <> 'paid'::text) OR ((paid_at IS NOT NULL) AND (amount_paid_minor = total_due_minor)))),
    CONSTRAINT invoices_payment_failed_shape CHECK (((status <> 'payment_failed'::text) OR (payment_failed_at IS NOT NULL))),
    CONSTRAINT invoices_period_range CHECK ((period_start_date <= period_end_date)),
    CONSTRAINT invoices_status_valid CHECK ((status = ANY (ARRAY['draft'::text, 'issued'::text, 'payment_failed'::text, 'paid'::text, 'overdue'::text, 'void'::text]))),
    CONSTRAINT invoices_subtotal_nonneg CHECK ((subtotal_minor >= 0)),
    CONSTRAINT invoices_total_due_nonneg CHECK ((total_due_minor >= 0)),
    CONSTRAINT invoices_void_shape CHECK (((status <> 'void'::text) OR ((voided_at IS NOT NULL) AND (void_reason IS NOT NULL) AND (btrim(void_reason) <> ''::text))))
);

CREATE TABLE invoice_lines (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    line_kind text NOT NULL,
    description text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    quantity_minutes integer,
    unit_amount_minor integer,
    line_amount_minor integer NOT NULL,
    raw_attended_minutes integer,
    rounded_attended_minutes integer,
    funded_allowance_minutes integer,
    funded_deduction_minutes integer,
    core_billable_minutes integer,
    session_count integer,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT invoice_lines_core_amount_nonneg CHECK (((line_kind <> 'core_childcare'::text) OR (line_amount_minor >= 0))),
    CONSTRAINT invoice_lines_core_billable_nonneg CHECK (((core_billable_minutes IS NULL) OR (core_billable_minutes >= 0))),
    CONSTRAINT invoice_lines_extra_amount_nonneg CHECK (((line_kind <> 'extra'::text) OR (line_amount_minor >= 0))),
    CONSTRAINT invoice_lines_funded_allowance_nonneg CHECK (((funded_allowance_minutes IS NULL) OR (funded_allowance_minutes >= 0))),
    CONSTRAINT invoice_lines_funded_deduction_nonneg CHECK (((funded_deduction_minutes IS NULL) OR (funded_deduction_minutes >= 0))),
    CONSTRAINT invoice_lines_funded_deduction_nonpos CHECK (((line_kind <> 'funded_deduction'::text) OR (line_amount_minor <= 0))),
    CONSTRAINT invoice_lines_line_kind_valid CHECK ((line_kind = ANY (ARRAY['core_childcare'::text, 'funded_deduction'::text, 'extra'::text, 'adjustment'::text, 'ad_hoc'::text, 'hourly'::text]))),
    CONSTRAINT invoice_lines_quantity_nonneg CHECK (((quantity_minutes IS NULL) OR (quantity_minutes >= 0))),
    CONSTRAINT invoice_lines_raw_attended_nonneg CHECK (((raw_attended_minutes IS NULL) OR (raw_attended_minutes >= 0))),
    CONSTRAINT invoice_lines_rounded_attended_nonneg CHECK (((rounded_attended_minutes IS NULL) OR (rounded_attended_minutes >= 0))),
    CONSTRAINT invoice_lines_session_count_nonneg CHECK (((session_count IS NULL) OR (session_count >= 0))),
    CONSTRAINT invoice_lines_unit_amount_nonneg CHECK (((unit_amount_minor IS NULL) OR (unit_amount_minor >= 0)))
);

CREATE TABLE ad_hoc_bookings (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    calendar_date date NOT NULL,
    session_type_id uuid NOT NULL,
    booked_by_membership_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ad_hoc_bookings_pkey PRIMARY KEY (id),
    CONSTRAINT ad_hoc_bookings_status_check CHECK (status IN ('active', 'cancelled'))
);

CREATE TABLE hourly_bookings (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    child_id uuid NOT NULL,
    calendar_date date NOT NULL,
    start_time_minutes integer NOT NULL CHECK (start_time_minutes >= 0 AND start_time_minutes <= 1439),
    duration_minutes integer NOT NULL CHECK (duration_minutes > 0),
    session_type_id uuid,
    booked_by_membership_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT hourly_bookings_status_check CHECK (status IN ('active', 'cancelled'))
);

CREATE UNIQUE INDEX hourly_bookings_unique_slot
    ON hourly_bookings (tenant_id, branch_id, child_id, calendar_date, start_time_minutes);

CREATE INDEX hourly_bookings_child_date_idx
    ON hourly_bookings (tenant_id, branch_id, child_id, calendar_date);

CREATE TABLE branch_closure_days (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    date date NOT NULL,
    reason text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE payment_attempts (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    initiated_by_user_id uuid NOT NULL,
    initiated_by_membership_id uuid NOT NULL,
    request_id text,
    status text NOT NULL,
    amount_minor integer NOT NULL,
    currency_code text DEFAULT 'GBP' NOT NULL,
    stripe_checkout_session_id text,
    stripe_checkout_url text,
    stripe_payment_intent_id text,
    stripe_expires_at timestamp with time zone,
    provider_error_code text,
    provider_error_message text,
    failure_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_payment_attempts_amount_positive CHECK ((amount_minor > 0)),
    CONSTRAINT chk_payment_attempts_created_has_session CHECK (((status <> 'checkout_created'::text) OR ((stripe_checkout_session_id IS NOT NULL) AND (stripe_checkout_session_id <> ''::text) AND (stripe_checkout_url IS NOT NULL) AND (stripe_checkout_url <> ''::text)))),
    CONSTRAINT chk_payment_attempts_currency_gbp CHECK (currency_code = 'GBP'),
    CONSTRAINT chk_payment_attempts_failed_has_reason CHECK (((status <> 'checkout_creation_failed'::text) OR ((failure_reason IS NOT NULL) AND (failure_reason <> ''::text)) OR ((provider_error_code IS NOT NULL) AND (provider_error_code <> ''::text)) OR ((provider_error_message IS NOT NULL) AND (provider_error_message <> ''::text)))),
    CONSTRAINT chk_payment_attempts_status CHECK ((status = ANY (ARRAY['checkout_creation_started'::text, 'checkout_created'::text, 'checkout_creation_failed'::text, 'paid'::text, 'payment_failed'::text, 'cancelled'::text, 'expired'::text])))
);

CREATE TABLE payment_reconciliation_records (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    payment_attempt_id uuid NOT NULL,
    stripe_webhook_event_id uuid NOT NULL,
    stripe_event_id text NOT NULL,
    stripe_event_type text NOT NULL,
    stripe_checkout_session_id text CONSTRAINT payment_reconciliation_reco_stripe_checkout_session_id_not_null NOT NULL,
    stripe_payment_intent_id text,
    outcome text NOT NULL,
    reason_code text NOT NULL,
    previous_invoice_status text,
    new_invoice_status text,
    attempt_previous_status text,
    attempt_new_status text,
    amount_minor integer,
    currency_code text,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_reconciliation_outcome CHECK ((outcome = ANY (ARRAY['paid'::text, 'payment_failed'::text, 'expired'::text, 'ignored'::text, 'rejected'::text])))
);

CREATE TABLE stripe_webhook_events (
    id uuid NOT NULL,
    stripe_event_id text NOT NULL,
    event_type text NOT NULL,
    livemode boolean NOT NULL,
    api_version text,
    provider_created_at timestamp with time zone,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    processed_at timestamp with time zone,
    processing_status text NOT NULL,
    processing_reason text,
    request_id text,
    raw_payload jsonb NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_webhook_events_processing_status CHECK ((processing_status = ANY (ARRAY['received'::text, 'processed'::text, 'ignored'::text, 'rejected'::text])))
);

CREATE TABLE manager_invites (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    email text NOT NULL,
    email_normalized text NOT NULL,
    role text NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    accepted_at timestamp with time zone,
    accepted_user_id uuid,
    accepted_membership_id uuid,
    revoked_at timestamp with time zone,
    revoked_by_user_id uuid,
    revoked_by_membership_id uuid,
    created_by_user_id uuid NOT NULL,
    created_by_membership_id uuid NOT NULL,
    resent_at timestamp with time zone,
    resent_by_user_id uuid,
    resent_by_membership_id uuid,
    send_count integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT manager_invites_accept_shape_check CHECK ((((accepted_at IS NULL) AND (accepted_user_id IS NULL) AND (accepted_membership_id IS NULL)) OR ((accepted_at IS NOT NULL) AND (accepted_user_id IS NOT NULL) AND (accepted_membership_id IS NOT NULL)))),
    CONSTRAINT manager_invites_revoke_shape_check CHECK ((((revoked_at IS NULL) AND (revoked_by_user_id IS NULL) AND (revoked_by_membership_id IS NULL)) OR ((revoked_at IS NOT NULL) AND (revoked_by_user_id IS NOT NULL) AND (revoked_by_membership_id IS NOT NULL)))),
    CONSTRAINT manager_invites_role_check CHECK ((role = ANY (ARRAY['manager'::text, 'practitioner'::text, 'parent'::text]))),
    CONSTRAINT manager_invites_send_count_check CHECK ((send_count >= 1)),
    CONSTRAINT manager_invites_terminal_state_check CHECK (((accepted_at IS NULL) OR (revoked_at IS NULL)))
);

CREATE TABLE refresh_tokens (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    user_agent text,
    ip_address text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    membership_id uuid NOT NULL,
    remember_me boolean DEFAULT true NOT NULL
);

CREATE TABLE password_reset_tokens (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    superseded_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT password_reset_tokens_consumed_shape_check CHECK (((used_at IS NULL) OR (superseded_at IS NULL)))
);

CREATE TABLE invoice_reminder_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    reminder_type text NOT NULL CHECK (reminder_type IN ('due_soon', 'due_today')),
    sent_at_date date NOT NULL DEFAULT CURRENT_DATE,
    sent_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (invoice_id, reminder_type, sent_at_date)
);

CREATE TABLE payment_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    stripe_payment_link_id text NOT NULL,
    stripe_payment_link_url text NOT NULL,
    amount_minor integer NOT NULL,
    currency_code text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_by_membership_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deactivated')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Primary keys

ALTER TABLE ONLY tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);

ALTER TABLE ONLY branches
    ADD CONSTRAINT branches_pkey PRIMARY KEY (id);

ALTER TABLE ONLY branches
    ADD CONSTRAINT branches_tenant_id_id_unique UNIQUE (tenant_id, id);

ALTER TABLE ONLY branches
    ADD CONSTRAINT branches_tenant_id_name_key UNIQUE (tenant_id, name);

ALTER TABLE ONLY site_profiles
    ADD CONSTRAINT site_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY site_profiles
    ADD CONSTRAINT site_profiles_branch_id_unique UNIQUE (branch_id);

ALTER TABLE ONLY academic_terms
    ADD CONSTRAINT academic_terms_pkey PRIMARY KEY (id);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_email_normalized_key UNIQUE (email_normalized);

ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_pkey PRIMARY KEY (id);

ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY rooms
    ADD CONSTRAINT rooms_pkey PRIMARY KEY (id);

ALTER TABLE ONLY children
    ADD CONSTRAINT children_pkey PRIMARY KEY (id);

ALTER TABLE ONLY children
    ADD CONSTRAINT children_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY child_profiles
    ADD CONSTRAINT child_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_profiles
    ADD CONSTRAINT child_profiles_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY child_contacts
    ADD CONSTRAINT child_contacts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_contacts
    ADD CONSTRAINT child_contacts_type_sort_unique UNIQUE (tenant_id, branch_id, child_id, contact_type, sort_order);

ALTER TABLE ONLY child_health_profiles
    ADD CONSTRAINT child_health_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_health_profiles
    ADD CONSTRAINT child_health_profiles_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY child_safeguarding_profiles
    ADD CONSTRAINT child_safeguarding_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_safeguarding_profiles
    ADD CONSTRAINT child_safeguarding_profiles_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY child_consent_records
    ADD CONSTRAINT child_consent_records_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_consent_records
    ADD CONSTRAINT child_consent_records_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY child_funding_records
    ADD CONSTRAINT child_funding_records_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_funding_records
    ADD CONSTRAINT child_funding_records_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY child_billing_profiles
    ADD CONSTRAINT child_billing_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_billing_profiles
    ADD CONSTRAINT child_billing_profiles_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY child_collection_settings
    ADD CONSTRAINT child_collection_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_collection_settings
    ADD CONSTRAINT child_collection_settings_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY child_room_assignments
    ADD CONSTRAINT child_room_assignments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_leaving_records
    ADD CONSTRAINT child_leaving_records_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_leaving_records
    ADD CONSTRAINT child_leaving_records_child_id_key UNIQUE (child_id);

ALTER TABLE ONLY attendance_sessions
    ADD CONSTRAINT attendance_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY attendance_sessions
    ADD CONSTRAINT attendance_sessions_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_events
    ADD CONSTRAINT attendance_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY absence_markers
    ADD CONSTRAINT absence_markers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_scope_child_month_unique UNIQUE (tenant_id, branch_id, child_id, billing_month);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_pkey PRIMARY KEY (id);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT bookings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_pkey PRIMARY KEY (id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_pkey PRIMARY KEY (id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_pkey PRIMARY KEY (id);

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY parent_membership_children
    ADD CONSTRAINT parent_membership_children_pkey PRIMARY KEY (id);

ALTER TABLE ONLY invoice_number_sequences
    ADD CONSTRAINT invoice_number_sequences_pkey PRIMARY KEY (tenant_id, branch_id, billing_year, billing_month);

ALTER TABLE ONLY invoice_runs
    ADD CONSTRAINT invoice_runs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY invoice_runs
    ADD CONSTRAINT invoice_runs_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY invoice_lines
    ADD CONSTRAINT invoice_lines_pkey PRIMARY KEY (id);

ALTER TABLE ONLY invoice_lines
    ADD CONSTRAINT invoice_lines_scope_id_unique UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY payment_attempts
    ADD CONSTRAINT payment_attempts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY payment_attempts
    ADD CONSTRAINT uq_payment_attempts_scoped_id UNIQUE (tenant_id, branch_id, id);

ALTER TABLE ONLY payment_reconciliation_records
    ADD CONSTRAINT payment_reconciliation_records_pkey PRIMARY KEY (id);

ALTER TABLE ONLY stripe_webhook_events
    ADD CONSTRAINT stripe_webhook_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY stripe_webhook_events
    ADD CONSTRAINT stripe_webhook_events_stripe_event_id_key UNIQUE (stripe_event_id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_pkey PRIMARY KEY (id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);

ALTER TABLE ONLY refresh_tokens
    ADD CONSTRAINT refresh_tokens_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);

ALTER TABLE ONLY password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_token_hash_key UNIQUE (token_hash);

-- Indexes

CREATE UNIQUE INDEX attendance_events_scope_id_unique ON attendance_events USING btree (tenant_id, branch_id, id);

CREATE INDEX idx_absence_markers_active_scope_date ON absence_markers USING btree (tenant_id, branch_id, local_date) WHERE (cleared_at IS NULL);

CREATE UNIQUE INDEX idx_absence_markers_active_unique ON absence_markers USING btree (tenant_id, branch_id, child_id, local_date) WHERE (cleared_at IS NULL);

CREATE UNIQUE INDEX idx_absence_markers_scope_id ON absence_markers USING btree (tenant_id, branch_id, id);

CREATE INDEX idx_attendance_events_child_date ON attendance_events USING btree (tenant_id, branch_id, child_id, local_date DESC);

CREATE INDEX idx_attendance_events_session_time ON attendance_events USING btree (tenant_id, branch_id, session_id, occurred_at);

CREATE INDEX idx_attendance_sessions_child_date ON attendance_sessions USING btree (tenant_id, branch_id, child_id, check_in_local_date DESC);

CREATE UNIQUE INDEX idx_attendance_sessions_one_open_child ON attendance_sessions USING btree (tenant_id, branch_id, child_id) WHERE (status = 'open'::text);

CREATE INDEX idx_attendance_sessions_open_scope ON attendance_sessions USING btree (tenant_id, branch_id, status, check_in_at) WHERE (status = 'open'::text);

CREATE INDEX idx_audit_logs_actor_membership ON audit_logs USING btree (actor_membership_id) WHERE (actor_membership_id IS NOT NULL);

CREATE INDEX idx_audit_logs_entity ON audit_logs USING btree (action_entity_type, action_entity_id);

CREATE INDEX idx_audit_logs_request_id ON audit_logs USING btree (request_id) WHERE (request_id IS NOT NULL);

CREATE INDEX idx_audit_logs_scope_time ON audit_logs USING btree (tenant_id, branch_id, created_at DESC);

CREATE INDEX idx_branches_active_tenant ON branches USING btree (tenant_id) WHERE (is_active = true);

CREATE INDEX idx_branches_core_hourly_rate ON branches USING btree (tenant_id) WHERE (core_hourly_rate_minor IS NOT NULL);

CREATE INDEX idx_branches_tenant_id ON branches USING btree (tenant_id);

CREATE INDEX idx_child_billing_profiles_child ON child_billing_profiles USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_child_collection_settings_child ON child_collection_settings USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_child_consent_records_child ON child_consent_records USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_child_contacts_child_type ON child_contacts USING btree (tenant_id, branch_id, child_id, contact_type, sort_order);

CREATE INDEX idx_child_funding_records_child ON child_funding_records USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_child_health_profiles_child ON child_health_profiles USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_child_leaving_records_child ON child_leaving_records USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_child_profiles_child ON child_profiles USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_child_room_assignments_child_current ON child_room_assignments USING btree (tenant_id, branch_id, child_id) WHERE is_current;

CREATE INDEX idx_child_room_assignments_room_current ON child_room_assignments USING btree (tenant_id, branch_id, room_id) WHERE is_current;

CREATE INDEX idx_child_safeguarding_profiles_child ON child_safeguarding_profiles USING btree (tenant_id, branch_id, child_id);

CREATE INDEX idx_children_active ON children USING btree (tenant_id, branch_id, is_active);

CREATE INDEX idx_children_scope ON children USING btree (tenant_id, branch_id);

CREATE INDEX idx_funding_profiles_child_month ON funding_profiles USING btree (tenant_id, branch_id, child_id, billing_month);

CREATE INDEX idx_funding_profiles_scope_month ON funding_profiles USING btree (tenant_id, branch_id, billing_month);

CREATE INDEX idx_invoice_lines_invoice_order ON invoice_lines USING btree (tenant_id, branch_id, invoice_id, sort_order);

CREATE INDEX idx_invoice_runs_billing_scope ON invoice_runs USING btree (tenant_id, branch_id, billing_month, run_type, started_at DESC);

CREATE INDEX idx_invoice_runs_request_id ON invoice_runs USING btree (request_id) WHERE (request_id IS NOT NULL);

CREATE INDEX idx_invoices_adjusts ON invoices USING btree (tenant_id, branch_id, adjusts_invoice_id) WHERE (adjusts_invoice_id IS NOT NULL);

CREATE INDEX idx_invoices_billing_status ON invoices USING btree (tenant_id, branch_id, billing_month, status);

CREATE INDEX idx_invoices_child_billing ON invoices USING btree (tenant_id, branch_id, child_id, billing_month DESC);

CREATE INDEX idx_invoices_due_at_outstanding ON invoices USING btree (tenant_id, branch_id, due_at) WHERE (status = ANY (ARRAY['issued'::text, 'overdue'::text]));

CREATE UNIQUE INDEX idx_invoices_invoice_number_unique ON invoices USING btree (tenant_id, branch_id, invoice_number) WHERE (invoice_number IS NOT NULL);

CREATE UNIQUE INDEX idx_invoices_monthly_unique ON invoices USING btree (tenant_id, branch_id, child_id, billing_month) WHERE (invoice_kind = 'monthly'::text);

CREATE INDEX idx_manager_invites_expires_at ON manager_invites USING btree (expires_at);

CREATE INDEX idx_manager_invites_pending_lookup ON manager_invites USING btree (tenant_id, branch_id, email_normalized, expires_at DESC) WHERE ((accepted_at IS NULL) AND (revoked_at IS NULL));

CREATE INDEX idx_manager_invites_pending_manager ON manager_invites USING btree (tenant_id, branch_id) WHERE ((role = 'manager'::text) AND (accepted_at IS NULL) AND (revoked_at IS NULL));

CREATE INDEX idx_manager_invites_scope_created ON manager_invites USING btree (tenant_id, branch_id, created_at DESC);

CREATE INDEX idx_manager_invites_scope_email ON manager_invites USING btree (tenant_id, branch_id, email_normalized);

CREATE INDEX idx_memberships_active_managers ON memberships USING btree (tenant_id, branch_id) WHERE ((role = 'manager'::text) AND (is_active = true) AND (ended_at IS NULL));

CREATE UNIQUE INDEX idx_memberships_branch_user ON memberships USING btree (tenant_id, branch_id, user_id) WHERE (role = ANY (ARRAY['manager'::text, 'practitioner'::text, 'parent'::text]));

CREATE UNIQUE INDEX idx_memberships_owner_tenant_user ON memberships USING btree (tenant_id, user_id) WHERE (role = 'owner'::text);

CREATE INDEX idx_memberships_scope ON memberships USING btree (tenant_id, branch_id, role);

CREATE INDEX idx_memberships_scope_active ON memberships USING btree (tenant_id, branch_id, role, is_active);

CREATE INDEX idx_memberships_user_active ON memberships USING btree (user_id, is_active, created_at);

CREATE INDEX idx_memberships_user_id ON memberships USING btree (user_id);

CREATE INDEX idx_password_reset_tokens_active_user ON password_reset_tokens USING btree (user_id, created_at DESC) WHERE ((used_at IS NULL) AND (superseded_at IS NULL));

CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens USING btree (expires_at);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens USING btree (user_id);

CREATE INDEX idx_payment_attempts_invoice_created ON payment_attempts USING btree (tenant_id, branch_id, invoice_id, created_at DESC);

CREATE INDEX idx_payment_attempts_open_attempts ON payment_attempts USING btree (tenant_id, branch_id, invoice_id, status) WHERE (status = ANY (ARRAY['checkout_creation_started'::text, 'checkout_created'::text]));

CREATE INDEX idx_reconciliation_attempt_created ON payment_reconciliation_records USING btree (tenant_id, branch_id, payment_attempt_id, created_at DESC);

CREATE INDEX idx_reconciliation_invoice_created ON payment_reconciliation_records USING btree (tenant_id, branch_id, invoice_id, created_at DESC);

CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens USING btree (expires_at);

CREATE INDEX idx_refresh_tokens_membership_id ON refresh_tokens USING btree (membership_id);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens USING btree (user_id);

CREATE INDEX idx_rooms_active ON rooms USING btree (branch_id, is_active);

CREATE UNIQUE INDEX idx_rooms_active_name_per_branch ON rooms USING btree (branch_id, name) WHERE (is_active = true);

CREATE INDEX idx_rooms_branch_id ON rooms USING btree (branch_id);

CREATE INDEX idx_rooms_tenant_id ON rooms USING btree (tenant_id);

CREATE INDEX idx_stripe_webhook_events_event_type ON stripe_webhook_events USING btree (event_type, received_at DESC);

CREATE INDEX idx_stripe_webhook_events_processing_status ON stripe_webhook_events USING btree (processing_status, received_at DESC);

CREATE UNIQUE INDEX uq_payment_attempts_stripe_session_id ON payment_attempts USING btree (stripe_checkout_session_id) WHERE (stripe_checkout_session_id IS NOT NULL);

CREATE UNIQUE INDEX uq_reconciliation_stripe_event_id ON payment_reconciliation_records USING btree (stripe_event_id);

CREATE UNIQUE INDEX session_types_active_name_unique ON session_types USING btree (tenant_id, branch_id, name) WHERE (is_active = true);

CREATE INDEX session_types_active_by_branch ON session_types USING btree (tenant_id, branch_id) WHERE (is_active = true);

CREATE INDEX session_types_branch_id ON session_types USING btree (branch_id);

CREATE INDEX session_types_tenant_id ON session_types USING btree (tenant_id);

CREATE UNIQUE INDEX child_booking_patterns_one_open_per_child ON child_booking_patterns USING btree (tenant_id, branch_id, child_id) WHERE is_current;

CREATE INDEX child_booking_patterns_by_child ON child_booking_patterns USING btree (tenant_id, branch_id, child_id, effective_from DESC);

CREATE INDEX child_booking_patterns_branch_id ON child_booking_patterns USING btree (branch_id);

CREATE INDEX child_booking_patterns_tenant_id ON child_booking_patterns USING btree (tenant_id);

CREATE UNIQUE INDEX child_booking_pattern_entries_unique_day ON child_booking_pattern_entries USING btree (tenant_id, branch_id, pattern_id, day_of_week);

CREATE INDEX child_booking_pattern_entries_by_pattern ON child_booking_pattern_entries USING btree (tenant_id, branch_id, pattern_id, day_of_week);

CREATE INDEX child_booking_pattern_entries_branch_id ON child_booking_pattern_entries USING btree (branch_id);

CREATE INDEX child_booking_pattern_entries_tenant_id ON child_booking_pattern_entries USING btree (tenant_id);

CREATE UNIQUE INDEX session_templates_active_name_unique
    ON session_templates USING btree (tenant_id, branch_id, name)
    WHERE (is_active = true);

CREATE INDEX session_templates_active_by_branch
    ON session_templates USING btree (tenant_id, branch_id)
    WHERE (is_active = true);

CREATE UNIQUE INDEX session_template_entries_unique_day_session
    ON session_template_entries USING btree (tenant_id, branch_id, template_id, day_of_week, session_type_id);

CREATE INDEX session_template_entries_by_template
    ON session_template_entries USING btree (tenant_id, branch_id, template_id, day_of_week);

CREATE INDEX idx_bookings_tenant_branch_child ON bookings USING btree (tenant_id, branch_id, child_id);
CREATE INDEX idx_bookings_tenant_branch_room_dates ON bookings USING btree (tenant_id, branch_id, room_id, effective_start_date);
CREATE INDEX idx_bookings_tenant_branch_status ON bookings USING btree (tenant_id, branch_id, status);

CREATE UNIQUE INDEX term_one_active_per_child ON term USING btree (tenant_id, branch_id, child_id)
    WHERE (status = ANY (ARRAY['pre_term'::text, 'active'::text, 'pending_renewal'::text]));

CREATE INDEX term_by_child ON term USING btree (tenant_id, branch_id, child_id, term_start_date DESC);

CREATE INDEX term_active_by_branch ON term USING btree (tenant_id, branch_id)
    WHERE (status = ANY (ARRAY['pre_term'::text, 'active'::text, 'pending_renewal'::text]));

CREATE INDEX term_ending_soon ON term USING btree (tenant_id, branch_id, term_end_date)
    WHERE (status = ANY (ARRAY['active'::text, 'pending_renewal'::text]));

CREATE INDEX term_branch_id ON term USING btree (branch_id);

CREATE INDEX term_tenant_id ON term USING btree (tenant_id);

CREATE UNIQUE INDEX invoice_run_advance_one_per_month ON invoice_run_advance USING btree (tenant_id, branch_id, billing_month);

CREATE INDEX invoice_run_advance_branch_id ON invoice_run_advance USING btree (branch_id);

CREATE INDEX invoice_run_advance_tenant_id ON invoice_run_advance USING btree (tenant_id);

CREATE UNIQUE INDEX idx_parent_membership_children_active_pair
    ON parent_membership_children USING btree (tenant_id, branch_id, membership_id, child_id)
    WHERE (ended_at IS NULL);

CREATE INDEX idx_parent_membership_children_child_active
    ON parent_membership_children USING btree (child_id) WHERE (ended_at IS NULL);

CREATE INDEX idx_parent_membership_children_membership_active
    ON parent_membership_children USING btree (membership_id) WHERE (ended_at IS NULL);

CREATE INDEX idx_site_profiles_tenant_branch ON site_profiles(tenant_id, branch_id);

CREATE UNIQUE INDEX idx_academic_terms_tenant_branch_name_active
    ON academic_terms(tenant_id, branch_id, name) WHERE is_active = true;

CREATE INDEX idx_academic_terms_tenant_branch ON academic_terms(tenant_id, branch_id);
CREATE INDEX idx_academic_terms_dates ON academic_terms(tenant_id, branch_id, start_date, end_date);

CREATE INDEX idx_ad_hoc_bookings_tenant_branch_child_date
    ON ad_hoc_bookings(tenant_id, branch_id, child_id, calendar_date);
CREATE INDEX idx_ad_hoc_bookings_tenant_branch_status
    ON ad_hoc_bookings(tenant_id, branch_id, status);

CREATE UNIQUE INDEX branch_closure_days_tenant_branch_date_idx
    ON branch_closure_days (tenant_id, branch_id, date);

CREATE INDEX branch_closure_days_tenant_branch_month_idx
    ON branch_closure_days (tenant_id, branch_id, date);

CREATE INDEX idx_funding_history_lookup
    ON child_funding_history (tenant_id, branch_id, child_id, changed_at DESC);

CREATE UNIQUE INDEX idx_payment_links_active_per_invoice
    ON payment_links (tenant_id, branch_id, invoice_id)
    WHERE status = 'active';

CREATE INDEX idx_payment_links_invoice_id
    ON payment_links (tenant_id, branch_id, invoice_id);

-- Triggers

CREATE TRIGGER trg_invoice_immutability BEFORE UPDATE ON invoices FOR EACH ROW EXECUTE FUNCTION protect_issued_invoice_immutability();

CREATE TRIGGER trg_invoice_lines_immutability BEFORE INSERT OR DELETE OR UPDATE ON invoice_lines FOR EACH ROW EXECUTE FUNCTION protect_issued_invoice_lines();

CREATE TRIGGER trg_invoice_status_transition BEFORE UPDATE OF status ON invoices FOR EACH ROW EXECUTE FUNCTION enforce_invoice_status_transition();

CREATE TRIGGER parent_membership_children_role_check
    BEFORE INSERT OR UPDATE ON parent_membership_children
    FOR EACH ROW EXECUTE FUNCTION enforce_parent_membership_child_role();

CREATE TRIGGER parent_membership_children_scope_check
    BEFORE INSERT OR UPDATE OF membership_id, child_id, ended_at ON parent_membership_children
    FOR EACH ROW EXECUTE FUNCTION enforce_parent_membership_child_scope();

CREATE TRIGGER memberships_end_cascade_children
    AFTER UPDATE OF ended_at ON memberships
    FOR EACH ROW EXECUTE FUNCTION cascade_parent_membership_child_end();

-- Foreign keys

ALTER TABLE ONLY branches
    ADD CONSTRAINT branches_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY site_profiles
    ADD CONSTRAINT site_profiles_branch_id_fkey FOREIGN KEY (branch_id)
        REFERENCES branches(id) ON DELETE CASCADE;

ALTER TABLE ONLY academic_terms
    ADD CONSTRAINT academic_terms_branch_id_fkey FOREIGN KEY (branch_id)
        REFERENCES branches(id) ON DELETE CASCADE;

ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);

ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY rooms
    ADD CONSTRAINT rooms_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY rooms
    ADD CONSTRAINT rooms_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);

ALTER TABLE ONLY children
    ADD CONSTRAINT children_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY children
    ADD CONSTRAINT children_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY children
    ADD CONSTRAINT children_current_term_fkey FOREIGN KEY (tenant_id, branch_id, current_term_id) REFERENCES term(tenant_id, branch_id, id);

ALTER TABLE ONLY child_profiles
    ADD CONSTRAINT child_profiles_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_profiles
    ADD CONSTRAINT child_profiles_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_contacts
    ADD CONSTRAINT child_contacts_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_health_profiles
    ADD CONSTRAINT child_health_profiles_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_health_profiles
    ADD CONSTRAINT child_health_profiles_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_safeguarding_profiles
    ADD CONSTRAINT child_safeguarding_profiles_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_safeguarding_profiles
    ADD CONSTRAINT child_safeguarding_profiles_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_consent_records
    ADD CONSTRAINT child_consent_records_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_consent_records
    ADD CONSTRAINT child_consent_records_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_consent_records
    ADD CONSTRAINT child_consent_records_entered_by_membership_id_fkey FOREIGN KEY (entered_by_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY child_consent_records
    ADD CONSTRAINT child_consent_records_entered_by_user_id_fkey FOREIGN KEY (entered_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY child_funding_records
    ADD CONSTRAINT child_funding_records_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_funding_records
    ADD CONSTRAINT child_funding_records_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_billing_profiles
    ADD CONSTRAINT child_billing_profiles_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_billing_profiles
    ADD CONSTRAINT child_billing_profiles_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_collection_settings
    ADD CONSTRAINT child_collection_settings_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_collection_settings
    ADD CONSTRAINT child_collection_settings_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_collection_settings
    ADD CONSTRAINT child_collection_settings_collection_password_updated_by_m_fkey FOREIGN KEY (collection_password_updated_by_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY child_collection_settings
    ADD CONSTRAINT child_collection_settings_collection_password_updated_by_u_fkey FOREIGN KEY (collection_password_updated_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY child_room_assignments
    ADD CONSTRAINT child_room_assignments_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_room_assignments
    ADD CONSTRAINT child_room_assignments_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_room_assignments
    ADD CONSTRAINT child_room_assignments_room_id_fkey FOREIGN KEY (room_id) REFERENCES rooms(id);

ALTER TABLE ONLY child_leaving_records
    ADD CONSTRAINT child_leaving_records_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_leaving_records
    ADD CONSTRAINT child_leaving_records_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_sessions
    ADD CONSTRAINT attendance_sessions_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY attendance_sessions
    ADD CONSTRAINT attendance_sessions_child_scope_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_sessions
    ADD CONSTRAINT attendance_sessions_check_in_event_fkey FOREIGN KEY (tenant_id, branch_id, check_in_event_id) REFERENCES attendance_events(tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_sessions
    ADD CONSTRAINT attendance_sessions_check_out_event_fkey FOREIGN KEY (tenant_id, branch_id, check_out_event_id) REFERENCES attendance_events(tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_sessions
    ADD CONSTRAINT attendance_sessions_corrected_by_event_fkey FOREIGN KEY (tenant_id, branch_id, corrected_by_event_id) REFERENCES attendance_events(tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_events
    ADD CONSTRAINT attendance_events_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY attendance_events
    ADD CONSTRAINT attendance_events_child_scope_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_events
    ADD CONSTRAINT attendance_events_session_scope_fkey FOREIGN KEY (tenant_id, branch_id, session_id) REFERENCES attendance_sessions(tenant_id, branch_id, id);

ALTER TABLE ONLY attendance_events
    ADD CONSTRAINT attendance_events_recorded_by_user_id_fkey FOREIGN KEY (recorded_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY attendance_events
    ADD CONSTRAINT attendance_events_membership_scope_fkey FOREIGN KEY (tenant_id, branch_id, recorded_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY absence_markers
    ADD CONSTRAINT absence_markers_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY absence_markers
    ADD CONSTRAINT absence_markers_child_scope_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY absence_markers
    ADD CONSTRAINT absence_markers_marked_by_user_id_fkey FOREIGN KEY (marked_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY absence_markers
    ADD CONSTRAINT absence_markers_marked_membership_scope_fkey FOREIGN KEY (tenant_id, branch_id, marked_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY absence_markers
    ADD CONSTRAINT absence_markers_cleared_membership_scope_fkey FOREIGN KEY (tenant_id, branch_id, cleared_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY funding_profiles
    ADD CONSTRAINT funding_profiles_child_scope_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY session_types
    ADD CONSTRAINT session_types_branch_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY child_booking_patterns
    ADD CONSTRAINT child_booking_patterns_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_pattern_fkey FOREIGN KEY (tenant_id, branch_id, pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);

ALTER TABLE ONLY child_booking_pattern_entries
    ADD CONSTRAINT child_booking_pattern_entries_session_type_fkey FOREIGN KEY (tenant_id, branch_id, session_type_id) REFERENCES session_types(tenant_id, branch_id, id);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY session_templates
    ADD CONSTRAINT session_templates_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_template_fkey FOREIGN KEY (tenant_id, branch_id, template_id) REFERENCES session_templates(tenant_id, branch_id, id);

ALTER TABLE ONLY session_template_entries
    ADD CONSTRAINT session_template_entries_session_type_fkey FOREIGN KEY (tenant_id, branch_id, session_type_id) REFERENCES session_types(tenant_id, branch_id, id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT bookings_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT bookings_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT bookings_child_id_fkey FOREIGN KEY (child_id) REFERENCES children(id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT bookings_session_template_id_fkey FOREIGN KEY (session_template_id) REFERENCES session_templates(id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT bookings_room_id_fkey FOREIGN KEY (room_id) REFERENCES rooms(id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_booking_pattern_fkey FOREIGN KEY (tenant_id, branch_id, booking_pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_created_by_membership_fkey FOREIGN KEY (tenant_id, branch_id, created_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_term_fkey FOREIGN KEY (tenant_id, branch_id, term_id) REFERENCES term(tenant_id, branch_id, id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_previous_pattern_fkey FOREIGN KEY (tenant_id, branch_id, previous_booking_pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_new_pattern_fkey FOREIGN KEY (tenant_id, branch_id, new_booking_pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);

ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES users(id);

ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_actor_membership_fkey FOREIGN KEY (actor_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY parent_membership_children
    ADD CONSTRAINT parent_membership_children_branch_scope_fkey
    FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY parent_membership_children
    ADD CONSTRAINT parent_membership_children_membership_scope_fkey
    FOREIGN KEY (tenant_id, branch_id, membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY parent_membership_children
    ADD CONSTRAINT parent_membership_children_child_scope_fkey
    FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY invoice_number_sequences
    ADD CONSTRAINT invoice_number_sequences_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY invoice_runs
    ADD CONSTRAINT invoice_runs_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY invoice_runs
    ADD CONSTRAINT invoice_runs_requested_by_user_id_fkey FOREIGN KEY (requested_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY invoice_runs
    ADD CONSTRAINT invoice_runs_membership_scope_fkey FOREIGN KEY (tenant_id, branch_id, requested_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_child_scope_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_generated_run_scope_fkey FOREIGN KEY (tenant_id, branch_id, generated_run_id) REFERENCES invoice_runs(tenant_id, branch_id, id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_issued_run_scope_fkey FOREIGN KEY (tenant_id, branch_id, issued_run_id) REFERENCES invoice_runs(tenant_id, branch_id, id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_adjusts_invoice_scope_fkey FOREIGN KEY (tenant_id, branch_id, adjusts_invoice_id) REFERENCES invoices(tenant_id, branch_id, id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_issued_by_user_id_fkey FOREIGN KEY (issued_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY invoices
    ADD CONSTRAINT invoices_issued_by_membership_scope_fkey FOREIGN KEY (tenant_id, branch_id, issued_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY invoice_lines
    ADD CONSTRAINT invoice_lines_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY invoice_lines
    ADD CONSTRAINT invoice_lines_invoice_scope_fkey FOREIGN KEY (tenant_id, branch_id, invoice_id) REFERENCES invoices(tenant_id, branch_id, id);

ALTER TABLE ONLY ad_hoc_bookings
    ADD CONSTRAINT ad_hoc_bookings_child_id_fkey FOREIGN KEY (child_id) REFERENCES children(id);

ALTER TABLE ONLY ad_hoc_bookings
    ADD CONSTRAINT ad_hoc_bookings_session_type_id_fkey FOREIGN KEY (session_type_id) REFERENCES session_types(id);

ALTER TABLE ONLY ad_hoc_bookings
    ADD CONSTRAINT ad_hoc_bookings_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;

ALTER TABLE ONLY hourly_bookings
    ADD CONSTRAINT hourly_bookings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY hourly_bookings
    ADD CONSTRAINT hourly_bookings_child_id_fkey FOREIGN KEY (child_id) REFERENCES children(id);

ALTER TABLE ONLY hourly_bookings
    ADD CONSTRAINT hourly_bookings_session_type_id_fkey FOREIGN KEY (session_type_id) REFERENCES session_types(id);

ALTER TABLE ONLY hourly_bookings
    ADD CONSTRAINT hourly_bookings_booked_by_membership_id_fkey FOREIGN KEY (booked_by_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY payment_attempts
    ADD CONSTRAINT fk_payment_attempts_branch FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY payment_attempts
    ADD CONSTRAINT fk_payment_attempts_invoice FOREIGN KEY (tenant_id, branch_id, invoice_id) REFERENCES invoices(tenant_id, branch_id, id);

ALTER TABLE ONLY payment_attempts
    ADD CONSTRAINT fk_payment_attempts_membership FOREIGN KEY (tenant_id, branch_id, initiated_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);

ALTER TABLE ONLY payment_attempts
    ADD CONSTRAINT payment_attempts_initiated_by_user_id_fkey FOREIGN KEY (initiated_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY payment_reconciliation_records
    ADD CONSTRAINT fk_reconciliation_branch FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY payment_reconciliation_records
    ADD CONSTRAINT fk_reconciliation_invoice FOREIGN KEY (tenant_id, branch_id, invoice_id) REFERENCES invoices(tenant_id, branch_id, id);

ALTER TABLE ONLY payment_reconciliation_records
    ADD CONSTRAINT fk_reconciliation_attempt FOREIGN KEY (tenant_id, branch_id, payment_attempt_id) REFERENCES payment_attempts(tenant_id, branch_id, id);

ALTER TABLE ONLY payment_reconciliation_records
    ADD CONSTRAINT fk_reconciliation_webhook_event FOREIGN KEY (stripe_webhook_event_id) REFERENCES stripe_webhook_events(id);

ALTER TABLE ONLY invoice_reminder_log
    ADD CONSTRAINT invoice_reminder_log_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES invoices(id);

ALTER TABLE ONLY payment_links
    ADD CONSTRAINT payment_links_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES invoices(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_branch_scope_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_accepted_user_id_fkey FOREIGN KEY (accepted_user_id) REFERENCES users(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_accepted_membership_id_fkey FOREIGN KEY (accepted_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_created_by_membership_id_fkey FOREIGN KEY (created_by_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_revoked_by_user_id_fkey FOREIGN KEY (revoked_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_revoked_by_membership_id_fkey FOREIGN KEY (revoked_by_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_resent_by_user_id_fkey FOREIGN KEY (resent_by_user_id) REFERENCES users(id);

ALTER TABLE ONLY manager_invites
    ADD CONSTRAINT manager_invites_resent_by_membership_id_fkey FOREIGN KEY (resent_by_membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY refresh_tokens
    ADD CONSTRAINT refresh_tokens_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES memberships(id);

ALTER TABLE ONLY password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
