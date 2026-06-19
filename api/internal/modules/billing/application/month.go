package application

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ParseBillingMonth parses a YYYY-MM string into the first day of that month
// at 00:00 UTC. Local London time is used elsewhere for display; this is the
// canonical monthly anchor.
func ParseBillingMonth(raw string) (time.Time, error) {
	parsed, err := time.Parse("2006-01", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid billing month format: %w", err)
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

// parseAndDedupeChildIDs parses raw UUID strings and returns a deduplicated slice.
func parseAndDedupeChildIDs(raw []string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	seen := make(map[uuid.UUID]struct{}, len(raw))
	result := make([]uuid.UUID, 0, len(raw))
	for _, r := range raw {
		id, err := uuid.Parse(r)
		if err != nil {
			return nil, fmt.Errorf("invalid child_id %q: %w", r, err)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}
