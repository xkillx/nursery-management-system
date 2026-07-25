package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/holiday_periods/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateHolidayPeriod struct {
	repo domain.Repository
}

func NewUpdateHolidayPeriod(repo domain.Repository) *UpdateHolidayPeriod {
	return &UpdateHolidayPeriod{repo: repo}
}

type UpdateHolidayPeriodParams struct {
	Name      *string
	Type      *string
	StartDate *time.Time
	EndDate   *time.Time
}

func (uc *UpdateHolidayPeriod) Execute(ctx context.Context, tenantID, branchID, id uuid.UUID, params UpdateHolidayPeriodParams) (domain.HolidayPeriod, error) {
	existing, found, err := uc.repo.GetByID(ctx, tenantID, branchID, id)
	if err != nil {
		return domain.HolidayPeriod{}, domainerrors.Internal(err)
	}
	if !found {
		return domain.HolidayPeriod{}, domain.ErrHolidayPeriodNotFound
	}

	fields := make(map[string]any)
	if params.Name != nil {
		if *params.Name == "" {
			return domain.HolidayPeriod{}, domain.ErrNameRequired
		}
		fields["name"] = *params.Name
	}
	if params.Type != nil {
		if !domain.ValidHolidayPeriodType(*params.Type) {
			return domain.HolidayPeriod{}, domain.ErrInvalidType
		}
		fields["type"] = *params.Type
	}
	if params.StartDate != nil {
		start := time.Date(params.StartDate.Year(), params.StartDate.Month(), params.StartDate.Day(), 0, 0, 0, 0, time.UTC)
		fields["start_date"] = start
	}
	if params.EndDate != nil {
		end := time.Date(params.EndDate.Year(), params.EndDate.Month(), params.EndDate.Day(), 0, 0, 0, 0, time.UTC)
		fields["end_date"] = end
	}

	if len(fields) == 0 {
		return existing, nil
	}

	_, err = uc.repo.Update(ctx, tenantID, branchID, id, fields)
	if err != nil {
		return domain.HolidayPeriod{}, domainerrors.Internal(err)
	}

	updated, _, err := uc.repo.GetByID(ctx, tenantID, branchID, id)
	if err != nil {
		return domain.HolidayPeriod{}, domainerrors.Internal(err)
	}

	return updated, nil
}
