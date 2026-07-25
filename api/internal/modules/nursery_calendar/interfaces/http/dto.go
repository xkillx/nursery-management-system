package http

import (
	"time"

	"github.com/google/uuid"
)

type calendarDayResponse struct {
	Date   string `json:"date"`
	IsOpen bool   `json:"is_open"`
	Reason string `json:"reason"`
}

type dateRangeQueryParams struct {
	From       string `form:"from"`
	To         string `form:"to"`
	IsTermTime bool   `form:"is_term_time"`
}

func toCalendarDayResponse(date time.Time, isOpen bool, reason string) calendarDayResponse {
	return calendarDayResponse{
		Date:   date.Format("2006-01-02"),
		IsOpen: isOpen,
		Reason: reason,
	}
}

func parseDatePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
