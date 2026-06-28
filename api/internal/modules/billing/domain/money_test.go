package domain

import (
	"encoding/json"
	"testing"
)

func TestGBP(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		m, err := GBP(0)
		if err != nil {
			t.Fatalf("GBP(0) unexpected error: %v", err)
		}
		if m.Minor() != 0 {
			t.Errorf("got %d, want 0", m.Minor())
		}
	})

	t.Run("positive value", func(t *testing.T) {
		m, err := GBP(100)
		if err != nil {
			t.Fatalf("GBP(100) unexpected error: %v", err)
		}
		if m.Minor() != 100 {
			t.Errorf("got %d, want 100", m.Minor())
		}
	})

	t.Run("negative value returns error", func(t *testing.T) {
		_, err := GBP(-1)
		if err == nil {
			t.Fatal("expected error for negative minor units")
		}
	})
}

func TestMoneyAdd(t *testing.T) {
	a, _ := GBP(100)
	b, _ := GBP(50)
	sum := a.Add(b)
	if sum.Minor() != 150 {
		t.Errorf("100 + 50 = %d, want 150", sum.Minor())
	}
}

func TestMoneyMultiply(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		m, _ := GBP(100)
		result := m.Multiply(3)
		if result.Minor() != 300 {
			t.Errorf("100 * 3 = %d, want 300", result.Minor())
		}
	})

	t.Run("zero", func(t *testing.T) {
		m, _ := GBP(0)
		result := m.Multiply(5)
		if result.Minor() != 0 {
			t.Errorf("0 * 5 = %d, want 0", result.Minor())
		}
	})
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		minor int
		want  string
	}{
		{0, "GBP 0.00"},
		{1, "GBP 0.01"},
		{100, "GBP 1.00"},
		{150, "GBP 1.50"},
		{12345, "GBP 123.45"},
	}
	for _, tc := range tests {
		m, _ := GBP(tc.minor)
		got := m.String()
		if got != tc.want {
			t.Errorf("GBP(%d).String() = %q, want %q", tc.minor, got, tc.want)
		}
	}
}

func TestMoneyJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		m, _ := GBP(150)
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if string(b) != "150" {
			t.Errorf("got %s, want 150", string(b))
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		var m Money
		err := json.Unmarshal([]byte("200"), &m)
		if err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if m.Minor() != 200 {
			t.Errorf("got %d, want 200", m.Minor())
		}
	})
}

func TestCalculateHourlyAmountMinor(t *testing.T) {
	tests := []struct {
		name            string
		minutes         int
		hourlyRateMinor int
		want            int
		wantErr         bool
	}{
		{name: "zero minutes", minutes: 0, hourlyRateMinor: 500, want: 0},
		{name: "zero rate", minutes: 60, hourlyRateMinor: 0, want: 0},
		{name: "exact hour", minutes: 60, hourlyRateMinor: 500, want: 500},
		{name: "exact half hour", minutes: 30, hourlyRateMinor: 500, want: 250},
		{name: "ceiling 1 min at 500 rate", minutes: 1, hourlyRateMinor: 500, want: 9},
		{name: "ceiling 61 min at 500 rate", minutes: 61, hourlyRateMinor: 500, want: 509},
		{name: "ceiling 45 min at 500 rate", minutes: 45, hourlyRateMinor: 500, want: 375},
		{name: "ceiling 46 min at 500 rate", minutes: 46, hourlyRateMinor: 500, want: 384},
		{name: "negative minutes", minutes: -1, hourlyRateMinor: 500, wantErr: true},
		{name: "negative rate", minutes: 60, hourlyRateMinor: -1, wantErr: true},
		{name: "both zero", minutes: 0, hourlyRateMinor: 0, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CalculateHourlyAmountMinor(tc.minutes, tc.hourlyRateMinor)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}
