package domain

import (
	"time"

	"github.com/google/uuid"
)

type BillingBasis string

const (
	BillingBasisSiteRate BillingBasis = "site_rate"
	BillingBasisCustom   BillingBasis = "custom"
)

type ChildBillingProfile struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	ChildID  uuid.UUID

	BillingBasis    BillingBasis
	CustomRateMinor *int
	EffectiveFrom   time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
