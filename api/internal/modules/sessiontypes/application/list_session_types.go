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
