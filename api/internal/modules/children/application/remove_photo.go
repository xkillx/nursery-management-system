package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type RemovePhoto struct {
	repo    domain.Repository
	storage domain.FileStorage
}

func NewRemovePhoto(repo domain.Repository, storage domain.FileStorage) *RemovePhoto {
	return &RemovePhoto{repo: repo, storage: storage}
}

func (uc *RemovePhoto) Execute(ctx context.Context, actor tenant.ActorContext, childID string) error {
	id, err := parseUUID(childID)
	if err != nil {
		return domainerrors.Validation("Invalid request payload.", "child_id")
	}

	child, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return domainerrors.Internal(fmt.Errorf("get child: %w", err))
	}
	if !found {
		return domainerrors.NotFound("child", "Resource not found.")
	}

	if child.ProfilePhotoPath != nil {
		if delErr := uc.storage.Delete(ctx, *child.ProfilePhotoPath); delErr != nil {
			return domainerrors.Internal(fmt.Errorf("delete photo: %w", delErr))
		}
	}

	if err := uc.repo.UpdatePhotoPath(ctx, actor.TenantID, actor.BranchID, id, nil); err != nil {
		return domainerrors.Internal(fmt.Errorf("clear photo path: %w", err))
	}

	return nil
}
