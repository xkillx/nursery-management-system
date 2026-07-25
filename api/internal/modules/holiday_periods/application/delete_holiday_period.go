package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/holiday_periods/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type DeleteHolidayPeriod struct {
	repo domain.Repository
}

func NewDeleteHolidayPeriod(repo domain.Repository) *DeleteHolidayPeriod {
	return &DeleteHolidayPeriod{repo: repo}
}

func (uc *DeleteHolidayPeriod) Execute(ctx context.Context, tenantID, branchID, id uuid.UUID) error {
	_, found, err := uc.repo.GetByID(ctx, tenantID, branchID, id)
	if err != nil {
		return domainerrors.Internal(err)
	}
	if !found {
		return domain.ErrHolidayPeriodNotFound
	}

	if err := uc.repo.Delete(ctx, tenantID, branchID, id); err != nil {
		return domainerrors.Internal(err)
	}
	return nil
}
