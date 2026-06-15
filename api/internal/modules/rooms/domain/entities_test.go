package domain_test

import (
	"testing"

	"nursery-management-system/api/internal/modules/rooms/domain"
)

func TestIsValidAgeGroup(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"baby", true},
		{"toddler", true},
		{"preschool", true},
		{"mixed", true},
		{"invalid", false},
		{"", false},
		{"BABY", false},
		{" pre-school ", false},
	}

	for _, tt := range tests {
		got := domain.IsValidAgeGroup(tt.input)
		if got != tt.want {
			t.Errorf("IsValidAgeGroup(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
