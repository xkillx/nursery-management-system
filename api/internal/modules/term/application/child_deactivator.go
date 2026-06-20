package application

import (
	"context"

	"github.com/google/uuid"
)

// ChildDeactivator is the consumer-side interface for marking a child
// inactive. Implemented by an adapter in the bootstrap layer that wraps
// the children.MarkInactive use case. Under Clean Architecture, the term
// module cannot import the children module directly.
//
// The default reason for system-initiated deactivation is "term ended
// without renewal"; the children.MarkInactive use case writes a child
// leaving record and a system audit event.
type ChildDeactivator interface {
	MarkChildInactive(ctx context.Context, tenantID, branchID, childID uuid.UUID, reasonCode, reasonNote string) error
}
