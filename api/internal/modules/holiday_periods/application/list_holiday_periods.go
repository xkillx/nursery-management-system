package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/holiday_periods/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type ListHolidayPeriods struct {
	repo domain.Repository
}

func NewListHolidayPeriods(repo domain.Repository) *ListHolidayPeriods {
	return &ListHolidayPeriods{repo: repo}
}

func (uc *ListHolidayPeriods) Execute(ctx context.Context, tenantID, branchID uuid.UUID) ([]domain.HolidayPeriod, error) {
	periods, err := uc.repo.ListByBranch(ctx, tenantID, branchID)
	if err != nil {
		return nil, domainerrors.Internal(err)
	}
	return periods, nil
}
