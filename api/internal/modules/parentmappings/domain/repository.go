package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	FindActiveByMembership(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (ParentMapping, bool, error)
	Create(ctx context.Context, tx pgx.Tx, mapping ParentMapping) error
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (ParentMapping, bool, error)
	End(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error
}

type MembershipChecker interface {
	GetForScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (MembershipInfo, bool, error)
}

type MembershipInfo struct {
	ID       uuid.UUID
	Role     string
	IsActive bool
}

type GuardianChecker interface {
	IsActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID) (bool, bool, error)
}
