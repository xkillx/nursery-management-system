package postgres

import (
	"time"

	"github.com/google/uuid"
)

type PreflightChildRow struct {
	ChildID               uuid.UUID
	FirstName             string
	MiddleName            *string
	LastName              *string
	DateOfBirth           time.Time
	StartDate             time.Time
	EndDate               *time.Time
	CoreHourlyRateMinor   *int
	HasParentCarerContact bool
	ExistingInvoiceID     *uuid.UUID
	ExistingInvoiceStatus *string
}

type PreflightAttendanceSessionRow struct {
	SessionID         uuid.UUID
	ChildID           uuid.UUID
	Status            string
	CheckInAt         time.Time
	CheckOutAt        *time.Time
	CheckInLocalDate  time.Time
	CheckOutLocalDate *time.Time
}
