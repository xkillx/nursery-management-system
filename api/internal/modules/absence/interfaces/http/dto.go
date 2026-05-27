package httpabsence

import (
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/absence/domain"
)

type markAbsentRequest struct {
	ChildID string `json:"child_id" binding:"required"`
}

type absenceMarkerResponse struct {
	ID        string  `json:"id"`
	ChildID   string  `json:"child_id"`
	LocalDate string  `json:"local_date"`
	MarkedAt  string  `json:"marked_at"`
	ClearedAt *string `json:"cleared_at"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func toMarkerResponse(m domain.AbsenceMarker) absenceMarkerResponse {
	resp := absenceMarkerResponse{
		ID:        m.ID.String(),
		ChildID:   m.ChildID.String(),
		LocalDate: m.LocalDate.Format("2006-01-02"),
		MarkedAt:  m.MarkedAt.UTC().Format(time.RFC3339),
		CreatedAt: m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if m.ClearedAt != nil {
		at := m.ClearedAt.UTC().Format(time.RFC3339)
		resp.ClearedAt = &at
	}
	return resp
}

func parseMarkerID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}
