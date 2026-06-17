package application

import (
	"strings"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

const (
	maxReasonNoteLen = 500
	defaultListLimit = 50
	maxListLimit     = 200
)

// ValidatePagination validates and normalizes pagination parameters.
func ValidatePagination(limit, offset int) (int, int, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		return 0, 0, domainerrors.Validation("Invalid request payload.", "limit")
	}
	if offset < 0 {
		return 0, 0, domainerrors.Validation("Invalid request payload.", "offset")
	}
	return limit, offset, nil
}

// ValidateStatusFilter returns a valid StatusFilter or a validation error.
func ValidateStatusFilter(v string) (domain.StatusFilter, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return domain.StatusActive, nil
	}
	switch domain.StatusFilter(v) {
	case domain.StatusActive, domain.StatusInactive, domain.StatusAll:
		return domain.StatusFilter(v), nil
	default:
		return "", domainerrors.Validation("Invalid request payload.", "status")
	}
}

// parseUUID validates and returns a uuid.UUID from a string.
func parseUUID(v string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(v))
}
