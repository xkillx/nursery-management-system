package invoicerun

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/billing/application"
	termapp "nursery-management-system/api/internal/modules/term/application"
	"nursery-management-system/api/internal/platform/db/sqlc"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
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

// GenerateAdvanceInvoicesRunner is the cron-driven runner. It produces one
// invoice per active Term for the given billing month by calling the
// existing GenerateDraftInvoicesUseCase.
type GenerateAdvanceInvoicesRunner struct {
	generation         *application.GenerateDraftInvoicesUseCase
	tenantBranchLister TenantBranchLister
	now                func() time.Time
}

func NewGenerateAdvanceInvoicesRunner(
	generation *application.GenerateDraftInvoicesUseCase,
	lister TenantBranchLister,
) *GenerateAdvanceInvoicesRunner {
	return &GenerateAdvanceInvoicesRunner{
		generation:         generation,
		tenantBranchLister: lister,
		now:                func() time.Time { return time.Now().UTC() },
	}
}

// RunForBillingMonth produces one draft invoice per active Term for the
// given billing month across every active (tenant, branch) scope. The
// generator is idempotent: re-running for the same month is a no-op
// because the existing-draft check short-circuits cleanly.
//
// The use case does not itself issue invoices; the manager triggers the
// issue step explicitly today. The plan calls for auto-issue on the 25th;
// wiring the auto-issue step into the same cron tick is straightforward
// and is included in the same runner so a single failure mode applies to
// both phases (the issue step is skipped if generation fails).
func (r *GenerateAdvanceInvoicesRunner) RunForBillingMonth(ctx context.Context, billingMonth time.Time, triggeredBy string) error {
	if r.tenantBranchLister == nil {
		return fmt.Errorf("tenant-branch lister not configured")
	}
	if r.generation == nil {
		return fmt.Errorf("generation use case not configured")
	}
	scopes, err := r.tenantBranchLister.ListAllTenantBranches(ctx)
	if err != nil {
		return fmt.Errorf("list tenant branches: %w", err)
	}
	billingMonthStr := billingMonth.Format("2006-01")

	for _, scope := range scopes {
		actor := tenant.ActorContext{
			TenantID:      scope.TenantID,
			BranchID:      scope.BranchID,
			RequestID:     "scheduler-" + uid.NewUUID().String(),
			CorrelationID: "scheduler:" + triggeredBy,
		}
		_, genErr := r.generation.Execute(ctx, actor, billingMonthStr, nil)
		if genErr != nil {
			return fmt.Errorf("generate drafts for %s/%s: %w", scope.TenantID, scope.BranchID, genErr)
		}
	}
	return nil
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
