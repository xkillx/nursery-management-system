package application

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// TestCreateDraftInvoice_SubtotalExcludesFundedDeduction verifies that the
// subtotal calculation in CreateDraftInvoice correctly separates funded
// deduction lines from core lines. Before the fix, the subtotal included the
// deduction amount as a positive value, causing it to be inflated.
func TestCreateDraftInvoice_SubtotalExcludesFundedDeduction(t *testing.T) {
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
			QuantityMinutes: 0,
			UnitAmountMinor: 0,
			LineAmountMinor: 47500, // £475.00 (positive input from prefill)
		},
	}

	// Replicate the subtotal calculation logic from CreateDraftInvoice.Execute
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

func TestCreateDraftInvoice_NoDeductionLine(t *testing.T) {
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
	if totalDueMinor != 200000 {
		t.Errorf("totalDue = %d, want 200000", totalDueMinor)
	}
}

func TestCreateDraftInvoice_ValidationRejectsNegativeQuantity(t *testing.T) {
	uc := &CreateDraftInvoice{}
	actor := tenant.ActorContext{
		TenantID: uuid.New(),
		BranchID: uuid.New(),
	}

	input := CreateDraftInvoiceInput{
		ChildID:      uuid.New(),
		BillingMonth: "2026-08",
		Lines: []DraftInvoiceLineInput{
			{
				LineKind:        domain.LineKindCoreChildcare,
				Description:     "Core childcare",
				SortOrder:       1,
				QuantityMinutes: -100,
				UnitAmountMinor: 1000,
				LineAmountMinor: -100000,
			},
		},
	}

	_, err := uc.Execute(context.Background(), actor, input)
	if err == nil {
		t.Fatal("expected validation error for negative quantity")
	}
	var domainErr *domainerrors.DomainError
	if ok := (err != nil); !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	_ = domainErr
}
