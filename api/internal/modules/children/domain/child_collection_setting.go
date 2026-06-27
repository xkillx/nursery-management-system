package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChildCollectionSetting struct {
	ID                                      uuid.UUID
	TenantID                                uuid.UUID
	BranchID                                uuid.UUID
	ChildID                                 uuid.UUID
	Over18CollectionAcknowledged            bool
	CollectionPassword                      string
	CollectionPasswordHint                  string
	CollectionPasswordUpdatedAt             *time.Time
	CollectionPasswordUpdatedByUserID       *uuid.UUID
	CollectionPasswordUpdatedByMembershipID *uuid.UUID
	CreatedAt                               time.Time
	UpdatedAt                               time.Time
}
