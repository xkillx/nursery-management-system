package domain

import (
	"time"

	"github.com/google/uuid"
)

type SessionType struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	BranchID     uuid.UUID
	Name         string
	StartMinutes int
	EndMinutes   int
	IsActive     bool
	Kind         string
	FlatFeeMinor *int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
