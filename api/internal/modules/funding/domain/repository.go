package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	Get(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (FundingProfile, bool, error)
	GetForUpdate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (FundingProfile, bool, error)
	Create(ctx context.Context, tx Tx, profile FundingProfile) (FundingProfile, error)
	UpdateAllowance(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time, minutes int) (FundingProfile, error)
	GetChildEnrollmentForUpdate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (ChildEnrollment, bool, error)
	ListOverview(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]OverviewRow, error)
}
