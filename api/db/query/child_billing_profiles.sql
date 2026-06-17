-- name: ChildBillingProfileGetByChild :one
SELECT id, tenant_id, branch_id, child_id,
       billing_basis, custom_rate_minor, effective_from,
       created_at, updated_at
FROM child_billing_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildBillingProfileUpsert :one
INSERT INTO child_billing_profiles (
    id, tenant_id, branch_id, child_id,
    billing_basis, custom_rate_minor, effective_from
)
VALUES ($1, $2, $3, $4, $5, $6, COALESCE(NULLIF($7, '')::date, CURRENT_DATE))
ON CONFLICT (child_id) DO UPDATE SET
    billing_basis = EXCLUDED.billing_basis,
    custom_rate_minor = EXCLUDED.custom_rate_minor,
    effective_from = EXCLUDED.effective_from,
    updated_at = now()
RETURNING id, tenant_id, branch_id, child_id,
          billing_basis, custom_rate_minor, effective_from,
          created_at, updated_at;
