package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	FindActiveByPair(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID, childID uuid.UUID) (GuardianChildLink, bool, error)
	Create(ctx context.Context, tx pgx.Tx, link GuardianChildLink) error
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (GuardianChildLink, bool, error)
	End(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error
}
