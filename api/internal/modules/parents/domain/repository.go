package domain

import (
	"context"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	List(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, limit, offset int) ([]Parent, error)
	ListFiltered(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, isActive *bool, search *string, limit, offset int) ([]Parent, int, error)
	GetByID(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (Parent, bool, error)
	GetByUserID(ctx context.Context, tx Tx, tenantID uuid.UUID, userID uuid.UUID) (Parent, bool, error)
	Create(ctx context.Context, tx Tx, parent Parent) error
	Update(ctx context.Context, tx Tx, parent Parent) error
	SoftDelete(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	SetUserID(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, userID *uuid.UUID) error

	ListChildrenByParent(ctx context.Context, tx Tx, tenantID, branchID, parentID uuid.UUID) ([]ParentChild, error)
	ListParentsByChild(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) ([]ParentChild, error)
	FindActiveByPair(ctx context.Context, tx Tx, tenantID, branchID, parentID, childID uuid.UUID) (ParentChild, bool, error)
	CreateLink(ctx context.Context, tx Tx, link ParentChild) error
	EndLink(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error
	HasActiveParentForChild(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
}

type ChildExistenceChecker interface {
	ExistsInScope(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
}
