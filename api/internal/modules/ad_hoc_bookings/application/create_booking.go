package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/ad_hoc_bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CreateAdHocBookingParams struct {
	ChildID       uuid.UUID
	CalendarDate  time.Time
	SessionTypeID uuid.UUID
}

type CreateAdHocBooking struct {
	repo          domain.Repository
	calendarQuery domain.CalendarQuery
	fundingLookup domain.ChildFundingLookup
}

func NewCreateAdHocBooking(repo domain.Repository, calendarQuery domain.CalendarQuery, fundingLookup domain.ChildFundingLookup) *CreateAdHocBooking {
	return &CreateAdHocBooking{repo: repo, calendarQuery: calendarQuery, fundingLookup: fundingLookup}
}

type CreateAdHocBookingResult struct {
	Booking  domain.AdHocBooking
	Warnings []domain.ClosureWarning
}

func (uc *CreateAdHocBooking) Execute(ctx context.Context, actor AdHocBookingActor, siteID uuid.UUID, params CreateAdHocBookingParams) (CreateAdHocBookingResult, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return CreateAdHocBookingResult{}, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if params.CalendarDate.Before(today) {
		return CreateAdHocBookingResult{}, domainerrors.Validation("Calendar date must be today or in the future.", "calendar_date")
	}

	if params.SessionTypeID == uuid.Nil {
		return CreateAdHocBookingResult{}, domainerrors.Validation("Session type is required.", "session_type_id")
	}

	if params.ChildID == uuid.Nil {
		return CreateAdHocBookingResult{}, domainerrors.Validation("Child is required.", "child_id")
	}

	booking := domain.AdHocBooking{
		ID:                   uuid.New(),
		TenantID:             actor.TenantID(),
		BranchID:             siteID,
		ChildID:              params.ChildID,
		CalendarDate:         params.CalendarDate,
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
		return CreateAdHocBookingResult{}, internalError(err)
	}

	return CreateAdHocBookingResult{Booking: booking, Warnings: warnings}, nil
}
