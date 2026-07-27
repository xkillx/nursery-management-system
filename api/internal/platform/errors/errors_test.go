package errors

import (
	"fmt"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "NotFound with entity prefix matches",
			err:  NotFound("site_profile", "Site profile not found."),
			want: true,
		},
		{
			name: "NotFound with invoice entity matches",
			err:  NotFound("invoice", "Invoice not found."),
			want: true,
		},
		{
			name: "exact not_found code matches",
			err:  &DomainError{Code: "not_found", Message: "Not found."},
			want: true,
		},
		{
			name: "validation_error does not match",
			err:  Validation("Invalid input.", "field"),
			want: false,
		},
		{
			name: "internal_error does not match",
			err:  Internal(nil),
			want: false,
		},
		{
			name: "non-DomainError does not match",
			err:  fmt.Errorf("some error"),
			want: false,
		},
		{
			name: "conflict code does not match",
			err:  Conflict("duplicate_child_month", "Duplicate."),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
