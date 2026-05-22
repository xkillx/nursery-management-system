package application

import (
	"context"

	"nursery-management-system/api/internal/modules/guardians/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListGuardians struct {
	repo domain.Repository
}

func NewListGuardians(repo domain.Repository) *ListGuardians {
	return &ListGuardians{repo: repo}
}

func (uc *ListGuardians) Execute(ctx context.Context, actor tenant.ActorContext, filter domain.StatusFilter, limit, offset int) ([]domain.Guardian, error) {
	return uc.repo.List(ctx, actor.TenantID, actor.BranchID, filter, limit, offset)
}
