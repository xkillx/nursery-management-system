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
	SessionTemplateID    *uuid.UUID
	RoomID               uuid.UUID
	DaysOfWeek           []int32
	EffectiveStartDate   time.Time
	EffectiveEndDate     *time.Time
	FundingType          *string
	FundingHoursPerWeek  *float64
	LaReference          *string
	SessionEntries       []SessionEntry
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
	case "none", "fifteen_hours", "thirty_hours", "two_year_old", "custom":
		return true
	}
	return false
}

func ValidDaysOfWeek(days []int32) bool {
	if len(days) == 0 {
		return false
	}
	for _, d := range days {
		if d < 1 || d > 7 {
			return false
		}
	}
	return true
}
