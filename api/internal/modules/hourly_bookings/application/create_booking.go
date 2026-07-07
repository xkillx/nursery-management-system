package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/hourly_bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

const maxDurationMinutes = 600

type CreateHourlyBookingParams struct {
	ChildID          uuid.UUID
	CalendarDate     time.Time
	StartTimeMinutes int
	DurationMinutes  int
	SessionTypeID    *uuid.UUID
}

type CreateHourlyBooking struct {
	repo domain.Repository
}

func NewCreateHourlyBooking(repo domain.Repository) *CreateHourlyBooking {
	return &CreateHourlyBooking{repo: repo}
}

func (uc *CreateHourlyBooking) Execute(ctx context.Context, actor HourlyBookingActor, siteID uuid.UUID, params CreateHourlyBookingParams) (domain.HourlyBooking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.HourlyBooking{}, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if params.CalendarDate.Before(today) {
		return domain.HourlyBooking{}, domainerrors.Validation("Calendar date must be today or in the future.", "calendar_date")
	}

	if params.ChildID == uuid.Nil {
		return domain.HourlyBooking{}, domainerrors.Validation("Child is required.", "child_id")
	}

	if params.DurationMinutes <= 0 {
		return domain.HourlyBooking{}, domainerrors.Validation("Duration must be positive.", "duration_minutes")
	}

	if params.DurationMinutes > maxDurationMinutes {
		return domain.HourlyBooking{}, domainerrors.Validation("Duration cannot exceed 10 hours.", "duration_minutes")
	}

	if params.StartTimeMinutes < 0 || params.StartTimeMinutes > 1439 {
		return domain.HourlyBooking{}, domainerrors.Validation("Start time must be between 0 and 1439.", "start_time_minutes")
	}

	booking := domain.HourlyBooking{
		ID:                   uuid.New(),
		TenantID:             actor.TenantID(),
		BranchID:             siteID,
		ChildID:              params.ChildID,
		CalendarDate:         params.CalendarDate,
		StartTimeMinutes:     params.StartTimeMinutes,
		DurationMinutes:      params.DurationMinutes,
		SessionTypeID:        params.SessionTypeID,
		BookedByMembershipID: actor.MembershipID(),
		Status:               domain.StatusActive,
	}

	if err := uc.repo.Create(ctx, booking); err != nil {
		return domain.HourlyBooking{}, internalError(err)
	}

	return booking, nil
}
