package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
)

type UnifiedBookingRow struct {
	BookingType       string
	ID                uuid.UUID
	TenantID          uuid.UUID
	BranchID          uuid.UUID
	ChildID           uuid.UUID
	StartDate         string
	EndDate           string
	RoomID            *uuid.UUID
	SessionTemplateID uuid.UUID
	Status            string
	CreatedAt         string
	UpdatedAt         string
	ChildFirstName    string
	ChildLastName     string
	RoomName          *string
}

type ListBookings struct {
	repo domain.Repository
}

func NewListBookings(repo domain.Repository) *ListBookings {
	return &ListBookings{repo: repo}
}

func (uc *ListBookings) ExecutePaginated(ctx context.Context, actor BookingActor, siteID uuid.UUID, filters domain.ListFilters, limit, offset int) ([]domain.Booking, int, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, 0, err
	}

	bookings, err := uc.repo.ListByBranchPaginated(ctx, actor.TenantID(), siteID, filters, limit, offset)
	if err != nil {
		return nil, 0, internalError(err)
	}

	total, err := uc.repo.CountByBranch(ctx, actor.TenantID(), siteID, filters)
	if err != nil {
		return nil, 0, internalError(err)
	}

	return bookings, total, nil
}

func (uc *ListBookings) ExecuteUnified(ctx context.Context, actor BookingActor, siteID uuid.UUID, filters domain.ListFilters, limit, offset int) ([]domain.UnifiedBookingRow, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	rows, err := uc.repo.ListUnifiedByBranchPaginated(ctx, actor.TenantID(), siteID, filters, limit, offset)
	if err != nil {
		return nil, internalError(err)
	}

	return rows, nil
}
