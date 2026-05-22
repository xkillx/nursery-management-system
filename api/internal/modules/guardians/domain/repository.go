package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatusFilter string

const (
	StatusActive   StatusFilter = "active"
	StatusInactive StatusFilter = "inactive"
	StatusAll      StatusFilter = "all"
)

type Repository interface {
	List(ctx context.Context, tenantID, branchID uuid.UUID, filter StatusFilter, limit, offset int) ([]Guardian, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (Guardian, error)
	Create(ctx context.Context, guardian Guardian) error
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (Guardian, error)
	GetActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, bool, error)
	Deactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error
	CascadeLinks(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID, reasonCode, reasonNote string) error
	CascadeMappings(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID, reasonCode, reasonNote string) error
	Reactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error
}

type PostgresRepo interface {
	Repository
	Pool() *pgxpool.Pool
}
