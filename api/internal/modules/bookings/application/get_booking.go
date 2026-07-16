package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
)

type GetBooking struct {
	repo domain.Repository
}

func NewGetBooking(repo domain.Repository) *GetBooking {
	return &GetBooking{repo: repo}
}

func (uc *GetBooking) Execute(ctx context.Context, actor BookingActor, siteID, bookingID uuid.UUID) (domain.Booking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Booking{}, err
	}

	return uc.repo.GetByID(ctx, actor.TenantID(), siteID, bookingID)
}
