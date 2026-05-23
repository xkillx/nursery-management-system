package application

import (
	"fmt"
	"time"
)

var londonLocation *time.Location

func init() {
	var err error
	londonLocation, err = time.LoadLocation("Europe/London")
	if err != nil {
		panic(fmt.Sprintf("load Europe/London timezone: %v", err))
	}
}

// Clock provides the current UTC instant. Injectable for tests.
type Clock func() time.Time

func RealClock() time.Time {
	return time.Now().UTC()
}

// AttendanceClock wraps clock + timezone helpers for attendance operations.
type AttendanceClock struct {
	now Clock
}

func NewAttendanceClock(now Clock) *AttendanceClock {
	return &AttendanceClock{now: now}
}

func (ac *AttendanceClock) Now() (utc time.Time, localDate time.Time) {
	utc = ac.now()
	localDate = utc.In(londonLocation)
	return utc, localDate
}

func (ac *AttendanceClock) LocalDate(utc time.Time) time.Time {
	return utc.In(londonLocation)
}

// LondonLocalDate derives the London local date from a UTC instant.
// Date-only; time portion carries the London wall-clock time at that instant.
func LondonLocalDate(utc time.Time) time.Time {
	return utc.In(londonLocation)
}

// LondonDateOnly returns the date-only portion of the London local date at the given UTC instant.
func LondonDateOnly(utc time.Time) time.Time {
	t := utc.In(londonLocation)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, londonLocation)
}
