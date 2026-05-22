package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetChild struct {
	repo domain.Repository
}

func NewGetChild(repo domain.Repository) *GetChild {
	return &GetChild{repo: repo}
}

func (uc *GetChild) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (domain.Child, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	child, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("get child: %w", err))
	}
	if !found {
		return domain.Child{}, domainerrors.NotFound("child", "Resource not found.")
	}

	return child, nil
}
