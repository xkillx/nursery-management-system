package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Tx = pgx.Tx

// Repository is the term domain's data-access port.
// All read methods scope to (tenant_id, branch_id); the postgres implementation
// enforces that scoping in the WHERE clause.
type Repository interface {
	// Lifecycle writes (transactional).
	Insert(ctx context.Context, tx Tx, t *Term) (*Term, error)
	Terminate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, terminatedAt time.Time, reasonCode, reasonNote string) (int64, error)
	UpdateStatus(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, status TermStatus) (int64, error)

	// Reads.
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (*Term, bool, error)
	GetActiveForChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*Term, bool, error)
	GetActiveForChildInTx(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (*Term, bool, error)
	GetByIDInTx(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (*Term, bool, error)
	ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]Term, error)
	ListActiveByBranch(ctx context.Context, tenantID, branchID uuid.UUID) ([]Term, error)
	ListExpiringWithin(ctx context.Context, tenantID, branchID uuid.UUID, maxTermEndDate time.Time) ([]Term, error)
	ListEndingOnOrBefore(ctx context.Context, tenantID, branchID uuid.UUID, endDate time.Time) ([]Term, error)
	ListActiveInBillingMonth(ctx context.Context, tenantID, branchID uuid.UUID, monthStart, monthEnd time.Time) ([]Term, error)
	ListActiveForChildUpdate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) ([]Term, error)

	// Child denormalisation (transactional).
	SetChildCurrentTermID(ctx context.Context, tx Tx, tenantID, branchID, childID, termID uuid.UUID) error
	ClearChildCurrentTermID(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) error
}

// ScheduleChangeRepository is the data-access port for term_schedule_change.
type ScheduleChangeRepository interface {
	Insert(ctx context.Context, tx Tx, c *TermScheduleChange) (*TermScheduleChange, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (*TermScheduleChange, bool, error)
	ListByTerm(ctx context.Context, tenantID, branchID, termID uuid.UUID) ([]TermScheduleChange, error)
	Approve(ctx context.Context, tx Tx, tenantID, branchID, id, approverMembershipID uuid.UUID) (int64, error)
	Reject(ctx context.Context, tx Tx, tenantID, branchID, id, approverMembershipID uuid.UUID) (int64, error)
}
