package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChildLeavingRecord struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	BranchID   uuid.UUID
	ChildID    uuid.UUID
	LeftAt     time.Time
	ReasonCode string
	ReasonNote *string
	CreatedAt  time.Time
}
