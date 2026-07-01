package domain

import (
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type FieldError struct {
	Field   string
	Message string
}

func ValidationError(fields []FieldError) *domainerrors.DomainError {
	errFields := make([]domainerrors.FieldError, 0, len(fields))
	for _, f := range fields {
		errFields = append(errFields, domainerrors.FieldError{
			Field:   f.Field,
			Message: f.Message,
		})
	}
	return domainerrors.ValidationWithFields("Validation failed.", errFields)
}
