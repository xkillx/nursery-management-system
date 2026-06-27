package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingPattern struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	BranchID      uuid.UUID
	ChildID       uuid.UUID
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	IsCurrent     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Entries       []BookingPatternEntry
}

type BookingPatternEntry struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	BranchID    uuid.UUID
	PatternID   uuid.UUID
	DayOfWeek   int
	SessionType *EntrySessionType
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EntrySessionType is the joined read-side view of a session type embedded
// in a booking pattern entry. It is not a domain entity; it is the denormalised
// representation returned for read endpoints.
type EntrySessionType struct {
	ID           uuid.UUID
	Name         string
	StartMinutes int
	EndMinutes   int
	IsActive     bool
}
