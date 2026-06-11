package domain

import "errors"

var (
	ErrSiteNotFound       = errors.New("site not found")
	ErrSiteInactive       = errors.New("site is inactive")
	ErrMembershipNotFound = errors.New("manager membership not found")
	ErrUserInactive       = errors.New("user inactive")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }
