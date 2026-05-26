package domain

import (
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	SessionStatusOpen      SessionStatus = "open"
	SessionStatusComplete  SessionStatus = "complete"
	SessionStatusCorrected SessionStatus = "corrected"
)

type EventType string

const (
	EventCheckIn    EventType = "check_in"
	EventCheckOut   EventType = "check_out"
	EventCorrection EventType = "correction"
)

type IncompleteSessionBlocker struct {
	ChildID          uuid.UUID
	ChildName        string
	SessionID        uuid.UUID
	CheckInAt        time.Time
	CheckInLocalDate time.Time
}

type Session struct {
	ID                uuid.UUID
	ChildID           uuid.UUID
	Status            SessionStatus
	CheckInAt         time.Time
	CheckOutAt        *time.Time
	CheckInLocalDate  time.Time
	CheckOutLocalDate *time.Time
	DurationMinutes   *int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
