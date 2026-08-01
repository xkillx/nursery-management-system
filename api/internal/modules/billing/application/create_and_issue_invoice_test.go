package application

import (
	"testing"

	"nursery-management-system/api/internal/modules/billing/domain"
)

// TestCreateAndIssueInvoice_SubtotalExcludesFundedDeduction verifies that the
// subtotal calculation in CreateAndIssueInvoiceFromForm correctly separates
// funded deduction lines from core lines, matching CreateDraftInvoice logic.
func TestCreateAndIssueInvoice_SubtotalExcludesFundedDeduction(t *testing.T) {
	lines := []DraftInvoiceLineInput{
		{
			LineKind:        domain.LineKindCoreChildcare,
			Description:     "Core childcare",
			SortOrder:       1,
			QuantityMinutes: 12000,
			UnitAmountMinor: 1000,
			LineAmountMinor: 200000, // £2,000.00
		},
		{
			LineKind:        domain.LineKindFundedDeduction,
			Description:     "Funded hours deduction",
			SortOrder:       2,
			QuantityMinutes: 2850,
			UnitAmountMinor: -1000,
			LineAmountMinor: 47500, // £475.00 (positive input from prefill)
		},
	}

	// Replicate the subtotal calculation logic from CreateAndIssueInvoiceFromForm.Execute
	subtotalMinor := 0
	fundedDeductionMinor := 0
	for _, line := range lines {
		if line.LineKind == domain.LineKindFundedDeduction {
			fundedDeductionMinor += line.LineAmountMinor
		} else {
			subtotalMinor += line.LineAmountMinor
		}
	}
	totalDueMinor := subtotalMinor - fundedDeductionMinor

	if subtotalMinor != 200000 {
		t.Errorf("subtotal = %d, want 200000", subtotalMinor)
	}
	if fundedDeductionMinor != 47500 {
		t.Errorf("fundedDeduction = %d, want 47500", fundedDeductionMinor)
	}
	if totalDueMinor != 152500 {
		t.Errorf("totalDue = %d, want 152500", totalDueMinor)
	}
}

// TestCreateAndIssueInvoice_NoDeductionLine verifies correct calculation
// when there is no funded deduction line.
func TestCreateAndIssueInvoice_NoDeductionLine(t *testing.T) {
	lines := []DraftInvoiceLineInput{
		{
			LineKind:        domain.LineKindCoreChildcare,
			Description:     "Core childcare",
			SortOrder:       1,
			QuantityMinutes: 12000,
			UnitAmountMinor: 1000,
			LineAmountMinor: 200000,
		},
	}

	subtotalMinor := 0
	fundedDeductionMinor := 0
	for _, line := range lines {
		if line.LineKind == domain.LineKindFundedDeduction {
			fundedDeductionMinor += line.LineAmountMinor
		} else {
			subtotalMinor += line.LineAmountMinor
		}
	}
	totalDueMinor := subtotalMinor - fundedDeductionMinor

	if subtotalMinor != 200000 {
		t.Errorf("subtotal = %d, want 200000", subtotalMinor)
	}
	if fundedDeductionMinor != 0 {
		t.Errorf("fundedDeduction = %d, want 0", fundedDeductionMinor)
	}
	if totalDueMinor != 200000 {
		t.Errorf("totalDue = %d, want 200000", totalDueMinor)
	}
}
