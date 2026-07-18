package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CreateBookingParams struct {
	ChildID             uuid.UUID
	SessionTemplateID   *uuid.UUID
	RoomID              uuid.UUID
	DaysOfWeek          []int32
	EffectiveStartDate  time.Time
	EffectiveEndDate    *time.Time
	FundingType         *string
	FundingHoursPerWeek *float64
	LaReference         *string
	SessionEntries      []domain.SessionEntry
}

type CreateBooking struct {
	repo domain.Repository
}

func NewCreateBooking(repo domain.Repository) *CreateBooking {
	return &CreateBooking{repo: repo}
}

func (uc *CreateBooking) Execute(ctx context.Context, actor BookingActor, siteID uuid.UUID, params CreateBookingParams) (domain.Booking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Booking{}, err
	}

	if params.ChildID == uuid.Nil {
		return domain.Booking{}, domainerrors.Validation("Child is required.", "child_id")
	}
	if params.SessionTemplateID == nil && len(params.SessionEntries) == 0 {
		return domain.Booking{}, domainerrors.Validation("Either session template or session entries is required.", "session_entries")
	}
	if params.RoomID == uuid.Nil {
		return domain.Booking{}, domainerrors.Validation("Room is required.", "room_id")
	}
	if params.SessionTemplateID != nil && !domain.ValidDaysOfWeek(params.DaysOfWeek) {
		return domain.Booking{}, domain.ErrInvalidDaysOfWeek
	}
	if params.EffectiveEndDate != nil && params.EffectiveEndDate.Before(params.EffectiveStartDate) {
		return domain.Booking{}, domain.ErrInvalidDateRange
	}
	if params.FundingType != nil && !domain.ValidFundingType(*params.FundingType) {
		return domain.Booking{}, domain.ErrInvalidFundingType
	}

	booking := domain.Booking{
		ID:                   uuid.New(),
		TenantID:             actor.TenantID(),
		BranchID:             siteID,
		ChildID:              params.ChildID,
		SessionTemplateID:    params.SessionTemplateID,
		RoomID:               params.RoomID,
		DaysOfWeek:           params.DaysOfWeek,
		EffectiveStartDate:   params.EffectiveStartDate,
		EffectiveEndDate:     params.EffectiveEndDate,
		FundingType:          params.FundingType,
		FundingHoursPerWeek:  params.FundingHoursPerWeek,
		LaReference:          params.LaReference,
		SessionEntries:       params.SessionEntries,
		Status:               domain.StatusActive,
		BookedByMembershipID: actor.MembershipID(),
	}

	if err := uc.repo.Create(ctx, booking); err != nil {
		return domain.Booking{}, internalError(err)
	}

	return booking, nil
}
