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

type IssuedInvoiceWarning struct {
	BillingMonth  string
	InvoiceID     uuid.UUID
	InvoiceNumber string
	Status        string
}

type CorrectionSessionContext struct {
	ChildID           uuid.UUID
	SelectedLocalDate time.Time
	InvoiceWarning    *IssuedInvoiceWarning
	Sessions          []Session
}

type CorrectionHistoryEvent struct {
	ID                     uuid.UUID
	EventType              EventType
	OccurredAt             time.Time
	LocalDate              time.Time
	RecordedByUserID       uuid.UUID
	RecordedByMembershipID uuid.UUID
	RecordedByLabel        *string
	ReasonCode             *string
	ReasonNote             *string
	PreviousCheckInAt      *time.Time
	PreviousCheckOutAt     *time.Time
	CorrectedCheckInAt     *time.Time
	CorrectedCheckOutAt    *time.Time
	CreatedByCorrection    bool
}

type IncompleteSessionBlocker struct {
	ChildID          uuid.UUID
	ChildFirstName   string
	ChildMiddleName  *string
	ChildLastName    *string
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
