package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidMembership  = errors.New("invalid membership")
)

var ErrMalformedMembershipID = ValidationError{Field: "membership_id", Message: "membership_id must be a valid UUID"}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

type MembershipSelectionRequiredError struct {
	Memberships   []Membership
	IsStaleChoice bool
}

func (e *MembershipSelectionRequiredError) Error() string {
	if e.IsStaleChoice {
		return "membership selection required: previous choice no longer available"
	}
	return "membership selection required"
}
