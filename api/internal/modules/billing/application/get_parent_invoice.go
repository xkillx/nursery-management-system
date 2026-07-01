package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetParentInvoice struct {
	repo     domain.BillingRepository
	spLookup SiteProfileLookup
}

func NewGetParentInvoice(repo domain.BillingRepository, spLookup SiteProfileLookup) *GetParentInvoice {
	return &GetParentInvoice{repo: repo, spLookup: spLookup}
}

type GetParentInvoiceResult struct {
	domain.ParentInvoiceDetail
	SiteProfile *domain.ParentInvoiceSiteProfile
}

func (uc *GetParentInvoice) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string) (GetParentInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return GetParentInvoiceResult{}, domainerrors.Validation("Invalid invoice_id format.", "invoice_id")
	}

	row, found, err := uc.repo.GetInvoiceForParent(ctx, actor.TenantID, actor.BranchID, actor.MembershipID, invoiceID)
	if err != nil {
		return GetParentInvoiceResult{}, domainerrors.Internal(err)
	}
	if !found {
		return GetParentInvoiceResult{}, domainerrors.NotFound("invoice", "Invoice not found.")
	}

	lines, err := uc.repo.ListInvoiceLinesForParent(ctx, actor.TenantID, actor.BranchID, actor.MembershipID, invoiceID)
	if err != nil {
		return GetParentInvoiceResult{}, domainerrors.Internal(err)
	}
	if lines == nil {
		lines = []domain.ParentInvoiceLineRow{}
	}

	calc, err := parseCalculationDetails(row.CalculationDetails)
	if err != nil {
		return GetParentInvoiceResult{}, domainerrors.Internal(err)
	}

	sp, lookupErr := uc.spLookup.GetForInvoice(ctx, actor.TenantID, actor.BranchID)
	if lookupErr != nil {
		return GetParentInvoiceResult{}, domainerrors.Internal(lookupErr)
	}

	var spDTO *domain.ParentInvoiceSiteProfile
	if sp != nil {
		spDTO = &domain.ParentInvoiceSiteProfile{
			NurseryName:     sp.NurseryName,
			Phone:           sp.Phone,
			Email:           sp.Email,
			Website:         sp.Website,
			AddressStreet:   sp.AddressStreet,
			AddressCity:     sp.AddressCity,
			AddressPostcode: sp.AddressPostcode,
		}
	}

	return GetParentInvoiceResult{
		ParentInvoiceDetail: domain.ParentInvoiceDetail{
			Invoice:     row,
			Lines:       lines,
			Calculation: calc,
		},
		SiteProfile: spDTO,
	}, nil
}
