package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/rooms/domain"
)

type GetRoom struct {
	repo domain.Repository
}

func NewGetRoom(repo domain.Repository) *GetRoom {
	return &GetRoom{repo: repo}
}

func (uc *GetRoom) Execute(ctx context.Context, actor RoomActor, siteID, roomID uuid.UUID) (domain.Room, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Room{}, err
	}

	room, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, roomID)
	if err != nil {
		return domain.Room{}, err
	}

	return room, nil
}
