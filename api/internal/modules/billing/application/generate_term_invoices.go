package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type GenerateTermInvoices struct {
	repo              domain.BillingRepository
	auditW            *audit.Writer
	termDateLookup    domain.TermDateLookup
	adHocLookup       domain.AdHocBookingLookup
	closureDateLookup domain.ClosureDateLookup
}

func NewGenerateTermInvoices(repo domain.BillingRepository, auditW *audit.Writer, termDateLookup domain.TermDateLookup, adHocLookup domain.AdHocBookingLookup, closureDateLookup domain.ClosureDateLookup) *GenerateTermInvoices {
	return &GenerateTermInvoices{repo: repo, auditW: auditW, termDateLookup: termDateLookup, adHocLookup: adHocLookup, closureDateLookup: closureDateLookup}
}

type GenerateTermInvoicesInput struct {
	Tx              pgx.Tx
	Actor           tenant.ActorContext
	BillingMonth    time.Time
	BillingMonthRaw string
	Period          domain.BillingPeriod
	RunID           uuid.UUID
}

type GenerateTermInvoicesOutput struct {
	Generated   []domain.DraftGenerationChildResult
	Blocked     []domain.DraftGenerationBlockedChild
	TotalDueSum int
}

func (uc *GenerateTermInvoices) Execute(ctx context.Context, in GenerateTermInvoicesInput, terms []domain.AdvancePayTermRow, requestedSet map[uuid.UUID]struct{}, isFullMonth bool) (GenerateTermInvoicesOutput, error) {
	var generated []domain.DraftGenerationChildResult
	var blocked []domain.DraftGenerationBlockedChild
	var totalDueSum int

	for _, termRow := range terms {
		if !isFullMonth {
			if _, ok := requestedSet[termRow.ChildID]; !ok {
				continue
			}
		}

		preflightBlockers := preflightBlockers(termRow)
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

		existingInvoice, invoiceFound, err := uc.repo.GetMonthlyInvoiceForUpdate(ctx, in.Tx, in.Actor.TenantID, in.Actor.BranchID, termRow.ChildID, in.BillingMonth)
		if err != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("get monthly invoice: %w", err)
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

		entries, err := uc.repo.ListBookingPatternEntries(ctx, in.Tx, in.Actor.TenantID, in.Actor.BranchID, termRow.BookingPatternID)
		if err != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("list booking pattern entries: %w", err)
		}

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
					Kind:            e.SessionTypeKind,
					FlatFeeMinor:    e.SessionTypeFlatFeeMinor,
				},
			})
		}

		var termDates []domain.TermDateRange
		var termDatesUsedLabels []string
		if termRow.TermTimeOnly && uc.termDateLookup != nil {
			termDates, err = uc.termDateLookup.GetTermDateRangesForBranchAndMonth(ctx, in.Actor.TenantID, in.Actor.BranchID, in.BillingMonth)
			if err != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("lookup term dates for term %s: %w", termRow.TermID, err)
			}
			for _, r := range termDates {
				termDatesUsedLabels = append(termDatesUsedLabels, fmt.Sprintf("%s to %s", r.StartDate.Format("2006-01-02"), r.EndDate.Format("2006-01-02")))
			}
		}

		var closureDates []time.Time
		var closureDaysExcludedLabels []string
		if uc.closureDateLookup != nil {
			closureDates, err = uc.closureDateLookup.GetClosureDatesForBranchAndMonth(ctx, in.Actor.TenantID, in.Actor.BranchID, in.BillingMonth)
			if err != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("lookup closure dates: %w", err)
			}
			for _, cd := range closureDates {
				closureDaysExcludedLabels = append(closureDaysExcludedLabels, cd.Format("2006-01-02"))
			}
		}

		calc, calcErr := domain.CalculateBookedCoreMinutesInMonth(
			termRow.BookingPatternID.String(),
			domainEntries,
			in.BillingMonth,
			termRow.SiteHourlyRateMinor,
			termDates,
			closureDates,
		)
		if calcErr != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("calculate booked minutes for term %s: %w", termRow.TermID, calcErr)
		}

		subtotalMinor := calc.Subtotal.Minor()

		fundedAllowance := 0
		fundingModel := termRow.FundingModel
		if fundingModel == "" {
			fundingModel = "unknown"
		}
		if termRow.FundedAllowanceMinutes != nil {
			fundedAllowance = *termRow.FundedAllowanceMinutes
		} else if fundingModel == "term_time_only" && termRow.FundedHoursPerWeek != nil {
			if len(termDates) > 0 {
				fundedAllowance = domain.CalculateTermTimeFundedAllowanceMinutes(*termRow.FundedHoursPerWeek, termDates, in.BillingMonth, closureDates)
			}
		}

		_, billableMinutes, _, billableMinor, err := domain.ComputeFundedDeductionMinor(
			calc.TotalMinutes, fundedAllowance, termRow.SiteHourlyRateMinor,
		)
		if err != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("compute funded deduction for term %s: %w", termRow.TermID, err)
		}
		fundedDeductionMinutes := minInt(calc.TotalMinutes, fundedAllowance)
		fundedDeductionMinor := subtotalMinor - billableMinor
		if fundedDeductionMinor < 0 {
			fundedDeductionMinor = 0
		}

		totalDueMinor := billableMinor

		extrasTotalMinor := 0
		var existingExtras []domain.ExtraLineRow
		if invoiceFound {
			existingExtras, err = uc.repo.ListDraftExtraLines(ctx, in.Tx, in.Actor.TenantID, in.Actor.BranchID, existingInvoice.ID)
			if err != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("list extra lines: %w", err)
			}
			for _, ex := range existingExtras {
				extrasTotalMinor += ex.LineAmountMinor
			}
		}

		var adHocLines []struct {
			description string
			minutes     int
			unitMinor   int
			lineMinor   int
		}
		adHocTotalMinor := 0
		if uc.adHocLookup != nil {
			monthEnd := in.BillingMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)
			adHocBookings, adHocErr := uc.adHocLookup.ListActiveBookingsForChildInMonth(ctx, in.Actor.TenantID, in.Actor.BranchID, termRow.ChildID, in.BillingMonth)
			if adHocErr != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("lookup ad-hoc bookings for term %s: %w", termRow.TermID, adHocErr)
			}
			for _, ab := range adHocBookings {
				if ab.CalendarDate.Before(in.BillingMonth) || ab.CalendarDate.After(monthEnd) {
					continue
				}
				duration := ab.EndMinutes - ab.StartMinutes
				if duration <= 0 {
					continue
				}
				var lineMinor int
				var lineDesc string
				sortOrder := 3
				_ = sortOrder
				if ab.FlatFeeMinor != nil {
					lineMinor = *ab.FlatFeeMinor
					lineDesc = fmt.Sprintf("Ad-hoc session: %s on %s (flat fee)", ab.SessionTypeName, ab.CalendarDate.Format("02 Jan"))
				} else {
					multiplier := termRow.AdHocRateMultiplier
					if multiplier <= 0 {
						multiplier = 1.50
					}
					chargedMinutes := domain.CalculateAdHocChargeMinutes(duration, multiplier)
					minor, hrErr := domain.CalculateHourlyAmountMinor(chargedMinutes, termRow.SiteHourlyRateMinor)
					if hrErr != nil {
						return GenerateTermInvoicesOutput{}, fmt.Errorf("calculate ad-hoc charge: %w", hrErr)
					}
					lineMinor = minor
					lineDesc = fmt.Sprintf("Ad-hoc session: %s on %s (×%.2f)", ab.SessionTypeName, ab.CalendarDate.Format("02 Jan"), multiplier)
				}
				adHocLines = append(adHocLines, struct {
					description string
					minutes     int
					unitMinor   int
					lineMinor   int
				}{
					description: lineDesc,
					minutes:     duration,
					unitMinor:   lineMinor,
					lineMinor:   lineMinor,
				})
				adHocTotalMinor += lineMinor
			}
		}

		subtotalMinor += adHocTotalMinor
		totalDueMinor += adHocTotalMinor

		calcDetails := domain.InvoiceCalculationDetails{
			BillingMonth:           in.BillingMonthRaw,
			ChildID:                termRow.ChildID,
			CoreHourlyRate:         domain.MustGBP(termRow.SiteHourlyRateMinor),
			CoreSubtotal:           domain.MustGBP(subtotalMinor),
			ExtrasTotal:            domain.MustGBP(extrasTotalMinor),
			ManualExtrasSupported:  true,
			FundingProfileID:       termRow.FundingProfileID,
			FundedAllowanceMinutes: fundedAllowance,
			FundedDeductionMinutes: fundedDeductionMinutes,
			CoreBillableMinutes:    billableMinutes,
			TermTimeOnly:           termRow.TermTimeOnly,
			FundingModel:           fundingModel,
			TermDatesUsed:          termDatesUsedLabels,
			ClosureDaysExcluded:    closureDaysExcludedLabels,
			TermID:                 termRow.TermID,
			BookingPatternID:       termRow.BookingPatternID,
			BookedCoreMinutes:      calc.TotalMinutes,
			BookedSessions:         calc.Sessions,
			BookedPerEntry:         calc.PerEntry,
		}
		calcDetailsJSON, jsonErr := domain.MarshalCalculationDetails(calcDetails)
		if jsonErr != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("marshal calc details: %w", jsonErr)
		}

		var invoiceID uuid.UUID
		var action domain.DraftInvoiceAction
		if invoiceFound {
			invoiceID = existingInvoice.ID
			action = domain.DraftUpdated
			if delErr := uc.repo.DeleteDraftSystemInvoiceLines(ctx, in.Tx, in.Actor.TenantID, in.Actor.BranchID, invoiceID); delErr != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("delete system lines: %w", delErr)
			}
			if updErr := uc.repo.UpdateDraftInvoice(ctx, in.Tx, domain.DraftInvoiceUpdateParams{
				ID:                 invoiceID,
				TenantID:           in.Actor.TenantID,
				BranchID:           in.Actor.BranchID,
				GeneratedRunID:     in.RunID,
				Subtotal:           domain.MustGBP(subtotalMinor + extrasTotalMinor),
				FundedDeduction:    domain.MustGBP(fundedDeductionMinor),
				TotalDue:           domain.MustGBP(totalDueMinor + extrasTotalMinor),
				CalculationDetails: calcDetailsJSON,
			}); updErr != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("update draft invoice: %w", updErr)
			}
		} else {
			invoiceID = uid.NewUUID()
			action = domain.DraftCreated
			if createErr := uc.repo.CreateDraftInvoice(ctx, in.Tx, domain.DraftInvoiceCreateParams{
				ID:                 invoiceID,
				TenantID:           in.Actor.TenantID,
				BranchID:           in.Actor.BranchID,
				ChildID:            termRow.ChildID,
				BillingMonth:       in.BillingMonth,
				GeneratedRunID:     in.RunID,
				CurrencyCode:       "GBP",
				Subtotal:           domain.MustGBP(subtotalMinor + extrasTotalMinor),
				FundedDeduction:    domain.MustGBP(fundedDeductionMinor),
				TotalDue:           domain.MustGBP(totalDueMinor + extrasTotalMinor),
				PeriodStartDate:    in.Period.StartLocal,
				PeriodEndDate:      in.Period.EndExclusiveLocal.AddDate(0, 0, -1),
				CalculationDetails: calcDetailsJSON,
			}); createErr != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("create draft invoice: %w", createErr)
			}
		}

		coreLineDetails := domain.CoreLineDetails{
			BookedCoreMinutes: calc.TotalMinutes,
			BookedSessions:    calc.Sessions,
			BookedPerEntry:    calc.PerEntry,
		}
		coreLineDetailsJSON, jsonErr := json.Marshal(coreLineDetails)
		if jsonErr != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("marshal core line details: %w", jsonErr)
		}
		if insErr := uc.repo.InsertInvoiceLine(ctx, in.Tx, domain.InvoiceLineCreateParams{
			ID:              uid.NewUUID(),
			TenantID:        in.Actor.TenantID,
			BranchID:        in.Actor.BranchID,
			InvoiceID:       invoiceID,
			LineKind:        domain.LineKindCoreChildcare,
			Description:     "Core childcare",
			SortOrder:       1,
			QuantityMinutes: calc.TotalMinutes,
			UnitAmount:      domain.MustGBP(termRow.SiteHourlyRateMinor),
			LineAmount:      domain.MustGBP(calc.Subtotal.Minor()),
			SessionCount:    len(calc.Sessions),
			Details:         coreLineDetailsJSON,
		}); insErr != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("insert core line: %w", insErr)
		}

		deductionLineAmount := -fundedDeductionMinor
		var deductionLineDetailsJSON []byte
		if termRow.FundingProfileID != nil || fundedAllowance > 0 {
			deductionDetails := domain.FundedDeductionLineDetails{
				FundedAllowanceMinutes: fundedAllowance,
				FundedDeductionMinutes: fundedDeductionMinutes,
				CoreBillableMinutes:    billableMinutes,
				FundingModel:           fundingModel,
			}
			if termRow.FundingProfileID != nil {
				deductionDetails.FundingProfileID = *termRow.FundingProfileID
			}
			deductionLineDetailsJSON, jsonErr = json.Marshal(deductionDetails)
			if jsonErr != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("marshal deduction line details: %w", jsonErr)
			}
		}
		deductionLineAmountAbs := deductionLineAmount
		if deductionLineAmountAbs < 0 {
			deductionLineAmountAbs = -deductionLineAmountAbs
		}
		if insErr := uc.repo.InsertInvoiceLine(ctx, in.Tx, domain.InvoiceLineCreateParams{
			ID:                     uid.NewUUID(),
			TenantID:               in.Actor.TenantID,
			BranchID:               in.Actor.BranchID,
			InvoiceID:              invoiceID,
			LineKind:               domain.LineKindFundedDeduction,
			Description:            "Funded hours deduction",
			SortOrder:              2,
			FundedAllowanceMinutes: fundedAllowance,
			FundedDeductionMinutes: fundedDeductionMinutes,
			CoreBillableMinutes:    billableMinutes,
			LineAmount:             domain.MustGBP(deductionLineAmountAbs),
			Details:                deductionLineDetailsJSON,
		}); insErr != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("insert deduction line: %w", insErr)
		}

		for i, ahLine := range adHocLines {
			if insErr := uc.repo.InsertInvoiceLine(ctx, in.Tx, domain.InvoiceLineCreateParams{
				ID:              uid.NewUUID(),
				TenantID:        in.Actor.TenantID,
				BranchID:        in.Actor.BranchID,
				InvoiceID:       invoiceID,
				LineKind:        domain.LineKindAdHoc,
				Description:     ahLine.description,
				SortOrder:       3 + i,
				QuantityMinutes: ahLine.minutes,
				UnitAmount:      domain.MustGBP(ahLine.unitMinor),
				LineAmount:      domain.MustGBP(ahLine.lineMinor),
				SessionCount:    1,
			}); insErr != nil {
				return GenerateTermInvoicesOutput{}, fmt.Errorf("insert ad-hoc line: %w", insErr)
			}
		}

		auditAction := domain.AuditInvoiceDraftGenerated
		if action == domain.DraftUpdated {
			auditAction = domain.AuditInvoiceDraftRegenerated
		}
		if auditErr := uc.auditW.WriteWithTx(ctx, in.Tx, in.Actor, audit.WriteParams{
			ActionType: auditAction,
			EntityType: domain.AuditEntityInvoice,
			EntityID:   invoiceID,
			Details: map[string]any{
				"term_id":              termRow.TermID.String(),
				"booking_pattern_id":   termRow.BookingPatternID.String(),
				"billing_month":        in.BillingMonthRaw,
				"booked_core_minutes":  calc.TotalMinutes,
				"funded_deduction_min": fundedDeductionMinor,
				"total_due_minor":      totalDueMinor + extrasTotalMinor,
				"ad_hoc_total_minor":   adHocTotalMinor,
			},
		}); auditErr != nil {
			return GenerateTermInvoicesOutput{}, fmt.Errorf("write audit: %w", auditErr)
		}

		generated = append(generated, domain.DraftGenerationChildResult{
			ChildID:         termRow.ChildID,
			ChildFirstName:  termRow.FirstName,
			ChildMiddleName: termRow.MiddleName,
			ChildLastName:   termRow.LastName,
			Action:          action,
			InvoiceID:       invoiceID,
			Subtotal:        domain.MustGBP(subtotalMinor + extrasTotalMinor),
			FundedDeduction: domain.MustGBP(fundedDeductionMinor),
			TotalDue:        domain.MustGBP(totalDueMinor + extrasTotalMinor),
		})
		totalDueSum += totalDueMinor + extrasTotalMinor
	}

	return GenerateTermInvoicesOutput{
		Generated:   generated,
		Blocked:     blocked,
		TotalDueSum: totalDueSum,
	}, nil
}
