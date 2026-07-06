package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
)

type ListSessionTemplates struct {
	repo domain.Repository
}

func NewListSessionTemplates(repo domain.Repository) *ListSessionTemplates {
	return &ListSessionTemplates{repo: repo}
}

func (uc *ListSessionTemplates) Execute(ctx context.Context, actor SessionTemplateActor, siteID uuid.UUID, includeArchived bool) ([]domain.SessionTemplate, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	templates, err := uc.repo.ListByBranch(ctx, actor.TenantID(), siteID, includeArchived)
	if err != nil {
		return nil, internalError(err)
	}
	return templates, nil
}

func (uc *ListSessionTemplates) ExecutePaginated(ctx context.Context, actor SessionTemplateActor, siteID uuid.UUID, includeArchived bool, limit, offset int) ([]domain.SessionTemplate, int, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, 0, err
	}

	templates, err := uc.repo.ListByBranchPaginated(ctx, actor.TenantID(), siteID, includeArchived, limit, offset)
	if err != nil {
		return nil, 0, internalError(err)
	}

	total, err := uc.repo.CountByBranch(ctx, actor.TenantID(), siteID, includeArchived)
	if err != nil {
		return nil, 0, internalError(err)
	}

	return templates, total, nil
}
