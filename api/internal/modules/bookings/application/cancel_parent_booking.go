package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CancelParentBooking struct {
	repo      domain.Repository
	childLook ParentChildLookup
	txm       TxManager
}

func NewCancelParentBooking(repo domain.Repository, childLook ParentChildLookup, txm TxManager) *CancelParentBooking {
	return &CancelParentBooking{repo: repo, childLook: childLook, txm: txm}
}

func (uc *CancelParentBooking) Execute(ctx context.Context, actor BookingActor, siteID, bookingID uuid.UUID) error {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return err
	}

	childIDs, err := uc.childLook.ListChildIDsForParent(ctx, actor.TenantID(), siteID, actor.MembershipID())
	if err != nil {
		return internalError(err)
	}

	return uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		booking, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, bookingID)
		if err != nil {
			return err
		}

		childAllowed := false
		for _, cid := range childIDs {
			if cid == booking.ChildID {
				childAllowed = true
				break
			}
		}
		if !childAllowed {
			return domainerrors.Forbidden("child_not_in_scope", "Booking does not belong to your child.")
		}

		if booking.Status == domain.StatusCancelled {
			return domainerrors.Conflict("booking_already_cancelled", "Booking is already cancelled.")
		}

		return uc.repo.Cancel(ctx, tx, actor.TenantID(), siteID, bookingID)
	})
}
