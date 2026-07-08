package domain

import (
	"encoding/json"
	"testing"
)

func TestNewHourlyRate(t *testing.T) {
	t.Run("positive value", func(t *testing.T) {
		r, err := NewHourlyRate(500)
		if err != nil {
			t.Fatalf("NewHourlyRate(500) unexpected error: %v", err)
		}
		if r.Minor() != 500 {
			t.Errorf("got %d, want 500", r.Minor())
		}
	})

	t.Run("zero value", func(t *testing.T) {
		r, err := NewHourlyRate(0)
		if err != nil {
			t.Fatalf("NewHourlyRate(0) unexpected error: %v", err)
		}
		if r.Minor() != 0 {
			t.Errorf("got %d, want 0", r.Minor())
		}
	})

	t.Run("negative value returns error", func(t *testing.T) {
		_, err := NewHourlyRate(-100)
		if err == nil {
			t.Fatal("expected error for negative rate")
		}
	})
}

func TestHourlyRatePerMinute(t *testing.T) {
	tests := []struct {
		name  string
		minor int
		want  int
	}{
		{name: "exact hour", minor: 500, want: 9},     // ceil(500/60) = 9
		{name: "zero rate", minor: 0, want: 0},
		{name: "one pence", minor: 1, want: 1},         // ceil(1/60) = 1
		{name: "60 pence", minor: 60, want: 1},         // ceil(60/60) = 1
		{name: "61 pence", minor: 61, want: 2},         // ceil(61/60) = 2
		{name: "500 pence", minor: 500, want: 9},       // ceil(500/60) = 9
		{name: "120 pence", minor: 120, want: 2},       // ceil(120/60) = 2
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewHourlyRate(tc.minor)
			if err != nil {
				t.Fatalf("NewHourlyRate(%d) unexpected error: %v", tc.minor, err)
			}
			got := r.PerMinute()
			if got.Minor() != tc.want {
				t.Errorf("PerMinute() = %d, want %d", got.Minor(), tc.want)
			}
		})
	}
}

func TestHourlyRateMultiply(t *testing.T) {
	tests := []struct {
		name    string
		minor   int
		minutes int
		want    int
	}{
		{name: "60 min at 500/hr", minor: 500, minutes: 60, want: 540}, // ceil(500/60)=9, 9*60=540
		{name: "zero minutes", minor: 500, minutes: 0, want: 0},
		{name: "zero rate", minor: 0, minutes: 60, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewHourlyRate(tc.minor)
			if err != nil {
				t.Fatalf("NewHourlyRate(%d) unexpected error: %v", tc.minor, err)
			}
			got := r.Multiply(tc.minutes)
			if got.Minor() != tc.want {
				t.Errorf("Multiply(%d) = %d, want %d", tc.minutes, got.Minor(), tc.want)
			}
		})
	}
}

func TestHourlyRateMultiply60MinAt500(t *testing.T) {
	r, _ := NewHourlyRate(500)
	got := r.Multiply(60)
	// PerMinute = ceil(500/60) = 9, Multiply(60) = 9*60 = 540
	if got.Minor() != 540 {
		t.Errorf("got %d, want 540", got.Minor())
	}
}

func TestHourlyRateJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		r, _ := NewHourlyRate(500)
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if string(b) != "500" {
			t.Errorf("got %s, want 500", string(b))
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		var r HourlyRate
		err := json.Unmarshal([]byte("200"), &r)
		if err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if r.Minor() != 200 {
			t.Errorf("got %d, want 200", r.Minor())
		}
	})
}

func TestHourlyRateString(t *testing.T) {
	r, _ := NewHourlyRate(500)
	got := r.String()
	want := "GBP 5.00/hour"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
