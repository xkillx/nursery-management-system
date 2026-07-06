package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/platform/tenant"
)

type Repository interface {
	CreateInvite(ctx context.Context, actor tenant.ActorContext, inv Invite, sendEmail func() error) (Invite, bool, error)
	ResendInvite(ctx context.Context, actor tenant.ActorContext, inviteID uuid.UUID, tokenHash string, expiresAt time.Time, sendEmail func() error) (Invite, error)
	ListInvites(ctx context.Context, tenantID, branchID uuid.UUID, status InviteStatus) ([]Invite, error)
	ListInvitesPaginated(ctx context.Context, tenantID, branchID uuid.UUID, status InviteStatus, limit, offset int, role *string) ([]Invite, error)
	ListInvitesPaginatedSorted(ctx context.Context, tenantID, branchID uuid.UUID, status InviteStatus, limit, offset int, role *string, sortField, sortDir string) ([]Invite, error)
	CountInvites(ctx context.Context, tenantID, branchID uuid.UUID, status InviteStatus, role *string) (int, error)
	GetInviteForUpdate(ctx context.Context, tenantID, branchID, inviteID uuid.UUID) (Invite, error)
	GetInviteByTokenHashForUpdate(ctx context.Context, tokenHash string) (Invite, error)
	RevokeInvite(ctx context.Context, actor tenant.ActorContext, inviteID uuid.UUID) (Invite, error)
	AcceptInvite(ctx context.Context, invite Invite, user CreatedUser, membership CreatedMembership) error
}
