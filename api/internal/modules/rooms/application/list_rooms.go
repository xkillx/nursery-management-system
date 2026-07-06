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

func (uc *ListRooms) Execute(ctx context.Context, actor RoomActor, siteID uuid.UUID, includeArchived, includeOccupancy bool) ([]domain.Room, map[uuid.UUID]int, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, nil, err
	}

	rooms, err := uc.repo.ListByBranch(ctx, actor.TenantID(), siteID, includeArchived)
	if err != nil {
		return nil, nil, internalError(err)
	}

	if !includeOccupancy {
		return rooms, nil, nil
	}

	counts, err := uc.repo.CountAssignedChildrenByBranch(ctx, actor.TenantID(), siteID)
	if err != nil {
		return nil, nil, internalError(err)
	}

	return rooms, counts, nil
}

func (uc *ListRooms) ExecutePaginated(ctx context.Context, actor RoomActor, siteID uuid.UUID, includeArchived, includeOccupancy bool, limit, offset int, sortField, sortDir string) ([]domain.Room, map[uuid.UUID]int, int, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, nil, 0, err
	}

	var rooms []domain.Room
	var err error
	if sortField != "" && sortDir != "" {
		rooms, err = uc.repo.ListByBranchPaginatedSorted(ctx, actor.TenantID(), siteID, includeArchived, limit, offset, sortField, sortDir)
	} else {
		rooms, err = uc.repo.ListByBranchPaginated(ctx, actor.TenantID(), siteID, includeArchived, limit, offset)
	}
	if err != nil {
		return nil, nil, 0, internalError(err)
	}

	total, err := uc.repo.CountByBranch(ctx, actor.TenantID(), siteID, includeArchived)
	if err != nil {
		return nil, nil, 0, internalError(err)
	}

	if !includeOccupancy {
		return rooms, nil, total, nil
	}

	counts, err := uc.repo.CountAssignedChildrenByBranch(ctx, actor.TenantID(), siteID)
	if err != nil {
		return nil, nil, 0, internalError(err)
	}

	return rooms, counts, total, nil
}
