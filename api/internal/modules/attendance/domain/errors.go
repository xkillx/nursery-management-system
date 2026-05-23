package domain

import "errors"

var (
	ErrSessionAlreadyOpen    = errors.New("attendance_session_already_open")
	ErrSessionNotOpen        = errors.New("attendance_session_not_open")
	ErrChildEnrollmentIncomplete = errors.New("child_enrollment_incomplete")
	ErrInvalidTimeOrder      = errors.New("attendance_invalid_time_order")
	ErrChildNotFound         = errors.New("child_not_found")
)
