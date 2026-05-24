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

func newTestCorrectAttendance() *CorrectAttendance {
	return NewCorrectAttendance(nil, nil, nil, audit.NewWriter(), fixedClock(time.Now().UTC()))
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
