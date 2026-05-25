CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    superseded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT password_reset_tokens_consumed_shape_check
        CHECK (used_at IS NULL OR superseded_at IS NULL)
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens (expires_at);
CREATE INDEX idx_password_reset_tokens_active_user ON password_reset_tokens (user_id, created_at DESC)
WHERE used_at IS NULL AND superseded_at IS NULL;
