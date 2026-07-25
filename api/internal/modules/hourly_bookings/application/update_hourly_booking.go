package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/hourly_bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateHourlyBookingParams struct {
	CalendarDate     *time.Time
	StartTimeMinutes *int
	DurationMinutes  *int
	SessionTypeID    *uuid.UUID
}

type UpdateHourlyBooking struct {
	repo domain.Repository
	txm  TxManager
}

func NewUpdateHourlyBooking(repo domain.Repository, txm TxManager) *UpdateHourlyBooking {
	return &UpdateHourlyBooking{repo: repo, txm: txm}
}

func (uc *UpdateHourlyBooking) Execute(ctx context.Context, actor HourlyBookingActor, siteID, bookingID uuid.UUID, params UpdateHourlyBookingParams) (domain.HourlyBooking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.HourlyBooking{}, err
	}

	if params.DurationMinutes != nil && *params.DurationMinutes <= 0 {
		return domain.HourlyBooking{}, domain.ErrInvalidDuration
	}
	if params.StartTimeMinutes != nil && (*params.StartTimeMinutes < 0 || *params.StartTimeMinutes > 1439) {
		return domain.HourlyBooking{}, domain.ErrInvalidStartTime
	}

	var result domain.HourlyBooking
	err := uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		booking, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != domain.StatusActive {
			return domainerrors.Conflict("booking_not_active", "Cannot update a booking that is not active.")
		}

		if err := uc.repo.Update(ctx, tx, actor.TenantID(), siteID, bookingID, params.CalendarDate, params.StartTimeMinutes, params.DurationMinutes, params.SessionTypeID); err != nil {
			return internalError(err)
		}

		if params.CalendarDate != nil {
			booking.CalendarDate = *params.CalendarDate
		}
		if params.StartTimeMinutes != nil {
			booking.StartTimeMinutes = *params.StartTimeMinutes
		}
		if params.DurationMinutes != nil {
			booking.DurationMinutes = *params.DurationMinutes
		}
		if params.SessionTypeID != nil {
			booking.SessionTypeID = params.SessionTypeID
		}

		result = booking
		return nil
	})

	return result, err
}
