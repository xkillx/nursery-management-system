package application

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/rooms/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateRoomParams struct {
	Name        string
	AgeGroup    string
	Capacity    int
	Description string
}

type CreateRoom struct {
	repo        domain.Repository
	siteChecker SiteExistsChecker
}

func NewCreateRoom(repo domain.Repository, siteChecker SiteExistsChecker) *CreateRoom {
	return &CreateRoom{repo: repo, siteChecker: siteChecker}
}

func (uc *CreateRoom) Execute(ctx context.Context, actor RoomActor, siteID uuid.UUID, params CreateRoomParams) (domain.Room, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return domain.Room{}, domainerrors.Validation("Invalid request payload.", "name")
	}
	if len(name) > 255 {
		return domain.Room{}, domainerrors.Validation("Invalid request payload.", "name")
	}

	if !domain.IsValidAgeGroup(params.AgeGroup) {
		return domain.Room{}, domainerrors.New("invalid_age_group", "Invalid request payload.", "age_group")
	}

	if params.Capacity <= 0 {
		return domain.Room{}, domainerrors.Validation("Invalid request payload.", "capacity")
	}

	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Room{}, err
	}

	if IsOwnerActor(actor) {
		exists, err := uc.siteChecker.SiteExists(ctx, actor.TenantID(), siteID)
		if err != nil {
			return domain.Room{}, internalError(err)
		}
		if !exists {
			return domain.Room{}, domainerrors.NotFound("site", "Site not found.")
		}
	}

	exists, err := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, name, nil)
	if err != nil {
		return domain.Room{}, internalError(err)
	}
	if exists {
		return domain.Room{}, domainerrors.Conflict("room_name_duplicate", "An active room with this name already exists in this site.")
	}

	var description *string
	if desc := strings.TrimSpace(params.Description); desc != "" {
		if len(desc) > 1000 {
			return domain.Room{}, domainerrors.Validation("Invalid request payload.", "description")
		}
		description = &desc
	}

	room := domain.Room{
		ID:          uid.NewUUID(),
		TenantID:    actor.TenantID(),
		BranchID:    siteID,
		Name:        name,
		Description: description,
		AgeGroup:    params.AgeGroup,
		Capacity:    params.Capacity,
		IsActive:    true,
	}

	if err := uc.repo.Create(ctx, room); err != nil {
		return domain.Room{}, internalError(err)
	}

	created, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, room.ID)
	if err != nil {
		return domain.Room{}, internalError(err)
	}

	return created, nil
}
