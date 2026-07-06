package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/ad_hoc_bookings/domain"
)

type ListAdHocBookings struct {
	repo domain.Repository
}

func NewListAdHocBookings(repo domain.Repository) *ListAdHocBookings {
	return &ListAdHocBookings{repo: repo}
}

func (uc *ListAdHocBookings) Execute(ctx context.Context, actor AdHocBookingActor, siteID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) ([]domain.AdHocBooking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	return uc.repo.ListByBranch(ctx, actor.TenantID(), siteID, childID, from, to, activeOnly)
}

func (uc *ListAdHocBookings) ExecutePaginated(ctx context.Context, actor AdHocBookingActor, siteID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool, limit, offset int) ([]domain.AdHocBooking, int, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, 0, err
	}

	bookings, err := uc.repo.ListByBranchPaginated(ctx, actor.TenantID(), siteID, childID, from, to, activeOnly, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.repo.CountByBranch(ctx, actor.TenantID(), siteID, childID, from, to, activeOnly)
	if err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}
