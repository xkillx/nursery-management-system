package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ClosureDayLookup interface {
	GetClosureDatesForBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]time.Time, error)
}

type HolidayPeriodLookup interface {
	GetHolidayPeriodsForBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]HolidayPeriodDateRange, error)
}
