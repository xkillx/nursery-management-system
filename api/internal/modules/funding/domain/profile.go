package domain

import (
	"time"

	"github.com/google/uuid"
)

type FundingProfile struct {
	ID                     uuid.UUID
	TenantID               uuid.UUID
	BranchID               uuid.UUID
	ChildID                uuid.UUID
	BillingMonth           time.Time
	FundedAllowanceMinutes int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type ChildEnrollment struct {
	ID        uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
}
