package domain

import "testing"

func TestCalculateHourlyAmountMinor(t *testing.T) {
	tests := []struct {
		name           string
		minutes        int
		hourlyRateMinor int
		want           int
		wantErr        bool
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
