package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BillingRepository interface {
	ListPreflightChildren(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth, nextBillingMonth time.Time) ([]PreflightChildRow, error)
	ListPreflightAttendanceSessions(ctx context.Context, tenantID, branchID uuid.UUID, periodStartLocalDate, periodEndExclusiveLocalDate time.Time) ([]PreflightAttendanceSessionRow, error)
}
