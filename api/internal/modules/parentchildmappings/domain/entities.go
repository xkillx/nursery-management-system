package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParentChildMapping struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	BranchID        uuid.UUID
	MembershipID    uuid.UUID
	ChildID         uuid.UUID
	EndedAt         *time.Time
	EndedReasonCode *string
	EndedReasonNote *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
