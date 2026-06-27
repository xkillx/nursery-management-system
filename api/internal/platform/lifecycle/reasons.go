package lifecycle

const (
	ReasonDuplicateRecord = "duplicate_record"
	ReasonEnteredInError  = "entered_in_error"
	ReasonLeftNursery     = "left_nursery"
	ReasonSafeguardingDir = "safeguarding_direction"
	ReasonContactUpdate   = "contact_update"
	ReasonAccessRevoked   = "access_revoked"
	ReasonOther           = "other"
	MaxReasonNoteLen      = 500
)

var ValidReasonCodes = map[string]struct{}{
	ReasonDuplicateRecord: {},
	ReasonEnteredInError:  {},
	ReasonLeftNursery:     {},
	ReasonSafeguardingDir: {},
	ReasonContactUpdate:   {},
	ReasonAccessRevoked:   {},
	ReasonOther:           {},
}

func IsValidReasonCode(code string) bool {
	_, ok := ValidReasonCodes[code]
	return ok
}
