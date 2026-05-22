package domain

import (
	"time"

	"github.com/google/uuid"
)

type Guardian struct {
	ID                     uuid.UUID
	TenantID               uuid.UUID
	BranchID               uuid.UUID
	FullName               string
	Email                  *string
	Phone                  *string
	Notes                  *string
	IsActive               bool
	DeactivatedAt          *time.Time
	DeactivationReasonCode *string
	DeactivationReasonNote *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
