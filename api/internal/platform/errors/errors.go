package errors

import "fmt"

type DomainError struct {
	Code    string
	Message string
	Field   string
	Details map[string]any
	cause   error
}

func (e *DomainError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error { return e.cause }

func NotFound(entity, message string) *DomainError {
	return &DomainError{Code: entity + "_not_found", Message: message}
}

func AlreadyExists(entity, message string) *DomainError {
	return &DomainError{Code: entity + "_already_exists", Message: message}
}

func Validation(message string, field string) *DomainError {
	return &DomainError{Code: "validation_error", Message: message, Field: field}
}

func Unauthorized(message string) *DomainError {
	return &DomainError{Code: "unauthorized", Message: message}
}

func Forbidden(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func Conflict(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func ConflictWithDetails(code, message string, details map[string]any) *DomainError {
	return &DomainError{Code: code, Message: message, Details: details}
}

func Internal(err error) *DomainError {
	return &DomainError{Code: "internal_error", Message: "Something went wrong.", cause: err}
}

// New creates a DomainError with a specific code, message, and optional field.
func New(code, message string, field ...string) *DomainError {
	var f string
	if len(field) > 0 {
		f = field[0]
	}
	return &DomainError{Code: code, Message: message, Field: f}
}

func IsNotFound(err error) bool {
	if d, ok := err.(*DomainError); ok {
		return d.Code == "not_found"
	}
	return false
}

func IsValidation(err error) bool {
	if d, ok := err.(*DomainError); ok {
		return d.Code == "validation_error"
	}
	return false
}
