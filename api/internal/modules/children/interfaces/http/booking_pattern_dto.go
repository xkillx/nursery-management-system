package httpchild

import (
	"fmt"
	"time"

	"nursery-management-system/api/internal/modules/children/domain"
)

type bookingPatternEntryInput struct {
	DayOfWeek     int    `json:"day_of_week"`
	SessionTypeID string `json:"session_type_id"`
}

type bookingPatternRequest struct {
	EffectiveFrom string                     `json:"effective_from" binding:"required"`
	Entries       []bookingPatternEntryInput `json:"entries" binding:"required"`
}

type bookingPatternUpdateRequest struct {
	EffectiveFrom *string                     `json:"effective_from"`
	Entries       *[]bookingPatternEntryInput `json:"entries"`
}

type sessionTypeRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	IsActive  bool   `json:"is_active"`
}

type bookedSessionResponse struct {
	DayOfWeek   int            `json:"day_of_week"`
	SessionType sessionTypeRef `json:"session_type"`
}

type bookingPatternResponse struct {
	ID            string                  `json:"id"`
	ChildID       string                  `json:"child_id"`
	EffectiveFrom string                  `json:"effective_from"`
	EffectiveTo   *string                 `json:"effective_to,omitempty"`
	IsCurrent     bool                    `json:"is_current"`
	CreatedAt     string                  `json:"created_at"`
	Entries       []bookedSessionResponse `json:"entries"`
}

func minutesToHHMMChild(m int) string {
	hh := m / 60
	mm := m % 60
	return fmt.Sprintf("%02d:%02d", hh, mm)
}

func toBookingPatternResponse(bp domain.BookingPattern) bookingPatternResponse {
	resp := bookingPatternResponse{
		ID:            bp.ID.String(),
		ChildID:       bp.ChildID.String(),
		EffectiveFrom: bp.EffectiveFrom.Format("2006-01-02"),
		EffectiveTo:   dateStringPtr(bp.EffectiveTo),
		IsCurrent:     bp.IsCurrent,
		CreatedAt:     bp.CreatedAt.UTC().Format(time.RFC3339),
		Entries:       make([]bookedSessionResponse, 0, len(bp.Entries)),
	}
	for _, e := range bp.Entries {
		bsr := bookedSessionResponse{DayOfWeek: e.DayOfWeek}
		if e.SessionType != nil {
			bsr.SessionType = sessionTypeRef{
				ID:        e.SessionType.ID.String(),
				Name:      e.SessionType.Name,
				StartTime: minutesToHHMMChild(e.SessionType.StartMinutes),
				EndTime:   minutesToHHMMChild(e.SessionType.EndMinutes),
				IsActive:  e.SessionType.IsActive,
			}
		}
		resp.Entries = append(resp.Entries, bsr)
	}
	return resp
}

func toBookingPatternListResponse(items []domain.BookingPattern) []bookingPatternResponse {
	out := make([]bookingPatternResponse, 0, len(items))
	for _, bp := range items {
		out = append(out, toBookingPatternResponse(bp))
	}
	return out
}
