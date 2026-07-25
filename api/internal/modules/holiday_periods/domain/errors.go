package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrHolidayPeriodNotFound = domainerrors.NotFound("holiday_period", "Holiday period not found.")
	ErrInvalidDateRange      = domainerrors.Validation("Start date must be before end date.", "date_range")
	ErrNameRequired          = domainerrors.Validation("Name is required.", "name")
	ErrTypeRequired          = domainerrors.Validation("Type is required.", "type")
	ErrInvalidType           = domainerrors.Validation("Invalid holiday period type.", "type")
	ErrStartDateRequired     = domainerrors.Validation("Start date is required.", "start_date")
	ErrEndDateRequired       = domainerrors.Validation("End date is required.", "end_date")
)
