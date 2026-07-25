-- Holiday periods: structured date ranges for school holidays within terms.
-- Used by billing to exclude half-term, Christmas, Easter, and summer breaks
-- from term-time minute calculations.

CREATE TABLE holiday_periods (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name varchar(200) NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    type varchar(20) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT holiday_periods_start_before_end CHECK (start_date < end_date)
);

CREATE INDEX idx_holiday_periods_tenant_branch
    ON holiday_periods (tenant_id, branch_id);

CREATE INDEX idx_holiday_periods_date_range
    ON holiday_periods (tenant_id, branch_id, start_date, end_date);
