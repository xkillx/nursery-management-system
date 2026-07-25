package domain

import (
	"context"

	"github.com/google/uuid"
)

type ChildFundingLookup interface {
	GetChildTermTimeOnly(ctx context.Context, tenantID, branchID, childID uuid.UUID) (bool, error)
}
