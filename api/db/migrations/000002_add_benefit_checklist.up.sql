ALTER TABLE child_funding_records
  ADD COLUMN benefit_universal_credit      boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_income_support        boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_jobseekers_allowance  boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_esa_income_related    boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_child_tax_credit      boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_other_support         boolean NOT NULL DEFAULT false,
  ADD COLUMN other_benefit_name             text;
