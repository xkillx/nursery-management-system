-- name: InsertDeliveryEvent :exec
INSERT INTO email_delivery (
    id, provider_message_id, email_outbox_id, status, response_json
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetDeliveryByProviderMessageID :many
SELECT id, provider_message_id, email_outbox_id, status, response_json, created_at
FROM email_delivery
WHERE provider_message_id = $1
ORDER BY created_at DESC;
