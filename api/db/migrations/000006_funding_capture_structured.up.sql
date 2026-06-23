-- Add new structured funding columns to child_funding_records
ALTER TABLE child_funding_records
    ADD COLUMN funding_enabled boolean NOT NULL DEFAULT false,
    ADD COLUMN funding_type text NOT NULL DEFAULT 'unknown',
    ADD COLUMN funding_model text NOT NULL DEFAULT 'unknown',
    ADD COLUMN funded_hours_per_week numeric(5,2),
    ADD COLUMN funding_start_date date,
    ADD COLUMN funding_end_date date,
    ADD COLUMN eligibility_code text,
    ADD COLUMN eligibility_code_validated boolean NOT NULL DEFAULT false,
    ADD COLUMN evidence_received boolean NOT NULL DEFAULT false,
    ADD COLUMN benefits_status text NOT NULL DEFAULT 'unknown',
    ADD COLUMN benefit_notes text,
    ADD COLUMN manager_notes text;

-- Add CHECK constraints for enums
ALTER TABLE child_funding_records
    ADD CONSTRAINT child_funding_records_funding_type_check
        CHECK (funding_type IN ('none','fifteen_hours','thirty_hours','two_year_old','custom','unknown')),
    ADD CONSTRAINT child_funding_records_funding_model_check
        CHECK (funding_model IN ('term_time_only','stretched','unknown')),
    ADD CONSTRAINT child_funding_records_benefits_status_check
        CHECK (benefits_status IN ('no','yes','unknown'));

-- Add CHECK constraint: end_date must be after start_date when both provided
ALTER TABLE child_funding_records
    ADD CONSTRAINT child_funding_records_end_after_start
        CHECK (funding_end_date IS NULL OR funding_start_date IS NULL OR funding_end_date > funding_start_date);

-- Migrate existing rows: carry funding_support_notes → manager_notes
UPDATE child_funding_records
SET manager_notes = funding_support_notes
WHERE funding_support_notes IS NOT NULL AND manager_notes IS NULL;

-- Drop old columns
ALTER TABLE child_funding_records
    DROP COLUMN benefits_contribute_to_fees,
    DROP COLUMN working_tax_credit,
    DROP COLUMN college_uni_paid_to_parent,
    DROP COLUMN college_uni_paid_to_nursery,
    DROP COLUMN funding_3yo_term_time,
    DROP COLUMN funding_2yo_term_time,
    DROP COLUMN funding_support_notes,
    DROP COLUMN funding_support_reviewed;
