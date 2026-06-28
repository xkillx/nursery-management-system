package domain

import (
	"context"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	FindActiveByPair(ctx context.Context, tx Tx, tenantID, branchID, membershipID, childID uuid.UUID) (ParentChildMapping, bool, error)
	ListActiveByMembership(ctx context.Context, tx Tx, tenantID, branchID, membershipID uuid.UUID) ([]ParentChildMapping, error)
	Create(ctx context.Context, tx Tx, mapping ParentChildMapping) error
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (ParentChildMapping, bool, error)
	End(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error
}

type MembershipChecker interface {
	GetForScope(ctx context.Context, tx Tx, tenantID, branchID, membershipID uuid.UUID) (MembershipInfo, bool, error)
}

type MembershipInfo struct {
	ID       uuid.UUID
	Role     string
	IsActive bool
}
