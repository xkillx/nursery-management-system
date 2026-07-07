package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrHourlyBookingNotFound = domainerrors.NotFound("hourly_booking", "Hourly booking not found.")
	ErrInvalidDuration       = domainerrors.Validation("Duration must be positive.", "duration_minutes")
	ErrInvalidStartTime      = domainerrors.Validation("Start time must be between 0 and 1439.", "start_time_minutes")
)
