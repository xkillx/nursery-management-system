package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	Create(ctx context.Context, booking HourlyBooking) error
	ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) ([]HourlyBooking, error)
	ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool, limit, offset int) ([]HourlyBooking, error)
	CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) (int, error)
	ListActiveByChildAndDateRange(ctx context.Context, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]HourlyBooking, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (HourlyBooking, error)
	Cancel(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
}
