package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/ad_hoc_bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateAdHocBookingParams struct {
	CalendarDate  *time.Time
	SessionTypeID *uuid.UUID
}

type UpdateAdHocBooking struct {
	repo domain.Repository
	txm  TxManager
}

func NewUpdateAdHocBooking(repo domain.Repository, txm TxManager) *UpdateAdHocBooking {
	return &UpdateAdHocBooking{repo: repo, txm: txm}
}

func (uc *UpdateAdHocBooking) Execute(ctx context.Context, actor AdHocBookingActor, siteID, bookingID uuid.UUID, params UpdateAdHocBookingParams) (domain.AdHocBooking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.AdHocBooking{}, err
	}

	var result domain.AdHocBooking
	err := uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		booking, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, bookingID)
		if err != nil {
			return err
		}

		if booking.Status != domain.StatusActive {
			return domainerrors.Conflict("booking_not_active", "Cannot update a booking that is not active.")
		}

		if err := uc.repo.Update(ctx, tx, actor.TenantID(), siteID, bookingID, params.CalendarDate, params.SessionTypeID); err != nil {
			return internalError(err)
		}

		if params.CalendarDate != nil {
			booking.CalendarDate = *params.CalendarDate
		}
		if params.SessionTypeID != nil {
			booking.SessionTypeID = *params.SessionTypeID
		}

		result = booking
		return nil
	})

	return result, err
}
