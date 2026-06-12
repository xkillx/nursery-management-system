-- Add site-level core hourly rate to branches.
-- Branches with exactly one distinct positive child rate are backfilled.
-- Branches with no positive child rate or multiple distinct positive rates remain null (owner setup exception).

ALTER TABLE branches
ADD COLUMN core_hourly_rate_minor INTEGER;

ALTER TABLE branches
ADD CONSTRAINT branches_core_hourly_rate_positive_check
CHECK (core_hourly_rate_minor IS NULL OR core_hourly_rate_minor > 0);

-- Backfill: seed site rate only when a branch has exactly one distinct positive non-null child rate.
-- NULL child rates are ignored (they represent absent data, not a zero or conflicting rate).
WITH branch_consistent_rate AS (
    SELECT
        c.branch_id,
        COUNT(DISTINCT c.core_hourly_rate_minor) FILTER (
            WHERE c.core_hourly_rate_minor IS NOT NULL AND c.core_hourly_rate_minor > 0
        ) AS distinct_positive_count,
        MIN(c.core_hourly_rate_minor) FILTER (
            WHERE c.core_hourly_rate_minor IS NOT NULL AND c.core_hourly_rate_minor > 0
        ) AS the_rate
    FROM children c
    WHERE c.core_hourly_rate_minor IS NOT NULL AND c.core_hourly_rate_minor > 0
    GROUP BY c.branch_id
    HAVING COUNT(DISTINCT c.core_hourly_rate_minor) FILTER (
        WHERE c.core_hourly_rate_minor IS NOT NULL AND c.core_hourly_rate_minor > 0
    ) = 1
)
UPDATE branches b
SET core_hourly_rate_minor = bcr.the_rate
FROM branch_consistent_rate bcr
WHERE b.id = bcr.branch_id;

CREATE INDEX idx_branches_core_hourly_rate
ON branches (tenant_id) WHERE core_hourly_rate_minor IS NOT NULL;
