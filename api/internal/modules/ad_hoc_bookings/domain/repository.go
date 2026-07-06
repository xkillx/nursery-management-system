package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) ([]AdHocBooking, error)
	ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool, limit, offset int) ([]AdHocBooking, error)
	CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) (int, error)
	ListActiveByChildAndDateRange(ctx context.Context, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]AdHocBooking, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (AdHocBooking, error)
	Create(ctx context.Context, booking AdHocBooking) error
	Cancel(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (AdHocBooking, error)
}
