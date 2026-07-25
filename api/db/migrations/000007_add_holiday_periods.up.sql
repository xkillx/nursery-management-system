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

-- Data migration: convert existing school-holiday closure days to holiday periods.
-- Groups consecutive dates by tenant, branch, and reason into date ranges.
-- Only converts reasons matching known school holidays: Christmas, Easter, Summer.
-- Bank holidays, inset days, and other reasons are left as individual closure days.

INSERT INTO holiday_periods (id, tenant_id, branch_id, name, start_date, end_date, type, created_at, updated_at)
SELECT
    gen_random_uuid(),
    tenant_id,
    branch_id,
    reason AS name,
    MIN(date) AS start_date,
    MAX(date) AS end_date,
    CASE
        WHEN reason ILIKE '%christmas%' THEN 'christmas'
        WHEN reason ILIKE '%easter%' THEN 'easter'
        WHEN reason ILIKE '%summer%' THEN 'summer'
        ELSE 'other'
    END AS type,
    now(),
    now()
FROM (
    SELECT
        tenant_id,
        branch_id,
        date,
        reason,
        date - INTERVAL '1 day' * ROW_NUMBER() OVER (
            PARTITION BY tenant_id, branch_id, reason
            ORDER BY date
        ) AS grp
    FROM branch_closure_days
    WHERE reason ILIKE '%christmas%'
       OR reason ILIKE '%easter%'
       OR reason ILIKE '%summer%'
) grouped
GROUP BY tenant_id, branch_id, reason, grp;

-- Delete the original school-holiday closure days that were migrated.
DELETE FROM branch_closure_days
WHERE reason ILIKE '%christmas%'
   OR reason ILIKE '%easter%'
   OR reason ILIKE '%summer%';
