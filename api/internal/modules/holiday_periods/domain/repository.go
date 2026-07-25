package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, hp HolidayPeriod) error
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (HolidayPeriod, bool, error)
	ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID) ([]HolidayPeriod, error)
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	Delete(ctx context.Context, tenantID, branchID, id uuid.UUID) error
	ListForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, monthStart, monthEnd time.Time) ([]HolidayPeriod, error)
}
