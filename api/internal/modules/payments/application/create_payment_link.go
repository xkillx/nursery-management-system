package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type CreatePaymentLink struct {
	repo             domain.ManagerPaymentRepository
	provider         domain.PaymentLinkProvider
	paymentLinkRepo  domain.PaymentLinkRepository
	stripeConfigured bool
	newUUID          func() uuid.UUID
	logger           *slog.Logger
	recorder         *metrics.Recorder
}

func NewCreatePaymentLink(
	repo domain.ManagerPaymentRepository,
	provider domain.PaymentLinkProvider,
	paymentLinkRepo domain.PaymentLinkRepository,
	stripeConfigured bool,
) *CreatePaymentLink {
	return &CreatePaymentLink{
		repo:             repo,
		provider:         provider,
		paymentLinkRepo:  paymentLinkRepo,
		stripeConfigured: stripeConfigured,
		newUUID:          uid.NewUUID,
	}
}

func (uc *CreatePaymentLink) WithObservability(logger *slog.Logger, recorder *metrics.Recorder) *CreatePaymentLink {
	return &CreatePaymentLink{
		repo:             uc.repo,
		provider:         uc.provider,
		paymentLinkRepo:  uc.paymentLinkRepo,
		stripeConfigured: uc.stripeConfigured,
		newUUID:          uc.newUUID,
		logger:           logger,
		recorder:         recorder,
	}
}

func (uc *CreatePaymentLink) logDebug(msg string, args ...any) {
	if uc.logger != nil {
		uc.logger.Debug(msg, args...)
	}
}

type CreatePaymentLinkResult struct {
	PaymentLinkID string
	URL           string
	Existing      bool
}

func (uc *CreatePaymentLink) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string) (CreatePaymentLinkResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return CreatePaymentLinkResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	if !uc.stripeConfigured {
		return CreatePaymentLinkResult{}, domainerrors.New("payment_provider_unconfigured", "Payment provider is not configured.")
	}

	tenantID := actor.TenantID.String()
	branchID := actor.BranchID.String()

	existing, found, err := uc.paymentLinkRepo.GetActivePaymentLinkForInvoice(ctx, tenantID, branchID, invoiceID.String())
	if err != nil {
		return CreatePaymentLinkResult{}, domainerrors.Internal(err)
	}
	if found {
		return CreatePaymentLinkResult{
			PaymentLinkID: existing.ID,
			URL:           existing.StripePaymentLinkURL,
			Existing:      true,
		}, nil
	}

	invoice, found, err := uc.repo.GetManagerInvoicePaymentStatus(ctx, tenantID, branchID, invoiceID.String())
	if err != nil {
		return CreatePaymentLinkResult{}, domainerrors.Internal(err)
	}
	if !found {
		return CreatePaymentLinkResult{}, domainerrors.NotFound("invoice", "Invoice not found.")
	}

	if !uc.isPayable(invoice) {
		return CreatePaymentLinkResult{}, domainerrors.Conflict("invoice_not_payable", "Invoice is not payable.")
	}

	productDesc := ""
	if invoice.InvoiceNumber != "" {
		productDesc = "Invoice " + invoice.InvoiceNumber
	}

	result, providerErr := uc.provider.CreatePaymentLink(ctx, domain.PaymentLinkCreateParams{
		AmountMinor:   invoice.TotalDueMinor,
		Currency:      "gbp",
		ProductName:   "Nursery invoice payment",
		Description:   productDesc,
		TenantID:      tenantID,
		BranchID:      branchID,
		InvoiceID:     invoiceID.String(),
		InvoiceNumber: invoice.InvoiceNumber,
	})
	if providerErr != nil {
		uc.logDebug("payment_link_provider",
			"operation", "create_payment_link",
			"invoice_id", invoiceID.String(),
			"error", providerErr,
		)
		return CreatePaymentLinkResult{}, domainerrors.New("payment_provider_error", "Payment provider failed to create payment link.")
	}

	linkID := uc.newUUID().String()
	if dbErr := uc.paymentLinkRepo.CreatePaymentLink(ctx, domain.PaymentLinkRecord{
		ID:                    linkID,
		TenantID:              tenantID,
		BranchID:              branchID,
		InvoiceID:             invoiceID.String(),
		StripePaymentLinkID:   result.ID,
		StripePaymentLinkURL:  result.URL,
		AmountMinor:           invoice.TotalDueMinor,
		CurrencyCode:          domain.CurrencyGBP,
		CreatedByUserID:       actor.UserID.String(),
		CreatedByMembershipID: actor.MembershipID.String(),
		Status:                domain.PaymentLinkStatusActive,
	}); dbErr != nil {
		uc.logDebug("payment_link_repo",
			"operation", "create_payment_link",
			"invoice_id", invoiceID.String(),
			"error", dbErr,
		)
		return CreatePaymentLinkResult{}, domainerrors.Internal(dbErr)
	}

	return CreatePaymentLinkResult{
		PaymentLinkID: linkID,
		URL:           result.URL,
		Existing:      false,
	}, nil
}

func (uc *CreatePaymentLink) isPayable(invoice domain.ManagerInvoicePaymentStatus) bool {
	if invoice.InvoiceKind != "monthly" {
		return false
	}
	if !payableStatuses[invoice.Status] {
		return false
	}
	if invoice.CurrencyCode != "GBP" {
		return false
	}
	if invoice.TotalDueMinor <= 0 {
		return false
	}
	if invoice.AmountPaidMinor != 0 {
		return false
	}
	return true
}
