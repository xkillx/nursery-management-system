ALTER TABLE refresh_tokens
ADD COLUMN membership_id UUID;

UPDATE refresh_tokens rt
SET membership_id = selected.id
FROM (
    SELECT DISTINCT ON (user_id) user_id, id
    FROM memberships
    ORDER BY user_id, created_at ASC
) AS selected
WHERE selected.user_id = rt.user_id
  AND rt.membership_id IS NULL;

DELETE FROM refresh_tokens
WHERE membership_id IS NULL;

ALTER TABLE refresh_tokens
ALTER COLUMN membership_id SET NOT NULL;

ALTER TABLE refresh_tokens
ADD CONSTRAINT refresh_tokens_membership_id_fkey
FOREIGN KEY (membership_id) REFERENCES memberships(id);

CREATE INDEX idx_refresh_tokens_membership_id ON refresh_tokens (membership_id);
