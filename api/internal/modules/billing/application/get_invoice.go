package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetInvoice struct {
	repo     domain.BillingRepository
	spLookup SiteProfileLookup
	pcLookup ParentContactLookup
}

func NewGetInvoice(repo domain.BillingRepository, spLookup SiteProfileLookup, pcLookup ParentContactLookup) *GetInvoice {
	return &GetInvoice{repo: repo, spLookup: spLookup, pcLookup: pcLookup}
}

type GetInvoiceResult struct {
	domain.InvoiceReviewDetail
	SiteProfile   *domain.InvoiceSiteProfile
	ParentContact *domain.ParentContact
}

func (uc *GetInvoice) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string) (GetInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return GetInvoiceResult{}, domainerrors.Validation("Invalid invoice_id format.", "invoice_id")
	}

	row, found, err := uc.repo.GetInvoiceForManagerReview(ctx, actor.TenantID, actor.BranchID, invoiceID)
	if err != nil {
		return GetInvoiceResult{}, domainerrors.Internal(err)
	}
	if !found {
		return GetInvoiceResult{}, domainerrors.NotFound("invoice", "Invoice not found.")
	}

	lines, err := uc.repo.ListInvoiceLinesForManagerReview(ctx, actor.TenantID, actor.BranchID, invoiceID)
	if err != nil {
		return GetInvoiceResult{}, domainerrors.Internal(err)
	}
	if lines == nil {
		lines = []domain.InvoiceReviewLineRow{}
	}

	calc, err := parseCalculationDetails(row.CalculationDetails)
	if err != nil {
		return GetInvoiceResult{}, domainerrors.Internal(fmt.Errorf("malformed calculation_details: %w", err))
	}

	exceptions, exceptionCount := parseRunExceptions(row.GeneratedRunDetails)

	sp, lookupErr := uc.spLookup.GetForInvoice(ctx, actor.TenantID, actor.BranchID)
	if lookupErr != nil {
		return GetInvoiceResult{}, domainerrors.Internal(lookupErr)
	}

	var spDTO *domain.InvoiceSiteProfile
	if sp != nil {
		spDTO = &domain.InvoiceSiteProfile{
			NurseryName:     sp.NurseryName,
			Phone:           sp.Phone,
			Email:           sp.Email,
			Website:         sp.Website,
			AddressStreet:   sp.AddressStreet,
			AddressCity:     sp.AddressCity,
			AddressPostcode: sp.AddressPostcode,
		}
	}

	pc, pcErr := uc.pcLookup.GetForInvoice(ctx, actor.TenantID, actor.BranchID, row.ChildID)
	if pcErr != nil {
		return GetInvoiceResult{}, domainerrors.Internal(pcErr)
	}

	return GetInvoiceResult{
		InvoiceReviewDetail: domain.InvoiceReviewDetail{
			Invoice:                    row,
			Lines:                      lines,
			Calculation:                calc,
			GeneratedRunExceptions:     exceptions,
			GeneratedRunExceptionCount: exceptionCount,
		},
		SiteProfile:   spDTO,
		ParentContact: pc,
	}, nil
}

func parseCalculationDetails(raw json.RawMessage) (domain.InvoiceReviewCalculation, error) {
	var calc domain.InvoiceReviewCalculation
	if len(raw) == 0 {
		return calc, nil
	}

	var details domain.InvoiceCalculationDetails
	if err := json.Unmarshal(raw, &details); err != nil {
		return calc, fmt.Errorf("unmarshal calculation_details: %w", err)
	}

	calc = domain.InvoiceReviewCalculation{
		CoreHourlyRate:         details.CoreHourlyRate,
		BookedCoreMinutes:      details.BookedCoreMinutes,
		BookedSessionCount:     len(details.BookedSessions),
		FundedAllowanceMinutes: details.FundedAllowanceMinutes,
		FundedDeductionMinutes: details.FundedDeductionMinutes,
		CoreBillableMinutes:    details.CoreBillableMinutes,
		CoreSubtotal:           details.CoreSubtotal,
		ExtrasTotal:            details.ExtrasTotal,
		TermID:                 details.TermID,
		BookingPatternID:       details.BookingPatternID,
		BookedSessions:         details.BookedSessions,
		BookedPerEntry:         details.BookedPerEntry,
	}

	return calc, nil
}

// runDetailsShape matches the relevant part of invoice_runs.details JSON.
type runDetailsShape struct {
	BlockedChildren []struct {
		ChildID         string   `json:"child_id"`
		ChildFirstName  string   `json:"child_first_name"`
		ChildMiddleName *string  `json:"child_middle_name"`
		ChildLastName   *string  `json:"child_last_name"`
		BlockerCodes    []string `json:"blocker_codes"`
	} `json:"blocked_children"`
}

func parseRunExceptions(raw json.RawMessage) ([]domain.InvoiceRunExceptionReference, int) {
	if len(raw) == 0 {
		return []domain.InvoiceRunExceptionReference{}, 0
	}

	var details runDetailsShape
	if err := json.Unmarshal(raw, &details); err != nil {
		return []domain.InvoiceRunExceptionReference{}, 0
	}

	exceptions := make([]domain.InvoiceRunExceptionReference, 0, len(details.BlockedChildren))
	for _, bc := range details.BlockedChildren {
		codes := bc.BlockerCodes
		if codes == nil {
			codes = []string{}
		}
		exceptions = append(exceptions, domain.InvoiceRunExceptionReference{
			ChildID:         bc.ChildID,
			ChildFirstName:  bc.ChildFirstName,
			ChildMiddleName: bc.ChildMiddleName,
			ChildLastName:   bc.ChildLastName,
			BlockerCodes:    codes,
		})
	}

	return exceptions, len(exceptions)
}
