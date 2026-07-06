package domain

import (
	"context"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]SessionType, error)
	ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool, limit, offset int) ([]SessionType, error)
	ListByBranchPaginatedSorted(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool, limit, offset int, sortField, sortDir string) ([]SessionType, error)
	CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) (int, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (SessionType, error)
	Create(ctx context.Context, st SessionType) error
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	Archive(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	Reactivate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
	Exists(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (bool, error)
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (SessionType, error)
}
