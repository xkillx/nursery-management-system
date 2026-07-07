package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive    = "active"
	StatusCancelled = "cancelled"
)

type HourlyBooking struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	BranchID             uuid.UUID
	ChildID              uuid.UUID
	CalendarDate         time.Time
	StartTimeMinutes     int
	DurationMinutes      int
	SessionTypeID        *uuid.UUID
	BookedByMembershipID uuid.UUID
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func ValidBookingStatus(s string) bool {
	switch s {
	case StatusActive, StatusCancelled:
		return true
	}
	return false
}
