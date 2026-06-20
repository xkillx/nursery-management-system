package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Tx is a transaction interface matching pgx.Tx for dependency injection.
type Tx = pgx.Tx

type Repository interface {
	ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]SessionTemplate, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (SessionTemplate, error)
	Create(ctx context.Context, t SessionTemplate) error
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	Archive(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	Reactivate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
	Exists(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (bool, error)
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (SessionTemplate, error)

	InsertEntry(ctx context.Context, tx Tx, entry SessionTemplateEntry) error
	DeleteEntriesByTemplate(ctx context.Context, tx Tx, tenantID, branchID, templateID uuid.UUID) error
	EntriesListByTemplate(ctx context.Context, tenantID, branchID, templateID uuid.UUID) ([]SessionTemplateEntry, error)
	EntriesListByTemplateTx(ctx context.Context, tx Tx, tenantID, branchID, templateID uuid.UUID) ([]SessionTemplateEntry, error)
}
