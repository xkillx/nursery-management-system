package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	"nursery-management-system/api/internal/platform/storage"
	"nursery-management-system/api/internal/platform/tenant"
)

type IssueInvoiceWithCheckout struct {
	issueUC       *IssueInvoice
	checkoutUC    *paymentsapp.CreateCheckoutSession
	pdfGen        *InvoicePDFGenerator
	storage       storage.Service
	parentContact ParentContactLookup
	siteProfile   SiteProfileLookup
}

func NewIssueInvoiceWithCheckout(
	issueUC *IssueInvoice,
	checkoutUC *paymentsapp.CreateCheckoutSession,
	pdfGen *InvoicePDFGenerator,
	storage storage.Service,
	parentContact ParentContactLookup,
	siteProfile SiteProfileLookup,
) *IssueInvoiceWithCheckout {
	return &IssueInvoiceWithCheckout{
		issueUC:       issueUC,
		checkoutUC:    checkoutUC,
		pdfGen:        pdfGen,
		storage:       storage,
		parentContact: parentContact,
		siteProfile:   siteProfile,
	}
}

func (uc *IssueInvoiceWithCheckout) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string, confirm bool) (domain.IssueInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return domain.IssueInvoiceResult{}, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Pre-work: create Stripe checkout session (idempotent)
	var checkoutURL string
	checkoutResult, checkoutErr := uc.checkoutUC.Execute(ctx,
		actor.TenantID.String(),
		actor.BranchID.String(),
		actor.MembershipID.String(),
		actor.UserID.String(),
		invoiceIDRaw,
		actor.RequestID,
	)
	if checkoutErr != nil {
		slog.WarnContext(ctx, "checkout_creation_failed_in_orchestrator",
			"invoice_id", invoiceID,
			"error", checkoutErr,
		)
		// Best-effort: invoice will still be issued, scheduler catches it
	} else {
		checkoutURL = checkoutResult.CheckoutURL
	}

	// Pre-work: generate PDF and upload to S3
	var attachmentS3Key string
	if uc.pdfGen != nil && uc.storage != nil {
		siteProfile, spErr := uc.siteProfile.GetForInvoice(ctx, actor.TenantID, actor.BranchID)
		if spErr != nil || siteProfile == nil {
			slog.WarnContext(ctx, "site_profile_lookup_failed",
				"invoice_id", invoiceID,
				"error", spErr,
			)
		} else {
			parent, parentErr := uc.parentContact.GetForInvoice(ctx, actor.TenantID, actor.BranchID, uuid.Nil)
			if parentErr != nil {
				slog.WarnContext(ctx, "parent_contact_lookup_failed",
					"invoice_id", invoiceID,
					"error", parentErr,
				)
			}

			parentName := ""
			childName := ""
			if parent != nil {
				parentName = parent.FullName
			}

			pdfData := InvoicePDFData{
				Invoice: domain.Invoice{
					ID: invoiceID,
				},
				SiteProfile: *siteProfile,
				ParentName:  parentName,
				ChildName:   childName,
				CheckoutURL: checkoutURL,
			}

			pdfBytes, pdfErr := uc.pdfGen.Generate(ctx, pdfData)
			if pdfErr != nil {
				slog.WarnContext(ctx, "pdf_generation_failed",
					"invoice_id", invoiceID,
					"error", pdfErr,
				)
			} else {
				s3Key := fmt.Sprintf("invoices/%s/invoice.pdf", invoiceID.String())
				if uploadErr := uc.storage.Upload(ctx, s3Key, pdfBytes, "application/pdf"); uploadErr != nil {
					slog.WarnContext(ctx, "s3_upload_failed",
						"invoice_id", invoiceID,
						"error", uploadErr,
					)
				} else {
					attachmentS3Key = s3Key
				}
			}
		}
	}

	// Call IssueInvoice.ExecuteWithContext to pass pre-computed data through to the event
	return uc.issueUC.ExecuteWithContext(ctx, actor, invoiceIDRaw, confirm, IssueContext{
		CheckoutURL:     checkoutURL,
		AttachmentS3Key: attachmentS3Key,
	})
}
