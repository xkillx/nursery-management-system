package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrTokenInvalid = domainerrors.New("password_reset_token_invalid", "Password reset token is invalid")
	ErrTokenExpired = domainerrors.New("password_reset_token_expired", "Password reset token has expired")
	ErrTokenUsed    = domainerrors.New("password_reset_token_used", "Password reset token has already been used")
	ErrUserNotFound = domainerrors.NotFound("password_reset_user", "User not found")
	ErrUserInactive = domainerrors.New("password_reset_user_inactive", "User is inactive")
)
