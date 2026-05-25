-- name: FindUserForPasswordResetByEmail :one
SELECT id, email, password_hash, is_active
FROM users
WHERE email_normalized = $1
LIMIT 1;

-- name: SupersedeUnusedPasswordResetTokensForUser :exec
UPDATE password_reset_tokens
SET superseded_at = now(), updated_at = now()
WHERE user_id = $1
  AND used_at IS NULL
  AND superseded_at IS NULL;

-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetPasswordResetTokenForUpdate :one
SELECT id, user_id, token_hash, expires_at, used_at, superseded_at
FROM password_reset_tokens
WHERE token_hash = $1
FOR UPDATE;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = now(), updated_at = now()
WHERE id = $1
  AND used_at IS NULL
  AND superseded_at IS NULL;

-- name: SupersedeOtherUnusedPasswordResetTokensForUser :exec
UPDATE password_reset_tokens
SET superseded_at = now(), updated_at = now()
WHERE user_id = $1
  AND id <> $2
  AND used_at IS NULL
  AND superseded_at IS NULL;

-- name: UpdateUserPasswordHash :execrows
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1 AND is_active = true;

-- name: RevokeRefreshTokensForUser :exec
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;
