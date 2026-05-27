package httpserver

import (
	"net/http"

	domainerrors "nursery-management-system/api/internal/platform/errors"
)

func MapDomainError(err error, requestID string) (int, ErrorResponse) {
	return mapDomainError(err, requestID)
}

func mapDomainError(err error, requestID string) (int, ErrorResponse) {
	domainErr, ok := err.(*domainerrors.DomainError)
	if !ok {
		return http.StatusInternalServerError, ErrorResponse{
			Code:      "internal_error",
			Message:   "Something went wrong.",
			RequestID: requestID,
		}
	}

	resp := ErrorResponse{
		Code:      domainErr.Code,
		Message:   domainErr.Message,
		RequestID: requestID,
	}

	if domainErr.Field != "" {
		resp.Details = map[string]string{"field": domainErr.Field}
	}

	status := http.StatusInternalServerError
	switch domainErr.Code {
	case "unauthorized":
		status = http.StatusUnauthorized
	case "forbidden_scope_selection", "forbidden_role", "forbidden_role_unknown", "forbidden_scope":
		status = http.StatusForbidden
	case "validation_error", "child_lifecycle_reason_required", "guardian_deactivation_reason_required",
		"relationship_reason_required", "lifecycle_reason_invalid", "reason_note_required_for_other",
		"guardian_not_active", "membership_not_parent", "membership_not_active",
			"attendance_correction_reason_required", "attendance_correction_reason_invalid",
			"password_reset_token_invalid", "password_reset_token_expired", "password_reset_token_used",
			"invite_token_invalid", "invite_token_expired", "invite_token_revoked", "invite_token_accepted":
		status = http.StatusBadRequest
	case "attendance_session_already_open", "attendance_session_not_open",
		"child_enrollment_incomplete", "attendance_invalid_time_order",
		"attendance_correction_future_time", "attendance_session_overlap",
		"attendance_outside_enrollment_window",
		"funding_month_outside_enrollment_window",
		"absence_attendance_exists", "absence_marker_exists":
		status = http.StatusConflict
	default:
		if len(domainErr.Code) > 10 && domainErr.Code[len(domainErr.Code)-10:] == "_not_found" {
			status = http.StatusNotFound
		}
		if len(domainErr.Code) > 8 && domainErr.Code[len(domainErr.Code)-8:] == "_conflict" {
			status = http.StatusConflict
		}
	}

	return status, resp
}
