package httpserver

import (
	"net/http"
	"testing"

	domainerrors "nursery-management-system/api/internal/platform/errors"
)

func TestMapDomainError_AttendanceSessionAlreadyOpen_Conflict(t *testing.T) {
	err := domainerrors.Conflict("attendance_session_already_open", "test")
	status, resp := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
	if resp.Code != "attendance_session_already_open" {
		t.Fatalf("expected attendance_session_already_open, got %s", resp.Code)
	}
}

func TestMapDomainError_AttendanceSessionNotOpen_Conflict(t *testing.T) {
	err := domainerrors.Conflict("attendance_session_not_open", "test")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_ChildEnrollmentIncomplete_Conflict(t *testing.T) {
	err := domainerrors.Conflict("child_enrollment_incomplete", "test")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_AttendanceInvalidTimeOrder_Conflict(t *testing.T) {
	err := domainerrors.Conflict("attendance_invalid_time_order", "test")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_CorrectionReasonRequired_BadRequest(t *testing.T) {
	err := domainerrors.New("attendance_correction_reason_required", "reason required")
	status, resp := MapDomainError(err, "req-1")
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
	if resp.Code != "attendance_correction_reason_required" {
		t.Fatalf("expected attendance_correction_reason_required, got %s", resp.Code)
	}
}

func TestMapDomainError_CorrectionReasonInvalid_BadRequest(t *testing.T) {
	err := domainerrors.New("attendance_correction_reason_invalid", "invalid reason")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
}

func TestMapDomainError_ReasonNoteRequiredForOther_BadRequest(t *testing.T) {
	err := domainerrors.New("reason_note_required_for_other", "note required", "reason_note")
	status, resp := MapDomainError(err, "req-1")
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
	if resp.Details.(map[string]string)["field"] != "reason_note" {
		t.Fatalf("expected field=reason_note, got %v", resp.Details)
	}
}

func TestMapDomainError_InternalError_500(t *testing.T) {
	err := domainerrors.Internal(nil)
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", status)
	}
}

func TestMapDomainError_AttendanceCorrectionFutureTime_Conflict(t *testing.T) {
	err := domainerrors.Conflict("attendance_correction_future_time", "future")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_AttendanceSessionOverlap_Conflict(t *testing.T) {
	err := domainerrors.Conflict("attendance_session_overlap", "overlap")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_AttendanceOutsideEnrollmentWindow_Conflict(t *testing.T) {
	err := domainerrors.Conflict("attendance_outside_enrollment_window", "outside")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_AbsenceAttendanceExists_Conflict(t *testing.T) {
	err := domainerrors.Conflict("absence_attendance_exists", "test")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_AbsenceMarkerExists_Conflict(t *testing.T) {
	err := domainerrors.Conflict("absence_marker_exists", "test")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestMapDomainError_AbsenceMarkerNotFound_404(t *testing.T) {
	err := domainerrors.NotFound("absence_marker", "test")
	status, _ := MapDomainError(err, "req-1")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestMapDomainError_InvoiceNotPayable_Conflict(t *testing.T) {
	err := domainerrors.Conflict("invoice_not_payable", "Invoice is not payable.")
	status, resp := MapDomainError(err, "req-1")
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
	if resp.Code != "invoice_not_payable" {
		t.Fatalf("expected invoice_not_payable, got %s", resp.Code)
	}
}

func TestMapDomainError_PaymentProviderUnconfigured_503(t *testing.T) {
	err := domainerrors.New("payment_provider_unconfigured", "Payment provider is not configured.")
	status, resp := MapDomainError(err, "req-1")
	if status != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", status)
	}
	if resp.Code != "payment_provider_unconfigured" {
		t.Fatalf("expected payment_provider_unconfigured, got %s", resp.Code)
	}
}

func TestMapDomainError_PaymentProviderError_502(t *testing.T) {
	err := domainerrors.New("payment_provider_error", "Payment provider failed to create checkout session.")
	status, resp := MapDomainError(err, "req-1")
	if status != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", status)
	}
	if resp.Code != "payment_provider_error" {
		t.Fatalf("expected payment_provider_error, got %s", resp.Code)
	}
}
