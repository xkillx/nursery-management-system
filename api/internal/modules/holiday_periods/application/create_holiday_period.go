package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/holiday_periods/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CreateHolidayPeriod struct {
	repo domain.Repository
}

func NewCreateHolidayPeriod(repo domain.Repository) *CreateHolidayPeriod {
	return &CreateHolidayPeriod{repo: repo}
}

type CreateHolidayPeriodParams struct {
	Name      string
	Type      string
	StartDate time.Time
	EndDate   time.Time
}

func (uc *CreateHolidayPeriod) Execute(ctx context.Context, tenantID, branchID uuid.UUID, params CreateHolidayPeriodParams) (domain.HolidayPeriod, error) {
	if params.Name == "" {
		return domain.HolidayPeriod{}, domain.ErrNameRequired
	}
	if params.Type == "" {
		return domain.HolidayPeriod{}, domain.ErrTypeRequired
	}
	if !domain.ValidHolidayPeriodType(params.Type) {
		return domain.HolidayPeriod{}, domain.ErrInvalidType
	}
	if params.StartDate.IsZero() {
		return domain.HolidayPeriod{}, domain.ErrStartDateRequired
	}
	if params.EndDate.IsZero() {
		return domain.HolidayPeriod{}, domain.ErrEndDateRequired
	}

	start := time.Date(params.StartDate.Year(), params.StartDate.Month(), params.StartDate.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(params.EndDate.Year(), params.EndDate.Month(), params.EndDate.Day(), 0, 0, 0, 0, time.UTC)

	if !end.After(start) {
		return domain.HolidayPeriod{}, domain.ErrInvalidDateRange
	}

	hp := domain.HolidayPeriod{
		ID:        uuid.New(),
		TenantID:  tenantID,
		BranchID:  branchID,
		Name:      params.Name,
		StartDate: start,
		EndDate:   end,
		Type:      domain.HolidayPeriodType(params.Type),
	}

	if err := uc.repo.Create(ctx, hp); err != nil {
		return domain.HolidayPeriod{}, domainerrors.Internal(err)
	}

	return hp, nil
}
