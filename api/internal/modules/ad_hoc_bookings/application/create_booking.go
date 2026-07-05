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
	repo domain.Repository
}

func NewCreateAdHocBooking(repo domain.Repository) *CreateAdHocBooking {
	return &CreateAdHocBooking{repo: repo}
}

func (uc *CreateAdHocBooking) Execute(ctx context.Context, actor AdHocBookingActor, siteID uuid.UUID, params CreateAdHocBookingParams) (domain.AdHocBooking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.AdHocBooking{}, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if params.CalendarDate.Before(today) {
		return domain.AdHocBooking{}, domainerrors.Validation("Calendar date must be today or in the future.", "calendar_date")
	}

	if params.SessionTypeID == uuid.Nil {
		return domain.AdHocBooking{}, domainerrors.Validation("Session type is required.", "session_type_id")
	}

	if params.ChildID == uuid.Nil {
		return domain.AdHocBooking{}, domainerrors.Validation("Child is required.", "child_id")
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

	if err := uc.repo.Create(ctx, booking); err != nil {
		return domain.AdHocBooking{}, internalError(err)
	}

	return booking, nil
}
