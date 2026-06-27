package application

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/rooms/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateRoomParams struct {
	Name        *string
	AgeGroup    *string
	Capacity    *int
	Description *string
}

type UpdateRoom struct {
	repo        domain.Repository
	siteChecker SiteExistsChecker
}

func NewUpdateRoom(repo domain.Repository, siteChecker SiteExistsChecker) *UpdateRoom {
	return &UpdateRoom{repo: repo, siteChecker: siteChecker}
}

func (uc *UpdateRoom) Execute(ctx context.Context, actor RoomActor, siteID, roomID uuid.UUID, params UpdateRoomParams) (domain.Room, error) {
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

	existing, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, roomID)
	if err != nil {
		return domain.Room{}, err
	}

	fields := make(map[string]any)

	if params.Name != nil {
		name := strings.TrimSpace(*params.Name)
		if name == "" {
			return domain.Room{}, domainerrors.Validation("Invalid request payload.", "name")
		}
		if len(name) > 255 {
			return domain.Room{}, domainerrors.Validation("Invalid request payload.", "name")
		}
		if name != existing.Name {
			exists, err := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, name, &roomID)
			if err != nil {
				return domain.Room{}, internalError(err)
			}
			if exists {
				return domain.Room{}, domainerrors.Conflict("room_name_duplicate", "An active room with this name already exists in this site.")
			}
			fields["name"] = name
		}
	}

	if params.AgeGroup != nil {
		if !domain.IsValidAgeGroup(*params.AgeGroup) {
			return domain.Room{}, domainerrors.New("invalid_age_group", "Invalid request payload.", "age_group")
		}
		fields["age_group"] = *params.AgeGroup
	}

	if params.Capacity != nil {
		if *params.Capacity <= 0 {
			return domain.Room{}, domainerrors.Validation("Invalid request payload.", "capacity")
		}
		fields["capacity"] = *params.Capacity
	}

	if params.Description != nil {
		desc := strings.TrimSpace(*params.Description)
		if len(desc) > 1000 {
			return domain.Room{}, domainerrors.Validation("Invalid request payload.", "description")
		}
		fields["description"] = desc
	}

	if len(fields) == 0 {
		return existing, nil
	}

	rowsAffected, err := uc.repo.Update(ctx, actor.TenantID(), siteID, roomID, fields)
	if err != nil {
		return domain.Room{}, internalError(err)
	}
	if rowsAffected == 0 {
		return domain.Room{}, domainerrors.NotFound("room", "Room not found.")
	}

	updated, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, roomID)
	if err != nil {
		return domain.Room{}, internalError(err)
	}

	return updated, nil
}
