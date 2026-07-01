package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
)

type ParentContactLookup interface {
	GetForInvoice(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ParentContact, error)
}
