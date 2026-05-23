package httpattendance

import (
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/domain"
)

type checkInRequest struct {
	ChildID string `json:"child_id" binding:"required"`
}

type checkOutRequest struct {
	ChildID string `json:"child_id" binding:"required"`
}

type sessionResponse struct {
	ID               string  `json:"id"`
	ChildID          string  `json:"child_id"`
	Status           string  `json:"status"`
	CheckInAt        string  `json:"check_in_at"`
	CheckOutAt       *string `json:"check_out_at"`
	CheckInLocalDate string  `json:"check_in_local_date"`
	CheckOutLocalDate *string `json:"check_out_local_date"`
	DurationMinutes  *int    `json:"duration_minutes"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func toSessionResponse(s domain.Session) sessionResponse {
	resp := sessionResponse{
		ID:               s.ID.String(),
		ChildID:          s.ChildID.String(),
		Status:           string(s.Status),
		CheckInAt:        s.CheckInAt.UTC().Format(time.RFC3339),
		CheckInLocalDate: s.CheckInLocalDate.Format("2006-01-02"),
		CreatedAt:        s.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        s.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if s.CheckOutAt != nil {
		v := s.CheckOutAt.UTC().Format(time.RFC3339)
		resp.CheckOutAt = &v
	}
	if s.CheckOutLocalDate != nil {
		v := s.CheckOutLocalDate.Format("2006-01-02")
		resp.CheckOutLocalDate = &v
	}
	if s.DurationMinutes != nil {
		resp.DurationMinutes = s.DurationMinutes
	}
	return resp
}

func parseChildID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}
