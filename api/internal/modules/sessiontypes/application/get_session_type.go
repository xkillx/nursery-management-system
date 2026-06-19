package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
)

type GetSessionType struct {
	repo domain.Repository
}

func NewGetSessionType(repo domain.Repository) *GetSessionType {
	return &GetSessionType{repo: repo}
}

func (uc *GetSessionType) Execute(ctx context.Context, actor SessionTypeActor, siteID, stID uuid.UUID) (domain.SessionType, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.SessionType{}, err
	}

	st, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, stID)
	if err != nil {
		return domain.SessionType{}, err
	}
	return st, nil
}
