package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusCancelled = "cancelled"
)

type SessionEntry struct {
	DayOfWeek     int32     `json:"day_of_week"`
	SessionTypeID uuid.UUID `json:"session_type_id"`
}

type Booking struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	BranchID             uuid.UUID
	ChildID              uuid.UUID
	EffectiveStartDate   time.Time
	EffectiveEndDate     *time.Time
	FundingType          *string
	FundingHoursPerWeek  *float64
	LaReference          *string
	SessionEntries       []SessionEntry
	TermTimeOnly         bool
	Status               string
	BookedByMembershipID uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func ValidBookingStatus(s string) bool {
	switch s {
	case StatusActive, StatusPaused, StatusCancelled:
		return true
	}
	return false
}

func ValidFundingType(ft string) bool {
	switch ft {
	case "none", "universal_15", "working_parent", "working_parent_under_3", "disadvantaged_2yo":
		return true
	}
	return false
}
