package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ClosureReason string

const (
	ClosureReasonNone          ClosureReason = "none"
	ClosureReasonClosureDay    ClosureReason = "closure_day"
	ClosureReasonHolidayPeriod ClosureReason = "holiday_period"
)

type CalendarDayInfo struct {
	Date   time.Time
	IsOpen bool
	Reason ClosureReason
}

type ClosureWarning struct {
	Date   time.Time
	Reason ClosureReason
}

type HolidayPeriodDateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

type DateRangeResult struct {
	Date   time.Time
	IsOpen bool
	Reason ClosureReason
}

type CalendarQuery interface {
	CheckDate(ctx context.Context, tenantID, branchID uuid.UUID, date time.Time, isTermTime bool) (*CalendarDayInfo, error)
	CheckDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time, isTermTime bool) ([]DateRangeResult, error)
}
