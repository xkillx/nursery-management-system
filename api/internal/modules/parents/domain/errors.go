package domain

import (
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

var (
	ErrParentNotFound    = domainerrors.NotFound("parent", "Parent not found")
	ErrParentInactive    = domainerrors.New("parent_inactive", "Parent is inactive")
	ErrLinkNotFound      = domainerrors.NotFound("parent_child_link", "Parent-child link not found")
	ErrLinkAlreadyExists = domainerrors.New("parent_child_link_exists", "Parent is already linked to this child")
	ErrChildNotFound     = domainerrors.NotFound("child", "Child not found")
	ErrUserAlreadyLinked = domainerrors.New("user_already_linked", "User account is already linked to another parent")
)
