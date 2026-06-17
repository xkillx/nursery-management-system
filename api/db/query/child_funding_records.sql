-- name: ChildFundingRecordGetByChild :one
SELECT id, tenant_id, branch_id, child_id,
       benefits_contribute_to_fees, working_tax_credit,
       college_uni_paid_to_parent, college_uni_paid_to_nursery,
       funding_3yo_term_time, funding_2yo_term_time,
       funding_support_notes, funding_support_reviewed,
       created_at, updated_at
FROM child_funding_records
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildFundingRecordUpsert :one
INSERT INTO child_funding_records (
    id, tenant_id, branch_id, child_id,
    benefits_contribute_to_fees, working_tax_credit,
    college_uni_paid_to_parent, college_uni_paid_to_nursery,
    funding_3yo_term_time, funding_2yo_term_time,
    funding_support_notes, funding_support_reviewed
)
VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9, $10,
    NULLIF($11, ''), $12
)
ON CONFLICT (child_id) DO UPDATE SET
    benefits_contribute_to_fees = EXCLUDED.benefits_contribute_to_fees,
    working_tax_credit = EXCLUDED.working_tax_credit,
    college_uni_paid_to_parent = EXCLUDED.college_uni_paid_to_parent,
    college_uni_paid_to_nursery = EXCLUDED.college_uni_paid_to_nursery,
    funding_3yo_term_time = EXCLUDED.funding_3yo_term_time,
    funding_2yo_term_time = EXCLUDED.funding_2yo_term_time,
    funding_support_notes = EXCLUDED.funding_support_notes,
    funding_support_reviewed = EXCLUDED.funding_support_reviewed,
    updated_at = now()
RETURNING id, tenant_id, branch_id, child_id,
          benefits_contribute_to_fees, working_tax_credit,
          college_uni_paid_to_parent, college_uni_paid_to_nursery,
          funding_3yo_term_time, funding_2yo_term_time,
          funding_support_notes, funding_support_reviewed,
          created_at, updated_at;
