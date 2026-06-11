-- name: RegistrationOfficeChecklistChildGet :one
SELECT
    c.id,
    c.full_name,
    c.date_of_birth,
    c.start_date,
    c.end_date
FROM children c
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.id = $3;

-- name: RegistrationOfficeChecklistGetByChild :one
SELECT
    croc.id,
    croc.tenant_id,
    croc.branch_id,
    croc.child_id,
    croc.deposit_status::text,
    croc.deposit_paid_date,
    croc.application_date_status::text,
    croc.application_date,
    croc.start_date_status::text,
    croc.date_left,
    croc.sessions_days_requested_status::text,
    croc.sessions_days_requested,
    croc.term_time_only_space_status::text,
    croc.contract_status::text,
    croc.contract_date,
    croc.handbook_status::text,
    croc.handbook_date,
    croc.red_book_status::text,
    croc.red_book_checked_date,
    croc.birth_certificate_passport_status::text,
    croc.birth_certificate_passport_checked_date,
    croc.proof_of_address_status::text,
    croc.proof_of_address_checked_date,
    croc.notes,
    croc.created_at,
    croc.updated_at
FROM child_registration_office_checklists croc
WHERE croc.tenant_id = $1
  AND croc.branch_id = $2
  AND croc.child_id = $3;

-- name: RegistrationOfficeChecklistGetForUpdateByChild :one
SELECT
    croc.id,
    croc.tenant_id,
    croc.branch_id,
    croc.child_id,
    croc.deposit_status::text,
    croc.deposit_paid_date,
    croc.application_date_status::text,
    croc.application_date,
    croc.start_date_status::text,
    croc.date_left,
    croc.sessions_days_requested_status::text,
    croc.sessions_days_requested,
    croc.term_time_only_space_status::text,
    croc.contract_status::text,
    croc.contract_date,
    croc.handbook_status::text,
    croc.handbook_date,
    croc.red_book_status::text,
    croc.red_book_checked_date,
    croc.birth_certificate_passport_status::text,
    croc.birth_certificate_passport_checked_date,
    croc.proof_of_address_status::text,
    croc.proof_of_address_checked_date,
    croc.notes,
    croc.created_at,
    croc.updated_at
FROM child_registration_office_checklists croc
WHERE croc.tenant_id = $1
  AND croc.branch_id = $2
  AND croc.child_id = $3
FOR UPDATE OF croc;

-- name: RegistrationOfficeChecklistCreate :one
INSERT INTO child_registration_office_checklists (
    id, tenant_id, branch_id, child_id,
    deposit_status, deposit_paid_date,
    application_date_status, application_date,
    start_date_status, date_left,
    sessions_days_requested_status, sessions_days_requested,
    term_time_only_space_status,
    contract_status, contract_date,
    handbook_status, handbook_date,
    red_book_status, red_book_checked_date,
    birth_certificate_passport_status, birth_certificate_passport_checked_date,
    proof_of_address_status, proof_of_address_checked_date,
    notes
) VALUES (
    $1, $2, $3, $4,
    $5, $6,
    $7, $8,
    $9, $10,
    $11, $12,
    $13,
    $14, $15,
    $16, $17,
    $18, $19,
    $20, $21,
    $22, $23,
    $24
)
RETURNING
    id, tenant_id, branch_id, child_id,
    deposit_status::text, deposit_paid_date,
    application_date_status::text, application_date,
    start_date_status::text, date_left,
    sessions_days_requested_status::text, sessions_days_requested,
    term_time_only_space_status::text,
    contract_status::text, contract_date,
    handbook_status::text, handbook_date,
    red_book_status::text, red_book_checked_date,
    birth_certificate_passport_status::text, birth_certificate_passport_checked_date,
    proof_of_address_status::text, proof_of_address_checked_date,
    notes,
    created_at, updated_at;

-- name: RegistrationOfficeChecklistUpdate :one
UPDATE child_registration_office_checklists croc
SET
    deposit_status = $4,
    deposit_paid_date = $5,
    application_date_status = $6,
    application_date = $7,
    start_date_status = $8,
    date_left = $9,
    sessions_days_requested_status = $10,
    sessions_days_requested = $11,
    term_time_only_space_status = $12,
    contract_status = $13,
    contract_date = $14,
    handbook_status = $15,
    handbook_date = $16,
    red_book_status = $17,
    red_book_checked_date = $18,
    birth_certificate_passport_status = $19,
    birth_certificate_passport_checked_date = $20,
    proof_of_address_status = $21,
    proof_of_address_checked_date = $22,
    notes = $23,
    updated_at = now()
WHERE croc.tenant_id = $1
  AND croc.branch_id = $2
  AND croc.child_id = $3
RETURNING
    id, tenant_id, branch_id, child_id,
    deposit_status::text, deposit_paid_date,
    application_date_status::text, application_date,
    start_date_status::text, date_left,
    sessions_days_requested_status::text, sessions_days_requested,
    term_time_only_space_status::text,
    contract_status::text, contract_date,
    handbook_status::text, handbook_date,
    red_book_status::text, red_book_checked_date,
    birth_certificate_passport_status::text, birth_certificate_passport_checked_date,
    proof_of_address_status::text, proof_of_address_checked_date,
    notes,
    created_at, updated_at;
