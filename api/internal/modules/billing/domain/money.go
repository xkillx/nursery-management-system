package domain

import (
	"fmt"
)

type Money struct {
	minor int
}

func GBP(minor int) (Money, error) {
	return Money{minor: minor}, nil
}

func MustGBP(minor int) Money {
	m, err := GBP(minor)
	if err != nil {
		panic(err)
	}
	return m
}

func (m Money) Minor() int {
	return m.minor
}

func (m Money) Add(other Money) Money {
	return Money{minor: m.minor + other.minor}
}

func (m Money) Multiply(factor int) Money {
	return Money{minor: m.minor * factor}
}

func (m Money) String() string {
	pounds := m.minor / 100
	pence := m.minor % 100
	return fmt.Sprintf("GBP %d.%02d", pounds, pence)
}

func (m Money) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%d", m.minor), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var v int
	_, err := fmt.Sscanf(string(data), "%d", &v)
	if err != nil {
		return fmt.Errorf("invalid Money value: %w", err)
	}
	m.minor = v
	return nil
}

func CalculateHourlyAmountMinor(minutes int, hourlyRateMinor int) (int, error) {
	if minutes < 0 || hourlyRateMinor < 0 {
		return 0, fmt.Errorf("minutes and hourly rate must not be negative")
	}
	numerator := minutes * hourlyRateMinor
	if numerator == 0 {
		return 0, nil
	}
	return (numerator + 59) / 60, nil
}
