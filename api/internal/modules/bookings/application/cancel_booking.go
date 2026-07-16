package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CancelBooking struct {
	repo domain.Repository
	txm  TxManager
}

func NewCancelBooking(repo domain.Repository, txm TxManager) *CancelBooking {
	return &CancelBooking{repo: repo, txm: txm}
}

func (uc *CancelBooking) Execute(ctx context.Context, actor BookingActor, siteID, bookingID uuid.UUID) error {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return err
	}

	return uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		booking, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, bookingID)
		if err != nil {
			return err
		}
		if booking.Status == domain.StatusCancelled {
			return domainerrors.Conflict("booking_already_cancelled", "Booking is already cancelled.")
		}
		return uc.repo.Cancel(ctx, tx, actor.TenantID(), siteID, bookingID)
	})
}
