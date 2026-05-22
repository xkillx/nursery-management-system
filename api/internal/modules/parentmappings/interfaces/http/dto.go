package httpmapping

import (
	"time"

	"nursery-management-system/api/internal/modules/parentmappings/domain"
)

type parentMappingResponse struct {
	ID              string  `json:"id"`
	MembershipID    string  `json:"membership_id"`
	GuardianID      string  `json:"guardian_id"`
	EndedAt         *string `json:"ended_at,omitempty"`
	EndedReasonCode *string `json:"ended_reason_code,omitempty"`
	EndedReasonNote *string `json:"ended_reason_note,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func toMappingResponse(m domain.ParentMapping) parentMappingResponse {
	resp := parentMappingResponse{
		ID:              m.ID.String(),
		MembershipID:    m.MembershipID.String(),
		GuardianID:      m.GuardianID.String(),
		EndedReasonCode: m.EndedReasonCode,
		EndedReasonNote: m.EndedReasonNote,
		CreatedAt:       m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if m.EndedAt != nil {
		v := m.EndedAt.UTC().Format(time.RFC3339)
		resp.EndedAt = &v
	}
	return resp
}
