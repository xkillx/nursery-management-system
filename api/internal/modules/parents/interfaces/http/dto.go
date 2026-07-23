package httpparents

import (
	"time"

	"nursery-management-system/api/internal/modules/parents/domain"
)

type parentResponse struct {
	ID                        string  `json:"id"`
	FirstName                 string  `json:"first_name"`
	LastName                  *string `json:"last_name,omitempty"`
	Email                     *string `json:"email,omitempty"`
	Phone                     *string `json:"phone,omitempty"`
	AddressLine1              *string `json:"address_line1,omitempty"`
	AddressLine2              *string `json:"address_line2,omitempty"`
	AddressCity               *string `json:"address_city,omitempty"`
	AddressPostcode           *string `json:"address_postcode,omitempty"`
	RelationshipToChild       *string `json:"relationship_to_child,omitempty"`
	HasParentalResponsibility bool    `json:"has_parental_responsibility"`
	CanPickUp                 bool    `json:"can_pick_up"`
	IsEmergencyContact        bool    `json:"is_emergency_contact"`
	Notes                     *string `json:"notes,omitempty"`
	UserID                    *string `json:"user_id,omitempty"`
	IsActive                  bool    `json:"is_active"`
	CreatedAt                 string  `json:"created_at"`
	UpdatedAt                 string  `json:"updated_at"`
}

type parentWithChildrenResponse struct {
	parentResponse
	Children []parentChildLinkResponse `json:"children"`
}

type parentChildLinkResponse struct {
	ID              string  `json:"id"`
	ChildID         string  `json:"child_id"`
	EndedAt         *string `json:"ended_at,omitempty"`
	EndedReasonCode *string `json:"ended_reason_code,omitempty"`
	EndedReasonNote *string `json:"ended_reason_note,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type parentListResponse struct {
	Parents    []parentResponse `json:"parents"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

func toParentResponse(p domain.Parent) parentResponse {
	resp := parentResponse{
		ID:                        p.ID.String(),
		FirstName:                 p.FirstName,
		LastName:                  p.LastName,
		Email:                     p.Email,
		Phone:                     p.Phone,
		AddressLine1:              p.AddressLine1,
		AddressLine2:              p.AddressLine2,
		AddressCity:               p.AddressCity,
		AddressPostcode:           p.AddressPostcode,
		RelationshipToChild:       p.RelationshipToChild,
		HasParentalResponsibility: p.HasParentalResponsibility,
		CanPickUp:                 p.CanPickUp,
		IsEmergencyContact:        p.IsEmergencyContact,
		Notes:                     p.Notes,
		IsActive:                  p.IsActive,
		CreatedAt:                 p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                 p.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if p.UserID != nil {
		s := p.UserID.String()
		resp.UserID = &s
	}
	return resp
}

func toParentChildLinkResponse(link domain.ParentChild) parentChildLinkResponse {
	resp := parentChildLinkResponse{
		ID:              link.ID.String(),
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
