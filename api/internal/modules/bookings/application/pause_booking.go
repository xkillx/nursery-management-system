package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/bookings/domain"
)

type PauseBooking struct {
	repo domain.Repository
	txm  TxManager
}

func NewPauseBooking(repo domain.Repository, txm TxManager) *PauseBooking {
	return &PauseBooking{repo: repo, txm: txm}
}

func (uc *PauseBooking) Execute(ctx context.Context, actor BookingActor, siteID, bookingID uuid.UUID) (domain.Booking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Booking{}, err
	}

	var result domain.Booking
	err := uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		booking, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, bookingID)
		if err != nil {
			return err
		}
		if booking.Status == domain.StatusPaused {
			return domain.ErrBookingAlreadyPaused
		}
		if booking.Status == domain.StatusCancelled {
			return domain.ErrBookingAlreadyCancelled
		}
		if err := uc.repo.Pause(ctx, tx, actor.TenantID(), siteID, bookingID); err != nil {
			return internalError(err)
		}
		booking.Status = domain.StatusPaused
		result = booking
		return nil
	})

	return result, err
}
