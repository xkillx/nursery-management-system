-- name: ChildProfileGetByChild :one
SELECT id, tenant_id, branch_id, child_id,
       sex, religion, ethnic_origin, first_language, other_languages,
       address_line1, address_line2, address_city, address_postcode, home_telephone,
       disability_status, disability_notes, access_requirements,
       routine_care_notes,
       gdpr_declared_by_name, gdpr_declared_at, gdpr_declaration_date,
       registration_date,
       demographics_home_reviewed, medical_dietary_reviewed,
       health_contacts_reviewed, social_development_reviewed,
       parent_responsibility_reviewed, emergency_collection_reviewed,
       routine_care_reviewed,
       created_at, updated_at
FROM child_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildProfileGetForUpdate :one
SELECT id, tenant_id, branch_id, child_id,
       sex, religion, ethnic_origin, first_language, other_languages,
       address_line1, address_line2, address_city, address_postcode, home_telephone,
       disability_status, disability_notes, access_requirements,
       routine_care_notes,
       gdpr_declared_by_name, gdpr_declared_at, gdpr_declaration_date,
       registration_date,
       demographics_home_reviewed, medical_dietary_reviewed,
       health_contacts_reviewed, social_development_reviewed,
       parent_responsibility_reviewed, emergency_collection_reviewed,
       routine_care_reviewed,
       created_at, updated_at
FROM child_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
FOR UPDATE;

-- name: ChildProfileInsert :one
INSERT INTO child_profiles (
    id, tenant_id, branch_id, child_id,
    sex, religion, ethnic_origin, first_language, other_languages,
    address_line1, address_line2, address_city, address_postcode, home_telephone,
    disability_status, disability_notes, access_requirements,
    routine_care_notes,
    gdpr_declared_by_name, gdpr_declared_at, gdpr_declaration_date,
    registration_date,
    demographics_home_reviewed, medical_dietary_reviewed,
    health_contacts_reviewed, social_development_reviewed,
    parent_responsibility_reviewed, emergency_collection_reviewed,
    routine_care_reviewed
)
VALUES (
    $1, $2, $3, $4,
    NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''),
    NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, ''), NULLIF($14, ''),
    $15, NULLIF($16, ''), NULLIF($17, ''),
    NULLIF($18, ''),
    NULLIF($19, ''), $20, $21,
    $22,
    $23, $24, $25, $26, $27, $28, $29
)
RETURNING id, tenant_id, branch_id, child_id,
          sex, religion, ethnic_origin, first_language, other_languages,
          address_line1, address_line2, address_city, address_postcode, home_telephone,
          disability_status, disability_notes, access_requirements,
          routine_care_notes,
          gdpr_declared_by_name, gdpr_declared_at, gdpr_declaration_date,
          registration_date,
          demographics_home_reviewed, medical_dietary_reviewed,
          health_contacts_reviewed, social_development_reviewed,
          parent_responsibility_reviewed, emergency_collection_reviewed,
          routine_care_reviewed,
          created_at, updated_at;

-- name: ChildProfileUpdate :one
UPDATE child_profiles SET
    sex = NULLIF($5, ''),
    religion = NULLIF($6, ''),
    ethnic_origin = NULLIF($7, ''),
    first_language = NULLIF($8, ''),
    other_languages = NULLIF($9, ''),
    address_line1 = NULLIF($10, ''),
    address_line2 = NULLIF($11, ''),
    address_city = NULLIF($12, ''),
    address_postcode = NULLIF($13, ''),
    home_telephone = NULLIF($14, ''),
    disability_status = $15,
    disability_notes = NULLIF($16, ''),
    access_requirements = NULLIF($17, ''),
    routine_care_notes = NULLIF($18, ''),
    gdpr_declared_by_name = NULLIF($19, ''),
    gdpr_declared_at = $20,
    gdpr_declaration_date = $21,
    registration_date = $22,
    demographics_home_reviewed = $23,
    medical_dietary_reviewed = $24,
    health_contacts_reviewed = $25,
    social_development_reviewed = $26,
    parent_responsibility_reviewed = $27,
    emergency_collection_reviewed = $28,
    routine_care_reviewed = $29,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND id = $4
RETURNING id, tenant_id, branch_id, child_id,
          sex, religion, ethnic_origin, first_language, other_languages,
          address_line1, address_line2, address_city, address_postcode, home_telephone,
          disability_status, disability_notes, access_requirements,
          routine_care_notes,
          gdpr_declared_by_name, gdpr_declared_at, gdpr_declaration_date,
          registration_date,
          demographics_home_reviewed, medical_dietary_reviewed,
          health_contacts_reviewed, social_development_reviewed,
          parent_responsibility_reviewed, emergency_collection_reviewed,
          routine_care_reviewed,
          created_at, updated_at;
