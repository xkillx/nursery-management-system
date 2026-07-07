package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/hourly_bookings/domain"
)

type ListHourlyBookings struct {
	repo domain.Repository
}

func NewListHourlyBookings(repo domain.Repository) *ListHourlyBookings {
	return &ListHourlyBookings{repo: repo}
}

func (uc *ListHourlyBookings) Execute(ctx context.Context, actor HourlyBookingActor, siteID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) ([]domain.HourlyBooking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	return uc.repo.ListByBranch(ctx, actor.TenantID(), siteID, childID, from, to, activeOnly)
}

func (uc *ListHourlyBookings) ExecutePaginated(ctx context.Context, actor HourlyBookingActor, siteID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool, limit, offset int) ([]domain.HourlyBooking, int, error) {
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
