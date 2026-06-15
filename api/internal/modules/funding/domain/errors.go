package domain

import "errors"

var (
	ErrFundingProfileNotFound = errors.New("funding profile not found")
	ErrChildNotFound          = errors.New("child not found")
	ErrMonthOutsideEnrollment = errors.New("billing month outside child enrollment window")
)
