package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	IsActive     bool
}

type Membership struct {
	TenantID uuid.UUID
	BranchID uuid.UUID
	Role     string
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindUserByEmail(ctx context.Context, emailNormalized string) (User, error) {
	const q = `
SELECT id, email, password_hash, is_active
FROM users
WHERE email_normalized = $1
LIMIT 1`

	var u User
	err := r.pool.QueryRow(ctx, q, emailNormalized).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}

	return u, nil
}

func (r *Repository) ListMembershipsByUserID(ctx context.Context, userID uuid.UUID) ([]Membership, error) {
	const q = `
SELECT tenant_id, branch_id, role
FROM memberships
WHERE user_id = $1
ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make([]Membership, 0)
	for rows.Next() {
		var m Membership
		if err := rows.Scan(&m.TenantID, &m.BranchID, &m.Role); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}

	return memberships, rows.Err()
}

func (r *Repository) CreateRefreshToken(ctx context.Context, token RefreshToken, userAgent, ipAddress string) error {
	const q = `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, q, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, userAgent, ipAddress)
	return err
}

func (r *Repository) FindActiveRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, User, error) {
	const q = `
SELECT rt.id, rt.user_id, rt.token_hash, rt.expires_at, rt.revoked_at,
       u.id, u.email, u.password_hash, u.is_active
FROM refresh_tokens rt
JOIN users u ON u.id = rt.user_id
WHERE rt.token_hash = $1
LIMIT 1`

	var token RefreshToken
	var user User
	err := r.pool.QueryRow(ctx, q, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefreshToken{}, User{}, ErrNotFound
	}
	if err != nil {
		return RefreshToken{}, User{}, err
	}

	if token.RevokedAt != nil || time.Now().UTC().After(token.ExpiresAt) || !user.IsActive {
		return RefreshToken{}, User{}, ErrNotFound
	}

	return token, user, nil
}

func (r *Repository) RotateRefreshToken(ctx context.Context, oldTokenID uuid.UUID, replacement RefreshToken, userAgent, ipAddress string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const revokeQ = `
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now()
WHERE id = $1 AND revoked_at IS NULL`

	if _, err := tx.Exec(ctx, revokeQ, oldTokenID); err != nil {
		return err
	}

	const insertQ = `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5, $6)`

	if _, err := tx.Exec(ctx, insertQ, replacement.ID, replacement.UserID, replacement.TokenHash, replacement.ExpiresAt, userAgent, ipAddress); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL`

	_, err := r.pool.Exec(ctx, q, tokenHash)
	return err
}
