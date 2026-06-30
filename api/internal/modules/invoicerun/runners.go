package invoicerun

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	termapp "nursery-management-system/api/internal/modules/term/application"
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

// ExpireTermsRunner iterates every (tenant, branch) and runs the term
// lifecycle steps in order:
//
//  1. MarkPendingRenewal — flip active terms to pending_renewal when
//     their term_end_date is within the next PendingRenewalThresholdDays days.
//  2. ExpireTerms — flip terms that reached their term_end_date + 1 day to
//     status='ended' and, if no renewal term exists, mark the child
//     inactive (Phase 5 acceptance criterion).
type ExpireTermsRunner struct {
	expireTerms        *termapp.ExpireTermsUseCase
	markPendingRenewal *termapp.MarkPendingRenewalUseCase
	tenantBranchLister TenantBranchLister
}

func NewExpireTermsRunner(
	expireTerms *termapp.ExpireTermsUseCase,
	markPendingRenewal *termapp.MarkPendingRenewalUseCase,
	lister TenantBranchLister,
) *ExpireTermsRunner {
	return &ExpireTermsRunner{
		expireTerms:        expireTerms,
		markPendingRenewal: markPendingRenewal,
		tenantBranchLister: lister,
	}
}

func (r *ExpireTermsRunner) RunForAllTenantsAndBranches(ctx context.Context) error {
	if r.tenantBranchLister == nil {
		return fmt.Errorf("tenant-branch lister not configured")
	}
	if r.expireTerms == nil {
		return fmt.Errorf("expire-terms use case not configured")
	}
	if r.markPendingRenewal == nil {
		return fmt.Errorf("mark-pending-renewal use case not configured")
	}
	scopes, err := r.tenantBranchLister.ListAllTenantBranches(ctx)
	if err != nil {
		return fmt.Errorf("list tenant branches: %w", err)
	}
	for _, scope := range scopes {
		if _, err := r.markPendingRenewal.RunForTenantBranch(ctx, scope.TenantID, scope.BranchID); err != nil {
			return fmt.Errorf("mark pending renewal for %s/%s: %w", scope.TenantID, scope.BranchID, err)
		}
		if _, err := r.expireTerms.RunForTenantBranch(ctx, scope.TenantID, scope.BranchID); err != nil {
			return fmt.Errorf("expire terms for %s/%s: %w", scope.TenantID, scope.BranchID, err)
		}
	}
	return nil
}
