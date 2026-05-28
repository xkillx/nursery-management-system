-- Invoice Runs: batch operations for draft generation and issue
CREATE TABLE invoice_runs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    billing_month DATE NOT NULL,
    run_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'started',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    requested_by_user_id UUID NOT NULL REFERENCES users(id),
    requested_by_membership_id UUID NOT NULL,
    request_id TEXT,
    eligible_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    blocked_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT invoice_runs_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT invoice_runs_membership_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, requested_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id),
    CONSTRAINT invoice_runs_billing_month_first_day
        CHECK (billing_month = date_trunc('month', billing_month)::date),
    CONSTRAINT invoice_runs_run_type_valid
        CHECK (run_type IN ('draft_generation', 'issue')),
    CONSTRAINT invoice_runs_status_valid
        CHECK (status IN ('started', 'completed', 'completed_with_exceptions', 'failed')),
    CONSTRAINT invoice_runs_completed_at_consistent
        CHECK ((status = 'started') = (completed_at IS NULL)),
    CONSTRAINT invoice_runs_eligible_count_nonneg
        CHECK (eligible_count >= 0),
    CONSTRAINT invoice_runs_success_count_nonneg
        CHECK (success_count >= 0),
    CONSTRAINT invoice_runs_blocked_count_nonneg
        CHECK (blocked_count >= 0),
    CONSTRAINT invoice_runs_failed_count_nonneg
        CHECK (failed_count >= 0),
    CONSTRAINT invoice_runs_scope_id_unique
        UNIQUE (tenant_id, branch_id, id)
);

CREATE INDEX idx_invoice_runs_billing_scope
    ON invoice_runs (tenant_id, branch_id, billing_month, run_type, started_at DESC);

CREATE INDEX idx_invoice_runs_request_id
    ON invoice_runs (request_id)
    WHERE request_id IS NOT NULL;


-- Invoices: per-child monthly billing statements and future adjustments
CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    billing_month DATE NOT NULL,
    invoice_kind TEXT NOT NULL DEFAULT 'monthly',
    status TEXT NOT NULL DEFAULT 'draft',

    -- Numbering
    invoice_number TEXT,
    issued_sequence INTEGER,

    -- Run links
    generated_run_id UUID,
    issued_run_id UUID,

    -- Issue/lock fields
    issued_at TIMESTAMPTZ,
    issued_by_user_id UUID REFERENCES users(id),
    issued_by_membership_id UUID,
    locked_at TIMESTAMPTZ,

    -- Due/payment-ready fields
    due_at TIMESTAMPTZ,
    currency_code CHAR(3) NOT NULL DEFAULT 'GBP',
    subtotal_minor INTEGER NOT NULL DEFAULT 0,
    funded_deduction_minor INTEGER NOT NULL DEFAULT 0,
    total_due_minor INTEGER NOT NULL DEFAULT 0,
    amount_paid_minor INTEGER NOT NULL DEFAULT 0,
    paid_at TIMESTAMPTZ,
    payment_failed_at TIMESTAMPTZ,
    payment_status_updated_at TIMESTAMPTZ,

    -- Adjustment hooks
    adjusts_invoice_id UUID,
    adjustment_reason_code TEXT,
    adjustment_reason_note TEXT,

    -- Snapshot metadata
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,
    calculation_details JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Foreign keys
    CONSTRAINT invoices_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT invoices_child_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT invoices_issued_by_membership_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, issued_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id),
    CONSTRAINT invoices_generated_run_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, generated_run_id) REFERENCES invoice_runs(tenant_id, branch_id, id),
    CONSTRAINT invoices_issued_run_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, issued_run_id) REFERENCES invoice_runs(tenant_id, branch_id, id),
    CONSTRAINT invoices_adjusts_invoice_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, adjusts_invoice_id) REFERENCES invoices(tenant_id, branch_id, id),

    -- Billing month must be first day
    CONSTRAINT invoices_billing_month_first_day
        CHECK (billing_month = date_trunc('month', billing_month)::date),
    -- Kind and status
    CONSTRAINT invoices_invoice_kind_valid
        CHECK (invoice_kind IN ('monthly', 'adjustment')),
    CONSTRAINT invoices_status_valid
        CHECK (status IN ('draft', 'issued', 'payment_failed', 'paid', 'overdue')),
    -- Currency
    CONSTRAINT invoices_currency_gbp
        CHECK (currency_code = 'GBP'),
    -- Money fields non-negative
    CONSTRAINT invoices_subtotal_nonneg
        CHECK (subtotal_minor >= 0),
    CONSTRAINT invoices_funded_deduction_nonneg
        CHECK (funded_deduction_minor >= 0),
    CONSTRAINT invoices_total_due_nonneg
        CHECK (total_due_minor >= 0),
    CONSTRAINT invoices_amount_paid_nonneg
        CHECK (amount_paid_minor >= 0),
    CONSTRAINT invoices_amount_paid_lte_total
        CHECK (amount_paid_minor <= total_due_minor),
    -- Draft shape: issue fields must be null
    CONSTRAINT invoices_draft_shape
        CHECK (
            (status <> 'draft') OR
            (issued_at IS NULL AND issued_by_user_id IS NULL AND issued_by_membership_id IS NULL
             AND locked_at IS NULL AND due_at IS NULL AND invoice_number IS NULL
             AND issued_sequence IS NULL AND issued_run_id IS NULL)
        ),
    -- Issued-or-later shape: issue fields must be non-null
    CONSTRAINT invoices_issued_shape
        CHECK (
            (status = 'draft') OR
            (issued_at IS NOT NULL AND issued_by_user_id IS NOT NULL AND issued_by_membership_id IS NOT NULL
             AND locked_at IS NOT NULL AND due_at IS NOT NULL AND invoice_number IS NOT NULL
             AND issued_sequence IS NOT NULL AND issued_run_id IS NOT NULL)
        ),
    -- Paid shape
    CONSTRAINT invoices_paid_shape
        CHECK (
            (status <> 'paid') OR
            (paid_at IS NOT NULL AND amount_paid_minor = total_due_minor)
        ),
    -- Payment failed shape
    CONSTRAINT invoices_payment_failed_shape
        CHECK (
            (status <> 'payment_failed') OR
            (payment_failed_at IS NOT NULL)
        ),
    -- Adjustment shape: adjustment kind requires adjusts_invoice_id and non-empty reason
    CONSTRAINT invoices_adjustment_shape
        CHECK (
            (invoice_kind <> 'adjustment') OR
            (adjusts_invoice_id IS NOT NULL
             AND adjustment_reason_code IS NOT NULL
             AND BTRIM(adjustment_reason_code) <> ''
             AND adjustment_reason_note IS NOT NULL
             AND BTRIM(adjustment_reason_note) <> '')
        ),
    -- Monthly shape: no adjustment fields
    CONSTRAINT invoices_monthly_shape
        CHECK (
            (invoice_kind <> 'monthly') OR
            (adjusts_invoice_id IS NULL AND adjustment_reason_code IS NULL AND adjustment_reason_note IS NULL)
        ),
    -- No self-reference
    CONSTRAINT invoices_no_self_adjust
        CHECK (adjusts_invoice_id IS NULL OR adjusts_invoice_id <> id),
    -- Period bounds
    CONSTRAINT invoices_period_range
        CHECK (period_start_date <= period_end_date),
    CONSTRAINT invoices_scope_id_unique
        UNIQUE (tenant_id, branch_id, id)
);

CREATE UNIQUE INDEX idx_invoices_monthly_unique
    ON invoices (tenant_id, branch_id, child_id, billing_month)
    WHERE invoice_kind = 'monthly';

CREATE UNIQUE INDEX idx_invoices_invoice_number_unique
    ON invoices (tenant_id, branch_id, invoice_number)
    WHERE invoice_number IS NOT NULL;

CREATE INDEX idx_invoices_billing_status
    ON invoices (tenant_id, branch_id, billing_month, status);

CREATE INDEX idx_invoices_child_billing
    ON invoices (tenant_id, branch_id, child_id, billing_month DESC);

CREATE INDEX idx_invoices_adjusts
    ON invoices (tenant_id, branch_id, adjusts_invoice_id)
    WHERE adjusts_invoice_id IS NOT NULL;

CREATE INDEX idx_invoices_due_at_outstanding
    ON invoices (tenant_id, branch_id, due_at)
    WHERE status IN ('issued', 'overdue');


-- Invoice Lines: accounting line items and calculation snapshots
CREATE TABLE invoice_lines (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    invoice_id UUID NOT NULL,
    line_kind TEXT NOT NULL,
    description TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,

    -- Money/quantity
    quantity_minutes INTEGER,
    unit_amount_minor INTEGER,
    line_amount_minor INTEGER NOT NULL,

    -- Calculation snapshot
    raw_attended_minutes INTEGER,
    rounded_attended_minutes INTEGER,
    funded_allowance_minutes INTEGER,
    funded_deduction_minutes INTEGER,
    core_billable_minutes INTEGER,
    session_count INTEGER,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT invoice_lines_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT invoice_lines_invoice_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, invoice_id) REFERENCES invoices(tenant_id, branch_id, id),

    -- Line kind
    CONSTRAINT invoice_lines_line_kind_valid
        CHECK (line_kind IN ('core_childcare', 'funded_deduction', 'extra', 'adjustment')),
    -- Quantity/amount bounds
    CONSTRAINT invoice_lines_quantity_nonneg
        CHECK (quantity_minutes IS NULL OR quantity_minutes >= 0),
    CONSTRAINT invoice_lines_unit_amount_nonneg
        CHECK (unit_amount_minor IS NULL OR unit_amount_minor >= 0),
    -- Minute snapshot bounds
    CONSTRAINT invoice_lines_raw_attended_nonneg
        CHECK (raw_attended_minutes IS NULL OR raw_attended_minutes >= 0),
    CONSTRAINT invoice_lines_rounded_attended_nonneg
        CHECK (rounded_attended_minutes IS NULL OR rounded_attended_minutes >= 0),
    CONSTRAINT invoice_lines_funded_allowance_nonneg
        CHECK (funded_allowance_minutes IS NULL OR funded_allowance_minutes >= 0),
    CONSTRAINT invoice_lines_funded_deduction_nonneg
        CHECK (funded_deduction_minutes IS NULL OR funded_deduction_minutes >= 0),
    CONSTRAINT invoice_lines_core_billable_nonneg
        CHECK (core_billable_minutes IS NULL OR core_billable_minutes >= 0),
    CONSTRAINT invoice_lines_session_count_nonneg
        CHECK (session_count IS NULL OR session_count >= 0),
    -- Core/extra lines non-negative
    CONSTRAINT invoice_lines_core_amount_nonneg
        CHECK (line_kind <> 'core_childcare' OR line_amount_minor >= 0),
    CONSTRAINT invoice_lines_extra_amount_nonneg
        CHECK (line_kind <> 'extra' OR line_amount_minor >= 0),
    -- Funded deduction lines non-positive
    CONSTRAINT invoice_lines_funded_deduction_nonpos
        CHECK (line_kind <> 'funded_deduction' OR line_amount_minor <= 0),
    CONSTRAINT invoice_lines_scope_id_unique
        UNIQUE (tenant_id, branch_id, id)
);

CREATE INDEX idx_invoice_lines_invoice_order
    ON invoice_lines (tenant_id, branch_id, invoice_id, sort_order);


-- Invoice Number Sequences: per-branch, per-month sequential numbering
CREATE TABLE invoice_number_sequences (
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    billing_year INTEGER NOT NULL,
    billing_month INTEGER NOT NULL,
    next_sequence INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT invoice_number_sequences_pkey
        PRIMARY KEY (tenant_id, branch_id, billing_year, billing_month),
    CONSTRAINT invoice_number_sequences_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT invoice_number_sequences_year_gte_2000
        CHECK (billing_year >= 2000),
    CONSTRAINT invoice_number_sequences_month_range
        CHECK (billing_month BETWEEN 1 AND 12),
    CONSTRAINT invoice_number_sequences_next_seq_gte_1
        CHECK (next_sequence >= 1)
);


-- Trigger: enforce legal status transitions on invoices
CREATE OR REPLACE FUNCTION enforce_invoice_status_transition()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF OLD.status = NEW.status THEN
            RETURN NEW;
        END IF;

        -- paid is terminal
        IF OLD.status = 'paid' THEN
            RAISE EXCEPTION 'invoice % is paid and cannot transition', OLD.id;
        END IF;

        -- cannot go back to draft
        IF NEW.status = 'draft' THEN
            RAISE EXCEPTION 'invoice % cannot transition back to draft', OLD.id;
        END IF;

        -- payment_failed cannot go to overdue
        IF OLD.status = 'payment_failed' AND NEW.status = 'overdue' THEN
            RAISE EXCEPTION 'invoice % is payment_failed and cannot transition to overdue', OLD.id;
        END IF;

        -- legal transitions
        CASE OLD.status
            WHEN 'draft' THEN
                IF NEW.status NOT IN ('issued') THEN
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
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_invoice_status_transition
    BEFORE UPDATE OF status ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION enforce_invoice_status_transition();


-- Trigger: protect issued invoice header immutability
CREATE OR REPLACE FUNCTION protect_issued_invoice_immutability()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status <> 'draft' THEN
        -- Only these fields may change after issue
        IF NEW.status IS DISTINCT FROM OLD.status
           OR NEW.amount_paid_minor IS DISTINCT FROM OLD.amount_paid_minor
           OR NEW.paid_at IS DISTINCT FROM OLD.paid_at
           OR NEW.payment_failed_at IS DISTINCT FROM OLD.payment_failed_at
           OR NEW.payment_status_updated_at IS DISTINCT FROM OLD.payment_status_updated_at
           OR NEW.updated_at IS DISTINCT FROM OLD.updated_at THEN
            -- Verify no other fields changed
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
            -- No permitted field changed — reject
            RAISE EXCEPTION 'invoice % is not draft and cannot be modified', OLD.id;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_invoice_immutability
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION protect_issued_invoice_immutability();


-- Trigger: protect invoice lines when parent invoice is not draft
CREATE OR REPLACE FUNCTION protect_issued_invoice_lines()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_invoice_lines_immutability
    BEFORE INSERT OR UPDATE OR DELETE ON invoice_lines
    FOR EACH ROW
    EXECUTE FUNCTION protect_issued_invoice_lines();
