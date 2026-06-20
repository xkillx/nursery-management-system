package domain

import (
	"time"

	"github.com/google/uuid"
)

type SessionTemplate struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	BranchID    uuid.UUID
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Entries     []SessionTemplateEntry
}

type SessionTemplateEntry struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	BranchID      uuid.UUID
	TemplateID    uuid.UUID
	DayOfWeek     int
	SessionTypeID uuid.UUID
	SessionType   *EntrySessionType
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// EntrySessionType is the joined read-side view of a session type embedded
// in a session template entry. It mirrors the booking-pattern projection
// so the wizard can render a template using the same UI primitives.
type EntrySessionType struct {
	ID           uuid.UUID
	Name         string
	StartMinutes int
	EndMinutes   int
	IsActive     bool
}
