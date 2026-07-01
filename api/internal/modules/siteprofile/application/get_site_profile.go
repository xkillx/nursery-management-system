package application

import (
	"context"

	"nursery-management-system/api/internal/modules/siteprofile/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetSiteProfileUseCase struct {
	repo domain.Repository
}

func NewGetSiteProfileUseCase(repo domain.Repository) *GetSiteProfileUseCase {
	return &GetSiteProfileUseCase{repo: repo}
}

func (uc *GetSiteProfileUseCase) Execute(ctx context.Context, actor tenant.ActorContext) (*domain.SiteProfile, error) {
	profile, err := uc.repo.GetByBranch(ctx, actor.TenantID, actor.BranchID)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}
