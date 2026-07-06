package domain

import (
	"context"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]Room, error)
	ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool, limit, offset int) ([]Room, error)
	CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) (int, error)
	GetByID(ctx context.Context, tenantID, branchID, roomID uuid.UUID) (Room, error)
	Create(ctx context.Context, room Room) error
	Update(ctx context.Context, tenantID, branchID, roomID uuid.UUID, fields map[string]any) (int64, error)
	Archive(ctx context.Context, tx Tx, tenantID, branchID, roomID uuid.UUID) error
	Reactivate(ctx context.Context, tx Tx, tenantID, branchID, roomID uuid.UUID) error
	ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeRoomID *uuid.UUID) (bool, error)
	CountActiveChildren(ctx context.Context, tx Tx, tenantID, branchID, roomID uuid.UUID) (int, error)
	Exists(ctx context.Context, tx Tx, tenantID, branchID, roomID uuid.UUID) (bool, error)
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, roomID uuid.UUID) (Room, error)
	CountAssignedChildrenByBranch(ctx context.Context, tenantID, branchID uuid.UUID) (map[uuid.UUID]int, error)
}
