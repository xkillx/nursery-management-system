package domain

import domainerrors "nursery-management-system/api/internal/platform/errors"

var (
	ErrChildNotFound = func() error { return domainerrors.NotFound("child", "Child not found.") }
)
