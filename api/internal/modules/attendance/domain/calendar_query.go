package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ClosureWarning struct {
	Date   time.Time `json:"date"`
	Reason string    `json:"reason"`
}

type CalendarQuery interface {
	CheckDate(ctx context.Context, tenantID, branchID uuid.UUID, date time.Time, isTermTime bool) (isClosed bool, reason string, err error)
}
