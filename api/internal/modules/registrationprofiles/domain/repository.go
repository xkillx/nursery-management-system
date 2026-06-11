package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Tx = pgx.Tx

type Repository interface {
	GetChildSummary(ctx context.Context, tenantID, branchID, childID uuid.UUID) (ChildSummary, bool, error)
	GetByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*Profile, error)
	GetForUpdateByChild(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (*Profile, error)
	Create(ctx context.Context, tx Tx, profile *Profile) (*Profile, error)
	Update(ctx context.Context, tx Tx, profile *Profile) (*Profile, error)
	SetCollectionPassword(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, hash string, updatedAt time.Time, updatedByUserID, updatedByMembershipID uuid.UUID) error
	ReplaceContactsForTypes(ctx context.Context, tx Tx, profileID uuid.UUID, contactTypes []ContactType, entries []ContactEntry) error
	ListContactsByProfile(ctx context.Context, tenantID, branchID, profileID uuid.UUID) ([]ContactEntry, error)
}
