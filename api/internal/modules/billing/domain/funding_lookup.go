package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FundedChildInfo struct {
	HasFunding             bool
	FundingType            string
	FundedAllowanceMinutes int
	FundedHourlyRateMinor  int
	FundedHoursPerWeek     float64
}

type FundingLookup interface {
	GetChildFunding(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (FundedChildInfo, error)
}
