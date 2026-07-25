package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/nursery_calendar/domain"
)

type QueryCalendarDay struct {
	closureLookup domain.ClosureDayLookup
	holidayLookup domain.HolidayPeriodLookup
}

func NewQueryCalendarDay(closureLookup domain.ClosureDayLookup, holidayLookup domain.HolidayPeriodLookup) *QueryCalendarDay {
	return &QueryCalendarDay{
		closureLookup: closureLookup,
		holidayLookup: holidayLookup,
	}
}

func (uc *QueryCalendarDay) Execute(ctx context.Context, tenantID, branchID uuid.UUID, date time.Time, isTermTime bool) (*domain.CalendarDayInfo, error) {
	closureDates, err := uc.closureLookup.GetClosureDatesForBranchAndDateRange(ctx, tenantID, branchID, date, date)
	if err != nil {
		return nil, fmt.Errorf("check closure dates: %w", err)
	}

	for _, cd := range closureDates {
		if sameDay(cd, date) {
			return &domain.CalendarDayInfo{
				Date:   date,
				IsOpen: false,
				Reason: domain.ClosureReasonClosureDay,
			}, nil
		}
	}

	if isTermTime {
		periods, err := uc.holidayLookup.GetHolidayPeriodsForBranchAndDateRange(ctx, tenantID, branchID, date, date)
		if err != nil {
			return nil, fmt.Errorf("check holiday periods: %w", err)
		}
		for _, p := range periods {
			if !date.Before(p.StartDate) && !date.After(p.EndDate) {
				return &domain.CalendarDayInfo{
					Date:   date,
					IsOpen: false,
					Reason: domain.ClosureReasonHolidayPeriod,
				}, nil
			}
		}
	}

	return &domain.CalendarDayInfo{
		Date:   date,
		IsOpen: true,
		Reason: domain.ClosureReasonNone,
	}, nil
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
