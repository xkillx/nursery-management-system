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

ALTER TABLE children ADD COLUMN current_term_id uuid;
ALTER TABLE children ADD CONSTRAINT children_current_term_fkey FOREIGN KEY (tenant_id, branch_id, current_term_id) REFERENCES term(tenant_id, branch_id, id);
