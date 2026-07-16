package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CloneBookingParams struct {
	ChildID *uuid.UUID
}

type CloneBooking struct {
	repo domain.Repository
}

func NewCloneBooking(repo domain.Repository) *CloneBooking {
	return &CloneBooking{repo: repo}
}

func (uc *CloneBooking) Execute(ctx context.Context, actor BookingActor, siteID, bookingID uuid.UUID, params CloneBookingParams) (domain.Booking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Booking{}, err
	}

	source, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, bookingID)
	if err != nil {
		return domain.Booking{}, err
	}

	if source.Status == domain.StatusCancelled {
		return domain.Booking{}, domainerrors.Conflict("booking_cancelled", "Cannot clone a cancelled booking.")
	}

	childID := source.ChildID
	if params.ChildID != nil {
		childID = *params.ChildID
	}

	cloned := domain.Booking{
		ID:                   uuid.New(),
		TenantID:             actor.TenantID(),
		BranchID:             siteID,
		ChildID:              childID,
		SessionTemplateID:    source.SessionTemplateID,
		RoomID:               source.RoomID,
		DaysOfWeek:           source.DaysOfWeek,
		EffectiveStartDate:   source.EffectiveStartDate,
		EffectiveEndDate:     source.EffectiveEndDate,
		FundingType:          source.FundingType,
		FundingHoursPerWeek:  source.FundingHoursPerWeek,
		LaReference:          source.LaReference,
		Status:               domain.StatusActive,
		BookedByMembershipID: actor.MembershipID(),
	}

	if err := uc.repo.Create(ctx, cloned); err != nil {
		return domain.Booking{}, internalError(err)
	}

	return cloned, nil
}
