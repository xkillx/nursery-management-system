package domain

import (
	"testing"
)

func TestValidateCorrectionReason_RejectsMissingReason(t *testing.T) {
	err := ValidateCorrectionReason("", "")
	if err != ErrCorrectionReasonRequired {
		t.Fatalf("expected ErrCorrectionReasonRequired, got %v", err)
	}
}

func TestValidateCorrectionReason_RejectsBlankReason(t *testing.T) {
	err := ValidateCorrectionReason("   ", "")
	if err != ErrCorrectionReasonRequired {
		t.Fatalf("expected ErrCorrectionReasonRequired, got %v", err)
	}
}

func TestValidateCorrectionReason_RejectsUnknownReason(t *testing.T) {
	err := ValidateCorrectionReason("not_a_real_code", "")
	if err != ErrCorrectionReasonInvalid {
		t.Fatalf("expected ErrCorrectionReasonInvalid, got %v", err)
	}
}

func TestValidateCorrectionReason_RejectsOtherWithWhitespaceOnlyNote(t *testing.T) {
	err := ValidateCorrectionReason("other", "   ")
	if err != ErrReasonNoteRequiredForOther {
		t.Fatalf("expected ErrReasonNoteRequiredForOther, got %v", err)
	}
}

func TestValidateCorrectionReason_RejectsOtherWithEmptyNote(t *testing.T) {
	err := ValidateCorrectionReason("other", "")
	if err != ErrReasonNoteRequiredForOther {
		t.Fatalf("expected ErrReasonNoteRequiredForOther, got %v", err)
	}
}

func TestValidateCorrectionReason_AcceptsMissedCheckIn(t *testing.T) {
	if err := ValidateCorrectionReason("missed_check_in", ""); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateCorrectionReason_AcceptsMissedCheckOut(t *testing.T) {
	if err := ValidateCorrectionReason("missed_check_out", ""); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateCorrectionReason_AcceptsIncorrectTime(t *testing.T) {
	if err := ValidateCorrectionReason("incorrect_time", ""); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateCorrectionReason_AcceptsDuplicateEntry(t *testing.T) {
	if err := ValidateCorrectionReason("duplicate_entry", ""); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateCorrectionReason_AcceptsOtherWithNote(t *testing.T) {
	if err := ValidateCorrectionReason("other", "explanation text"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateCorrectionReason_RejectsLifecycleReasonCode(t *testing.T) {
	err := ValidateCorrectionReason("left_nursery", "")
	if err != ErrCorrectionReasonInvalid {
		t.Fatalf("expected ErrCorrectionReasonInvalid for lifecycle code, got %v", err)
	}
}
