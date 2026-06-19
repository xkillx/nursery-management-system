-- 000003_terms: Term (12-month fixed-term advance-pay invoicing) + schedule changes + invoice run advance log.
-- Replaces the attendance-actuals billing source. Adds:
--   * term: per-child commercial commitment, 12-month fixed term.
--   * term_schedule_change: in-term adjustment audit trail (decrease/increase).
--   * invoice_run_advance: scheduled monthly generation log.
--   * children.current_term_id: denormalisation for fast lookups.

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

ALTER TABLE ONLY term
    ADD CONSTRAINT term_pkey PRIMARY KEY (id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_scope_id_unique UNIQUE (tenant_id, branch_id, id);

-- At most one non-historical (pre_term/active/pending_renewal) term per child.
CREATE UNIQUE INDEX term_one_active_per_child ON term USING btree (tenant_id, branch_id, child_id)
    WHERE (status = ANY (ARRAY['pre_term'::text, 'active'::text, 'pending_renewal'::text]));

CREATE INDEX term_by_child ON term USING btree (tenant_id, branch_id, child_id, term_start_date DESC);

CREATE INDEX term_active_by_branch ON term USING btree (tenant_id, branch_id)
    WHERE (status = ANY (ARRAY['pre_term'::text, 'active'::text, 'pending_renewal'::text]));

CREATE INDEX term_ending_soon ON term USING btree (tenant_id, branch_id, term_end_date)
    WHERE (status = ANY (ARRAY['active'::text, 'pending_renewal'::text]));

CREATE INDEX term_branch_id ON term USING btree (branch_id);

CREATE INDEX term_tenant_id ON term USING btree (tenant_id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_child_fkey FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_booking_pattern_fkey FOREIGN KEY (tenant_id, branch_id, booking_pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);

ALTER TABLE ONLY term
    ADD CONSTRAINT term_created_by_membership_fkey FOREIGN KEY (tenant_id, branch_id, created_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id);


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

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_pkey PRIMARY KEY (id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE INDEX term_schedule_change_by_term ON term_schedule_change USING btree (tenant_id, branch_id, term_id, requested_at DESC);

CREATE INDEX term_schedule_change_branch_id ON term_schedule_change USING btree (branch_id);

CREATE INDEX term_schedule_change_tenant_id ON term_schedule_change USING btree (tenant_id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_term_fkey FOREIGN KEY (tenant_id, branch_id, term_id) REFERENCES term(tenant_id, branch_id, id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_previous_pattern_fkey FOREIGN KEY (tenant_id, branch_id, previous_booking_pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);

ALTER TABLE ONLY term_schedule_change
    ADD CONSTRAINT term_schedule_change_new_pattern_fkey FOREIGN KEY (tenant_id, branch_id, new_booking_pattern_id) REFERENCES child_booking_patterns(tenant_id, branch_id, id);


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

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_pkey PRIMARY KEY (id);

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_scope_id_unique UNIQUE (tenant_id, branch_id, id);

CREATE UNIQUE INDEX invoice_run_advance_one_per_month ON invoice_run_advance USING btree (tenant_id, branch_id, billing_month);

CREATE INDEX invoice_run_advance_branch_id ON invoice_run_advance USING btree (branch_id);

CREATE INDEX invoice_run_advance_tenant_id ON invoice_run_advance USING btree (tenant_id);

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_branch_fkey FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id);

ALTER TABLE ONLY invoice_run_advance
    ADD CONSTRAINT invoice_run_advance_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id);


-- Denormalisation: cheap read for "does this child have an active term?".
ALTER TABLE children
    ADD COLUMN current_term_id uuid;

ALTER TABLE ONLY children
    ADD CONSTRAINT children_current_term_fkey FOREIGN KEY (tenant_id, branch_id, current_term_id) REFERENCES term(tenant_id, branch_id, id);
