ALTER TABLE invoice_lines DROP CONSTRAINT IF EXISTS invoice_lines_line_kind_valid;
ALTER TABLE invoice_lines ADD CONSTRAINT invoice_lines_line_kind_valid
    CHECK (line_kind IN ('core_childcare', 'funded_deduction', 'extra', 'adjustment'));
