-- Re-add old columns with defaults
ALTER TABLE child_funding_records
    ADD COLUMN benefits_contribute_to_fees text NOT NULL DEFAULT 'unknown',
    ADD COLUMN working_tax_credit text NOT NULL DEFAULT 'unknown',
    ADD COLUMN college_uni_paid_to_parent text NOT NULL DEFAULT 'unknown',
    ADD COLUMN college_uni_paid_to_nursery text NOT NULL DEFAULT 'unknown',
    ADD COLUMN funding_3yo_term_time text NOT NULL DEFAULT 'unknown',
    ADD COLUMN funding_2yo_term_time text NOT NULL DEFAULT 'unknown',
    ADD COLUMN funding_support_notes text,
    ADD COLUMN funding_support_reviewed boolean NOT NULL DEFAULT false;

-- Copy manager_notes → funding_support_notes for existing rows
UPDATE child_funding_records
SET funding_support_notes = manager_notes
WHERE manager_notes IS NOT NULL AND funding_support_notes IS NULL;

-- Drop new columns and constraints
ALTER TABLE child_funding_records
    DROP CONSTRAINT child_funding_records_funding_type_check,
    DROP CONSTRAINT child_funding_records_funding_model_check,
    DROP CONSTRAINT child_funding_records_benefits_status_check,
    DROP CONSTRAINT child_funding_records_end_after_start,
    DROP COLUMN funding_enabled,
    DROP COLUMN funding_type,
    DROP COLUMN funding_model,
    DROP COLUMN funded_hours_per_week,
    DROP COLUMN funding_start_date,
    DROP COLUMN funding_end_date,
    DROP COLUMN eligibility_code,
    DROP COLUMN eligibility_code_validated,
    DROP COLUMN evidence_received,
    DROP COLUMN benefits_status,
    DROP COLUMN benefit_notes,
    DROP COLUMN manager_notes;
