package application

import (
	"context"

	"github.com/google/uuid"

	siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"
)

type SiteProfileLookup interface {
	GetForInvoice(ctx context.Context, tenantID, branchID uuid.UUID) (*siteprofiledomain.SiteProfile, error)
}
