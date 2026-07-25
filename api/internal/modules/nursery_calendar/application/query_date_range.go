package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/nursery_calendar/domain"
)

type QueryDateRange struct {
	closureLookup domain.ClosureDayLookup
	holidayLookup domain.HolidayPeriodLookup
}

func NewQueryDateRange(closureLookup domain.ClosureDayLookup, holidayLookup domain.HolidayPeriodLookup) *QueryDateRange {
	return &QueryDateRange{
		closureLookup: closureLookup,
		holidayLookup: holidayLookup,
	}
}

func (uc *QueryDateRange) Execute(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time, isTermTime bool) ([]domain.DateRangeResult, error) {
	closureDates, err := uc.closureLookup.GetClosureDatesForBranchAndDateRange(ctx, tenantID, branchID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query closure dates: %w", err)
	}

	closureSet := make(map[string]bool)
	for _, d := range closureDates {
		closureSet[d.Format("2006-01-02")] = true
	}

	var holidayPeriods []domain.HolidayPeriodDateRange
	if isTermTime {
		holidayPeriods, err = uc.holidayLookup.GetHolidayPeriodsForBranchAndDateRange(ctx, tenantID, branchID, from, to)
		if err != nil {
			return nil, fmt.Errorf("query holiday periods: %w", err)
		}
	}

	var results []domain.DateRangeResult
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		if closureSet[key] {
			results = append(results, domain.DateRangeResult{
				Date:   d,
				IsOpen: false,
				Reason: domain.ClosureReasonClosureDay,
			})
			continue
		}
		if isTermTime && isInAnyHolidayPeriod(d, holidayPeriods) {
			results = append(results, domain.DateRangeResult{
				Date:   d,
				IsOpen: false,
				Reason: domain.ClosureReasonHolidayPeriod,
			})
			continue
		}
		results = append(results, domain.DateRangeResult{
			Date:   d,
			IsOpen: true,
			Reason: domain.ClosureReasonNone,
		})
	}

	return results, nil
}

func isInAnyHolidayPeriod(t time.Time, periods []domain.HolidayPeriodDateRange) bool {
	for _, p := range periods {
		if !t.Before(p.StartDate) && !t.After(p.EndDate) {
			return true
		}
	}
	return false
}
