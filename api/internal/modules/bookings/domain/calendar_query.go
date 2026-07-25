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

type ClosureWarning struct {
	Date   time.Time     `json:"date"`
	Reason ClosureReason `json:"reason"`
}

type CalendarQuery interface {
	CheckDate(ctx context.Context, tenantID, branchID uuid.UUID, date time.Time, isTermTime bool) (isClosed bool, reason ClosureReason, err error)
}
