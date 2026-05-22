package httplink

import (
	"time"

	"nursery-management-system/api/internal/modules/guardianlinks/domain"
)

type guardianChildLinkResponse struct {
	ID              string  `json:"id"`
	GuardianID      string  `json:"guardian_id"`
	ChildID         string  `json:"child_id"`
	EndedAt         *string `json:"ended_at,omitempty"`
	EndedReasonCode *string `json:"ended_reason_code,omitempty"`
	EndedReasonNote *string `json:"ended_reason_note,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func toLinkResponse(link domain.GuardianChildLink) guardianChildLinkResponse {
	resp := guardianChildLinkResponse{
		ID:              link.ID.String(),
		GuardianID:      link.GuardianID.String(),
		ChildID:         link.ChildID.String(),
		EndedReasonCode: link.EndedReasonCode,
		EndedReasonNote: link.EndedReasonNote,
		CreatedAt:       link.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       link.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if link.EndedAt != nil {
		v := link.EndedAt.UTC().Format(time.RFC3339)
		resp.EndedAt = &v
	}
	return resp
}
