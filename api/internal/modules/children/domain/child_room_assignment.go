package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChildRoomAssignment struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	ChildID   uuid.UUID
	RoomID    uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
	IsCurrent bool
	CreatedAt time.Time
}
