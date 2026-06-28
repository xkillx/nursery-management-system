package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrFundingProfileNotFound = domainerrors.NotFound("funding_profile", "funding profile not found")
	ErrChildNotFound          = domainerrors.NotFound("funding_child", "child not found")
	ErrMonthOutsideEnrollment = domainerrors.Conflict("funding_month_outside_enrollment_window", "billing month outside child enrollment window")
)
