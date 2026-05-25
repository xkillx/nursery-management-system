package domain

import "errors"

var (
	ErrTokenInvalid  = errors.New("password reset token is invalid")
	ErrTokenExpired  = errors.New("password reset token has expired")
	ErrTokenUsed     = errors.New("password reset token has already been used")
	ErrUserNotFound  = errors.New("user not found")
	ErrUserInactive  = errors.New("user is inactive")
)
