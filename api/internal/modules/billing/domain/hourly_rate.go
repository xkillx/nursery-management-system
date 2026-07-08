package domain

import "fmt"

// HourlyRate is a value object representing a rate in pence per hour.
// It wraps Money to reuse its arithmetic and JSON marshaling.
type HourlyRate struct {
	money Money
}

// NewHourlyRate creates an HourlyRate from a minor-unit amount (pence per hour).
// Returns an error if the value is negative.
func NewHourlyRate(minor int) (HourlyRate, error) {
	if minor < 0 {
		return HourlyRate{}, fmt.Errorf("hourly rate must not be negative: %d", minor)
	}
	m, err := GBP(minor)
	if err != nil {
		return HourlyRate{}, err
	}
	return HourlyRate{money: m}, nil
}

// Minor returns the rate in pence per hour.
func (r HourlyRate) Minor() int {
	return r.money.Minor()
}

// PerMinute returns the per-minute rate as Money, using ceiling division.
func (r HourlyRate) PerMinute() Money {
	minor := r.money.Minor()
	if minor == 0 {
		return Money{minor: 0}
	}
	perMinute := (minor + 59) / 60
	return Money{minor: perMinute}
}

// Multiply returns the total amount for the given number of minutes.
func (r HourlyRate) Multiply(minutes int) Money {
	perMinute := r.PerMinute()
	return perMinute.Multiply(minutes)
}

// MarshalJSON delegates to the underlying Money value.
func (r HourlyRate) MarshalJSON() ([]byte, error) {
	return r.money.MarshalJSON()
}

// UnmarshalJSON delegates to the underlying Money value.
func (r *HourlyRate) UnmarshalJSON(data []byte) error {
	return r.money.UnmarshalJSON(data)
}

// String returns a human-readable representation.
func (r HourlyRate) String() string {
	return fmt.Sprintf("%s/hour", r.money.String())
}
