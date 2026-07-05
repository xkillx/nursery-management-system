package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrClosureDayNotFound = domainerrors.NotFound("closure_day", "Closure day not found.")
	ErrDuplicateDate      = domainerrors.Conflict("closure_date_taken", "A closure day for this date already exists at this site.")
	ErrDateRequired       = domainerrors.Validation("Date is required.", "date")
)
