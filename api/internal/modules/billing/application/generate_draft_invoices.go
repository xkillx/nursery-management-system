package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

// GenerateDraftInvoicesUseCase implements the advance-pay generation rule.
// One draft invoice per active Term per month, computed from the Term's
// Booking Pattern. This replaces the attendance-actuals-based generator.
type GenerateDraftInvoicesUseCase struct {
	repo         domain.BillingRepository
	txMgr        *transaction.Manager
	auditW       *audit.Writer
	logger       *slog.Logger
	recorder     *metrics.Recorder
	termInvoices *GenerateTermInvoices
	completeRun  *CompleteInvoiceRun
	metricsRec   *InvoiceMetrics
}

func NewGenerateDraftInvoices(
	repo domain.BillingRepository,
	txMgr *transaction.Manager,
	auditW *audit.Writer,
	logger *slog.Logger,
	recorder *metrics.Recorder,
	termDateLookup domain.TermDateLookup,
	adHocLookup domain.AdHocBookingLookup,
	hourlyLookup domain.HourlyBookingLookup,
	closureDateLookup domain.ClosureDateLookup,
	fundingLookup domain.FundingLookup,
	bookingEntriesLookup domain.BookingEntriesLookup,
) *GenerateDraftInvoicesUseCase {
	return &GenerateDraftInvoicesUseCase{
		repo:         repo,
		txMgr:        txMgr,
		auditW:       auditW,
		logger:       logger,
		recorder:     recorder,
		termInvoices: NewGenerateTermInvoices(repo, auditW, termDateLookup, adHocLookup, hourlyLookup, closureDateLookup, fundingLookup, bookingEntriesLookup),
		completeRun:  NewCompleteInvoiceRun(repo),
		metricsRec:   NewInvoiceMetrics(recorder, logger),
	}
}

// Execute runs the full-month advance-pay generation. The billing month is
// the first day of the calendar month being invoiced. The system iterates
// every active Term in the branch whose term range covers the billing month
// and produces one draft (or issued) invoice per Term.
func (uc *GenerateDraftInvoicesUseCase) Execute(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string, rawChildIDs []string) (domain.DraftGenerationResult, error) {
	startedAt := time.Now()

	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.DraftGenerationResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	childIDs, err := parseAndDedupeChildIDs(rawChildIDs)
	if err != nil {
		return domain.DraftGenerationResult{}, domainerrors.Validation("Invalid child ID format.", "child_ids")
	}
	isFullMonth := len(rawChildIDs) == 0

	year := billingMonth.Year()
	month := billingMonth.Month()

	period, err := domain.NewBillingPeriod(year, month)
	if err != nil {
		return domain.DraftGenerationResult{}, domainerrors.Internal(fmt.Errorf("billing period: %w", err))
	}

	runID := uid.NewUUID()
	var result domain.DraftGenerationResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if runErr := uc.repo.CreateInvoiceRun(ctx, tx, domain.InvoiceRunCreateParams{
			ID:                      runID,
			TenantID:                actor.TenantID,
			BranchID:                actor.BranchID,
			BillingMonth:            billingMonth,
			RunType:                 domain.InvoiceRunTypeDraftGeneration,
			Status:                  domain.InvoiceRunStatusStarted,
			RequestedByUserID:       actor.UserID,
			RequestedByMembershipID: actor.MembershipID,
			RequestID:               actor.RequestID,
		}); runErr != nil {
			return fmt.Errorf("create invoice run: %w", runErr)
		}

		candidateTerms, err := uc.repo.ListActiveTermsForGeneration(ctx, tx, actor.TenantID, actor.BranchID, billingMonth)
		if err != nil {
			return fmt.Errorf("list active terms: %w", err)
		}

		blocked := buildSelectedChildBlockers(candidateTerms, childIDs, isFullMonth)

		genInput := GenerateTermInvoicesInput{
			Tx:              tx,
			Actor:           actor,
			BillingMonth:    billingMonth,
			BillingMonthRaw: billingMonthRaw,
			Period:          period,
			RunID:           runID,
		}
		requestedSet := buildRequestedSet(childIDs)
		genOutput, genErr := uc.termInvoices.Execute(ctx, genInput, candidateTerms, requestedSet, isFullMonth)
		if genErr != nil {
			return genErr
		}
		blocked = append(blocked, genOutput.Blocked...)

		compInput := CompleteInvoiceRunInput{
			Tx:              tx,
			Actor:           actor,
			RunID:           runID,
			BillingMonthRaw: billingMonthRaw,
			Generated:       genOutput.Generated,
			Blocked:         blocked,
			IsFullMonth:     isFullMonth,
			ChildIDs:        childIDs,
		}
		result, err = uc.completeRun.Execute(ctx, compInput)
		if err != nil {
			return err
		}
		return nil
	})

	if txErr != nil {
		uc.metricsRec.RecordRun("full_month", "error", startedAt, result, actor)
		return domain.DraftGenerationResult{}, domainerrors.Internal(txErr)
	}

	mode := "full_month"
	if len(rawChildIDs) > 0 {
		mode = "selected_children"
	}
	outcome := "completed"
	if result.RunStatus == domain.InvoiceRunStatusCompletedWithExceptions {
		outcome = "completed_with_exceptions"
	}
	uc.metricsRec.RecordRun(mode, outcome, startedAt, result, actor)
	return result, nil
}

func buildSelectedChildBlockers(candidateTerms []domain.AdvancePayTermRow, childIDs []uuid.UUID, isFullMonth bool) []domain.DraftGenerationBlockedChild {
	if isFullMonth {
		return nil
	}
	foundTermByChild := make(map[uuid.UUID]domain.AdvancePayTermRow, len(candidateTerms))
	for _, t := range candidateTerms {
		foundTermByChild[t.ChildID] = t
	}
	var blocked []domain.DraftGenerationBlockedChild
	for _, id := range childIDs {
		if _, found := foundTermByChild[id]; !found {
			blocked = append(blocked, domain.DraftGenerationBlockedChild{
				ChildID: id,
				Blockers: []domain.PreflightBlocker{
					{Code: domain.BlockerChildNotFound, Message: "Child has no active term for this billing month."},
				},
			})
		}
	}
	return blocked
}

func buildRequestedSet(childIDs []uuid.UUID) map[uuid.UUID]struct{} {
	requestedSet := make(map[uuid.UUID]struct{}, len(childIDs))
	for _, id := range childIDs {
		requestedSet[id] = struct{}{}
	}
	return requestedSet
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
