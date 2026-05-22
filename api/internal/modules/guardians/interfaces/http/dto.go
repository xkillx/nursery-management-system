package httpguardian

import (
	"time"

	"nursery-management-system/api/internal/modules/guardians/domain"
)

type guardianResponse struct {
	ID                     string  `json:"id"`
	FullName               string  `json:"full_name"`
	Email                  *string `json:"email,omitempty"`
	Phone                  *string `json:"phone,omitempty"`
	Notes                  *string `json:"notes,omitempty"`
	IsActive               bool    `json:"is_active"`
	DeactivatedAt          *string `json:"deactivated_at,omitempty"`
	DeactivationReasonCode *string `json:"deactivation_reason_code,omitempty"`
	DeactivationReasonNote *string `json:"deactivation_reason_note,omitempty"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

type guardianWriteRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Notes    string `json:"notes"`
}

type reasonRequest struct {
	ReasonCode string `json:"reason_code"`
	ReasonNote string `json:"reason_note"`
}

func toGuardianResponse(g domain.Guardian) guardianResponse {
	resp := guardianResponse{
		ID:                     g.ID.String(),
		FullName:               g.FullName,
		Email:                  g.Email,
		Phone:                  g.Phone,
		Notes:                  g.Notes,
		IsActive:               g.IsActive,
		DeactivationReasonCode: g.DeactivationReasonCode,
		DeactivationReasonNote: g.DeactivationReasonNote,
		CreatedAt:              g.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:              g.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if g.DeactivatedAt != nil {
		v := g.DeactivatedAt.UTC().Format(time.RFC3339)
		resp.DeactivatedAt = &v
	}
	return resp
}
