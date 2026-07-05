package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/branch_closures/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CreateClosureDay struct {
	repo domain.Repository
}

func NewCreateClosureDay(repo domain.Repository) *CreateClosureDay {
	return &CreateClosureDay{repo: repo}
}

type CreateClosureDayParams struct {
	Date   time.Time
	Reason *string
}

func (uc *CreateClosureDay) Execute(ctx context.Context, tenantID, branchID uuid.UUID, params CreateClosureDayParams) (domain.BranchClosureDay, error) {
	if params.Date.IsZero() {
		return domain.BranchClosureDay{}, domain.ErrDateRequired
	}

	dateOnly := time.Date(params.Date.Year(), params.Date.Month(), params.Date.Day(), 0, 0, 0, 0, time.UTC)

	exists, err := uc.repo.DateExists(ctx, tenantID, branchID, dateOnly)
	if err != nil {
		return domain.BranchClosureDay{}, domainerrors.Internal(err)
	}
	if exists {
		return domain.BranchClosureDay{}, domain.ErrDuplicateDate
	}

	closure := domain.BranchClosureDay{
		ID:       uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
		Date:     dateOnly,
		Reason:   params.Reason,
	}

	if err := uc.repo.Create(ctx, closure); err != nil {
		return domain.BranchClosureDay{}, domainerrors.Internal(err)
	}

	return closure, nil
}
