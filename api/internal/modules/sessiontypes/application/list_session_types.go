package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
)

type ListSessionTypes struct {
	repo domain.Repository
}

func NewListSessionTypes(repo domain.Repository) *ListSessionTypes {
	return &ListSessionTypes{repo: repo}
}

func (uc *ListSessionTypes) Execute(ctx context.Context, actor SessionTypeActor, siteID uuid.UUID, includeArchived bool) ([]domain.SessionType, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	types, err := uc.repo.ListByBranch(ctx, actor.TenantID(), siteID, includeArchived)
	if err != nil {
		return nil, internalError(err)
	}
	return types, nil
}

func (uc *ListSessionTypes) ExecutePaginated(ctx context.Context, actor SessionTypeActor, siteID uuid.UUID, includeArchived bool, limit, offset int, sortField, sortDir string) ([]domain.SessionType, int, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, 0, err
	}

	var types []domain.SessionType
	var err error
	if sortField != "" && sortDir != "" {
		types, err = uc.repo.ListByBranchPaginatedSorted(ctx, actor.TenantID(), siteID, includeArchived, limit, offset, sortField, sortDir)
	} else {
		types, err = uc.repo.ListByBranchPaginated(ctx, actor.TenantID(), siteID, includeArchived, limit, offset)
	}
	if err != nil {
		return nil, 0, internalError(err)
	}

	total, err := uc.repo.CountByBranch(ctx, actor.TenantID(), siteID, includeArchived)
	if err != nil {
		return nil, 0, internalError(err)
	}

	return types, total, nil
}
