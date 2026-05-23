package domain

import "strings"

type CorrectionReasonCode string

const (
	CorrectionReasonMissedCheckIn  CorrectionReasonCode = "missed_check_in"
	CorrectionReasonMissedCheckOut CorrectionReasonCode = "missed_check_out"
	CorrectionReasonIncorrectTime  CorrectionReasonCode = "incorrect_time"
	CorrectionReasonDuplicateEntry CorrectionReasonCode = "duplicate_entry"
	CorrectionReasonOther          CorrectionReasonCode = "other"
)

var validCorrectionReasons = map[CorrectionReasonCode]struct{}{
	CorrectionReasonMissedCheckIn:  {},
	CorrectionReasonMissedCheckOut: {},
	CorrectionReasonIncorrectTime:  {},
	CorrectionReasonDuplicateEntry: {},
	CorrectionReasonOther:          {},
}

// ValidateCorrectionReason validates the reason code and note for an attendance correction event.
func ValidateCorrectionReason(code string, note string) error {
	if strings.TrimSpace(code) == "" {
		return ErrCorrectionReasonRequired
	}

	reason := CorrectionReasonCode(code)
	if _, ok := validCorrectionReasons[reason]; !ok {
		return ErrCorrectionReasonInvalid
	}

	if reason == CorrectionReasonOther && strings.TrimSpace(note) == "" {
		return ErrReasonNoteRequiredForOther
	}

	return nil
}
