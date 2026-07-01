package domain

import (
	"context"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	GetByBranch(ctx context.Context, tenantID, branchID uuid.UUID) (SiteProfile, error)
	Upsert(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, profile SiteProfile) error
}
