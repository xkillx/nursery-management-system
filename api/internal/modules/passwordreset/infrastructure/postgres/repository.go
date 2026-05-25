package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/passwordreset/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	"nursery-management-system/api/internal/platform/uid"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindUserByEmail(ctx context.Context, emailNormalized string) (domain.User, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.FindUserForPasswordResetByEmail(ctx, emailNormalized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	return domain.User{
		ID:           row.ID.Bytes,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
	}, true, nil
}

func (r *Repository) IssueResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time, sendEmail func() error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)

	if err := q.SupersedeUnusedPasswordResetTokensForUser(ctx, pgUUID(userID)); err != nil {
		return err
	}

	tokenID := uid.NewUUID()
	var pgTokenID pgtype.UUID
	copy(pgTokenID.Bytes[:], tokenID[:])

	var pgExpiresAt pgtype.Timestamptz
	pgExpiresAt.Time = expiresAt
	pgExpiresAt.Valid = true

	if err := q.CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
		ID:        pgTokenID,
		UserID:    pgUUID(userID),
		TokenHash: tokenHash,
		ExpiresAt: pgExpiresAt,
	}); err != nil {
		return err
	}

	if err := sendEmail(); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) ResetPassword(ctx context.Context, tokenHash string, newPasswordHash string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)

	row, err := q.GetPasswordResetTokenForUpdate(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTokenInvalid
		}
		return err
	}

	if row.UsedAt.Valid {
		return domain.ErrTokenUsed
	}
	if row.SupersededAt.Valid {
		return domain.ErrTokenUsed
	}
	if row.ExpiresAt.Time.Before(time.Now().UTC()) {
		return domain.ErrTokenExpired
	}

	rowsAffected, err := q.UpdateUserPasswordHash(ctx, sqlc.UpdateUserPasswordHashParams{
		ID:           row.UserID,
		PasswordHash: newPasswordHash,
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrTokenInvalid
	}

	if err := q.MarkPasswordResetTokenUsed(ctx, row.ID); err != nil {
		return err
	}

	if err := q.SupersedeOtherUnusedPasswordResetTokensForUser(ctx, sqlc.SupersedeOtherUnusedPasswordResetTokensForUserParams{
		UserID: row.UserID,
		ID:     row.ID,
	}); err != nil {
		return err
	}

	if err := q.RevokeRefreshTokensForUser(ctx, row.UserID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	var pg pgtype.UUID
	copy(pg.Bytes[:], id[:])
	pg.Valid = true
	return pg
}
