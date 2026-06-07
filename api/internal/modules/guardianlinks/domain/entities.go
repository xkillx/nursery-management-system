package domain

import (
	"time"

	"github.com/google/uuid"
)

type GuardianChildLink struct {
	ID              uuid.UUID
	GuardianID      uuid.UUID
	ChildID         uuid.UUID
	TenantID        uuid.UUID
	BranchID        uuid.UUID
	EndedAt         *time.Time
	EndedReasonCode *string
	EndedReasonNote *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type LinkedGuardianChildLink struct {
	ID         uuid.UUID
	GuardianID uuid.UUID
	ChildID    uuid.UUID
	Guardian   LinkedGuardianSummary
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type LinkedGuardianSummary struct {
	ID       uuid.UUID
	FullName string
	Email    *string
	Phone    *string
	IsActive bool
}
