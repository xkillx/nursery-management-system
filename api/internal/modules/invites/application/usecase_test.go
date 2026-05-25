package application

import "testing"

func TestParseStatusDefaults(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"", "pending", true},
		{"pending", "pending", true},
		{"accepted", "accepted", true},
		{"revoked", "revoked", true},
		{"expired", "expired", true},
		{"all", "all", true},
		{"invalid", "", false},
	}

	for _, tc := range tests {
		status, ok := ParseStatus(tc.input)
		if ok != tc.ok {
			t.Errorf("ParseStatus(%q) ok = %v, want %v", tc.input, ok, tc.ok)
		}
		if ok && string(status) != tc.want {
			t.Errorf("ParseStatus(%q) = %q, want %q", tc.input, status, tc.want)
		}
	}
}
