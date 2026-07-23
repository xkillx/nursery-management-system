package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParentChild struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	BranchID        uuid.UUID
	ParentID        uuid.UUID
	ChildID         uuid.UUID
	EndedAt         *time.Time
	EndedReasonCode *string
	EndedReasonNote *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
