package domain

import (
	"strings"
	"testing"
)

func TestSetLineDescriptionOverride_PreservesExistingDetails(t *testing.T) {
	details := []byte(`{"booked_core_minutes":480,"booked_sessions":[{"day":"mon","minutes":300}]}`)
	out, err := SetLineDescriptionOverride(details)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !HasLineDescriptionOverride(out) {
		t.Errorf("expected description_override marker, got %s", out)
	}
	for _, key := range []string{"booked_core_minutes", "booked_sessions"} {
		if !strings.Contains(string(out), key) {
			t.Errorf("expected existing key %q preserved in %s", key, out)
		}
	}
}

func TestSetLineDescriptionOverride_NilDetails(t *testing.T) {
	out, err := SetLineDescriptionOverride(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !HasLineDescriptionOverride(out) {
		t.Errorf("expected description_override marker on empty details, got %s", out)
	}
}

func TestHasLineDescriptionOverride_EmptyAndMissing(t *testing.T) {
	if HasLineDescriptionOverride(nil) {
		t.Error("expected false for nil details")
	}
	if HasLineDescriptionOverride([]byte(`{"booked_core_minutes":480}`)) {
		t.Error("expected false when marker absent")
	}
	if HasLineDescriptionOverride([]byte(`{"description_override":false}`)) {
		t.Error("expected false when marker is false")
	}
	if HasLineDescriptionOverride([]byte("not-json")) {
		t.Error("expected false for malformed details")
	}
}

func TestSetLineDescriptionOverride_StickyRoundTrip(t *testing.T) {
	// KTD2: renaming back to the default keeps the marker (sticky).
	first, err := SetLineDescriptionOverride([]byte(`{"booked_sessions":[]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := SetLineDescriptionOverride(first)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !HasLineDescriptionOverride(second) {
		t.Errorf("expected marker to remain after re-merge, got %s", second)
	}
}
