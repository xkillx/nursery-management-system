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
	EffectiveStartDate  time.Time
	EffectiveEndDate    *time.Time
	FundingType         *string
	FundingHoursPerWeek *float64
	LaReference         *string
	SessionEntries      []domain.SessionEntry
	TermTimeOnly        bool
}

type CreateBooking struct {
	repo          domain.Repository
	fundingLookup domain.FundingLookup
}

func NewCreateBooking(repo domain.Repository, fundingLookup domain.FundingLookup) *CreateBooking {
	return &CreateBooking{repo: repo, fundingLookup: fundingLookup}
}

func (uc *CreateBooking) Execute(ctx context.Context, actor BookingActor, siteID uuid.UUID, params CreateBookingParams) (domain.Booking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Booking{}, err
	}

	if params.ChildID == uuid.Nil {
		return domain.Booking{}, domainerrors.Validation("Child is required.", "child_id")
	}
	if len(params.SessionEntries) == 0 {
		return domain.Booking{}, domainerrors.Validation("Session entries are required.", "session_entries")
	}
	if params.EffectiveEndDate != nil && params.EffectiveEndDate.Before(params.EffectiveStartDate) {
		return domain.Booking{}, domain.ErrInvalidDateRange
	}
	if params.FundingType != nil && !domain.ValidFundingType(*params.FundingType) {
		return domain.Booking{}, domain.ErrInvalidFundingType
	}

	fundingType := params.FundingType
	fundingHours := params.FundingHoursPerWeek
	laReference := params.LaReference
	termTimeOnly := params.TermTimeOnly

	if fundingType == nil {
		fi, err := uc.fundingLookup.GetChildFunding(ctx, actor.TenantID(), siteID, params.ChildID)
		if err != nil {
			return domain.Booking{}, internalError(err)
		}
		if fi.HasFunding {
			fundingType = &fi.FundingType
			fundingHours = fi.FundedHoursPerWeek
			laReference = fi.LaReference
			termTimeOnly = fi.TermTimeOnly
		} else {
			none := "none"
			fundingType = &none
		}
	}

	booking := domain.Booking{
		ID:                   uuid.New(),
		TenantID:             actor.TenantID(),
		BranchID:             siteID,
		ChildID:              params.ChildID,
		EffectiveStartDate:   params.EffectiveStartDate,
		EffectiveEndDate:     params.EffectiveEndDate,
		FundingType:          fundingType,
		FundingHoursPerWeek:  fundingHours,
		LaReference:          laReference,
		SessionEntries:       params.SessionEntries,
		TermTimeOnly:         termTimeOnly,
		Status:               domain.StatusActive,
		BookedByMembershipID: actor.MembershipID(),
	}

	if err := uc.repo.Create(ctx, booking); err != nil {
		return domain.Booking{}, internalError(err)
	}

	return booking, nil
}
