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
		"guardian_not_active", "membership_not_parent", "membership_not_active":
		status = http.StatusBadRequest
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
