package domain

import (
	"time"

	"github.com/google/uuid"
)

type HolidayPeriodType string

const (
	HolidayPeriodTypeHalfTerm  HolidayPeriodType = "half_term"
	HolidayPeriodTypeChristmas HolidayPeriodType = "christmas"
	HolidayPeriodTypeEaster    HolidayPeriodType = "easter"
	HolidayPeriodTypeSummer    HolidayPeriodType = "summer"
	HolidayPeriodTypeOther     HolidayPeriodType = "other"
)

func ValidHolidayPeriodType(t string) bool {
	switch HolidayPeriodType(t) {
	case HolidayPeriodTypeHalfTerm, HolidayPeriodTypeChristmas, HolidayPeriodTypeEaster, HolidayPeriodTypeSummer, HolidayPeriodTypeOther:
		return true
	}
	return false
}

type HolidayPeriod struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Type      HolidayPeriodType
	CreatedAt time.Time
	UpdatedAt time.Time
}
