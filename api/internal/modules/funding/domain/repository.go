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
	ListOverviewPaginated(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time, limit, offset int) ([]OverviewRow, error)
	CountOverview(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) (int, error)
	ListExpiringSoon(ctx context.Context, tenantID, branchID uuid.UUID, withinDays int) ([]ExpiringFundingRecord, error)
	GetFundedChildrenCount(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) (EnhancedOverviewMetrics, error)
	GetBookedHoursThisWeek(ctx context.Context, tenantID, branchID uuid.UUID) (float64, error)
	GetExpiringSoonCount(ctx context.Context, tenantID, branchID uuid.UUID, withinDays int) (int, error)
	GetChildAllocation(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonthStart, billingMonthEnd time.Time) ([]AllocationEntry, error)
}

type HistoryRepository interface {
	Create(ctx context.Context, history FundingHistory) error
	ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]FundingHistory, error)
}

type FundingRecordRepository interface {
	GetFundingRecord(ctx context.Context, tenantID, branchID, childID uuid.UUID) (FundingRecord, bool, error)
	UpsertFundingRecord(ctx context.Context, tx Tx, record FundingRecord) (FundingRecord, error)
}
