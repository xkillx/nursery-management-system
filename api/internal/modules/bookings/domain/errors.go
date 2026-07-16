package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrBookingNotFound         = domainerrors.NotFound("booking", "Booking not found.")
	ErrInvalidDaysOfWeek       = domainerrors.Validation("Days of week must contain at least one valid day (1-7).", "days_of_week")
	ErrInvalidFundingType      = domainerrors.Validation("Invalid funding type.", "funding_type")
	ErrInvalidDateRange        = domainerrors.Validation("Effective end date must be on or after start date.", "effective_end_date")
	ErrBookingNotActive        = domainerrors.Conflict("booking", "Booking is not active.")
	ErrBookingAlreadyPaused    = domainerrors.Conflict("booking", "Booking is already paused.")
	ErrBookingAlreadyCancelled = domainerrors.Conflict("booking", "Booking is already cancelled.")
)
