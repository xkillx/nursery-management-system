package domain

import "errors"

var (
	ErrRoleNotAllowed         = errors.New("invite role not allowed")
	ErrEmailAlreadyRegistered = errors.New("invite email already registered")
	ErrScopeConflict          = errors.New("invite scope conflict")
	ErrInviteNotFound         = errors.New("invite not found")
	ErrInviteNotPending       = errors.New("invite not pending")
	ErrInviteAccepted         = errors.New("invite already accepted")
	ErrTokenInvalid           = errors.New("invite token invalid")
	ErrTokenExpired           = errors.New("invite token expired")
	ErrTokenRevoked           = errors.New("invite token revoked")
	ErrTokenAccepted          = errors.New("invite token already accepted")
)
