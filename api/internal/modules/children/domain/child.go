package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Child struct {
	ID                      uuid.UUID
	FirstName               string
	MiddleName              *string
	LastName                *string
	DateOfBirth             time.Time
	StartDate               time.Time
	EndDate                 *time.Time
	SiteCoreHourlyRateMinor *int
	Notes                   *string
	IsActive                bool
	PrimaryRoomID           *uuid.UUID
	HasCurrentRoom          bool
	HasParentCarerContact   bool
	HasBookingPattern       bool
	ProfilePhotoPath        *string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type ReasonCode string

const (
	ReasonDuplicateRecord ReasonCode = "duplicate_record"
	ReasonEnteredInError  ReasonCode = "entered_in_error"
	ReasonLeftNursery     ReasonCode = "left_nursery"
	ReasonSafeguardingDir ReasonCode = "safeguarding_direction"
	ReasonContactUpdate   ReasonCode = "contact_update"
	ReasonAccessRevoked   ReasonCode = "access_revoked"
	ReasonOther           ReasonCode = "other"
)

var ValidReasonCodes = map[ReasonCode]struct{}{
	ReasonDuplicateRecord: {},
	ReasonEnteredInError:  {},
	ReasonLeftNursery:     {},
	ReasonSafeguardingDir: {},
	ReasonContactUpdate:   {},
	ReasonAccessRevoked:   {},
	ReasonOther:           {},
}

// NewReasonCode creates a validated ReasonCode.
// Returns an error if the raw string is empty or not a valid reason code.
func NewReasonCode(raw string) (ReasonCode, error) {
	if raw == "" {
		return "", fmt.Errorf("reason code must not be empty")
	}
	code := ReasonCode(raw)
	if _, ok := ValidReasonCodes[code]; !ok {
		return "", fmt.Errorf("invalid reason code: %s", raw)
	}
	return code, nil
}

func (c Child) MissingRequirements() []string {
	missing := make([]string, 0)
	if c.FirstName == "" {
		missing = append(missing, "first_name")
	}
	if c.DateOfBirth.IsZero() {
		missing = append(missing, "date_of_birth")
	}
	if c.StartDate.IsZero() {
		missing = append(missing, "start_date")
	}
	if !c.HasParentCarerContact {
		missing = append(missing, "parent_carer_contact")
	}
	return missing
}

func (c Child) EnrollmentComplete() bool {
	return len(c.MissingRequirements()) == 0
}

func (c *Child) Activate(startDate time.Time, hourlyRateMinor int) error {
	if c.IsActive {
		return fmt.Errorf("child is already active")
	}
	now := time.Now().UTC()
	if startDate.Before(now.Truncate(24 * time.Hour)) {
		return fmt.Errorf("start date must not be in the past")
	}
	c.StartDate = startDate
	c.SiteCoreHourlyRateMinor = &hourlyRateMinor
	c.IsActive = true
	c.EndDate = nil
	return nil
}

func (c *Child) Deactivate(reasonCode ReasonCode, deactivatedAt time.Time) error {
	if !c.IsActive {
		return fmt.Errorf("child is already inactive")
	}
	if _, ok := ValidReasonCodes[reasonCode]; !ok {
		return fmt.Errorf("invalid reason code: %s", reasonCode)
	}
	c.EndDate = &deactivatedAt
	c.IsActive = false
	return nil
}

func (c *Child) ChangeName(firstName string, lastName *string) error {
	if firstName == "" {
		return fmt.Errorf("first name must not be empty")
	}
	if c.FirstName == firstName && stringPtrEqual(c.LastName, lastName) {
		return fmt.Errorf("no change in name")
	}
	c.FirstName = firstName
	c.LastName = lastName
	return nil
}

func (c Child) IsEligibleForAttendance(localDate time.Time) bool {
	if !c.IsActive {
		return false
	}
	if localDate.Before(c.StartDate) {
		return false
	}
	if c.EndDate != nil && !localDate.Before(*c.EndDate) {
		return false
	}
	return true
}

func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

type AttendanceChild struct {
	ID                   uuid.UUID
	FirstName            string
	MiddleName           *string
	LastName             *string
	EnrollmentComplete   bool
	AttendanceState      string
	OpenSessionID        *uuid.UUID
	CheckedInAt          *time.Time
	HasIncompleteSession bool
	AbsenceMarkerID      *uuid.UUID
	AbsenceMarkedAt      *time.Time
}
