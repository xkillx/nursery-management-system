package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AttendanceSessionStatus string

const (
	AttendanceSessionStatusOpen      AttendanceSessionStatus = "open"
	AttendanceSessionStatusComplete  AttendanceSessionStatus = "complete"
	AttendanceSessionStatusCorrected AttendanceSessionStatus = "corrected"
)

type AttendanceSessionInput struct {
	SessionID  uuid.UUID
	Status     AttendanceSessionStatus
	CheckInAt  time.Time
	CheckOutAt *time.Time
}

type BillingMonth struct {
	Year  int
	Month time.Month
}

type BillingPeriod struct {
	Month             BillingMonth
	StartLocal        time.Time
	EndExclusiveLocal time.Time
	StartUTC          time.Time
	EndExclusiveUTC   time.Time
}

type AttendanceMinuteCalculation struct {
	Period                     BillingPeriod
	RawElapsedMinutes          int
	RoundedBillableMinutes     int
	IncludedSessionCount       int
	ExcludedIncompleteSessions []ExcludedIncompleteSession
	Sessions                   []SessionMinuteCalculation
}

type SessionMinuteCalculation struct {
	SessionID              uuid.UUID
	Status                 AttendanceSessionStatus
	CheckInAt              time.Time
	CheckOutAt             time.Time
	AllocationYear         int
	AllocationMonth        time.Month
	RawElapsedDuration     time.Duration
	RawElapsedMinutes      int
	RoundedBillableMinutes int
}

type ExcludedIncompleteSession struct {
	SessionID       uuid.UUID
	Status          AttendanceSessionStatus
	CheckInAt       time.Time
	AllocationYear  int
	AllocationMonth time.Month
}

const billingBlock = 15 * time.Minute

var londonLoc *time.Location

func init() {
	var err error
	londonLoc, err = time.LoadLocation("Europe/London")
	if err != nil {
		panic("billing: failed to load Europe/London location: " + err.Error())
	}
}

func NewBillingPeriod(year int, month time.Month) (BillingPeriod, error) {
	if year < 1 || month < time.January || month > time.December {
		return BillingPeriod{}, fmt.Errorf("invalid billing period: year=%d month=%d", year, month)
	}

	startLocal := time.Date(year, month, 1, 0, 0, 0, 0, londonLoc)

	nextMonth := month + 1
	nextYear := year
	if nextMonth > time.December {
		nextMonth = time.January
		nextYear++
	}
	endExclusiveLocal := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, londonLoc)

	return BillingPeriod{
		Month: BillingMonth{
			Year:  year,
			Month: month,
		},
		StartLocal:        startLocal,
		EndExclusiveLocal: endExclusiveLocal,
		StartUTC:          startLocal.UTC(),
		EndExclusiveUTC:   endExclusiveLocal.UTC(),
	}, nil
}

func CalculateAttendanceMinutes(year int, month time.Month, sessions []AttendanceSessionInput) (AttendanceMinuteCalculation, error) {
	period, err := NewBillingPeriod(year, month)
	if err != nil {
		return AttendanceMinuteCalculation{}, err
	}

	result := AttendanceMinuteCalculation{
		Period: period,
	}

	for _, s := range sessions {
		if err := validateSession(s); err != nil {
			return AttendanceMinuteCalculation{}, fmt.Errorf("session %s: %w", s.SessionID, err)
		}

		allocYear, allocMonth := allocationMonth(s.CheckInAt)
		if allocYear != year || allocMonth != month {
			continue
		}

		if s.CheckOutAt == nil {
			result.ExcludedIncompleteSessions = append(result.ExcludedIncompleteSessions, ExcludedIncompleteSession{
				SessionID:       s.SessionID,
				Status:          s.Status,
				CheckInAt:       s.CheckInAt,
				AllocationYear:  allocYear,
				AllocationMonth: allocMonth,
			})
			continue
		}

		if !s.CheckOutAt.After(s.CheckInAt) {
			return AttendanceMinuteCalculation{}, fmt.Errorf("session %s: check_out_at must be after check_in_at", s.SessionID)
		}

		elapsed := s.CheckOutAt.Sub(s.CheckInAt)
		rawMinutes := int(elapsed / time.Minute)
		rounded := roundDurationUpToBlockMinutes(elapsed)

		row := SessionMinuteCalculation{
			SessionID:              s.SessionID,
			Status:                 s.Status,
			CheckInAt:              s.CheckInAt,
			CheckOutAt:             *s.CheckOutAt,
			AllocationYear:         allocYear,
			AllocationMonth:        allocMonth,
			RawElapsedDuration:     elapsed,
			RawElapsedMinutes:      rawMinutes,
			RoundedBillableMinutes: rounded,
		}

		result.Sessions = append(result.Sessions, row)
		result.RawElapsedMinutes += rawMinutes
		result.RoundedBillableMinutes += rounded
		result.IncludedSessionCount++
	}

	return result, nil
}

func allocationMonth(t time.Time) (int, time.Month) {
	local := t.In(londonLoc)
	return local.Year(), local.Month()
}

func validateSession(s AttendanceSessionInput) error {
	switch s.Status {
	case AttendanceSessionStatusOpen, AttendanceSessionStatusComplete, AttendanceSessionStatusCorrected:
	default:
		return fmt.Errorf("unknown status %q", s.Status)
	}

	if s.Status == AttendanceSessionStatusOpen && s.CheckOutAt != nil {
		return errors.New("open session must not have check_out_at")
	}

	if s.CheckOutAt != nil && !s.CheckOutAt.After(s.CheckInAt) {
		return errors.New("check_out_at must be after check_in_at")
	}

	return nil
}

func roundDurationUpToBlockMinutes(d time.Duration) int {
	blocks := d / billingBlock
	if d%billingBlock != 0 {
		blocks++
	}
	return int(blocks * 15)
}
