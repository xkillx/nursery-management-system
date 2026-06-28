package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrRoleNotAllowed         = domainerrors.New("invite_role_not_allowed", "Role not allowed for this invite")
	ErrEmailAlreadyRegistered = domainerrors.Conflict("invite_email_already_registered", "Email already registered")
	ErrScopeConflict          = domainerrors.Conflict("invite_scope_conflict", "Invite scope conflict")
	ErrInviteNotFound         = domainerrors.NotFound("invite", "Invite not found")
	ErrInviteNotPending       = domainerrors.Conflict("invite_not_pending", "Invite is not pending")
	ErrInviteAccepted         = domainerrors.Conflict("invite_already_accepted", "Invite already accepted")
	ErrTokenInvalid           = domainerrors.New("invite_token_invalid", "Invite token invalid")
	ErrTokenExpired           = domainerrors.New("invite_token_expired", "Invite token expired")
	ErrTokenRevoked           = domainerrors.New("invite_token_revoked", "Invite token revoked")
	ErrTokenAccepted          = domainerrors.New("invite_token_accepted", "Invite token already accepted")
)
