package application

import (
	"fmt"
	"time"
)

func ParseBillingMonth(raw string) (time.Time, error) {
	parsed, err := time.Parse("2006-01", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid billing month format: %w", err)
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}
