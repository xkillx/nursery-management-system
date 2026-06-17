package domain

import (
	"time"

	"github.com/google/uuid"
)

type ContactType string

const (
	ContactTypeParentCarer         ContactType = "parent_carer"
	ContactTypeEmergencyContact    ContactType = "emergency_contact"
	ContactTypeAuthorisedCollector ContactType = "authorised_collector"
)

type ChildContact struct {
	ID                        uuid.UUID
	TenantID                  uuid.UUID
	BranchID                  uuid.UUID
	ChildID                   uuid.UUID
	ContactType               ContactType
	SortOrder                 int
	FullName                  string
	RelationshipToChild       *string
	Address                   map[string]any
	Telephone                 *string
	Email                     *string
	WorkAddress               map[string]any
	HasParentalResponsibility *bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}
