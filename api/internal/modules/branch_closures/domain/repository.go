package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	Create(ctx context.Context, closure BranchClosureDay) error
	ListByBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]BranchClosureDay, error)
	ListByBranchAndDateRangePaginated(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time, limit, offset int) ([]BranchClosureDay, error)
	CountByBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) (int, error)
	Delete(ctx context.Context, tenantID, branchID, id uuid.UUID) error
	DateExists(ctx context.Context, tenantID, branchID uuid.UUID, date time.Time) (bool, error)
	ListClosureDatesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]time.Time, error)
}
