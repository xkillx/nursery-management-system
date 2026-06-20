package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
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
	repo     domain.BillingRepository
	txMgr    *transaction.Manager
	auditW   *audit.Writer
	logger   *slog.Logger
	recorder *metrics.Recorder
}

func NewGenerateDraftInvoices(
	repo domain.BillingRepository,
	txMgr *transaction.Manager,
	auditW *audit.Writer,
) *GenerateDraftInvoicesUseCase {
	return &GenerateDraftInvoicesUseCase{repo: repo, txMgr: txMgr, auditW: auditW}
}

func (uc *GenerateDraftInvoicesUseCase) WithObservability(logger *slog.Logger, recorder *metrics.Recorder) *GenerateDraftInvoicesUseCase {
	return &GenerateDraftInvoicesUseCase{
		repo:     uc.repo,
		txMgr:    uc.txMgr,
		auditW:   uc.auditW,
		logger:   logger,
		recorder: recorder,
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
		// 1. Create the invoice run.
		runErr := uc.repo.CreateInvoiceRun(ctx, tx, domain.InvoiceRunCreateParams{
			ID:                      runID,
			TenantID:                actor.TenantID,
			BranchID:                actor.BranchID,
			BillingMonth:            billingMonth,
			RunType:                 domain.InvoiceRunTypeDraftGeneration,
			Status:                  domain.InvoiceRunStatusStarted,
			RequestedByUserID:       actor.UserID,
			RequestedByMembershipID: actor.MembershipID,
			RequestID:               actor.RequestID,
		})
		if runErr != nil {
			return fmt.Errorf("create invoice run: %w", runErr)
		}

		// 2. List candidate terms for the month.
		candidateTerms, err := uc.repo.ListActiveTermsForGeneration(ctx, tx, actor.TenantID, actor.BranchID, billingMonth)
		if err != nil {
			return fmt.Errorf("list active terms: %w", err)
		}

		// 3. If selected-children mode, build the requested set and the exceptions.
		blocked := make([]domain.DraftGenerationBlockedChild, 0)
		requestedSet := make(map[uuid.UUID]struct{}, len(childIDs))
		for _, id := range childIDs {
			requestedSet[id] = struct{}{}
		}
		if !isFullMonth {
			foundTermByChild := make(map[uuid.UUID]domain.AdvancePayTermRow, len(candidateTerms))
			for _, t := range candidateTerms {
				foundTermByChild[t.ChildID] = t
			}
			for id := range requestedSet {
				_, found := foundTermByChild[id]
				if !found {
					blocked = append(blocked, domain.DraftGenerationBlockedChild{
						ChildID: id,
						Blockers: []domain.PreflightBlocker{
							{Code: domain.BlockerChildNotFound, Message: "Child has no active term for this billing month."},
						},
					})
				}
			}
		}

		// 4. For each candidate term, compute + write the draft invoice.
		var generated []domain.DraftGenerationChildResult
		var totalDueSum int

		for _, termRow := range candidateTerms {
			if !isFullMonth {
				if _, ok := requestedSet[termRow.ChildID]; !ok {
					continue
				}
			}

			// Preflight: Term exists; check funding profile presence + guardian link.
			preflightBlockers := uc.preflightTerm(termRow)
			if len(preflightBlockers) > 0 {
				blocked = append(blocked, domain.DraftGenerationBlockedChild{
					ChildID:         termRow.ChildID,
					ChildFirstName:  termRow.FirstName,
					ChildMiddleName: termRow.MiddleName,
					ChildLastName:   termRow.LastName,
					Blockers:        preflightBlockers,
				})
				continue
			}

			// Existing invoice? Lock under transaction.
			existingInvoice, invoiceFound, err := uc.repo.GetMonthlyInvoiceForUpdate(ctx, tx, actor.TenantID, actor.BranchID, termRow.ChildID, billingMonth)
			if err != nil {
				return fmt.Errorf("get monthly invoice: %w", err)
			}
			if invoiceFound && existingInvoice.Status != domain.InvoiceStatusDraft {
				blocked = append(blocked, domain.DraftGenerationBlockedChild{
					ChildID:         termRow.ChildID,
					ChildFirstName:  termRow.FirstName,
					ChildMiddleName: termRow.MiddleName,
					ChildLastName:   termRow.LastName,
					Blockers: []domain.PreflightBlocker{
						{
							Code:    domain.BlockerInvoiceAlreadyIssued,
							Message: "A monthly invoice has already been issued for this child and billing month.",
						},
					},
				})
				continue
			}

			// 5. Load booking pattern entries for this Term.
			entries, err := uc.repo.ListBookingPatternEntries(ctx, tx, actor.TenantID, actor.BranchID, termRow.BookingPatternID)
			if err != nil {
				return fmt.Errorf("list booking pattern entries: %w", err)
			}

			// 6. Build domain entries + run the booking-driven calculation.
			domainEntries := make([]domain.BookedPatternEntry, 0, len(entries))
			for _, e := range entries {
				domainEntries = append(domainEntries, domain.BookedPatternEntry{
					DayOfWeek: e.DayOfWeek,
					SessionType: domain.BookedSessionType{
						ID:              e.SessionTypeID.String(),
						Name:            e.SessionTypeName,
						StartMinutes:    e.StartMinutes,
						EndMinutes:      e.EndMinutes,
						DurationMinutes: e.EndMinutes - e.StartMinutes,
					},
				})
			}

			calc, calcErr := domain.CalculateBookedCoreMinutesInMonth(
				termRow.BookingPatternID.String(),
				domainEntries,
				billingMonth,
				termRow.SiteHourlyRateMinor,
			)
			if calcErr != nil {
				return fmt.Errorf("calculate booked minutes for term %s: %w", termRow.TermID, calcErr)
			}

			fundedAllowance := 0
			if termRow.FundedAllowanceMinutes != nil {
				fundedAllowance = *termRow.FundedAllowanceMinutes
			}
			_, billableMinutes, _, billableMinor, err := domain.ComputeFundedDeductionMinor(
				calc.TotalMinutes, fundedAllowance, termRow.SiteHourlyRateMinor,
			)
			if err != nil {
				return fmt.Errorf("compute funded deduction for term %s: %w", termRow.TermID, err)
			}
			fundedDeductionMinutes := minInt(calc.TotalMinutes, fundedAllowance)
			fundedDeductionMinor := calc.SubtotalMinor - billableMinor
			if fundedDeductionMinor < 0 {
				fundedDeductionMinor = 0
			}

			subtotalMinor := calc.SubtotalMinor
			totalDueMinor := billableMinor

			// 7. Pre-existing extra line items (if regenerating an existing draft).
			extrasTotalMinor := 0
			var existingExtras []domain.ExtraLineRow
			if invoiceFound {
				existingExtras, err = uc.repo.ListDraftExtraLines(ctx, tx, actor.TenantID, actor.BranchID, existingInvoice.ID)
				if err != nil {
					return fmt.Errorf("list extra lines: %w", err)
				}
				for _, ex := range existingExtras {
					extrasTotalMinor += ex.LineAmountMinor
				}
			}

			calcDetails := domain.InvoiceCalculationDetails{
				BillingMonth:           billingMonthRaw,
				ChildID:                termRow.ChildID,
				CoreHourlyRateMinor:    termRow.SiteHourlyRateMinor,
				CoreSubtotalMinor:      subtotalMinor,
				ExtrasTotalMinor:       extrasTotalMinor,
				ManualExtrasSupported:  true,
				FundingProfileID:       termRow.FundingProfileID,
				FundedAllowanceMinutes: fundedAllowance,
				FundedDeductionMinutes: fundedDeductionMinutes,
				CoreBillableMinutes:    billableMinutes,
				TermID:                 termRow.TermID,
				BookingPatternID:       termRow.BookingPatternID,
				BookedCoreMinutes:      calc.TotalMinutes,
				BookedSessions:         calc.Sessions,
				BookedPerEntry:         calc.PerEntry,
			}
			calcDetailsJSON, jsonErr := domain.MarshalCalculationDetails(calcDetails)
			if jsonErr != nil {
				return fmt.Errorf("marshal calc details: %w", jsonErr)
			}

			var invoiceID uuid.UUID
			var action domain.DraftInvoiceAction
			if invoiceFound {
				invoiceID = existingInvoice.ID
				action = domain.DraftUpdated
				if delErr := uc.repo.DeleteDraftSystemInvoiceLines(ctx, tx, actor.TenantID, actor.BranchID, invoiceID); delErr != nil {
					return fmt.Errorf("delete system lines: %w", delErr)
				}
				if updErr := uc.repo.UpdateDraftInvoice(ctx, tx, domain.DraftInvoiceUpdateParams{
					ID:                   invoiceID,
					TenantID:             actor.TenantID,
					BranchID:             actor.BranchID,
					GeneratedRunID:       runID,
					SubtotalMinor:        subtotalMinor + extrasTotalMinor,
					FundedDeductionMinor: fundedDeductionMinor,
					TotalDueMinor:        totalDueMinor + extrasTotalMinor,
					CalculationDetails:   calcDetailsJSON,
				}); updErr != nil {
					return fmt.Errorf("update draft invoice: %w", updErr)
				}
			} else {
				invoiceID = uid.NewUUID()
				action = domain.DraftCreated
				if createErr := uc.repo.CreateDraftInvoice(ctx, tx, domain.DraftInvoiceCreateParams{
					ID:                   invoiceID,
					TenantID:             actor.TenantID,
					BranchID:             actor.BranchID,
					ChildID:              termRow.ChildID,
					BillingMonth:         billingMonth,
					GeneratedRunID:       runID,
					CurrencyCode:         "GBP",
					SubtotalMinor:        subtotalMinor + extrasTotalMinor,
					FundedDeductionMinor: fundedDeductionMinor,
					TotalDueMinor:        totalDueMinor + extrasTotalMinor,
					PeriodStartDate:      period.StartLocal,
					PeriodEndDate:        period.EndExclusiveLocal.AddDate(0, 0, -1),
					CalculationDetails:   calcDetailsJSON,
				}); createErr != nil {
					return fmt.Errorf("create draft invoice: %w", createErr)
				}
			}

			// 8. Insert core_childcare line.
			coreLineDetails := domain.CoreLineDetails{
				BookedCoreMinutes: calc.TotalMinutes,
				BookedSessions:    calc.Sessions,
				BookedPerEntry:    calc.PerEntry,
			}
			coreLineDetailsJSON, jsonErr := json.Marshal(coreLineDetails)
			if jsonErr != nil {
				return fmt.Errorf("marshal core line details: %w", jsonErr)
			}
			if insErr := uc.repo.InsertInvoiceLine(ctx, tx, domain.InvoiceLineCreateParams{
				ID:              uid.NewUUID(),
				TenantID:        actor.TenantID,
				BranchID:        actor.BranchID,
				InvoiceID:       invoiceID,
				LineKind:        domain.LineKindCoreChildcare,
				Description:     "Core childcare",
				SortOrder:       1,
				QuantityMinutes: calc.TotalMinutes,
				UnitAmountMinor: termRow.SiteHourlyRateMinor,
				LineAmountMinor: subtotalMinor,
				SessionCount:    len(calc.Sessions),
				Details:         coreLineDetailsJSON,
			}); insErr != nil {
				return fmt.Errorf("insert core line: %w", insErr)
			}

			// 9. Insert funded_deduction line.
			deductionLineAmount := -fundedDeductionMinor
			var deductionLineDetailsJSON []byte
			if termRow.FundingProfileID != nil {
				deductionDetails := domain.FundedDeductionLineDetails{
					FundingProfileID:       *termRow.FundingProfileID,
					FundedAllowanceMinutes: fundedAllowance,
					FundedDeductionMinutes: fundedDeductionMinutes,
					CoreBillableMinutes:    billableMinutes,
				}
				deductionLineDetailsJSON, jsonErr = json.Marshal(deductionDetails)
				if jsonErr != nil {
					return fmt.Errorf("marshal deduction line details: %w", jsonErr)
				}
			}
			if insErr := uc.repo.InsertInvoiceLine(ctx, tx, domain.InvoiceLineCreateParams{
				ID:                     uid.NewUUID(),
				TenantID:               actor.TenantID,
				BranchID:               actor.BranchID,
				InvoiceID:              invoiceID,
				LineKind:               domain.LineKindFundedDeduction,
				Description:            "Funded hours deduction",
				SortOrder:              2,
				FundedAllowanceMinutes: fundedAllowance,
				FundedDeductionMinutes: fundedDeductionMinutes,
				CoreBillableMinutes:    billableMinutes,
				LineAmountMinor:        deductionLineAmount,
				Details:                deductionLineDetailsJSON,
			}); insErr != nil {
				return fmt.Errorf("insert deduction line: %w", insErr)
			}

			// 10. Audit.
			auditAction := domain.AuditInvoiceDraftGenerated
			if action == domain.DraftUpdated {
				auditAction = domain.AuditInvoiceDraftRegenerated
			}
			if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: auditAction,
				EntityType: domain.AuditEntityInvoice,
				EntityID:   invoiceID,
				Details: map[string]any{
					"term_id":              termRow.TermID.String(),
					"booking_pattern_id":   termRow.BookingPatternID.String(),
					"billing_month":        billingMonthRaw,
					"booked_core_minutes":  calc.TotalMinutes,
					"funded_deduction_min": fundedDeductionMinor,
					"total_due_minor":      totalDueMinor + extrasTotalMinor,
				},
			}); auditErr != nil {
				return fmt.Errorf("write audit: %w", auditErr)
			}

			generated = append(generated, domain.DraftGenerationChildResult{
				ChildID:              termRow.ChildID,
				ChildFirstName:       termRow.FirstName,
				ChildMiddleName:      termRow.MiddleName,
				ChildLastName:        termRow.LastName,
				Action:               action,
				InvoiceID:            invoiceID,
				SubtotalMinor:        subtotalMinor + extrasTotalMinor,
				FundedDeductionMinor: fundedDeductionMinor,
				TotalDueMinor:        totalDueMinor + extrasTotalMinor,
			})
			totalDueSum += totalDueMinor + extrasTotalMinor
		}

		// 11. Sort the generated slice by structured child name for stable ordering.
		sort.Slice(generated, func(i, j int) bool {
			if generated[i].ChildFirstName != generated[j].ChildFirstName {
				return generated[i].ChildFirstName < generated[j].ChildFirstName
			}
			if stringPtrValue(generated[i].ChildMiddleName) != stringPtrValue(generated[j].ChildMiddleName) {
				return stringPtrValue(generated[i].ChildMiddleName) < stringPtrValue(generated[j].ChildMiddleName)
			}
			if stringPtrValue(generated[i].ChildLastName) != stringPtrValue(generated[j].ChildLastName) {
				return stringPtrValue(generated[i].ChildLastName) < stringPtrValue(generated[j].ChildLastName)
			}
			return generated[i].ChildID.String() < generated[j].ChildID.String()
		})

		// 12. Complete the invoice run.
		runStatus := domain.InvoiceRunStatusCompleted
		if len(blocked) > 0 {
			runStatus = domain.InvoiceRunStatusCompletedWithExceptions
		}
		runDetails := map[string]any{
			"mode":            "full_month",
			"billing_month":   billingMonthRaw,
			"generated_count": len(generated),
			"blocked_count":   len(blocked),
		}
		if !isFullMonth {
			runDetails["mode"] = "selected_children"
			runDetails["requested_child_count"] = len(childIDs)
		}
		if len(blocked) > 0 {
			blockedDetails := make([]map[string]any, 0, len(blocked))
			for _, b := range blocked {
				codes := make([]string, 0, len(b.Blockers))
				for _, bl := range b.Blockers {
					codes = append(codes, string(bl.Code))
				}
				blockedDetails = append(blockedDetails, map[string]any{
					"child_id":          b.ChildID.String(),
					"child_first_name":  b.ChildFirstName,
					"child_middle_name": b.ChildMiddleName,
					"child_last_name":   b.ChildLastName,
					"blockers":          codes,
				})
			}
			runDetails["blocked_children"] = blockedDetails
		}
		detailsJSON, _ := json.Marshal(runDetails)

		if compErr := uc.repo.CompleteInvoiceRun(ctx, tx, domain.InvoiceRunCompleteParams{
			ID:            runID,
			TenantID:      actor.TenantID,
			BranchID:      actor.BranchID,
			Status:        runStatus,
			EligibleCount: len(generated) + len(blocked),
			SuccessCount:  len(generated),
			BlockedCount:  len(blocked),
			Details:       detailsJSON,
		}); compErr != nil {
			return fmt.Errorf("complete invoice run: %w", compErr)
		}

		result = domain.DraftGenerationResult{
			RunID:        runID,
			BillingMonth: billingMonthRaw,
			RunStatus:    runStatus,
			Generated:    generated,
			Blocked:      blocked,
			Summary: domain.DraftGenerationSummary{
				EligibleCount: len(generated) + len(blocked),
				SuccessCount:  len(generated),
				BlockedCount:  len(blocked),
				TotalDueMinor: totalDueSum,
			},
		}
		return nil
	})

	if txErr != nil {
		uc.recordOutcome("full_month", "error", startedAt, result, actor)
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
	uc.recordOutcome(mode, outcome, startedAt, result, actor)
	return result, nil
}

func (uc *GenerateDraftInvoicesUseCase) preflightTerm(row domain.AdvancePayTermRow) []domain.PreflightBlocker {
	var blockers []domain.PreflightBlocker
	if row.FirstName == "" {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildName, Message: "Child first name is missing.",
		})
	}
	if row.DateOfBirth.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildDateOfBirth, Message: "Child date of birth is missing.",
		})
	}
	if row.StartDate.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildStartDate, Message: "Child start date is missing.",
		})
	}
	if !row.HasParentCarerContact {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingGuardianLink, Message: "No active guardian linked to this child.",
		})
	}
	if row.SiteHourlyRateMinor <= 0 {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingBillingRate, Message: "Site billing rate is missing or invalid.",
		})
	}
	if row.FundingProfileID == nil {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingFundingProfile,
			Message: "Funding profile is missing for this billing month.",
			Field:   strPtr("funding_profile"),
		})
	}
	return blockers
}

func (uc *GenerateDraftInvoicesUseCase) recordOutcome(mode, outcome string, startedAt time.Time, result domain.DraftGenerationResult, actor tenant.ActorContext) {
	elapsed := time.Since(startedAt).Seconds()
	if uc.recorder != nil {
		uc.recorder.InvoiceGenerationRun(mode, outcome, elapsed)
		for _, b := range result.Blocked {
			for _, bl := range b.Blockers {
				uc.recorder.InvoiceGenerationBlocker(string(bl.Code), 1)
			}
		}
	}
	if uc.logger != nil {
		args := []any{
			"operation", "advance_pay_draft_generation",
			"outcome", outcome,
			"run_id", result.RunID.String(),
			"billing_month", result.BillingMonth,
			"mode", mode,
			"eligible_count", result.Summary.EligibleCount,
			"success_count", result.Summary.SuccessCount,
			"blocked_count", result.Summary.BlockedCount,
			"total_due_minor", result.Summary.TotalDueMinor,
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"request_id", actor.RequestID,
			"correlation_id", actor.CorrelationID,
		}
		if actor.TraceID != "" {
			args = append(args, "trace_id", actor.TraceID)
		}
		uc.logger.Info("advance_pay_draft_generation", args...)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
