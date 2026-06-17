package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProfessionalReferral struct {
	Type              string  `json:"type"`
	ReferredDate      *string `json:"referred_date,omitempty"`
	ReferredBy        *string `json:"referred_by,omitempty"`
	WaitingListStatus string  `json:"waiting_list_status"`
	Notes             *string `json:"notes,omitempty"`
}

type ChildSafeguardingProfile struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	ChildID   uuid.UUID

	SocialServicesStatus      YesNoUnknown
	SocialServicesNotes       *string
	SocialWorkerName          *string
	SocialWorkerPhone         *string
	SocialWorkerEmail         *string
	ConcernWalking            YesNoUnknown
	ConcernSpeechLanguage     YesNoUnknown
	ConcernHearing            YesNoUnknown
	ConcernSight              YesNoUnknown
	ConcernEmotionalWellbeing YesNoUnknown
	ConcernBehaviour          YesNoUnknown
	ProfessionalReferrals     []ProfessionalReferral
	RestrictedNotes           *string

	CreatedAt time.Time
	UpdatedAt time.Time
}
