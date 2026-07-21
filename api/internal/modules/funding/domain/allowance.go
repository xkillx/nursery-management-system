package domain

import (
	"fmt"
	"math"
	"time"
)

type TermDateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

// ComputeAllowanceMinutes computes the funded allowance in minutes for a billing month.
func ComputeAllowanceMinutes(
	fundedHoursPerWeek float64,
	fundingModel FundingModel,
	termDates []TermDateRange,
	billingMonth time.Time,
	closureDates []time.Time,
	fundingStartDate *time.Time,
	fundingEndDate *time.Time,
) (int, error) {
	if fundedHoursPerWeek <= 0 {
		return 0, nil
	}

	switch fundingModel {
	case FundingModelTermTimeOnly:
		return computeTermTimeAllowance(fundedHoursPerWeek, termDates, billingMonth, closureDates, fundingStartDate, fundingEndDate), nil
	case FundingModelStretched:
		return computeStretchedAllowance(fundedHoursPerWeek, fundingStartDate, fundingEndDate, billingMonth), nil
	default:
		return 0, fmt.Errorf("unknown funding model: %s", fundingModel)
	}
}

func computeTermTimeAllowance(
	fundedHoursPerWeek float64,
	termDates []TermDateRange,
	billingMonth time.Time,
	closureDates []time.Time,
	fundingStartDate *time.Time,
	fundingEndDate *time.Time,
) int {
	billingMonthStart := time.Date(billingMonth.Year(), billingMonth.Month(), 1, 0, 0, 0, 0, billingMonth.Location())
	billingMonthEnd := billingMonthStart.AddDate(0, 1, 0).Add(-time.Second)

	// Determine effective date range
	effectiveStart := billingMonthStart
	if fundingStartDate != nil && fundingStartDate.After(effectiveStart) {
		effectiveStart = *fundingStartDate
	}
	effectiveEnd := billingMonthEnd
	if fundingEndDate != nil && fundingEndDate.Before(effectiveEnd) {
		effectiveEnd = *fundingEndDate
	}

	if effectiveStart.After(effectiveEnd) {
		return 0
	}

	// Count term days in the effective range
	termDayCount := 0
	for _, td := range termDates {
		termStart := td.StartDate
		termEnd := td.EndDate

		// Clamp to effective range
		if termStart.Before(effectiveStart) {
			termStart = effectiveStart
		}
		if termEnd.After(effectiveEnd) {
			termEnd = effectiveEnd
		}

		if termStart.After(termEnd) {
			continue
		}

		for d := termStart; !d.After(termEnd); d = d.AddDate(0, 0, 1) {
			if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
				termDayCount++
			}
		}
	}

	// Subtract closure dates that fall within effective range
	closureSet := make(map[string]bool)
	for _, cd := range closureDates {
		if !cd.Before(effectiveStart) && !cd.After(effectiveEnd) {
			if cd.Weekday() != time.Saturday && cd.Weekday() != time.Sunday {
				closureSet[cd.Format("2006-01-02")] = true
			}
		}
	}
	termDayCount -= len(closureSet)

	if termDayCount <= 0 {
		return 0
	}

	// Formula: hours_per_week * 60 * term_days / 5
	return int(math.Round(fundedHoursPerWeek * 60 * float64(termDayCount) / 5))
}

func computeStretchedAllowance(
	fundedHoursPerWeek float64,
	fundingStartDate *time.Time,
	fundingEndDate *time.Time,
	billingMonth time.Time,
) int {
	billingMonthStart := time.Date(billingMonth.Year(), billingMonth.Month(), 1, 0, 0, 0, 0, billingMonth.Location())
	billingMonthEnd := billingMonthStart.AddDate(0, 1, 0).Add(-time.Second)

	// Check if funding is active for this month
	if fundingStartDate != nil && fundingStartDate.After(billingMonthEnd) {
		return 0
	}
	if fundingEndDate != nil && fundingEndDate.Before(billingMonthStart) {
		return 0
	}

	// Formula: hours_per_week * 60 * 38 / 12
	fullMonthMinutes := int(math.Round(fundedHoursPerWeek * 60 * 38 / 12))

	// Pro-rating: if funding starts/ends mid-month
	if fundingStartDate != nil && fundingStartDate.After(billingMonthStart) {
		remainingDays := int(billingMonthEnd.Sub(*fundingStartDate).Hours()/24) + 1
		totalDays := int(billingMonthEnd.Sub(billingMonthStart).Hours()/24) + 1
		return int(math.Round(float64(fullMonthMinutes) * float64(remainingDays) / float64(totalDays)))
	}
	if fundingEndDate != nil && fundingEndDate.Before(billingMonthEnd) {
		activeDays := int(fundingEndDate.Sub(billingMonthStart).Hours()/24) + 1
		totalDays := int(billingMonthEnd.Sub(billingMonthStart).Hours()/24) + 1
		return int(math.Round(float64(fullMonthMinutes) * float64(activeDays) / float64(totalDays)))
	}

	return fullMonthMinutes
}
