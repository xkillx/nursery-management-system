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
	repo          domain.Repository
	calendarQuery domain.CalendarQuery
	fundingLookup domain.ChildFundingLookup
}

func NewCreateHourlyBooking(repo domain.Repository, calendarQuery domain.CalendarQuery, fundingLookup domain.ChildFundingLookup) *CreateHourlyBooking {
	return &CreateHourlyBooking{repo: repo, calendarQuery: calendarQuery, fundingLookup: fundingLookup}
}

type CreateHourlyBookingResult struct {
	Booking  domain.HourlyBooking
	Warnings []domain.ClosureWarning
}

func (uc *CreateHourlyBooking) Execute(ctx context.Context, actor HourlyBookingActor, siteID uuid.UUID, params CreateHourlyBookingParams) (CreateHourlyBookingResult, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return CreateHourlyBookingResult{}, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if params.CalendarDate.Before(today) {
		return CreateHourlyBookingResult{}, domainerrors.Validation("Calendar date must be today or in the future.", "calendar_date")
	}

	if params.ChildID == uuid.Nil {
		return CreateHourlyBookingResult{}, domainerrors.Validation("Child is required.", "child_id")
	}

	if params.DurationMinutes <= 0 {
		return CreateHourlyBookingResult{}, domainerrors.Validation("Duration must be positive.", "duration_minutes")
	}

	if params.DurationMinutes > maxDurationMinutes {
		return CreateHourlyBookingResult{}, domainerrors.Validation("Duration cannot exceed 10 hours.", "duration_minutes")
	}

	if params.StartTimeMinutes < 0 || params.StartTimeMinutes > 1439 {
		return CreateHourlyBookingResult{}, domainerrors.Validation("Start time must be between 0 and 1439.", "start_time_minutes")
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

	var warnings []domain.ClosureWarning
	if uc.calendarQuery != nil && uc.fundingLookup != nil {
		isTermTime, _ := uc.fundingLookup.GetChildTermTimeOnly(ctx, actor.TenantID(), siteID, params.ChildID)
		isClosed, reason, err := uc.calendarQuery.CheckDate(ctx, actor.TenantID(), siteID, params.CalendarDate, isTermTime)
		if err == nil && isClosed {
			warnings = append(warnings, domain.ClosureWarning{Date: params.CalendarDate, Reason: reason})
		}
	}

	if err := uc.repo.Create(ctx, booking); err != nil {
		return CreateHourlyBookingResult{}, internalError(err)
	}

	return CreateHourlyBookingResult{Booking: booking, Warnings: warnings}, nil
}
