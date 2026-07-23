package domain

import (
	"time"

	"github.com/google/uuid"
)

type Parent struct {
	ID                      uuid.UUID
	TenantID                uuid.UUID
	BranchID                uuid.UUID
	FirstName               string
	LastName                *string
	Email                   *string
	Phone                   *string
	AddressLine1            *string
	AddressLine2            *string
	AddressCity             *string
	AddressPostcode         *string
	RelationshipToChild     *string
	HasParentalResponsibility bool
	CanPickUp               bool
	IsEmergencyContact      bool
	Notes                   *string
	UserID                  *uuid.UUID
	IsActive                bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
