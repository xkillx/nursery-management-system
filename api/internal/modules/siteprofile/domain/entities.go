package domain

import (
	"time"

	"github.com/google/uuid"
)

type SiteProfile struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	BranchID        uuid.UUID
	NurseryName     string
	Description     string
	Phone           string
	Email           string
	Website         string
	AddressStreet   string
	AddressCity     string
	AddressPostcode string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
