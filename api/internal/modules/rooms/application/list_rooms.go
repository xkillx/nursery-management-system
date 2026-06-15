package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/rooms/domain"
)

type ListRooms struct {
	repo domain.Repository
}

func NewListRooms(repo domain.Repository) *ListRooms {
	return &ListRooms{repo: repo}
}

func (uc *ListRooms) Execute(ctx context.Context, actor RoomActor, siteID uuid.UUID, includeArchived bool) ([]domain.Room, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	rooms, err := uc.repo.ListByBranch(ctx, actor.TenantID(), siteID, includeArchived)
	if err != nil {
		return nil, internalError(err)
	}

	return rooms, nil
}
