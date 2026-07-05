package httpsessiontypes

import (
	"fmt"
	"time"

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
)

func minutesToHHMM(m int) string {
	hh := m / 60
	mm := m % 60
	return fmt.Sprintf("%02d:%02d", hh, mm)
}

type sessionTypeResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	IsActive     bool   `json:"is_active"`
	Kind         string `json:"kind"`
	FlatFeeMinor *int   `json:"flat_fee_minor"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type createSessionTypeRequest struct {
	Name         string `json:"name" binding:"required"`
	StartTime    string `json:"start_time" binding:"required"`
	EndTime      string `json:"end_time" binding:"required"`
	Kind         string `json:"kind"`
	FlatFeeMinor *int   `json:"flat_fee_minor"`
}

type updateSessionTypeRequest struct {
	Name         *string `json:"name"`
	StartTime    *string `json:"start_time"`
	EndTime      *string `json:"end_time"`
	Kind         *string `json:"kind"`
	FlatFeeMinor **int   `json:"flat_fee_minor"`
}

func toSessionTypeResponse(st domain.SessionType) sessionTypeResponse {
	return sessionTypeResponse{
		ID:           st.ID.String(),
		Name:         st.Name,
		StartTime:    minutesToHHMM(st.StartMinutes),
		EndTime:      minutesToHHMM(st.EndMinutes),
		IsActive:     st.IsActive,
		Kind:         st.Kind,
		FlatFeeMinor: st.FlatFeeMinor,
		CreatedAt:    st.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    st.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toSessionTypeListResponse(items []domain.SessionType) []sessionTypeResponse {
	out := make([]sessionTypeResponse, 0, len(items))
	for _, st := range items {
		out = append(out, toSessionTypeResponse(st))
	}
	return out
}
