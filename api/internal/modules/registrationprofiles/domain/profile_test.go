package domain

import (
	"testing"
	"time"
)

func TestProfile_RegistrationDateRecorded(t *testing.T) {
	now := time.Now()

	cases := []struct {
		name string
		date *time.Time
		want bool
	}{
		{"nil date", nil, false},
		{"today", &now, true},
		{
			name: "30 days ago",
			date: ptrTime(now.AddDate(0, 0, -30)),
			want: true,
		},
		{
			name: "1 day in the future",
			date: ptrTime(now.AddDate(0, 0, 1)),
			want: false,
		},
		{
			name: "2 years ago",
			date: ptrTime(now.AddDate(-2, 0, 0)),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Profile{RegistrationDate: tc.date}
			if got := p.RegistrationDateRecorded(); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
