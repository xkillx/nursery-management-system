package domain

import (
	"fmt"
	"strings"
)

// EmailAddress is a validated, normalized email value object.
// Construction enforces format rules and normalizes to lowercase.
type EmailAddress struct {
	value string
}

// NewEmailAddress creates a validated, normalized EmailAddress.
// It trims whitespace, validates format (has @, domain part, no whitespace),
// and lowercases both local and domain parts.
func NewEmailAddress(raw string) (EmailAddress, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return EmailAddress{}, fmt.Errorf("email address must not be empty")
	}
	if strings.ContainsAny(trimmed, " \t\n\r") {
		return EmailAddress{}, fmt.Errorf("email address must not contain whitespace")
	}
	at := strings.Index(trimmed, "@")
	if at < 1 {
		return EmailAddress{}, fmt.Errorf("email address must contain @ with a local part")
	}
	if at == len(trimmed)-1 {
		return EmailAddress{}, fmt.Errorf("email address must have a domain part after @")
	}
	local := strings.ToLower(trimmed[:at])
	domain := strings.ToLower(trimmed[at+1:])
	if !strings.Contains(domain, ".") {
		return EmailAddress{}, fmt.Errorf("email domain must contain at least one dot")
	}
	return EmailAddress{value: local + "@" + domain}, nil
}

// String returns the normalized email address.
func (e EmailAddress) String() string {
	return e.value
}

// LocalPart returns the part before @.
func (e EmailAddress) LocalPart() string {
	at := strings.Index(e.value, "@")
	if at < 0 {
		return ""
	}
	return e.value[:at]
}

// DomainPart returns the part after @.
func (e EmailAddress) DomainPart() string {
	at := strings.Index(e.value, "@")
	if at < 0 {
		return ""
	}
	return e.value[at+1:]
}

// MarshalJSON returns the email as a JSON string.
func (e EmailAddress) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%q", e.value), nil
}

// UnmarshalJSON reads a JSON string into an EmailAddress.
func (e *EmailAddress) UnmarshalJSON(data []byte) error {
	var s string
	_, err := fmt.Sscanf(string(data), "%q", &s)
	if err != nil {
		return fmt.Errorf("invalid email JSON: %w", err)
	}
	parsed, err := NewEmailAddress(s)
	if err != nil {
		return err
	}
	*e = parsed
	return nil
}
