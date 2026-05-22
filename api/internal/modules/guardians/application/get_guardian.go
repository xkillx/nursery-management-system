package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/guardians/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetGuardian struct {
	repo domain.Repository
}

func NewGetGuardian(repo domain.Repository) *GetGuardian {
	return &GetGuardian{repo: repo}
}

func (uc *GetGuardian) Execute(ctx context.Context, actor tenant.ActorContext, guardianID uuid.UUID) (domain.Guardian, error) {
	return uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, guardianID)
}
