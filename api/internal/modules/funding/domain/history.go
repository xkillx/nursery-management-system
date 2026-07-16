package domain

import (
	"time"

	"github.com/google/uuid"
)

type FundingHistory struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	BranchID           uuid.UUID
	ChildID            uuid.UUID
	FundingType        *string
	FundingModel       *string
	FundedHoursPerWeek *float64
	FundingStartDate   *time.Time
	FundingEndDate     *time.Time
	ChangedAt          time.Time
	ChangedByUserID    uuid.UUID
}
