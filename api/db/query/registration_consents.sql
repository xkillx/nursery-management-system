-- name: ConsentGetLatestByChild :one
SELECT
    id, tenant_id, branch_id, child_id, version, source,
    signer_name, signed_date, paper_form_on_file,
    urgent_medical_treatment, urgent_medical_treatment_exceptions,
    plasters, safeguarding_reporting_acknowledgement,
    information_sharing_consent,
    area_senco_liaison, health_visitor_liaison,
    transition_documents, local_outings, face_painting,
    parent_supplied_sun_cream, parent_supplied_nappy_cream,
    development_profile_photos, nursery_display_boards,
    promotional_literature, nursery_website,
    staff_student_coursework, social_media, social_media_channel_notes,
    notes_exceptions,
    entered_by_user_id, entered_by_membership_id,
    created_at
FROM child_registration_consent_records
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY version DESC
LIMIT 1;

-- name: ConsentListByChild :many
SELECT
    id, tenant_id, branch_id, child_id, version, source,
    signer_name, signed_date, paper_form_on_file,
    urgent_medical_treatment, urgent_medical_treatment_exceptions,
    plasters, safeguarding_reporting_acknowledgement,
    information_sharing_consent,
    area_senco_liaison, health_visitor_liaison,
    transition_documents, local_outings, face_painting,
    parent_supplied_sun_cream, parent_supplied_nappy_cream,
    development_profile_photos, nursery_display_boards,
    promotional_literature, nursery_website,
    staff_student_coursework, social_media, social_media_channel_notes,
    notes_exceptions,
    entered_by_user_id, entered_by_membership_id,
    created_at
FROM child_registration_consent_records
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY version DESC;

-- name: ConsentGetCurrentVersion :one
SELECT COALESCE(MAX(version), 0) AS current_version
FROM child_registration_consent_records
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ConsentCreate :exec
INSERT INTO child_registration_consent_records (
    id, tenant_id, branch_id, child_id, version, source,
    signer_name, signed_date, paper_form_on_file,
    urgent_medical_treatment, urgent_medical_treatment_exceptions,
    plasters, safeguarding_reporting_acknowledgement,
    information_sharing_consent,
    area_senco_liaison, health_visitor_liaison,
    transition_documents, local_outings, face_painting,
    parent_supplied_sun_cream, parent_supplied_nappy_cream,
    development_profile_photos, nursery_display_boards,
    promotional_literature, nursery_website,
    staff_student_coursework, social_media, social_media_channel_notes,
    notes_exceptions,
    entered_by_user_id, entered_by_membership_id
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11,
    $12, $13,
    $14,
    $15, $16,
    $17, $18, $19,
    $20, $21,
    $22, $23,
    $24, $25,
    $26, $27, $28,
    $29,
    $30, $31
);
