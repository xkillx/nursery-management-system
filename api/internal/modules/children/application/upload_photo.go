package application

import (
	"context"
	"fmt"
	"io"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type UploadPhoto struct {
	repo    domain.Repository
	storage domain.FileStorage
}

func NewUploadPhoto(repo domain.Repository, storage domain.FileStorage) *UploadPhoto {
	return &UploadPhoto{repo: repo, storage: storage}
}

type UploadPhotoResult struct {
	PhotoURL string `json:"photo_url"`
}

func (uc *UploadPhoto) Execute(ctx context.Context, actor tenant.ActorContext, childID string, data io.Reader, ext string) (*UploadPhotoResult, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	child, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}

	if child.ProfilePhotoPath != nil {
		if delErr := uc.storage.Delete(ctx, *child.ProfilePhotoPath); delErr != nil {
			return nil, domainerrors.Internal(fmt.Errorf("delete old photo: %w", delErr))
		}
	}

	path, err := uc.storage.Save(ctx, actor.TenantID, actor.BranchID, id, data, ext)
	if err != nil {
		return nil, domainerrors.Validation(err.Error(), "photo")
	}

	if err := uc.repo.UpdatePhotoPath(ctx, actor.TenantID, actor.BranchID, id, &path); err != nil {
		uc.storage.Delete(ctx, path)
		return nil, domainerrors.Internal(fmt.Errorf("update photo path: %w", err))
	}

	url, err := uc.storage.GetURL(ctx, path)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get photo URL: %w", err))
	}

	return &UploadPhotoResult{PhotoURL: url}, nil
}
