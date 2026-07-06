package application

import (
	"context"

	"nursery-management-system/api/internal/modules/invites/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListInvitesUseCase struct {
	repo domain.Repository
}

func NewListInvitesUseCase(repo domain.Repository) *ListInvitesUseCase {
	return &ListInvitesUseCase{repo: repo}
}

type ListInvitesResult struct {
	Invites []domain.Invite
}

func (uc *ListInvitesUseCase) Execute(ctx context.Context, actor tenant.ActorContext, status domain.InviteStatus) (ListInvitesResult, error) {
	invites, err := uc.repo.ListInvites(ctx, actor.TenantID, actor.BranchID, status)
	if err != nil {
		return ListInvitesResult{}, err
	}
	if invites == nil {
		invites = []domain.Invite{}
	}
	return ListInvitesResult{Invites: invites}, nil
}

func (uc *ListInvitesUseCase) ExecutePaginated(ctx context.Context, actor tenant.ActorContext, status domain.InviteStatus, limit, offset int) (ListInvitesResult, int, error) {
	invites, err := uc.repo.ListInvitesPaginated(ctx, actor.TenantID, actor.BranchID, status, limit, offset)
	if err != nil {
		return ListInvitesResult{}, 0, err
	}
	if invites == nil {
		invites = []domain.Invite{}
	}

	total, err := uc.repo.CountInvites(ctx, actor.TenantID, actor.BranchID, status)
	if err != nil {
		return ListInvitesResult{}, 0, err
	}

	return ListInvitesResult{Invites: invites}, total, nil
}

// ParseStatus returns the InviteStatus for the query value, defaulting to pending.
func ParseStatus(v string) (domain.InviteStatus, bool) {
	switch v {
	case "", "pending":
		return domain.StatusPending, true
	case "accepted":
		return domain.StatusAccepted, true
	case "revoked":
		return domain.StatusRevoked, true
	case "expired":
		return domain.StatusExpired, true
	case "all":
		return domain.StatusAll, true
	default:
		return "", false
	}
}
