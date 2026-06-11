package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindUserByEmail(ctx context.Context, emailNormalized string) (domain.User, error) {
	q := sqlc.New(r.pool)
	row, err := q.AuthFindUserByEmail(ctx, emailNormalized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return domain.User{
		ID:           pgtypeUUIDToUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
	}, nil
}

func (r *Repository) ListMembershipsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error) {
	q := sqlc.New(r.pool)
	rows, err := q.AuthListMembershipsByUserID(ctx, uuidToPgtype(userID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.Membership, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Membership{
			ID:         pgtypeUUIDToUUID(row.ID),
			TenantID:   pgtypeUUIDToUUID(row.TenantID),
			TenantName: row.TenantName,
			BranchID:   pgtypeUUIDToUUID(row.BranchID),
			BranchName: pgtypeTextToString(row.BranchName),
			Role:       row.Role,
			IsActive:   row.IsActive,
		})
	}
	return out, nil
}

func (r *Repository) CreateRefreshToken(ctx context.Context, token domain.RefreshToken, userAgent, ipAddress string) error {
	q := sqlc.New(r.pool)
	return q.AuthCreateRefreshToken(ctx, sqlc.AuthCreateRefreshTokenParams{
		ID:           uuidToPgtype(token.ID),
		UserID:       uuidToPgtype(token.UserID),
		MembershipID: uuidToPgtype(token.MembershipID),
		TokenHash:    token.TokenHash,
		ExpiresAt:    timeToPgtypeTimestamptz(token.ExpiresAt),
		UserAgent:    stringToPgtypeText(userAgent),
		IpAddress:    stringToPgtypeText(ipAddress),
	})
}

func (r *Repository) FindActiveRefreshToken(ctx context.Context, tokenHash string) (domain.RefreshToken, domain.User, domain.Membership, error) {
	q := sqlc.New(r.pool)
	row, err := q.AuthFindActiveRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RefreshToken{}, domain.User{}, domain.Membership{}, domain.ErrNotFound
		}
		return domain.RefreshToken{}, domain.User{}, domain.Membership{}, err
	}

	token := domain.RefreshToken{
		ID:           pgtypeUUIDToUUID(row.ID),
		UserID:       pgtypeUUIDToUUID(row.UserID),
		MembershipID: pgtypeUUIDToUUID(row.MembershipID),
		TokenHash:    row.TokenHash,
		ExpiresAt:    pgtypeTimestamptzToTime(row.ExpiresAt),
		RevokedAt:    pgtypeTimestamptzToTimePtr(row.RevokedAt),
	}
	user := domain.User{
		ID:           pgtypeUUIDToUUID(row.UserTableID),
		Email:        row.UserEmail,
		PasswordHash: row.UserPasswordHash,
		IsActive:     row.UserIsActive,
	}
	membership := domain.Membership{
		ID:         pgtypeUUIDToUUID(row.MembershipTableID),
		TenantID:   pgtypeUUIDToUUID(row.MembershipTenantID),
		TenantName: row.MembershipTenantName,
		BranchID:   pgtypeUUIDToUUID(row.MembershipBranchID),
		BranchName: pgtypeTextToString(row.MembershipBranchName),
		Role:       row.MembershipRole,
		IsActive:   row.MembershipIsActive,
	}

	if token.RevokedAt != nil || time.Now().UTC().After(token.ExpiresAt) || !user.IsActive {
		return domain.RefreshToken{}, domain.User{}, domain.Membership{}, domain.ErrNotFound
	}

	return token, user, membership, nil
}

func (r *Repository) RotateRefreshToken(ctx context.Context, oldTokenID uuid.UUID, replacement domain.RefreshToken, userAgent, ipAddress string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)

	if err := q.AuthRevokeRefreshTokenByID(ctx, uuidToPgtype(oldTokenID)); err != nil {
		return err
	}

	if err := q.AuthCreateRefreshToken(ctx, sqlc.AuthCreateRefreshTokenParams{
		ID:           uuidToPgtype(replacement.ID),
		UserID:       uuidToPgtype(replacement.UserID),
		MembershipID: uuidToPgtype(replacement.MembershipID),
		TokenHash:    replacement.TokenHash,
		ExpiresAt:    timeToPgtypeTimestamptz(replacement.ExpiresAt),
		UserAgent:    stringToPgtypeText(userAgent),
		IpAddress:    stringToPgtypeText(ipAddress),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	q := sqlc.New(r.pool)
	return q.AuthRevokeRefreshTokenByHash(ctx, tokenHash)
}

func (r *Repository) CreateScopeSwitchAuditLog(ctx context.Context, actorUserID uuid.UUID, fromMembership, toMembership domain.Membership, requestID string) error {
	auditBranchID := toMembership.BranchID
	if auditBranchID == uuid.Nil {
		auditBranchID = fromMembership.BranchID
	}

	q := sqlc.New(r.pool)
	return q.AuthCreateScopeSwitchAuditLog(ctx, sqlc.AuthCreateScopeSwitchAuditLogParams{
		ID:                uuidToPgtype(uuid.New()),
		TenantID:          uuidToPgtype(toMembership.TenantID),
		BranchID:          uuidToPgtype(auditBranchID),
		ActorUserID:       uuidToPgtype(actorUserID),
		ActorMembershipID: uuidToPgtype(fromMembership.ID),
		ActionType:        "session_scope_switched",
		ActionEntityType:  "membership",
		ActionEntityID:    uuidToPgtype(toMembership.ID),
		RequestID:         stringToPgtypeText(requestID),
		Column10:          uuidToPgtype(fromMembership.ID),
		Column11:          uuidToPgtype(toMembership.ID),
	})
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func timeToPgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func stringToPgtypeText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func pgtypeTextToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}
