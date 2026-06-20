package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

// TermStatusTransition is a per-term result row from a status-transition job.
type TermStatusTransition struct {
	TermID           uuid.UUID
	ChildID          uuid.UUID
	From             domain.TermStatus
	To               domain.TermStatus
	BookingPatternID uuid.UUID
}

// MarkPendingRenewalUseCase flips terms from active to pending_renewal when
// their term_end_date is within the next PendingRenewalThresholdDays days.
type MarkPendingRenewalUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr *transaction.Manager
	now   func() time.Time
}

func NewMarkPendingRenewalUseCase(
	repo domain.Repository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *MarkPendingRenewalUseCase {
	return &MarkPendingRenewalUseCase{repo: repo, audit: auditWriter, txMgr: txMgr, now: func() time.Time { return time.Now().UTC() }}
}

// RunForTenantBranch scans a single (tenant, branch) for active terms within
// the renewal window and flips them to pending_renewal. The transition is
// auditable and the function is idempotent (terms already at pending_renewal
// are skipped via the WHERE clause).
func (uc *MarkPendingRenewalUseCase) RunForTenantBranch(ctx context.Context, tenantID, branchID uuid.UUID) ([]TermStatusTransition, error) {
	today := uc.now().UTC().Truncate(24 * time.Hour)
	maxEnd := today.AddDate(0, 0, domain.PendingRenewalThresholdDays)

	var transitions []TermStatusTransition
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		terms, err := uc.repo.ListExpiringWithin(ctx, tenantID, branchID, maxEnd)
		if err != nil {
			return fmt.Errorf("list expiring: %w", err)
		}
		for _, t := range terms {
			if t.Status != domain.TermStatusActive {
				continue
			}
			rows, err := uc.repo.UpdateStatus(ctx, tx, tenantID, branchID, t.ID, domain.TermStatusPendingRenewal)
			if err != nil {
				return fmt.Errorf("update status to pending_renewal: %w", err)
			}
			if rows == 0 {
				continue
			}
			if err := uc.audit.WriteSystemWithTx(ctx, tx, tenantID, branchID, "", audit.WriteParams{
				ActionType: domain.AuditTermStatusTransitioned,
				EntityType: domain.AuditEntityTerm,
				EntityID:   t.ID,
				Details: map[string]any{
					"child_id": t.ChildID.String(),
					"from":     string(domain.TermStatusActive),
					"to":       string(domain.TermStatusPendingRenewal),
					"trigger":  "mark_pending_renewal",
				},
			}); err != nil {
				return fmt.Errorf("audit term status transition: %w", err)
			}
			transitions = append(transitions, TermStatusTransition{
				TermID:           t.ID,
				ChildID:          t.ChildID,
				From:             domain.TermStatusActive,
				To:               domain.TermStatusPendingRenewal,
				BookingPatternID: t.BookingPatternID,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transitions, nil
}

// RunForAllTenantsAndBranches is the cron entry point: iterates every active
// (tenant, branch) and runs the pending-renewal scan.
func (uc *MarkPendingRenewalUseCase) RunForAllTenantsAndBranches(ctx context.Context, lister interface {
	ListAllTenantBranches(ctx context.Context) ([]TenantBranch, error)
}) error {
	scopes, err := lister.ListAllTenantBranches(ctx)
	if err != nil {
		return fmt.Errorf("list tenant branches: %w", err)
	}
	for _, scope := range scopes {
		if _, err := uc.RunForTenantBranch(ctx, scope.TenantID, scope.BranchID); err != nil {
			return fmt.Errorf("mark pending renewal for %s/%s: %w", scope.TenantID, scope.BranchID, err)
		}
	}
	return nil
}

// TenantBranch is a minimal projection shared with the invoicerun lister.
type TenantBranch struct {
	TenantID uuid.UUID
	BranchID uuid.UUID
}

// ExpireTermsUseCase flips terms that reached their term_end_date + 1 day to status='ended',
// and (per Phase 5) marks the child as inactive if no renewal Term is on file.
// Both transitions are written to the audit log.
type ExpireTermsUseCase struct {
	repo        domain.Repository
	audit       *audit.Writer
	txMgr       *transaction.Manager
	deactivator ChildDeactivator
	now         func() time.Time
}

func NewExpireTermsUseCase(
	repo domain.Repository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *ExpireTermsUseCase {
	return &ExpireTermsUseCase{repo: repo, audit: auditWriter, txMgr: txMgr, now: func() time.Time { return time.Now().UTC() }}
}

// WithDeactivator injects the ChildDeactivator adapter. Optional; without
// it the expire run still flips the term but does not deactivate the child.
func (uc *ExpireTermsUseCase) WithDeactivator(d ChildDeactivator) *ExpireTermsUseCase {
	uc.deactivator = d
	return uc
}

// ExpireResult is the per-term outcome of the expire run.
type ExpireResult struct {
	Transition          TermStatusTransition
	ChildMarkedInactive bool
}

// RunForTenantBranch expires terms for a single (tenant, branch) scope.
// Each flipped term triggers (a) status='ended', (b) clear of the child
// current_term_id denormalisation, and (c) if no renewal term exists for
// the child, mark the child inactive.
func (uc *ExpireTermsUseCase) RunForTenantBranch(ctx context.Context, tenantID, branchID uuid.UUID) ([]ExpireResult, error) {
	today := uc.now().UTC().Truncate(24 * time.Hour)
	cutoff := today.AddDate(0, 0, -1) // term_end_date <= cutoff (i.e. ended at least 1 day ago)

	var results []ExpireResult
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		terms, err := uc.repo.ListEndingOnOrBefore(ctx, tenantID, branchID, cutoff)
		if err != nil {
			return fmt.Errorf("list ending: %w", err)
		}
		for _, t := range terms {
			if t.Status != domain.TermStatusActive && t.Status != domain.TermStatusPendingRenewal {
				continue
			}
			prev := t.Status
			rows, err := uc.repo.UpdateStatus(ctx, tx, tenantID, branchID, t.ID, domain.TermStatusEnded)
			if err != nil {
				return fmt.Errorf("update status to ended: %w", err)
			}
			if rows == 0 {
				continue
			}
			if err := uc.repo.ClearChildCurrentTermID(ctx, tx, tenantID, branchID, t.ChildID); err != nil {
				return fmt.Errorf("clear child current term: %w", err)
			}
			// System audit (no actor).
			if err := uc.audit.WriteSystemWithTx(ctx, tx, tenantID, branchID, "", audit.WriteParams{
				ActionType: domain.AuditTermStatusTransitioned,
				EntityType: domain.AuditEntityTerm,
				EntityID:   t.ID,
				Details: map[string]any{
					"child_id": t.ChildID.String(),
					"from":     string(prev),
					"to":       string(domain.TermStatusEnded),
					"trigger":  "expire_terms",
				},
			}); err != nil {
				return fmt.Errorf("audit term status transition: %w", err)
			}

			result := ExpireResult{
				Transition: TermStatusTransition{
					TermID:           t.ID,
					ChildID:          t.ChildID,
					From:             prev,
					To:               domain.TermStatusEnded,
					BookingPatternID: t.BookingPatternID,
				},
			}

			// If no renewal term exists for this child, mark the child inactive
			// (the existing Child Enrollment Lifecycle flow). This is the
			// Phase 5 acceptance criterion: a child whose term ends without
			// renewal transitions to inactive on term_end_date + 1.
			hasRenewal, err := uc.hasRenewalTerm(ctx, tx, tenantID, branchID, t.ChildID)
			if err != nil {
				return fmt.Errorf("check renewal term: %w", err)
			}
			if !hasRenewal && uc.deactivator != nil {
				if err := uc.deactivator.MarkChildInactive(ctx, tenantID, branchID, t.ChildID, "other", "term ended without renewal"); err != nil {
					return fmt.Errorf("mark child inactive after term expiry: %w", err)
				}
				result.ChildMarkedInactive = true
			}

			results = append(results, result)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// hasRenewalTerm returns true if the child has any pre_term, active, or
// pending_renewal term on file. The check is best-effort and only used
// to decide whether to cascade to child-inactive.
func (uc *ExpireTermsUseCase) hasRenewalTerm(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	_, found, err := uc.repo.GetActiveForChildInTx(ctx, tx, tenantID, branchID, childID)
	if err != nil {
		return false, fmt.Errorf("get active term in tx: %w", err)
	}
	return found, nil
}

// systemActor returns a tenant.ActorContext with no user/membership, used
// for system-initiated state transitions written to the audit log.
func systemActor(tenantID, branchID uuid.UUID) tenant.ActorContext {
	return tenant.ActorContext{
		TenantID:      tenantID,
		BranchID:      branchID,
		RequestID:     "system:expire_terms",
		CorrelationID: "system:expire_terms",
	}
}
