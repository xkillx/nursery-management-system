package domain

import (
	"time"

	"github.com/google/uuid"
)

type AbsenceMarker struct {
	ID                    uuid.UUID
	TenantID              uuid.UUID
	BranchID              uuid.UUID
	ChildID               uuid.UUID
	LocalDate             time.Time
	MarkedAt              time.Time
	MarkedByUserID        uuid.UUID
	MarkedByMembershipID  uuid.UUID
	ClearedAt             *time.Time
	ClearedByUserID       *uuid.UUID
	ClearedByMembershipID *uuid.UUID
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
