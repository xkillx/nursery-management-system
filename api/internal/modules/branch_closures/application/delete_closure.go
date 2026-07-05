package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/branch_closures/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type DeleteClosureDay struct {
	repo domain.Repository
}

func NewDeleteClosureDay(repo domain.Repository) *DeleteClosureDay {
	return &DeleteClosureDay{repo: repo}
}

func (uc *DeleteClosureDay) Execute(ctx context.Context, tenantID, branchID, id uuid.UUID) error {
	if err := uc.repo.Delete(ctx, tenantID, branchID, id); err != nil {
		return domainerrors.Internal(err)
	}
	return nil
}
