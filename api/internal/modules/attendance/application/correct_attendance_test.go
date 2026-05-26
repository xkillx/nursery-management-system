package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/audit"
)

func TestCorrection_RejectsNoTarget(t *testing.T) {
	uc := newTestCorrectAttendance()
	_, err := uc.Execute(context.Background(), makeActor(), domain.CorrectionParams{
		CheckInAt:  time.Now().Add(-2 * time.Hour),
		CheckOutAt: time.Now().Add(-time.Hour),
		ReasonCode: "missed_check_in",
	})
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %v", err)
	}
}

func TestCorrection_RejectsBothTargets(t *testing.T) {
	uc := newTestCorrectAttendance()
	_, err := uc.Execute(context.Background(), makeActor(), domain.CorrectionParams{
		SessionID:  ptrUUID(uuid.New()),
		ChildID:    ptrUUID(uuid.New()),
		CheckInAt:  time.Now().Add(-2 * time.Hour),
		CheckOutAt: time.Now().Add(-time.Hour),
		ReasonCode: "missed_check_in",
	})
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %v", err)
	}
}

func TestCorrection_RejectsInvalidReason(t *testing.T) {
	uc := newTestCorrectAttendance()
	_, err := uc.Execute(context.Background(), makeActor(), domain.CorrectionParams{
		SessionID:  ptrUUID(uuid.New()),
		CheckInAt:  time.Now().Add(-2 * time.Hour),
		CheckOutAt: time.Now().Add(-time.Hour),
		ReasonCode: "not_real",
	})
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_correction_reason_invalid" {
		t.Fatalf("expected attendance_correction_reason_invalid, got %v", err)
	}
}

func TestCorrection_RejectsOtherWithoutNote(t *testing.T) {
	uc := newTestCorrectAttendance()
	_, err := uc.Execute(context.Background(), makeActor(), domain.CorrectionParams{
		SessionID:  ptrUUID(uuid.New()),
		CheckInAt:  time.Now().Add(-2 * time.Hour),
		CheckOutAt: time.Now().Add(-time.Hour),
		ReasonCode: "other",
	})
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "reason_note_required_for_other" {
		t.Fatalf("expected reason_note_required_for_other, got %v", err)
	}
}

func TestCorrection_RejectsInvalidTimeOrder(t *testing.T) {
	uc := newTestCorrectAttendance()
	_, err := uc.Execute(context.Background(), makeActor(), domain.CorrectionParams{
		SessionID:  ptrUUID(uuid.New()),
		CheckInAt:  time.Now().Add(-time.Hour),
		CheckOutAt: time.Now().Add(-2 * time.Hour),
		ReasonCode: "incorrect_time",
	})
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_invalid_time_order" {
		t.Fatalf("expected attendance_invalid_time_order, got %v", err)
	}
}

func TestCorrection_RejectsFutureTime(t *testing.T) {
	uc := newTestCorrectAttendance()
	_, err := uc.Execute(context.Background(), makeActor(), domain.CorrectionParams{
		SessionID:  ptrUUID(uuid.New()),
		CheckInAt:  time.Now().Add(-time.Hour),
		CheckOutAt: time.Now().Add(time.Hour),
		ReasonCode: "incorrect_time",
	})
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_correction_future_time" {
		t.Fatalf("expected attendance_correction_future_time, got %v", err)
	}
}

func TestCorrection_RejectsBlankReason(t *testing.T) {
	uc := newTestCorrectAttendance()
	_, err := uc.Execute(context.Background(), makeActor(), domain.CorrectionParams{
		SessionID:  ptrUUID(uuid.New()),
		CheckInAt:  time.Now().Add(-2 * time.Hour),
		CheckOutAt: time.Now().Add(-time.Hour),
		ReasonCode: "",
	})
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_correction_reason_required" {
		t.Fatalf("expected attendance_correction_reason_required, got %v", err)
	}
}

func TestMapCorrectionError_ReasonRequired(t *testing.T) {
	err := mapCorrectionError(domain.ErrCorrectionReasonRequired)
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_correction_reason_required" {
		t.Fatalf("expected attendance_correction_reason_required, got %v", err)
	}
}

func TestMapCorrectionError_ReasonInvalid(t *testing.T) {
	err := mapCorrectionError(domain.ErrCorrectionReasonInvalid)
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_correction_reason_invalid" {
		t.Fatalf("expected attendance_correction_reason_invalid, got %v", err)
	}
}

func TestMapCorrectionError_ReasonNoteRequired(t *testing.T) {
	err := mapCorrectionError(domain.ErrReasonNoteRequiredForOther)
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "reason_note_required_for_other" {
		t.Fatalf("expected reason_note_required_for_other, got %v", err)
	}
}

func TestValidateTargets_SessionOnly(t *testing.T) {
	err := validateTargets(domain.CorrectionParams{SessionID: ptrUUID(uuid.New())})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateTargets_ChildOnly(t *testing.T) {
	err := validateTargets(domain.CorrectionParams{ChildID: ptrUUID(uuid.New())})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCorrection_NonOtherReasonDoesNotRequireNote(t *testing.T) {
	for _, code := range []string{"missed_check_in", "missed_check_out", "incorrect_time", "duplicate_entry"} {
		err := domain.ValidateCorrectionReason(code, "")
		if err != nil {
			t.Fatalf("reason_code=%s: expected nil for empty note, got %v", code, err)
		}
	}
}

func TestCorrection_ActionLocalDateDistinctFromCorrectedInterval(t *testing.T) {
	// Verify the clock computes distinct local dates for the corrected interval vs the action.
	actionTime := time.Date(2025, 5, 20, 14, 0, 0, 0, time.UTC)
	clock := fixedClock(actionTime)

	checkInLocalDate := clock.LocalDate(time.Date(2025, 5, 15, 8, 0, 0, 0, time.UTC))
	correctionActionLocalDate := clock.LocalDate(actionTime)

	y, m, d := checkInLocalDate.Date()
	if y != 2025 || m != 5 || d != 15 {
		t.Fatalf("checkInLocalDate = %d-%02d-%02d, want 2025-05-15", y, m, d)
	}

	y, m, d = correctionActionLocalDate.Date()
	if y != 2025 || m != 5 || d != 20 {
		t.Fatalf("correctionActionLocalDate = %d-%02d-%02d, want 2025-05-20", y, m, d)
	}

	if checkInLocalDate.Equal(correctionActionLocalDate) {
		t.Fatal("expected corrected interval and action local dates to differ")
	}
}

func TestCorrection_ActionLocalDateLondonBoundary(t *testing.T) {
	// Action at 2025-06-15 23:30 UTC = 2025-06-16 00:30 BST
	actionTime := time.Date(2025, 6, 15, 23, 30, 0, 0, time.UTC)
	clock := fixedClock(actionTime)

	correctionActionLocalDate := clock.LocalDate(actionTime)
	y, m, d := correctionActionLocalDate.Date()
	if y != 2025 || m != 6 || d != 16 {
		t.Fatalf("correctionActionLocalDate = %d-%02d-%02d, want 2025-06-16 (BST)", y, m, d)
	}

	checkInLocalDate := clock.LocalDate(time.Date(2025, 6, 10, 8, 0, 0, 0, time.UTC))
	y, m, d = checkInLocalDate.Date()
	if y != 2025 || m != 6 || d != 10 {
		t.Fatalf("checkInLocalDate = %d-%02d-%02d, want 2025-06-10", y, m, d)
	}
}

func newTestCorrectAttendance() *CorrectAttendance {
	return NewCorrectAttendance(nil, nil, nil, audit.NewWriter(), fixedClock(time.Now().UTC()))
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
