package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

// BookingPatternLookup is the consumer-side interface for validating that a booking pattern
// exists in the same tenant+branch scope. Implemented by an adapter in the bootstrap layer.
type BookingPatternLookup interface {
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID) (bool, error)
}

// SiteRateProvider returns the branch's hourly rate at a given point in time.
// For Phase 1 this is the branch-level rate (no per-child overrides under advance-pay).
type SiteRateProvider interface {
	SiteHourlyRateMinor(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID) (int, bool, error)
}

// CreateTermInput is the validated input for creating a new Term.
type CreateTermInput struct {
	ChildID           uuid.UUID
	TermStartDate     time.Time
	BookingPatternID  uuid.UUID
	EnrollImmediately bool // if true, status starts as active when start_date <= today
}

type CreateTermUseCase struct {
	repo          domain.Repository
	txMgr         *transaction.Manager
	audit         *audit.Writer
	patternLookup BookingPatternLookup
	rateProvider  SiteRateProvider
}

func NewCreateTermUseCase(
	repo domain.Repository,
	txMgr *transaction.Manager,
	auditWriter *audit.Writer,
	patternLookup BookingPatternLookup,
	rateProvider SiteRateProvider,
) *CreateTermUseCase {
	return &CreateTermUseCase{
		repo:          repo,
		txMgr:         txMgr,
		audit:         auditWriter,
		patternLookup: patternLookup,
		rateProvider:  rateProvider,
	}
}

func (uc *CreateTermUseCase) Execute(ctx context.Context, actor tenant.ActorContext, in CreateTermInput) (*domain.Term, error) {
	if in.ChildID == uuid.Nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in.BookingPatternID == uuid.Nil {
		return nil, domainerrors.Validation("Invalid request payload.", "booking_pattern_id")
	}
	if err := domain.ValidateTermStartDate(in.TermStartDate); err != nil {
		return nil, domainerrors.New("term_invalid_start_date", "Invalid request payload.", "term_start_date")
	}

	var result *domain.Term
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		// 1. Ensure no active term already exists for this child.
		_, found, err := uc.repo.GetActiveForChildInTx(ctx, tx, actor.TenantID, actor.BranchID, in.ChildID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("check existing term: %w", err))
		}
		if found {
			return domainerrors.Conflict("term_already_exists", "An active term already exists for this child.")
		}

		// 2. Validate the booking pattern exists in scope.
		patternExists, err := uc.patternLookup.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, in.BookingPatternID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("lookup booking pattern: %w", err))
		}
		if !patternExists {
			return domainerrors.NotFound("booking_pattern", "Resource not found.")
		}

		// 3. Snapshot the site hourly rate at term creation.
		rate, rateFound, err := uc.rateProvider.SiteHourlyRateMinor(ctx, tx, actor.TenantID, actor.BranchID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("lookup site rate: %w", err))
		}
		if !rateFound || rate <= 0 {
			return domainerrors.New("site_rate_missing", "Invalid request payload.", "site_hourly_rate")
		}

		// 4. Build the Term.
		termID := uid.NewUUID()
		term, err := domain.NewTerm(
			termID,
			actor.TenantID,
			actor.BranchID,
			in.ChildID,
			in.TermStartDate,
			in.BookingPatternID,
			rate,
			actor.MembershipID,
		)
		if err != nil {
			return domainerrors.Internal(err)
		}

		// 5. Persist.
		saved, err := uc.repo.Insert(ctx, tx, term)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("insert term: %w", err))
		}

		// 6. Update child denormalisation.
		if err := uc.repo.SetChildCurrentTermID(ctx, tx, actor.TenantID, actor.BranchID, in.ChildID, saved.ID); err != nil {
			return domainerrors.Internal(fmt.Errorf("set child current term: %w", err))
		}

		// 7. Audit.
		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: domain.AuditTermCreated,
			EntityType: domain.AuditEntityTerm,
			EntityID:   saved.ID,
			Details: map[string]any{
				"child_id":               saved.ChildID.String(),
				"term_start_date":        saved.TermStartDate.Format("2006-01-02"),
				"term_end_date":          saved.TermEndDate.Format("2006-01-02"),
				"booking_pattern_id":     saved.BookingPatternID.String(),
				"site_hourly_rate_minor": saved.SiteHourlyRateMinor,
				"status":                 string(saved.Status),
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit term_created: %w", err))
		}

		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
