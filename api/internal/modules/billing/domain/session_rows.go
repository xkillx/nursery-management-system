package domain

import (
	"encoding/json"
	"sort"
	"time"
)

// SessionRow is a display-ready per-session breakdown row for a core
// childcare line. It is snake_case-tagged for the API/PDF consumers (KTD3),
// distinct from the untagged storage keys of BookedSession. It carries no
// unit field — the hourly-rate Unit cell is supplied by the caller from the
// line's unit_amount_minor.
type SessionRow struct {
	OccurrenceDate     time.Time `json:"occurrence_date"`
	StartMinutes       int       `json:"start_minutes"`
	EndMinutes         int       `json:"end_minutes"`
	DurationMinutes    int       `json:"duration_minutes"`
	SessionTypeName    string    `json:"session_type_name"`
	SessionAmountMinor int       `json:"session_amount_minor"`
}

// AllocateLineAmount distributes totalMinor across weights using the
// largest-remainder method, so the per-row amounts always sum exactly to the
// line total (R7). A total <= 0 yields all-zero amounts; negative weights are
// treated as absolute values.
func AllocateLineAmount(totalMinor int, weights []int) []int {
	result := make([]int, len(weights))
	if len(weights) == 0 || totalMinor <= 0 {
		return result
	}

	absWeights := make([]int, len(weights))
	totalWeight := 0
	for i, w := range weights {
		if w < 0 {
			w = -w
		}
		absWeights[i] = w
		totalWeight += w
	}
	if totalWeight == 0 {
		return result
	}

	allocated := 0
	type remainder struct {
		index int
		value int
	}
	remainders := make([]remainder, len(weights))
	for i, w := range absWeights {
		share := totalMinor * w / totalWeight
		result[i] = share
		allocated += share
		remainders[i] = remainder{index: i, value: totalMinor*w - share*totalWeight}
	}

	// Distribute the leftover pennies to the rows with the largest remainders.
	// leftover < len(weights) always holds because each fractional part is < 1.
	leftover := totalMinor - allocated
	sort.SliceStable(remainders, func(i, j int) bool {
		return remainders[i].value > remainders[j].value
	})
	for i := 0; i < leftover && i < len(remainders); i++ {
		result[remainders[i].index]++
	}

	return result
}

// BuildSessionRows turns a core line's persisted details into display-ready
// per-session rows: parse -> chronological sort -> drop invalid dates ->
// allocate the line total (R6, R14). An empty, malformed, session-less or
// non-core details yields nil so the caller falls back to the single
// aggregate row.
func BuildSessionRows(details json.RawMessage, lineKind string, quantityMinutes *int, lineAmountMinor int) []SessionRow {
	if lineKind != LineKindCoreChildcare {
		return nil
	}

	var coreLine CoreLineDetails
	if err := json.Unmarshal(details, &coreLine); err != nil {
		return nil
	}

	sessions := make([]BookedSession, 0, len(coreLine.BookedSessions))
	for _, s := range coreLine.BookedSessions {
		if s.OccurrenceDate.IsZero() {
			continue
		}
		sessions = append(sessions, s)
	}
	if len(sessions) == 0 {
		return nil
	}

	// Chronological sort: occurrence date, then start time ascending; rows
	// lacking a start time sort after timed rows on the same date, preserving
	// stored order among themselves (R5).
	sort.SliceStable(sessions, func(i, j int) bool {
		if !sessions[i].OccurrenceDate.Equal(sessions[j].OccurrenceDate) {
			return sessions[i].OccurrenceDate.Before(sessions[j].OccurrenceDate)
		}
		ti := sessions[i].StartMinutes > 0
		tj := sessions[j].StartMinutes > 0
		if ti != tj {
			return ti
		}
		return sessions[i].StartMinutes < sessions[j].StartMinutes
	})

	weights := make([]int, len(sessions))
	for i, s := range sessions {
		weights[i] = s.DurationMinutes
	}
	allocated := AllocateLineAmount(lineAmountMinor, weights)

	rows := make([]SessionRow, len(sessions))
	for i, s := range sessions {
		rows[i] = SessionRow{
			OccurrenceDate:     s.OccurrenceDate,
			StartMinutes:       s.StartMinutes,
			EndMinutes:         s.EndMinutes,
			DurationMinutes:    s.DurationMinutes,
			SessionTypeName:    s.SessionTypeName,
			SessionAmountMinor: allocated[i],
		}
	}
	return rows
}
