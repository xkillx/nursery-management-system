package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateBookingParams struct {
	DaysOfWeek          []int32
	EffectiveStartDate  *time.Time
	EffectiveEndDate    *time.Time
	FundingType         *string
	FundingHoursPerWeek *float64
	LaReference         *string
	TermTimeOnly        *bool
}

type UpdateBooking struct {
	repo domain.Repository
	txm  TxManager
}

func NewUpdateBooking(repo domain.Repository, txm TxManager) *UpdateBooking {
	return &UpdateBooking{repo: repo, txm: txm}
}

func (uc *UpdateBooking) Execute(ctx context.Context, actor BookingActor, siteID, bookingID uuid.UUID, params UpdateBookingParams) (domain.Booking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Booking{}, err
	}

	var result domain.Booking
	err := uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		booking, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, bookingID)
		if err != nil {
			return err
		}

		if booking.Status == domain.StatusCancelled {
			return domainerrors.Conflict("booking_cancelled", "Cannot update a cancelled booking.")
		}

		if params.DaysOfWeek != nil {
			if !domain.ValidDaysOfWeek(params.DaysOfWeek) {
				return domain.ErrInvalidDaysOfWeek
			}
			booking.DaysOfWeek = params.DaysOfWeek
		}
		if params.EffectiveStartDate != nil {
			booking.EffectiveStartDate = *params.EffectiveStartDate
		}
		if params.EffectiveEndDate != nil {
			booking.EffectiveEndDate = params.EffectiveEndDate
		}
		if booking.EffectiveEndDate != nil && booking.EffectiveEndDate.Before(booking.EffectiveStartDate) {
			return domain.ErrInvalidDateRange
		}
		if params.FundingType != nil {
			if !domain.ValidFundingType(*params.FundingType) {
				return domain.ErrInvalidFundingType
			}
			booking.FundingType = params.FundingType
		}
		if params.FundingHoursPerWeek != nil {
			booking.FundingHoursPerWeek = params.FundingHoursPerWeek
		}
		if params.LaReference != nil {
			booking.LaReference = params.LaReference
		}
		if params.TermTimeOnly != nil {
			booking.TermTimeOnly = *params.TermTimeOnly
		}

		if err := uc.repo.Update(ctx, tx, booking); err != nil {
			return internalError(err)
		}

		result = booking
		return nil
	})

	return result, err
}
