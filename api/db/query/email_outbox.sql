-- name: InsertEmailOutbox :one
INSERT INTO email_outbox (
    id, tenant_id, branch_id, idempotency_key, event_type,
    recipient, recipient_name, subject, template_name, template_version,
    payload_json, attachments, entity_id, status, max_attempts
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, 'pending', $14
)
RETURNING id, tenant_id, branch_id, idempotency_key, event_type,
    recipient, recipient_name, subject, template_name, template_version,
    payload_json, attachments, entity_id, status, attempts, max_attempts, next_retry_at,
    last_error, provider_message_id, created_at, sent_at, updated_at;

-- name: GetPendingEmails :many
SELECT id, tenant_id, branch_id, idempotency_key, event_type,
    recipient, recipient_name, subject, template_name, template_version,
    payload_json, attachments, entity_id, status, attempts, max_attempts, next_retry_at,
    last_error, provider_message_id, created_at, sent_at, updated_at
FROM email_outbox
WHERE status = 'pending'
  AND next_retry_at <= now()
ORDER BY next_retry_at ASC
FOR UPDATE SKIP LOCKED
LIMIT $1;

-- name: UpdateEmailStatus :exec
UPDATE email_outbox
SET status = $2,
    attempts = $3,
    next_retry_at = $4,
    last_error = $5,
    provider_message_id = $6,
    sent_at = CASE WHEN $2 = 'sent' THEN now() ELSE sent_at END,
    updated_at = now()
WHERE id = $1;

-- name: GetEmailByID :one
SELECT id, tenant_id, branch_id, idempotency_key, event_type,
    recipient, recipient_name, subject, template_name, template_version,
    payload_json, attachments, entity_id, status, attempts, max_attempts, next_retry_at,
    last_error, provider_message_id, created_at, sent_at, updated_at
FROM email_outbox
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3;

-- name: ListEmails :many
SELECT id, tenant_id, branch_id, idempotency_key, event_type,
    recipient, recipient_name, subject, template_name, template_version,
    payload_json, attachments, entity_id, status, attempts, max_attempts, next_retry_at,
    last_error, provider_message_id, created_at, sent_at, updated_at
FROM email_outbox
WHERE tenant_id = $1 AND branch_id = $2
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
ORDER BY created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountEmails :one
SELECT COUNT(*)
FROM email_outbox
WHERE tenant_id = $1 AND branch_id = $2
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text);

-- name: GetEmailStats :many
SELECT status, COUNT(*)::integer AS count
FROM email_outbox
WHERE tenant_id = $1 AND branch_id = $2
GROUP BY status;

-- name: ResetEmailToPending :exec
UPDATE email_outbox
SET status = 'pending',
    attempts = 0,
    next_retry_at = now(),
    last_error = NULL
WHERE id = $1 AND status = 'dead_letter';
