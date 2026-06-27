-- name: ChildFundingRecordGetByChild :one
SELECT id, tenant_id, branch_id, child_id,
       funding_enabled, funding_type, funding_model,
       funded_hours_per_week, funding_start_date, funding_end_date,
       eligibility_code, eligibility_code_validated,
       evidence_received, benefits_status,
       benefit_universal_credit, benefit_income_support,
       benefit_jobseekers_allowance, benefit_esa_income_related,
       benefit_child_tax_credit, benefit_other_support,
       other_benefit_name,
       benefit_notes, manager_notes,
       created_at, updated_at
FROM child_funding_records
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildFundingRecordUpsert :one
INSERT INTO child_funding_records (
    id, tenant_id, branch_id, child_id,
    funding_enabled, funding_type, funding_model,
    funded_hours_per_week, funding_start_date, funding_end_date,
    eligibility_code, eligibility_code_validated,
    evidence_received, benefits_status,
    benefit_universal_credit, benefit_income_support,
    benefit_jobseekers_allowance, benefit_esa_income_related,
    benefit_child_tax_credit, benefit_other_support,
    other_benefit_name,
    benefit_notes, manager_notes
)
VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8::numeric, $9::date, $10::date,
    NULLIF($11, ''), $12,
    $13, $14,
    $15, $16,
    $17, $18,
    $19, $20,
    NULLIF($21, ''),
    NULLIF($22, ''), NULLIF($23, '')
)
ON CONFLICT (child_id) DO UPDATE SET
    funding_enabled = EXCLUDED.funding_enabled,
    funding_type = EXCLUDED.funding_type,
    funding_model = EXCLUDED.funding_model,
    funded_hours_per_week = EXCLUDED.funded_hours_per_week,
    funding_start_date = EXCLUDED.funding_start_date,
    funding_end_date = EXCLUDED.funding_end_date,
    eligibility_code = EXCLUDED.eligibility_code,
    eligibility_code_validated = EXCLUDED.eligibility_code_validated,
    evidence_received = EXCLUDED.evidence_received,
    benefits_status = EXCLUDED.benefits_status,
    benefit_universal_credit = EXCLUDED.benefit_universal_credit,
    benefit_income_support = EXCLUDED.benefit_income_support,
    benefit_jobseekers_allowance = EXCLUDED.benefit_jobseekers_allowance,
    benefit_esa_income_related = EXCLUDED.benefit_esa_income_related,
    benefit_child_tax_credit = EXCLUDED.benefit_child_tax_credit,
    benefit_other_support = EXCLUDED.benefit_other_support,
    other_benefit_name = EXCLUDED.other_benefit_name,
    benefit_notes = EXCLUDED.benefit_notes,
    manager_notes = EXCLUDED.manager_notes,
    updated_at = now()
RETURNING id, tenant_id, branch_id, child_id,
          funding_enabled, funding_type, funding_model,
          funded_hours_per_week, funding_start_date, funding_end_date,
          eligibility_code, eligibility_code_validated,
          evidence_received, benefits_status,
          benefit_universal_credit, benefit_income_support,
          benefit_jobseekers_allowance, benefit_esa_income_related,
          benefit_child_tax_credit, benefit_other_support,
          other_benefit_name,
          benefit_notes, manager_notes,
          created_at, updated_at;
