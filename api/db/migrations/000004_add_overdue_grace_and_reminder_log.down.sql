DROP TABLE IF EXISTS invoice_reminder_log;

ALTER TABLE branches
    DROP CONSTRAINT IF EXISTS branches_reminder_days_before_range;

ALTER TABLE branches
    DROP COLUMN IF EXISTS reminder_days_before;

ALTER TABLE branches
    DROP CONSTRAINT IF EXISTS branches_overdue_grace_days_range;

ALTER TABLE branches
    DROP COLUMN IF EXISTS overdue_grace_days;
