package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/invites/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/db/sqlc"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type Repository struct {
	pool        *pgxpool.Pool
	auditWriter *audit.Writer
}

func NewRepository(pool *pgxpool.Pool, auditWriter *audit.Writer) *Repository {
	return &Repository{pool: pool, auditWriter: auditWriter}
}

func (r *Repository) CreateInvite(ctx context.Context, actor tenant.ActorContext, inv domain.Invite, sendEmail func() error) (domain.Invite, bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Invite{}, false, err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)

	lockKey := fmt.Sprintf("%s|%s|%s", actor.TenantID, actor.BranchID, inv.EmailNormalized)
	_, err = q.InviteAcquireEmailScopeLock(ctx, lockKey)
	if err != nil {
		return domain.Invite{}, false, err
	}

	_, err = q.InviteFindUserByEmailNormalized(ctx, inv.EmailNormalized)
	if err == nil {
		return domain.Invite{}, false, domain.ErrEmailAlreadyRegistered
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Invite{}, false, err
	}

	existing, err := q.InviteFindLivePendingByEmailForUpdate(ctx, sqlc.InviteFindLivePendingByEmailForUpdateParams{
		TenantID:        pgUUID(actor.TenantID),
		BranchID:        pgUUID(actor.BranchID),
		EmailNormalized: inv.EmailNormalized,
	})
	if err == nil {
		existInv := scanManagerInvite(existing)
		if existInv.Role != inv.Role {
			return domain.Invite{}, false, domain.ErrScopeConflict
		}

		if err := q.InviteRefreshPending(ctx, sqlc.InviteRefreshPendingParams{
			ID:                   pgUUID(existInv.ID),
			TokenHash:            inv.TokenHash,
			ExpiresAt:            pgTimestamptz(inv.ExpiresAt),
			ResentByUserID:       pgUUID(actor.UserID),
			ResentByMembershipID: pgUUID(actor.MembershipID),
		}); err != nil {
			return domain.Invite{}, false, err
		}

		if err := sendEmail(); err != nil {
			return domain.Invite{}, false, err
		}

		if err := r.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "invite_resent",
			EntityType: "manager_invite",
			EntityID:   existInv.ID,
			Details:    map[string]any{"email_normalized": inv.EmailNormalized, "role": inv.Role},
		}); err != nil {
			return domain.Invite{}, false, err
		}

		existInv.TokenHash = inv.TokenHash
		existInv.ExpiresAt = inv.ExpiresAt
		existInv.SendCount++

		if err := tx.Commit(ctx); err != nil {
			return domain.Invite{}, false, err
		}
		return existInv, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Invite{}, false, err
	}

	if err := q.InviteCreate(ctx, sqlc.InviteCreateParams{
		ID:                    pgUUID(inv.ID),
		TenantID:              pgUUID(inv.TenantID),
		BranchID:              pgUUID(inv.BranchID),
		Email:                 inv.Email,
		EmailNormalized:       inv.EmailNormalized,
		Role:                  inv.Role,
		TokenHash:             inv.TokenHash,
		ExpiresAt:             pgTimestamptz(inv.ExpiresAt),
		CreatedByUserID:       pgUUID(inv.CreatedByUserID),
		CreatedByMembershipID: pgUUID(inv.CreatedByMembershipID),
	}); err != nil {
		return domain.Invite{}, false, err
	}

	if err := sendEmail(); err != nil {
		return domain.Invite{}, false, err
	}

	if err := r.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
		ActionType: "invite_created",
		EntityType: "manager_invite",
		EntityID:   inv.ID,
		Details:    map[string]any{"email_normalized": inv.EmailNormalized, "role": inv.Role},
	}); err != nil {
		return domain.Invite{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Invite{}, false, err
	}
	return inv, true, nil
}

func (r *Repository) ResendInvite(ctx context.Context, actor tenant.ActorContext, inviteID uuid.UUID, tokenHash string, expiresAt time.Time, sendEmail func() error) (domain.Invite, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Invite{}, err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)

	row, err := q.InviteGetByIDForUpdate(ctx, sqlc.InviteGetByIDForUpdateParams{
		ID:       pgUUID(inviteID),
		TenantID: pgUUID(actor.TenantID),
		BranchID: pgUUID(actor.BranchID),
	})
	if err != nil {
		return domain.Invite{}, err
	}
	inv := scanManagerInvite(row)

	if err := q.InviteRefreshPending(ctx, sqlc.InviteRefreshPendingParams{
		ID:                   pgUUID(inviteID),
		TokenHash:            tokenHash,
		ExpiresAt:            pgTimestamptz(expiresAt),
		ResentByUserID:       pgUUID(actor.UserID),
		ResentByMembershipID: pgUUID(actor.MembershipID),
	}); err != nil {
		return domain.Invite{}, err
	}

	if err := sendEmail(); err != nil {
		return domain.Invite{}, err
	}

	if err := r.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
		ActionType: "invite_resent",
		EntityType: "manager_invite",
		EntityID:   inviteID,
	}); err != nil {
		return domain.Invite{}, err
	}

	inv.TokenHash = tokenHash
	inv.ExpiresAt = expiresAt
	inv.SendCount++

	return inv, tx.Commit(ctx)
}

func (r *Repository) ListInvites(ctx context.Context, tenantID, branchID uuid.UUID, status domain.InviteStatus) ([]domain.Invite, error) {
	q := sqlc.New(r.pool)
	pgTenant := pgUUID(tenantID)
	pgBranch := pgUUID(branchID)

	switch status {
	case domain.StatusAccepted:
		rows, err := q.InviteListAccepted(ctx, sqlc.InviteListAcceptedParams{TenantID: pgTenant, BranchID: pgBranch})
		if err != nil {
			return nil, err
		}
		return mapRows(rows), nil
	case domain.StatusRevoked:
		rows, err := q.InviteListRevoked(ctx, sqlc.InviteListRevokedParams{TenantID: pgTenant, BranchID: pgBranch})
		if err != nil {
			return nil, err
		}
		return mapRows(rows), nil
	case domain.StatusExpired:
		rows, err := q.InviteListExpired(ctx, sqlc.InviteListExpiredParams{TenantID: pgTenant, BranchID: pgBranch})
		if err != nil {
			return nil, err
		}
		return mapRows(rows), nil
	case domain.StatusAll:
		rows, err := q.InviteListAll(ctx, sqlc.InviteListAllParams{TenantID: pgTenant, BranchID: pgBranch})
		if err != nil {
			return nil, err
		}
		return mapRows(rows), nil
	default:
		rows, err := q.InviteListPending(ctx, sqlc.InviteListPendingParams{TenantID: pgTenant, BranchID: pgBranch})
		if err != nil {
			return nil, err
		}
		return mapRows(rows), nil
	}
}

func (r *Repository) GetInviteForUpdate(ctx context.Context, tenantID, branchID, inviteID uuid.UUID) (domain.Invite, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Invite{}, err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)
	row, err := q.InviteGetByIDForUpdate(ctx, sqlc.InviteGetByIDForUpdateParams{
		ID:       pgUUID(inviteID),
		TenantID: pgUUID(tenantID),
		BranchID: pgUUID(branchID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Invite{}, domain.ErrInviteNotFound
		}
		return domain.Invite{}, err
	}

	inv := scanManagerInvite(row)
	return inv, tx.Commit(ctx)
}

func (r *Repository) GetInviteByTokenHashForUpdate(ctx context.Context, tokenHash string) (domain.Invite, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Invite{}, err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)
	row, err := q.InviteGetByTokenHashForUpdate(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Invite{}, domain.ErrTokenInvalid
		}
		return domain.Invite{}, err
	}

	inv := scanManagerInvite(row)
	return inv, tx.Commit(ctx)
}

func (r *Repository) RevokeInvite(ctx context.Context, actor tenant.ActorContext, inviteID uuid.UUID) (domain.Invite, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Invite{}, err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)
	row, err := q.InviteGetByIDForUpdate(ctx, sqlc.InviteGetByIDForUpdateParams{
		ID:       pgUUID(inviteID),
		TenantID: pgUUID(actor.TenantID),
		BranchID: pgUUID(actor.BranchID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Invite{}, domain.ErrInviteNotFound
		}
		return domain.Invite{}, err
	}

	inv := scanManagerInvite(row)
	if inv.AcceptedAt != nil {
		return domain.Invite{}, domain.ErrInviteAccepted
	}
	if inv.RevokedAt != nil {
		return inv, tx.Commit(ctx)
	}

	if err := q.InviteRevoke(ctx, sqlc.InviteRevokeParams{
		ID:                    pgUUID(inviteID),
		RevokedByUserID:       pgUUID(actor.UserID),
		RevokedByMembershipID: pgUUID(actor.MembershipID),
	}); err != nil {
		return domain.Invite{}, err
	}

	if err := r.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
		ActionType: "invite_revoked",
		EntityType: "manager_invite",
		EntityID:   inviteID,
	}); err != nil {
		return domain.Invite{}, err
	}

	inv.RevokedAt = ptrTime(time.Now().UTC())
	inv.RevokedByUserID = actor.UserID
	inv.RevokedByMembershipID = actor.MembershipID

	return inv, tx.Commit(ctx)
}

func (r *Repository) AcceptInvite(ctx context.Context, invite domain.Invite, user domain.CreatedUser, membership domain.CreatedMembership) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)

	_, err = q.InviteFindUserByEmailNormalized(ctx, invite.EmailNormalized)
	if err == nil {
		return domain.ErrEmailAlreadyRegistered
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	_, err = q.InviteCreateUser(ctx, sqlc.InviteCreateUserParams{
		ID:              pgUUID(user.ID),
		Email:           user.Email,
		EmailNormalized: normalizeEmail(user.Email),
		PasswordHash:    user.PasswordHash,
	})
	if err != nil {
		return err
	}

	if err := q.InviteCreateMembership(ctx, sqlc.InviteCreateMembershipParams{
		ID:       pgUUID(membership.ID),
		TenantID: pgUUID(membership.TenantID),
		BranchID: pgUUID(membership.BranchID),
		UserID:   pgUUID(membership.UserID),
		Role:     membership.Role,
	}); err != nil {
		return err
	}

	if err := q.InviteAccept(ctx, sqlc.InviteAcceptParams{
		ID:                   pgUUID(invite.ID),
		AcceptedUserID:       pgUUID(user.ID),
		AcceptedMembershipID: pgUUID(membership.ID),
	}); err != nil {
		return err
	}

	actor := tenant.ActorContext{
		TenantID:     invite.TenantID,
		BranchID:     invite.BranchID,
		UserID:       user.ID,
		MembershipID: membership.ID,
	}

	if err := r.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
		ActionType: "invite_accepted",
		EntityType: "manager_invite",
		EntityID:   invite.ID,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func scanManagerInvite(r sqlc.ManagerInvite) domain.Invite {
	return domain.Invite{
		ID:                    r.ID.Bytes,
		TenantID:              r.TenantID.Bytes,
		BranchID:              r.BranchID.Bytes,
		Email:                 r.Email,
		EmailNormalized:       r.EmailNormalized,
		Role:                  r.Role,
		TokenHash:             r.TokenHash,
		ExpiresAt:             r.ExpiresAt.Time,
		AcceptedAt:            pgTimePtr(r.AcceptedAt),
		AcceptedUserID:        r.AcceptedUserID.Bytes,
		AcceptedMembershipID:  r.AcceptedMembershipID.Bytes,
		RevokedAt:             pgTimePtr(r.RevokedAt),
		RevokedByUserID:       r.RevokedByUserID.Bytes,
		RevokedByMembershipID: r.RevokedByMembershipID.Bytes,
		CreatedByUserID:       r.CreatedByUserID.Bytes,
		CreatedByMembershipID: r.CreatedByMembershipID.Bytes,
		ResentAt:              pgTimePtr(r.ResentAt),
		ResentByUserID:        r.ResentByUserID.Bytes,
		ResentByMembershipID:  r.ResentByMembershipID.Bytes,
		SendCount:             int(r.SendCount),
		CreatedAt:             r.CreatedAt.Time,
		UpdatedAt:             r.UpdatedAt.Time,
	}
}

func mapRows(rows []sqlc.ManagerInvite) []domain.Invite {
	out := make([]domain.Invite, len(rows))
	for i, r := range rows {
		out[i] = scanManagerInvite(r)
	}
	return out
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	var pg pgtype.UUID
	copy(pg.Bytes[:], id[:])
	pg.Valid = true
	return pg
}

func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

var _ = uid.NewUUID
var _ domain.InviteStatus
