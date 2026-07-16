package application

import (
	"context"

	"github.com/google/uuid"
)

// ParentChildLookup returns the child IDs linked to a parent's membership.
type ParentChildLookup interface {
	ListChildIDsForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID) ([]uuid.UUID, error)
}
