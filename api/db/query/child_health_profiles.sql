-- name: ChildHealthProfileGetByChild :one
SELECT id, tenant_id, branch_id, child_id,
       medical_conditions_status, medical_conditions_notes,
       prescribed_medication_status, medication_notes,
       immunisation_status, immunisation_country, illness_diagnosis_history,
       dietary_requirements_status, dietary_requirements_notes, dietary_side_effects,
       doctor_name, doctor_address, doctor_phone,
       health_visitor_name, health_visitor_address, health_visitor_phone,
       created_at, updated_at
FROM child_health_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildHealthProfileUpsert :one
INSERT INTO child_health_profiles (
    id, tenant_id, branch_id, child_id,
    medical_conditions_status, medical_conditions_notes,
    prescribed_medication_status, medication_notes,
    immunisation_status, immunisation_country, illness_diagnosis_history,
    dietary_requirements_status, dietary_requirements_notes, dietary_side_effects,
    doctor_name, doctor_address, doctor_phone,
    health_visitor_name, health_visitor_address, health_visitor_phone
)
VALUES (
    $1, $2, $3, $4,
    $5, NULLIF($6, ''),
    $7, NULLIF($8, ''),
    $9, NULLIF($10, ''), NULLIF($11, ''),
    $12, NULLIF($13, ''), NULLIF($14, ''),
    NULLIF($15, ''), NULLIF($16, ''), NULLIF($17, ''),
    NULLIF($18, ''), NULLIF($19, ''), NULLIF($20, '')
)
ON CONFLICT (child_id) DO UPDATE SET
    medical_conditions_status = EXCLUDED.medical_conditions_status,
    medical_conditions_notes = EXCLUDED.medical_conditions_notes,
    prescribed_medication_status = EXCLUDED.prescribed_medication_status,
    medication_notes = EXCLUDED.medication_notes,
    immunisation_status = EXCLUDED.immunisation_status,
    immunisation_country = EXCLUDED.immunisation_country,
    illness_diagnosis_history = EXCLUDED.illness_diagnosis_history,
    dietary_requirements_status = EXCLUDED.dietary_requirements_status,
    dietary_requirements_notes = EXCLUDED.dietary_requirements_notes,
    dietary_side_effects = EXCLUDED.dietary_side_effects,
    doctor_name = EXCLUDED.doctor_name,
    doctor_address = EXCLUDED.doctor_address,
    doctor_phone = EXCLUDED.doctor_phone,
    health_visitor_name = EXCLUDED.health_visitor_name,
    health_visitor_address = EXCLUDED.health_visitor_address,
    health_visitor_phone = EXCLUDED.health_visitor_phone,
    updated_at = now()
RETURNING id, tenant_id, branch_id, child_id,
          medical_conditions_status, medical_conditions_notes,
          prescribed_medication_status, medication_notes,
          immunisation_status, immunisation_country, illness_diagnosis_history,
          dietary_requirements_status, dietary_requirements_notes, dietary_side_effects,
          doctor_name, doctor_address, doctor_phone,
          health_visitor_name, health_visitor_address, health_visitor_phone,
          created_at, updated_at;
