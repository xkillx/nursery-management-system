package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	Create(ctx context.Context, booking Booking) error
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (Booking, error)
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (Booking, error)
	ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, filters ListFilters, limit, offset int) ([]Booking, error)
	CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, filters ListFilters) (int, error)
	Update(ctx context.Context, tx Tx, booking Booking) error
	Cancel(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	Pause(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	ListByChildAndDateRange(ctx context.Context, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]Booking, error)
	ListUnifiedByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, filters ListFilters, limit, offset int) ([]UnifiedBookingRow, error)
}

type UnifiedBookingRow struct {
	BookingType       string
	ID                uuid.UUID
	TenantID          uuid.UUID
	BranchID          uuid.UUID
	ChildID           uuid.UUID
	StartDate         time.Time
	EndDate           *time.Time
	RoomID            *uuid.UUID
	SessionTemplateID *uuid.UUID
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ChildFirstName    string
	ChildLastName     string
	RoomName          *string
}

type ListFilters struct {
	ChildID       *uuid.UUID
	RoomID        *uuid.UUID
	SessionTypeID *uuid.UUID
	Status        *string
	FundingType   *string
	Search        *string
	From          *time.Time
	To            *time.Time
	ActiveOnly    bool
}
