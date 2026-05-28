package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type GenerateDraftInvoices struct {
	repo   domain.BillingRepository
	txMgr  *transaction.Manager
	auditW *audit.Writer
}

func NewGenerateDraftInvoices(
	repo domain.BillingRepository,
	txMgr *transaction.Manager,
	auditW *audit.Writer,
) *GenerateDraftInvoices {
	return &GenerateDraftInvoices{repo: repo, txMgr: txMgr, auditW: auditW}
}

func (uc *GenerateDraftInvoices) Execute(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string, rawChildIDs []string) (domain.DraftGenerationResult, error) {
	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.DraftGenerationResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	childIDs, err := parseAndDedupeChildIDs(rawChildIDs)
	if err != nil {
		return domain.DraftGenerationResult{}, domainerrors.Validation("Invalid child ID format.", "child_ids")
	}

	year := billingMonth.Year()
	month := billingMonth.Month()

	period, err := domain.NewBillingPeriod(year, month)
	if err != nil {
		return domain.DraftGenerationResult{}, domainerrors.Internal(fmt.Errorf("billing period: %w", err))
	}

	isFullMonth := len(rawChildIDs) == 0
	nextMonth := month + 1
	nextYear := year
	if nextMonth > time.December {
		nextMonth = time.January
		nextYear++
	}
	nextBillingMonth := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, time.UTC)

	runID := uid.NewUUID()

	var result domain.DraftGenerationResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx domain.Tx) error {
		// Create invoice run.
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

		// Load children.
		var children []domain.PreflightChildRow
		if isFullMonth {
			children, runErr = uc.repo.ListCandidateChildrenForUpdate(ctx, tx, actor.TenantID, actor.BranchID, billingMonth, nextBillingMonth)
		} else {
			children, runErr = uc.repo.ListSelectedChildrenForUpdate(ctx, tx, actor.TenantID, actor.BranchID, childIDs)
		}
		if runErr != nil {
			return fmt.Errorf("list children: %w", runErr)
		}

		// Build lookup of found children for selected-child exceptions.
		foundChildSet := make(map[uuid.UUID]domain.PreflightChildRow, len(children))
		for _, c := range children {
			foundChildSet[c.ChildID] = c
		}

		// For selected-child mode, build exceptions for unknown / out-of-month children.
		var blocked []domain.DraftGenerationBlockedChild
		if !isFullMonth {
			for _, id := range childIDs {
				child, found := foundChildSet[id]
				if !found {
					// Could be unknown or out-of-scope. Check if child exists at all.
					// Since ListSelectedChildrenForUpdate already scopes to tenant/branch,
					// not found means either child doesn't exist or wrong scope.
					blocked = append(blocked, domain.DraftGenerationBlockedChild{
						ChildID:   id,
						ChildName: "",
						Blockers: []domain.PreflightBlocker{
							{Code: domain.BlockerChildNotFound, Message: "Child not found or not in scope."},
						},
					})
					continue
				}
				// Check if child overlaps billing month.
				if !childInBillingMonth(child, billingMonth, nextBillingMonth) {
					blocked = append(blocked, domain.DraftGenerationBlockedChild{
						ChildID:   child.ChildID,
						ChildName: child.FullName,
						Blockers: []domain.PreflightBlocker{
							{Code: domain.BlockerChildNotInBillingMonth, Message: "Child is not active during this billing month."},
						},
					})
					delete(foundChildSet, id)
				}
			}
		}

		// Load attendance sessions.
		sessions, runErr := uc.repo.ListAttendanceSessions(ctx, tx, actor.TenantID, actor.BranchID, period.StartLocal, period.EndExclusiveLocal)
		if runErr != nil {
			return fmt.Errorf("list attendance sessions: %w", runErr)
		}

		sessionsByChild := make(map[uuid.UUID][]domain.AttendanceSessionInput)
		for _, s := range sessions {
			sessionsByChild[s.ChildID] = append(sessionsByChild[s.ChildID], domain.AttendanceSessionInput{
				SessionID:  s.SessionID,
				Status:     s.Status,
				CheckInAt:  s.CheckInAt,
				CheckOutAt: s.CheckOutAt,
			})
		}

		var generated []domain.DraftGenerationChildResult
		var totalDueSum int

		for _, child := range children {
			if _, ok := foundChildSet[child.ChildID]; !ok && !isFullMonth {
				continue // already in blocked from selected-child exceptions
			}

			childSessions := sessionsByChild[child.ChildID]
			readiness, readinessErr := EvaluateChildReadiness(child, childSessions, year, int(month))
			if readinessErr != nil {
				return fmt.Errorf("evaluate readiness for child %s: %w", child.ChildID, readinessErr)
			}

			if len(readiness.Blockers) > 0 {
				// If there's an existing draft, and the child is now blocked, leave it in place.
				blocked = append(blocked, domain.DraftGenerationBlockedChild{
					ChildID:   child.ChildID,
					ChildName: child.FullName,
					Blockers:  readiness.Blockers,
				})
				continue
			}

			// Check for existing monthly invoice under lock.
			existingInvoice, invoiceFound, runErr := uc.repo.GetMonthlyInvoiceForUpdate(ctx, tx, actor.TenantID, actor.BranchID, child.ChildID, billingMonth)
			if runErr != nil {
				return fmt.Errorf("get monthly invoice for child %s: %w", child.ChildID, runErr)
			}

			if invoiceFound && existingInvoice.Status != domain.InvoiceStatusDraft {
				// Non-draft invoice exists — block.
				blocked = append(blocked, domain.DraftGenerationBlockedChild{
					ChildID:   child.ChildID,
					ChildName: child.FullName,
					Blockers: []domain.PreflightBlocker{
						{
							Code:    domain.BlockerInvoiceAlreadyIssued,
							Message: "A monthly invoice has already been issued for this child and billing month.",
						},
					},
				})
				continue
			}

			// Build calculation details and source sessions.
			calcDetails := uc.buildCalculationDetails(child, readiness, billingMonthRaw)
			calcDetailsJSON, jsonErr := domain.MarshalCalculationDetails(calcDetails)
			if jsonErr != nil {
				return fmt.Errorf("marshal calc details for child %s: %w", child.ChildID, jsonErr)
			}

			// Get existing extra lines (for update case) or empty.
			var extrasTotalMinor int
			var existingExtras []domain.ExtraLineRow
			if invoiceFound {
				existingExtras, runErr = uc.repo.ListDraftExtraLines(ctx, tx, actor.TenantID, actor.BranchID, existingInvoice.ID)
				if runErr != nil {
					return fmt.Errorf("list extra lines for child %s: %w", child.ChildID, runErr)
				}
			}
			for _, ex := range existingExtras {
				extrasTotalMinor += ex.LineAmountMinor
			}

			coreSubtotal := readiness.SubtotalMinor
			fundedDeductionMinor := readiness.FundedDeductionMinor
			subtotalMinor := coreSubtotal + extrasTotalMinor
			totalDueMinor := max(0, subtotalMinor-fundedDeductionMinor)

			var invoiceID uuid.UUID
			var action domain.DraftInvoiceAction

			if invoiceFound {
				// Update existing draft.
				invoiceID = existingInvoice.ID
				action = domain.DraftUpdated

				// Delete system lines, then re-insert.
				if delErr := uc.repo.DeleteDraftSystemInvoiceLines(ctx, tx, actor.TenantID, actor.BranchID, invoiceID); delErr != nil {
					return fmt.Errorf("delete system lines for child %s: %w", child.ChildID, delErr)
				}

				// Update calculation_details in the JSON to include extras_total_minor.
				calcDetails.ExtrasTotalMinor = extrasTotalMinor
				calcDetailsJSON, jsonErr = domain.MarshalCalculationDetails(calcDetails)
				if jsonErr != nil {
					return fmt.Errorf("re-marshal calc details: %w", jsonErr)
				}

				if updErr := uc.repo.UpdateDraftInvoice(ctx, tx, domain.DraftInvoiceUpdateParams{
					ID:                   invoiceID,
					TenantID:             actor.TenantID,
					BranchID:             actor.BranchID,
					GeneratedRunID:       runID,
					SubtotalMinor:        subtotalMinor,
					FundedDeductionMinor: fundedDeductionMinor,
					TotalDueMinor:        totalDueMinor,
					CalculationDetails:   calcDetailsJSON,
				}); updErr != nil {
					return fmt.Errorf("update draft invoice for child %s: %w", child.ChildID, updErr)
				}
			} else {
				// Create new draft.
				invoiceID = uid.NewUUID()
				action = domain.DraftCreated

				calcDetails.ExtrasTotalMinor = extrasTotalMinor
				calcDetailsJSON, jsonErr = domain.MarshalCalculationDetails(calcDetails)
				if jsonErr != nil {
					return fmt.Errorf("re-marshal calc details: %w", jsonErr)
				}

				if createErr := uc.repo.CreateDraftInvoice(ctx, tx, domain.DraftInvoiceCreateParams{
					ID:                   invoiceID,
					TenantID:             actor.TenantID,
					BranchID:             actor.BranchID,
					ChildID:              child.ChildID,
					BillingMonth:         billingMonth,
					GeneratedRunID:       runID,
					CurrencyCode:         "GBP",
					SubtotalMinor:        subtotalMinor,
					FundedDeductionMinor: fundedDeductionMinor,
					TotalDueMinor:        totalDueMinor,
					PeriodStartDate:      period.StartLocal,
					PeriodEndDate:        period.EndExclusiveLocal.AddDate(0, 0, -1),
					CalculationDetails:   calcDetailsJSON,
				}); createErr != nil {
					return fmt.Errorf("create draft invoice for child %s: %w", child.ChildID, createErr)
				}
			}

			// Insert core_childcare line.
			coreLineDetails := domain.CoreLineDetails{
				RawAttendedMinutes:     readiness.RawAttendedMinutes,
				RoundedAttendedMinutes: readiness.RoundedAttendedMinutes,
				IncludedSessionCount:   readiness.IncludedSessionCount,
				CoreBillableMinutes:    readiness.FundingCalc.CoreBillableMinutes,
				SourceSessions:         calcDetails.SourceSessions,
			}
			coreLineDetailsJSON, jsonErr := json.Marshal(coreLineDetails)
			if jsonErr != nil {
				return fmt.Errorf("marshal core line details: %w", jsonErr)
			}

			if insErr := uc.repo.InsertInvoiceLine(ctx, tx, domain.InvoiceLineCreateParams{
				ID:                     uid.NewUUID(),
				TenantID:               actor.TenantID,
				BranchID:               actor.BranchID,
				InvoiceID:              invoiceID,
				LineKind:               domain.LineKindCoreChildcare,
				Description:            "Core childcare",
				SortOrder:              1,
				QuantityMinutes:        readiness.RoundedAttendedMinutes,
				UnitAmountMinor:        child.CoreHourlyRateMinor,
				LineAmountMinor:        coreSubtotal,
				RawAttendedMinutes:     readiness.RawAttendedMinutes,
				RoundedAttendedMinutes: readiness.RoundedAttendedMinutes,
				SessionCount:           readiness.IncludedSessionCount,
				Details:                coreLineDetailsJSON,
			}); insErr != nil {
				return fmt.Errorf("insert core line for child %s: %w", child.ChildID, insErr)
			}

			// Insert funded_deduction line (negative or zero amount).
			deductionLineAmount := -fundedDeductionMinor
			var deductionLineDetailsJSON []byte
			if child.FundingProfileID != nil {
				deductionDetails := domain.FundedDeductionLineDetails{
					FundingProfileID:       *child.FundingProfileID,
					FundedAllowanceMinutes: readiness.FundedAllowanceMinutes,
					FundedDeductionMinutes: readiness.FundingCalc.FundedDeductionMinutes,
					CoreBillableMinutes:    readiness.FundingCalc.CoreBillableMinutes,
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
				FundedAllowanceMinutes: readiness.FundedAllowanceMinutes,
				FundedDeductionMinutes: readiness.FundingCalc.FundedDeductionMinutes,
				CoreBillableMinutes:    readiness.FundingCalc.CoreBillableMinutes,
				LineAmountMinor:        deductionLineAmount,
				Details:                deductionLineDetailsJSON,
			}); insErr != nil {
				return fmt.Errorf("insert deduction line for child %s: %w", child.ChildID, insErr)
			}

			// Write audit.
			auditAction := domain.AuditInvoiceDraftGenerated
			if action == domain.DraftUpdated {
				auditAction = domain.AuditInvoiceDraftRegenerated
			}
			if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: auditAction,
				EntityType: domain.AuditEntityInvoice,
				EntityID:   invoiceID,
			}); auditErr != nil {
				return fmt.Errorf("write audit for invoice %s: %w", invoiceID, auditErr)
			}

			generated = append(generated, domain.DraftGenerationChildResult{
				ChildID:              child.ChildID,
				ChildName:            child.FullName,
				Action:               action,
				InvoiceID:            invoiceID,
				SubtotalMinor:        subtotalMinor,
				FundedDeductionMinor: fundedDeductionMinor,
				TotalDueMinor:        totalDueMinor,
			})
			totalDueSum += totalDueMinor
		}

		// Complete invoice run.
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
					"child_id":   b.ChildID.String(),
					"child_name": b.ChildName,
					"blockers":   codes,
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
		return domain.DraftGenerationResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func (uc *GenerateDraftInvoices) buildCalculationDetails(child domain.PreflightChildRow, readiness ChildReadinessResult, billingMonth string) domain.InvoiceCalculationDetails {
	sourceSessions := make([]domain.SourceSessionSnapshot, 0, len(readiness.AttendanceCalc.Sessions))
	for _, s := range readiness.AttendanceCalc.Sessions {
		sourceSessions = append(sourceSessions, domain.SourceSessionSnapshot{
			SessionID:              s.SessionID,
			Status:                 string(s.Status),
			CheckInAt:              s.CheckInAt,
			CheckOutAt:             &s.CheckOutAt,
			RawElapsedMinutes:      s.RawElapsedMinutes,
			RoundedBillableMinutes: s.RoundedBillableMinutes,
		})
	}

	return domain.InvoiceCalculationDetails{
		BillingMonth:           billingMonth,
		ChildID:                child.ChildID,
		CoreHourlyRateMinor:    child.CoreHourlyRateMinor,
		CoreSubtotalMinor:      readiness.SubtotalMinor,
		ManualExtrasSupported:  true,
		FundingProfileID:       child.FundingProfileID,
		FundedAllowanceMinutes: readiness.FundedAllowanceMinutes,
		FundedDeductionMinutes: readiness.FundingCalc.FundedDeductionMinutes,
		CoreBillableMinutes:    readiness.FundingCalc.CoreBillableMinutes,
		RawAttendedMinutes:     readiness.RawAttendedMinutes,
		RoundedAttendedMinutes: readiness.RoundedAttendedMinutes,
		IncludedSessionCount:   readiness.IncludedSessionCount,
		SourceSessions:         sourceSessions,
	}
}

func parseAndDedupeChildIDs(raw []string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	seen := make(map[uuid.UUID]struct{}, len(raw))
	result := make([]uuid.UUID, 0, len(raw))
	for _, r := range raw {
		id, err := uuid.Parse(r)
		if err != nil {
			return nil, fmt.Errorf("invalid child_id %q: %w", r, err)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func childInBillingMonth(child domain.PreflightChildRow, billingMonth, nextBillingMonth time.Time) bool {
	if child.StartDate.IsZero() {
		return false
	}
	if !child.StartDate.Before(nextBillingMonth) {
		return false
	}
	if child.EndDate != nil && child.EndDate.Before(billingMonth) {
		return false
	}
	return true
}
