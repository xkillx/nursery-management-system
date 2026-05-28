package domain

import "fmt"

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
