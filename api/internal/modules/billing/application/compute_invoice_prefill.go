package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type PrefillTxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type ComputeInvoicePrefill struct {
	repo  domain.BillingRepository
	txMgr PrefillTxManager
}

func NewComputeInvoicePrefill(repo domain.BillingRepository, txMgr PrefillTxManager) *ComputeInvoicePrefill {
	return &ComputeInvoicePrefill{repo: repo, txMgr: txMgr}
}

type ComputeInvoicePrefillResult struct {
	ChildID                uuid.UUID
	ChildFirstName         string
	ChildMiddleName        *string
	ChildLastName          *string
	BillingMonth           string
	FundingProfileID       *uuid.UUID
	FundedAllowanceMinutes int
	Lines                  []PrefillLine
	SubtotalMinor          int
	FundedDeductionMinor   int
	TotalDueMinor          int
	Warnings               []string
}

type PrefillLine struct {
	LineKind               string
	Description            string
	SortOrder              int
	QuantityMinutes        int
	UnitAmountMinor        int
	LineAmountMinor        int
	FundedAllowanceMinutes int
	FundedDeductionMinutes int
	CoreBillableMinutes    int
	SessionCount           int
}

func (uc *ComputeInvoicePrefill) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw, billingMonthRaw string) (ComputeInvoicePrefillResult, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return ComputeInvoicePrefillResult{}, domainerrors.Validation("Invalid child ID format.", "child_id")
	}

	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return ComputeInvoicePrefillResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	var result ComputeInvoicePrefillResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		terms, listErr := uc.repo.ListActiveTermsForGeneration(ctx, tx, actor.TenantID, actor.BranchID, billingMonth)
		if listErr != nil {
			return fmt.Errorf("list active terms: %w", listErr)
		}

		var termRow *domain.AdvancePayTermRow
		for i, t := range terms {
			if t.ChildID == childID {
				termRow = &terms[i]
				break
			}
		}
		if termRow == nil {
			return domainerrors.NotFound("child", "Child not found for this billing month.")
		}

		warnings := prefillWarnings(*termRow)

		entries, entriesErr := uc.repo.ListBookingPatternEntries(ctx, tx, actor.TenantID, actor.BranchID, termRow.BookingPatternID)
		if entriesErr != nil {
			return fmt.Errorf("list booking pattern entries: %w", entriesErr)
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
			return fmt.Errorf("calculate booked minutes: %w", calcErr)
		}

		fundedAllowance := 0
		if termRow.FundedAllowanceMinutes != nil {
			fundedAllowance = *termRow.FundedAllowanceMinutes
		}

		subtotalMinor := calc.Subtotal.Minor()
		fundedDeductionMinor := 0
		fundedDeductionMinutes := 0
		billableMinutes := calc.TotalMinutes

		if termRow.FundingProfileID != nil {
			var fundErr error
			fundedDeductionMinutes, billableMinutes, fundedDeductionMinor, _, fundErr = domain.ComputeFundedDeductionMinor(
				calc.TotalMinutes, fundedAllowance, termRow.SiteHourlyRateMinor,
			)
			if fundErr != nil {
				return fmt.Errorf("compute funded deduction: %w", fundErr)
			}
		}

		totalDueMinor := subtotalMinor - fundedDeductionMinor
		if totalDueMinor < 0 {
			totalDueMinor = 0
		}

		lines := make([]PrefillLine, 0, 2)
		lines = append(lines, PrefillLine{
			LineKind:               domain.LineKindCoreChildcare,
			Description:            "Core childcare",
			SortOrder:              1,
			QuantityMinutes:        calc.TotalMinutes,
			UnitAmountMinor:        termRow.SiteHourlyRateMinor,
			LineAmountMinor:        subtotalMinor,
			FundedAllowanceMinutes: fundedAllowance,
			FundedDeductionMinutes: fundedDeductionMinutes,
			CoreBillableMinutes:    billableMinutes,
			SessionCount:           len(calc.Sessions),
		})

		if termRow.FundingProfileID != nil && fundedDeductionMinor > 0 {
			lines = append(lines, PrefillLine{
				LineKind:               domain.LineKindFundedDeduction,
				Description:            "Funded hours deduction",
				SortOrder:              2,
				FundedAllowanceMinutes: fundedAllowance,
				FundedDeductionMinutes: fundedDeductionMinutes,
				CoreBillableMinutes:    billableMinutes,
				LineAmountMinor:        fundedDeductionMinor,
			})
		}

		if termRow.SiteHourlyRateMinor <= 0 {
			warnings = append(warnings, "site_rate_not_set")
		}
		if termRow.FundingProfileID == nil {
			warnings = append(warnings, "missing_funding_profile")
		}
		if fundedDeductionMinor > 0 && subtotalMinor > 0 {
			threshold := subtotalMinor / 4
			if fundedDeductionMinor > threshold {
				warnings = append(warnings, "significant_funding_deduction")
			}
		}

		result = ComputeInvoicePrefillResult{
			ChildID:                termRow.ChildID,
			ChildFirstName:         termRow.FirstName,
			ChildMiddleName:        termRow.MiddleName,
			ChildLastName:          termRow.LastName,
			BillingMonth:           billingMonthRaw,
			FundingProfileID:       termRow.FundingProfileID,
			FundedAllowanceMinutes: fundedAllowance,
			Lines:                  lines,
			SubtotalMinor:          subtotalMinor,
			FundedDeductionMinor:   fundedDeductionMinor,
			TotalDueMinor:          totalDueMinor,
			Warnings:               warnings,
		}

		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return ComputeInvoicePrefillResult{}, txErr
		}
		return ComputeInvoicePrefillResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func prefillWarnings(t domain.AdvancePayTermRow) []string {
	var w []string
	if t.FirstName == "" {
		w = append(w, "missing_child_name")
	}
	if t.DateOfBirth.IsZero() {
		w = append(w, "missing_date_of_birth")
	}
	if t.StartDate.IsZero() {
		w = append(w, "missing_start_date")
	}
	if !t.HasParentCarerContact {
		w = append(w, "missing_guardian_link")
	}
	return w
}
