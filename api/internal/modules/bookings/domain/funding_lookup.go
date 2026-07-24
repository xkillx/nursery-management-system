package domain

import (
	"context"

	"github.com/google/uuid"
)

type FundingInfo struct {
	HasFunding         bool
	FundingType        string
	FundedHoursPerWeek *float64
	LaReference        *string
	TermTimeOnly       bool
}

type FundingLookup interface {
	GetChildFunding(ctx context.Context, tenantID, branchID, childID uuid.UUID) (FundingInfo, error)
}
