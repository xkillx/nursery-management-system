package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrNotFound           = domainerrors.NotFound("not_found", "Not found")
	ErrInvalidCredentials = domainerrors.Unauthorized("Invalid credentials or session.")
	ErrInvalidToken       = domainerrors.Unauthorized("Invalid credentials or session.")
	ErrInvalidMembership  = domainerrors.Forbidden("forbidden_scope_selection", "Invalid membership selection.")
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
