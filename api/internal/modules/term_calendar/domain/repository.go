package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]AcademicTerm, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (AcademicTerm, error)
	Create(ctx context.Context, term AcademicTerm) error
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	Archive(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (AcademicTerm, error)
	ListActiveDateRanges(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]TermDateRange, error)
}
