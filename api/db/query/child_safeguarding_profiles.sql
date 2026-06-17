-- name: ChildSafeguardingProfileGetByChild :one
SELECT id, tenant_id, branch_id, child_id,
       social_services_status, social_services_notes,
       social_worker_name, social_worker_phone, social_worker_email,
       concern_walking, concern_speech_language, concern_hearing,
       concern_sight, concern_emotional_wellbeing, concern_behaviour,
       professional_referrals, restricted_notes,
       created_at, updated_at
FROM child_safeguarding_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;

-- name: ChildSafeguardingProfileUpsert :one
INSERT INTO child_safeguarding_profiles (
    id, tenant_id, branch_id, child_id,
    social_services_status, social_services_notes,
    social_worker_name, social_worker_phone, social_worker_email,
    concern_walking, concern_speech_language, concern_hearing,
    concern_sight, concern_emotional_wellbeing, concern_behaviour,
    professional_referrals, restricted_notes
)
VALUES (
    $1, $2, $3, $4,
    $5, NULLIF($6, ''),
    NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''),
    $10, $11, $12, $13, $14, $15,
    $16, NULLIF($17, '')
)
ON CONFLICT (child_id) DO UPDATE SET
    social_services_status = EXCLUDED.social_services_status,
    social_services_notes = EXCLUDED.social_services_notes,
    social_worker_name = EXCLUDED.social_worker_name,
    social_worker_phone = EXCLUDED.social_worker_phone,
    social_worker_email = EXCLUDED.social_worker_email,
    concern_walking = EXCLUDED.concern_walking,
    concern_speech_language = EXCLUDED.concern_speech_language,
    concern_hearing = EXCLUDED.concern_hearing,
    concern_sight = EXCLUDED.concern_sight,
    concern_emotional_wellbeing = EXCLUDED.concern_emotional_wellbeing,
    concern_behaviour = EXCLUDED.concern_behaviour,
    professional_referrals = EXCLUDED.professional_referrals,
    restricted_notes = EXCLUDED.restricted_notes,
    updated_at = now()
RETURNING id, tenant_id, branch_id, child_id,
          social_services_status, social_services_notes,
          social_worker_name, social_worker_phone, social_worker_email,
          concern_walking, concern_speech_language, concern_hearing,
          concern_sight, concern_emotional_wellbeing, concern_behaviour,
          professional_referrals, restricted_notes,
          created_at, updated_at;
