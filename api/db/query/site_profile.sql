-- name: SiteProfileGetByBranch :one
SELECT id, tenant_id, branch_id, nursery_name, description, phone, email, website, address_street, address_city, address_postcode, created_at, updated_at
FROM site_profiles
WHERE tenant_id = $1
  AND branch_id = $2;

-- name: SiteProfileGetByBranchForUpdate :one
SELECT id, tenant_id, branch_id, nursery_name, description, phone, email, website, address_street, address_city, address_postcode, created_at, updated_at
FROM site_profiles
WHERE tenant_id = $1
  AND branch_id = $2
FOR UPDATE;

-- name: SiteProfileUpsert :exec
INSERT INTO site_profiles (id, tenant_id, branch_id, nursery_name, description, phone, email, website, address_street, address_city, address_postcode, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), now())
ON CONFLICT (branch_id) DO UPDATE
SET nursery_name = EXCLUDED.nursery_name,
    description = EXCLUDED.description,
    phone = EXCLUDED.phone,
    email = EXCLUDED.email,
    website = EXCLUDED.website,
    address_street = EXCLUDED.address_street,
    address_city = EXCLUDED.address_city,
    address_postcode = EXCLUDED.address_postcode,
    updated_at = now();
