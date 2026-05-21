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
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	Role     string
	IsActive bool
}

type RefreshToken struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	MembershipID uuid.UUID
	TokenHash    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
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
SELECT id, tenant_id, branch_id, role, is_active
FROM memberships
WHERE user_id = $1 AND is_active = true AND ended_at IS NULL
ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make([]Membership, 0)
	for rows.Next() {
		var m Membership
		if err := rows.Scan(&m.ID, &m.TenantID, &m.BranchID, &m.Role, &m.IsActive); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}

	return memberships, rows.Err()
}

func (r *Repository) CreateRefreshToken(ctx context.Context, token RefreshToken, userAgent, ipAddress string) error {
	const q = `
INSERT INTO refresh_tokens (id, user_id, membership_id, token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, q, token.ID, token.UserID, token.MembershipID, token.TokenHash, token.ExpiresAt, userAgent, ipAddress)
	return err
}

func (r *Repository) FindActiveRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, User, Membership, error) {
	const q = `
SELECT rt.id, rt.user_id, rt.membership_id, rt.token_hash, rt.expires_at, rt.revoked_at,
       u.id, u.email, u.password_hash, u.is_active,
       m.id, m.tenant_id, m.branch_id, m.role, m.is_active
FROM refresh_tokens rt
JOIN users u ON u.id = rt.user_id
JOIN memberships m ON m.id = rt.membership_id AND m.user_id = u.id AND m.is_active = true AND m.ended_at IS NULL
WHERE rt.token_hash = $1
LIMIT 1`

	var token RefreshToken
	var user User
	var membership Membership
	err := r.pool.QueryRow(ctx, q, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.MembershipID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&membership.ID,
		&membership.TenantID,
		&membership.BranchID,
		&membership.Role,
		&membership.IsActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefreshToken{}, User{}, Membership{}, ErrNotFound
	}
	if err != nil {
		return RefreshToken{}, User{}, Membership{}, err
	}

	if token.RevokedAt != nil || time.Now().UTC().After(token.ExpiresAt) || !user.IsActive {
		return RefreshToken{}, User{}, Membership{}, ErrNotFound
	}

	return token, user, membership, nil
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
INSERT INTO refresh_tokens (id, user_id, membership_id, token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if _, err := tx.Exec(ctx, insertQ, replacement.ID, replacement.UserID, replacement.MembershipID, replacement.TokenHash, replacement.ExpiresAt, userAgent, ipAddress); err != nil {
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

func (r *Repository) CreateScopeSwitchAuditLog(ctx context.Context, actorUserID uuid.UUID, fromMembership, toMembership Membership, requestID string) error {
	const q = `
INSERT INTO audit_logs (id, tenant_id, branch_id, actor_user_id, actor_membership_id, action_type, action_entity_type, action_entity_id, request_id, details)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, jsonb_build_object('from_membership_id', $10, 'to_membership_id', $11))`

	_, err := r.pool.Exec(
		ctx,
		q,
		uuid.New(),
		toMembership.TenantID,
		toMembership.BranchID,
		actorUserID,
		fromMembership.ID,
		"session_scope_switched",
		"membership",
		toMembership.ID,
		requestID,
		fromMembership.ID,
		toMembership.ID,
	)
	return err
}
