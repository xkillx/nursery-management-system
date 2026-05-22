package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParentMapping struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	BranchID     uuid.UUID
	MembershipID uuid.UUID
	GuardianID   uuid.UUID
	EndedAt      *time.Time
	EndedReasonCode *string
	EndedReasonNote *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
