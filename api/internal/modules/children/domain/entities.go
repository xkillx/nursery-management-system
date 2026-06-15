package domain

import (
	"strings"
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
	CoreHourlyRateMinor     *int
	SiteCoreHourlyRateMinor *int
	Notes                   *string
	IsActive                bool
	LeftAt                  *time.Time
	LeftReasonCode          *string
	LeftReasonNote          *string
	HasGuardianLink         bool
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

func (c Child) MissingRequirements() []string {
	missing := make([]string, 0)
	if strings.TrimSpace(c.FirstName) == "" {
		missing = append(missing, "first_name")
	}
	if c.DateOfBirth.IsZero() {
		missing = append(missing, "date_of_birth")
	}
	if c.StartDate.IsZero() {
		missing = append(missing, "start_date")
	}
	if !c.HasGuardianLink {
		missing = append(missing, "guardian_link")
	}
	return missing
}

func (c Child) EnrollmentComplete() bool {
	return len(c.MissingRequirements()) == 0
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
