package domain

import (
	"encoding/json"
	"testing"
)

func TestNewEmailAddress(t *testing.T) {
	t.Run("normalizes to lowercase", func(t *testing.T) {
		e, err := NewEmailAddress("User@Example.COM")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.String() != "user@example.com" {
			t.Errorf("got %q, want %q", e.String(), "user@example.com")
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		e, err := NewEmailAddress(" user@ex.com ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.String() != "user@ex.com" {
			t.Errorf("got %q, want %q", e.String(), "user@ex.com")
		}
	})

	t.Run("local and domain parts", func(t *testing.T) {
		e, _ := NewEmailAddress("user@example.com")
		if e.LocalPart() != "user" {
			t.Errorf("LocalPart = %q, want %q", e.LocalPart(), "user")
		}
		if e.DomainPart() != "example.com" {
			t.Errorf("DomainPart = %q, want %q", e.DomainPart(), "example.com")
		}
	})
}

func TestNewEmailAddressErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "whitespace only", input: "   "},
		{name: "no at sign", input: "no-at-sign"},
		{name: "at start", input: "@no-local"},
		{name: "at end", input: "no-domain@"},
		{name: "no domain dot", input: "user@nodot"},
		{name: "internal whitespace", input: "user @example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewEmailAddress(tc.input)
			if err == nil {
				t.Fatalf("expected error for %q", tc.input)
			}
		})
	}
}

func TestEmailAddressJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		e, _ := NewEmailAddress("user@example.com")
		b, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if string(b) != `"user@example.com"` {
			t.Errorf("got %s, want %q", string(b), "user@example.com")
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		var e EmailAddress
		err := json.Unmarshal([]byte(`"User@Example.COM"`), &e)
		if err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if e.String() != "user@example.com" {
			t.Errorf("got %q, want %q", e.String(), "user@example.com")
		}
	})
}
