package invoicerun

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/platform/db/sqlc"
)

// TenantBranchLister exposes the (tenant, branch) pairs to the scheduler
// so it can iterate every scope. In the single-tenant dev DB this is a
// single pair; in multi-tenant production this comes from a system-level
// table scan.
type TenantBranchLister interface {
	ListAllTenantBranches(ctx context.Context) ([]TenantBranch, error)
}

type TenantBranch struct {
	TenantID uuid.UUID
	BranchID uuid.UUID
}

// SystemTenantBranchLister is the default implementation backed by the
// branches table. The pool is shared with the rest of the application.
type SystemTenantBranchLister struct {
	pool *pgxpool.Pool
}

func NewSystemTenantBranchLister(pool *pgxpool.Pool) *SystemTenantBranchLister {
	return &SystemTenantBranchLister{pool: pool}
}

func (l *SystemTenantBranchLister) ListAllTenantBranches(ctx context.Context) ([]TenantBranch, error) {
	rows, err := sqlc.New(l.pool).ListAllTenantBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all tenant branches: %w", err)
	}
	out := make([]TenantBranch, 0, len(rows))
	for _, r := range rows {
		out = append(out, TenantBranch{
			TenantID: uuid.UUID(r.TenantID.Bytes),
			BranchID: uuid.UUID(r.BranchID.Bytes),
		})
	}
	return out, nil
}

// ExpireTermsRunner is a no-op since the term module has been removed.
// The billing module now derives billable children from bookings directly.
type ExpireTermsRunner struct {
	tenantBranchLister TenantBranchLister
}

func NewExpireTermsRunner(lister TenantBranchLister) *ExpireTermsRunner {
	return &ExpireTermsRunner{
		tenantBranchLister: lister,
	}
}

func (r *ExpireTermsRunner) RunForAllTenantsAndBranches(ctx context.Context) error {
	// No-op: term module has been removed.
	return nil
}
