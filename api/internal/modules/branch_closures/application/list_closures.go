package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/branch_closures/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type ListClosureDays struct {
	repo domain.Repository
}

func NewListClosureDays(repo domain.Repository) *ListClosureDays {
	return &ListClosureDays{repo: repo}
}

func (uc *ListClosureDays) Execute(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]domain.BranchClosureDay, error) {
	if from.IsZero() || to.IsZero() {
		return nil, domainerrors.Validation("Both from and to dates are required.", "date_range")
	}

	if to.Before(from) {
		return nil, domainerrors.Validation("To date must not be before from date.", "date_range")
	}

	closures, err := uc.repo.ListByBranchAndDateRange(ctx, tenantID, branchID, from, to)
	if err != nil {
		return nil, domainerrors.Internal(err)
	}

	return closures, nil
}
