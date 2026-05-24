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

type correctionRequest struct {
	SessionID  *string `json:"session_id"`
	ChildID    *string `json:"child_id"`
	CheckInAt  string  `json:"check_in_at" binding:"required"`
	CheckOutAt string  `json:"check_out_at" binding:"required"`
	ReasonCode string  `json:"reason_code" binding:"required"`
	ReasonNote string  `json:"reason_note"`
}

type sessionResponse struct {
	ID                string  `json:"id"`
	ChildID           string  `json:"child_id"`
	Status            string  `json:"status"`
	CheckInAt         string  `json:"check_in_at"`
	CheckOutAt        *string `json:"check_out_at"`
	CheckInLocalDate  string  `json:"check_in_local_date"`
	CheckOutLocalDate *string `json:"check_out_local_date"`
	DurationMinutes   *int    `json:"duration_minutes"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
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

func parseCorrectionRequest(req correctionRequest) (domain.CorrectionParams, error) {
	var params domain.CorrectionParams

	if req.SessionID != nil && *req.SessionID != "" {
		id, err := uuid.Parse(*req.SessionID)
		if err != nil {
			return domain.CorrectionParams{}, err
		}
		params.SessionID = &id
	}

	if req.ChildID != nil && *req.ChildID != "" {
		id, err := uuid.Parse(*req.ChildID)
		if err != nil {
			return domain.CorrectionParams{}, err
		}
		params.ChildID = &id
	}

	checkInAt, err := time.Parse(time.RFC3339, req.CheckInAt)
	if err != nil {
		return domain.CorrectionParams{}, err
	}
	params.CheckInAt = checkInAt.UTC()

	checkOutAt, err := time.Parse(time.RFC3339, req.CheckOutAt)
	if err != nil {
		return domain.CorrectionParams{}, err
	}
	params.CheckOutAt = checkOutAt.UTC()

	params.ReasonCode = req.ReasonCode
	params.ReasonNote = req.ReasonNote

	return params, nil
}
