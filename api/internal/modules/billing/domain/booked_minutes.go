package domain

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// BookedSessionType is a minimal projection of a session type used for
// invoice-line explainability: the duration (in minutes) and the human
// label.
type BookedSessionType struct {
	ID              string
	Name            string
	StartMinutes    int // minutes since midnight, local London time
	EndMinutes      int
	DurationMinutes int // EndMinutes - StartMinutes
}

// BookedPatternEntry is one row in a Booking Pattern (day-of-week + session type).
type BookedPatternEntry struct {
	DayOfWeek   int
	SessionType BookedSessionType
}

// BookedSession is a single booked occurrence in the billing month — used
// for invoice line explainability.
type BookedSession struct {
	DayOfWeek       int
	OccurrenceDate  time.Time
	DurationMinutes int
	SessionTypeID   string
	SessionTypeName string
}

// BookedCoreCalculation is the per-term result of the advance-pay calculation.
type BookedCoreCalculation struct {
	BookingPatternID string
	TotalMinutes     int
	Subtotal         Money
	PerEntry         []BookedEntryBreakdown
	Sessions         []BookedSession
}

// BookedEntryBreakdown is the per-(day,session) subtotal.
type BookedEntryBreakdown struct {
	DayOfWeek          int
	SessionTypeID      string
	SessionTypeName    string
	DurationMinutes    int
	OccurrencesInMonth int
	TotalMinutes       int
}

type TermDateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

type TermDateLookup interface {
	GetTermDateRangesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]TermDateRange, error)
}

type AdHocBookingRow struct {
	ID              uuid.UUID
	ChildID         uuid.UUID
	CalendarDate    time.Time
	SessionTypeID   uuid.UUID
	SessionTypeName string
	StartMinutes    int
	EndMinutes      int
}

type AdHocBookingLookup interface {
	ListActiveBookingsForChildInMonth(ctx context.Context, tenantID, branchID, childID uuid.UUID, month time.Time) ([]AdHocBookingRow, error)
}

type HourlyBookingRow struct {
	ID               uuid.UUID
	ChildID          uuid.UUID
	CalendarDate     time.Time
	StartTimeMinutes int
	DurationMinutes  int
}

type HourlyBookingLookup interface {
	ListActiveByChildAndMonth(ctx context.Context, tenantID, branchID, childID uuid.UUID, monthStart, monthEnd time.Time) ([]HourlyBookingRow, error)
}

type ClosureDateLookup interface {
	GetClosureDatesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]time.Time, error)
}

type FundingRecordRow struct {
	FundingModel       string
	FundedHoursPerWeek *float64
}

type FundingRecordLookup interface {
	GetFundingRecord(ctx context.Context, tenantID, branchID, childID uuid.UUID) (FundingRecordRow, bool, error)
}

func isDateInAnyTermRange(t time.Time, ranges []TermDateRange) bool {
	for _, r := range ranges {
		if !t.Before(r.StartDate) && !t.After(r.EndDate) {
			return true
		}
	}
	return false
}

func isWeekday(t time.Time) bool {
	d := t.Weekday()
	return d != time.Saturday && d != time.Sunday
}

func isClosureDate(t time.Time, closures []time.Time) bool {
	for _, c := range closures {
		if t.Year() == c.Year() && t.Month() == c.Month() && t.Day() == c.Day() {
			return true
		}
	}
	return false
}

// CalculateTermTimeFundedAllowanceMinutes computes the funded allowance for a
// term-time-only child in a given billing month.
//
// The formula:
//
//	termDaysInMonth  = count of weekdays falling within any term date range
//	                   that lie inside billingMonth, minus closure dates
//	allowance        = fundedHoursPerWeek × 60 × termDaysInMonth / 5
//
// Returns 0 when termDaysInMonth is 0 or fundedHoursPerWeek is 0.
func CalculateTermTimeFundedAllowanceMinutes(fundedHoursPerWeek float64, termDateRanges []TermDateRange, billingMonth time.Time, closureDates []time.Time) int {
	if fundedHoursPerWeek <= 0 {
		return 0
	}
	if billingMonth.Day() != 1 {
		return 0
	}
	nextMonth := billingMonth.AddDate(0, 1, 0)

	termDaysInMonth := 0
	for d := billingMonth; d.Before(nextMonth); d = d.AddDate(0, 0, 1) {
		if isWeekday(d) && isDateInAnyTermRange(d, termDateRanges) && !isClosureDate(d, closureDates) {
			termDaysInMonth++
		}
	}
	if termDaysInMonth == 0 {
		return 0
	}
	return int(fundedHoursPerWeek * 60 * float64(termDaysInMonth) / 5)
}

// CalculateAdHocChargeMinutes returns the billable minutes for an ad-hoc
// booking: ceil(sessionDuration × multiplier).
func CalculateAdHocChargeMinutes(sessionDurationMinutes int, adHocRateMultiplier float64) int {
	return int(math.Ceil(float64(sessionDurationMinutes) * adHocRateMultiplier))
}

// CalculateBookedCoreMinutesInMonth computes the monthly booked core minutes
// for a Booking Pattern, by counting occurrences of each (day_of_week,
// session_type) entry in the calendar month.
//
// When termDates is nil or empty: counts all matching day-of-week occurrences
// (standard 52-week behaviour).
//
// When termDates is non-empty: only counts dates that fall within at least
// one term date range (term-time-only billing).
//
// The session type duration is in minutes; occurrences are integer counts of
// that day-of-week falling inside the billing month AND in any term range.
//
// The plan's "always bill 52 weeks" rule means we do NOT pro-rate for partial
// weeks at the month boundary — we simply count how many times each
// day-of-week occurs in the calendar month (filtered by term dates when
// term_time_only).
func CalculateBookedCoreMinutesInMonth(
	patternID string,
	entries []BookedPatternEntry,
	billingMonthStart time.Time,
	siteHourlyRateMinor int,
	termDates []TermDateRange,
	closureDates []time.Time,
) (BookedCoreCalculation, error) {
	if siteHourlyRateMinor < 0 {
		return BookedCoreCalculation{}, fmt.Errorf("site_hourly_rate_minor must be >= 0")
	}

	monthStart := billingMonthStart.UTC()
	// End-exclusive first day of the next month.
	nextMonth := monthStart.AddDate(0, 1, 0)
	if monthStart.Day() != 1 {
		return BookedCoreCalculation{}, fmt.Errorf("billing_month_start must be the 1st of a month (got day=%d)", monthStart.Day())
	}

	calc := BookedCoreCalculation{
		BookingPatternID: patternID,
		PerEntry:         make([]BookedEntryBreakdown, 0, len(entries)),
		Sessions:         make([]BookedSession, 0),
	}

	for _, e := range entries {
		if e.DayOfWeek < 1 || e.DayOfWeek > 7 {
			return BookedCoreCalculation{}, fmt.Errorf("day_of_week out of range: %d", e.DayOfWeek)
		}
		if e.SessionType.DurationMinutes <= 0 {
			return BookedCoreCalculation{}, fmt.Errorf("session_type %q has non-positive duration", e.SessionType.Name)
		}

		// Iterate over each calendar date in the month and count the
		// matching day-of-week occurrences.
		occurrences := 0
		for d := monthStart; d.Before(nextMonth); d = d.AddDate(0, 0, 1) {
			// time.Weekday(): Sunday=0 ... Saturday=6.
			// Our day_of_week: Monday=1 ... Sunday=7.
			if int(d.Weekday()) == e.DayOfWeek%7 {
				if (len(termDates) == 0 || isDateInAnyTermRange(d, termDates)) && !isClosureDate(d, closureDates) {
					occurrences++
				}
			}
		}

		if occurrences == 0 {
			continue
		}

		totalMinutes := occurrences * e.SessionType.DurationMinutes
		calc.PerEntry = append(calc.PerEntry, BookedEntryBreakdown{
			DayOfWeek:          e.DayOfWeek,
			SessionTypeID:      e.SessionType.ID,
			SessionTypeName:    e.SessionType.Name,
			DurationMinutes:    e.SessionType.DurationMinutes,
			OccurrencesInMonth: occurrences,
			TotalMinutes:       totalMinutes,
		})
		calc.TotalMinutes += totalMinutes

		// Per-session breakdown for explainability.
		for d := monthStart; d.Before(nextMonth); d = d.AddDate(0, 0, 1) {
			if int(d.Weekday()) == e.DayOfWeek%7 {
				if (len(termDates) == 0 || isDateInAnyTermRange(d, termDates)) && !isClosureDate(d, closureDates) {
					calc.Sessions = append(calc.Sessions, BookedSession{
						DayOfWeek:       e.DayOfWeek,
						OccurrenceDate:  d,
						DurationMinutes: e.SessionType.DurationMinutes,
						SessionTypeID:   e.SessionType.ID,
						SessionTypeName: e.SessionType.Name,
					})
				}
			}
		}
	}

	subtotal, err := CalculateHourlyAmountMinor(calc.TotalMinutes, siteHourlyRateMinor)
	if err != nil {
		return BookedCoreCalculation{}, err
	}
	calc.Subtotal = Money{minor: subtotal}
	return calc, nil
}

// ComputeFundedDeductionMinor returns the funded deduction (in minor units) and
// the billable minutes after the deduction is applied.
//
// bookedCoreMinutes is the booking-driven monthly total.
// fundedAllowanceMinutes comes from the funding module via FundingLookup.
// siteHourlyRateMinor is the term's snapshotted rate.
//
// fundedDeductionMinutes = min(bookedCoreMinutes, fundedAllowanceMinutes)
// billableMinutes = max(0, bookedCoreMinutes - fundedAllowanceMinutes)
func ComputeFundedDeductionMinor(
	bookedCoreMinutes int,
	fundedAllowanceMinutes int,
	siteHourlyRateMinor int,
) (fundedDeductionMinutes, billableMinutes, fundedDeductionMinor, billableMinor int, err error) {
	if bookedCoreMinutes < 0 {
		return 0, 0, 0, 0, fmt.Errorf("booked_core_minutes must be >= 0")
	}
	if fundedAllowanceMinutes < 0 {
		return 0, 0, 0, 0, fmt.Errorf("funded_allowance_minutes must be >= 0")
	}
	if siteHourlyRateMinor < 0 {
		return 0, 0, 0, 0, fmt.Errorf("site_hourly_rate_minor must be >= 0")
	}
	fundedDeductionMinutes = minInt(bookedCoreMinutes, fundedAllowanceMinutes)
	billableMinutes = maxInt(0, bookedCoreMinutes-fundedAllowanceMinutes)
	fundedDeductionMinor, err = CalculateHourlyAmountMinor(fundedDeductionMinutes, siteHourlyRateMinor)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	billableMinor, err = CalculateHourlyAmountMinor(billableMinutes, siteHourlyRateMinor)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return fundedDeductionMinutes, billableMinutes, fundedDeductionMinor, billableMinor, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
