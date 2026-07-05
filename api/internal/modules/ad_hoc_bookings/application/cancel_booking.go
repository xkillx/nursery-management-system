package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/ad_hoc_bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CancelAdHocBooking struct {
	repo domain.Repository
	txm  TxManager
}

func NewCancelAdHocBooking(repo domain.Repository, txm TxManager) *CancelAdHocBooking {
	return &CancelAdHocBooking{repo: repo, txm: txm}
}

func (uc *CancelAdHocBooking) Execute(ctx context.Context, actor AdHocBookingActor, siteID, bookingID uuid.UUID) error {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return err
	}

	return uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		booking, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, bookingID)
		if err != nil {
			return err
		}
		if booking.Status != domain.StatusActive {
			return domainerrors.Conflict("booking_not_active", "Cannot cancel a booking that is not active.")
		}
		return uc.repo.Cancel(ctx, tx, actor.TenantID(), siteID, bookingID)
	})
}
