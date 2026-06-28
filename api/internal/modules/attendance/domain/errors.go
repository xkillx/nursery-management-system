package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrSessionAlreadyOpen         = domainerrors.Conflict("attendance_session_already_open", "Open session exists for this child")
	ErrSessionNotOpen             = domainerrors.Conflict("attendance_session_not_open", "Session is not open")
	ErrChildEnrollmentIncomplete  = domainerrors.Conflict("child_enrollment_incomplete", "Child enrollment incomplete")
	ErrInvalidTimeOrder           = domainerrors.Conflict("attendance_invalid_time_order", "Check-out must be after check-in")
	ErrChildNotFound              = domainerrors.NotFound("child", "Child not found")
	ErrCorrectionReasonRequired   = domainerrors.Conflict("attendance_correction_reason_required", "Correction reason required")
	ErrCorrectionReasonInvalid    = domainerrors.Conflict("attendance_correction_reason_invalid", "Invalid correction reason")
	ErrReasonNoteRequiredForOther = domainerrors.New("reason_note_required_for_other", "Reason note required for 'other'", "reason_note")
	ErrCorrectionFutureTime       = domainerrors.Conflict("attendance_correction_future_time", "Cannot correct future time")
	ErrSessionOverlap             = domainerrors.Conflict("attendance_session_overlap", "Session overlap detected")
	ErrOutsideEnrollmentWindow    = domainerrors.Conflict("attendance_outside_enrollment_window", "Outside enrollment window")
	ErrSessionNotFound            = domainerrors.NotFound("attendance_session", "Session not found")
	ErrAbsenceMarkerExists        = domainerrors.Conflict("absence_marker_exists", "Absence marker already exists")
)
