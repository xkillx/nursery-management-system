package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	"nursery-management-system/api/internal/platform/events"
	"nursery-management-system/api/internal/platform/storage"
)

type MarkOverdueInvoices struct {
	repo          domain.BillingRepository
	dispatcher    *events.EventDispatcher
	now           func() time.Time
	london        *time.Location
	checkoutUC    *paymentsapp.CreateCheckoutSession
	pdfGen        *InvoicePDFGenerator
	storage       storage.Service
	parentContact ParentContactLookup
	siteProfile   SiteProfileLookup
}

func NewMarkOverdueInvoices(
	repo domain.BillingRepository,
	dispatcher *events.EventDispatcher,
	now func() time.Time,
	checkoutUC *paymentsapp.CreateCheckoutSession,
	pdfGen *InvoicePDFGenerator,
	storage storage.Service,
	parentContact ParentContactLookup,
	siteProfile SiteProfileLookup,
) *MarkOverdueInvoices {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(fmt.Sprintf("failed to load Europe/London timezone: %v", err))
	}
	return &MarkOverdueInvoices{
		repo:          repo,
		dispatcher:    dispatcher,
		now:           now,
		london:        london,
		checkoutUC:    checkoutUC,
		pdfGen:        pdfGen,
		storage:       storage,
		parentContact: parentContact,
		siteProfile:   siteProfile,
	}
}

func (uc *MarkOverdueInvoices) Execute(ctx context.Context) (domain.OverdueTransitionResult, error) {
	nowUTC := uc.now().UTC()

	currentLondonDate := nowUTC.In(uc.london)
	londonMidnight := time.Date(
		currentLondonDate.Year(),
		currentLondonDate.Month(),
		currentLondonDate.Day(),
		0, 0, 0, 0,
		uc.london,
	)
	cutoffUTC := londonMidnight.UTC()

	var result domain.OverdueTransitionResult
	result.CurrentLondonDate = londonMidnight
	result.CutoffUTC = cutoffUTC

	txErr := uc.dispatcher.DispatchInTx(ctx, func(tx pgx.Tx, emitter events.Emitter) error {
		acquired, lockErr := uc.repo.TryAcquireOverdueTransitionJobLock(ctx, tx)
		if lockErr != nil {
			return fmt.Errorf("acquire overdue job lock: %w", lockErr)
		}
		if !acquired {
			result.LockAcquired = false
			return nil
		}
		result.LockAcquired = true

		transitioned, markErr := uc.repo.MarkIssuedInvoicesOverdue(ctx, tx, cutoffUTC)
		if markErr != nil {
			return fmt.Errorf("mark invoices overdue: %w", markErr)
		}
		result.Transitioned = transitioned

		if len(transitioned) > 0 {
			// Per-invoice pre-work: create fresh checkout and PDF for each overdue invoice
			for i := range transitioned {
				inv := &transitioned[i]
				uc.enrichOverdueInvoice(ctx, inv)
			}

			emitter.Emit(domain.InvoiceMarkedOverdue{
				Transitioned: transitioned,
				Occurred:     nowUTC,
			})
		}
		return nil
	})

	if txErr != nil {
		return domain.OverdueTransitionResult{}, txErr
	}

	return result, nil
}

func (uc *MarkOverdueInvoices) enrichOverdueInvoice(ctx context.Context, inv *domain.OverdueTransitionedInvoice) {
	if uc.checkoutUC == nil {
		return
	}

	// Create fresh Stripe checkout session (idempotent — expired sessions are replaced)
	checkoutResult, checkoutErr := uc.checkoutUC.Execute(ctx,
		inv.TenantID.String(),
		inv.BranchID.String(),
		"", // membershipID — system-initiated
		"", // userID — system-initiated
		inv.ID.String(),
		"overdue-transition",
	)
	if checkoutErr != nil {
		slog.WarnContext(ctx, "overdue_checkout_creation_failed",
			"invoice_id", inv.ID,
			"error", checkoutErr,
		)
		return
	}
	inv.CheckoutURL = checkoutResult.CheckoutURL

	// Generate PDF and upload to S3
	if uc.pdfGen != nil && uc.storage != nil {
		siteProfile, spErr := uc.siteProfile.GetForInvoice(ctx, inv.TenantID, inv.BranchID)
		if spErr != nil || siteProfile == nil {
			slog.WarnContext(ctx, "overdue_site_profile_lookup_failed",
				"invoice_id", inv.ID,
				"error", spErr,
			)
			return
		}

		parent, parentErr := uc.parentContact.GetForInvoice(ctx, inv.TenantID, inv.BranchID, uuid.Nil)
		if parentErr != nil {
			slog.WarnContext(ctx, "overdue_parent_contact_lookup_failed",
				"invoice_id", inv.ID,
				"error", parentErr,
			)
		}

		parentName := ""
		if parent != nil {
			parentName = parent.FullName
		}

		pdfData := InvoicePDFData{
			Invoice:     domain.Invoice{ID: inv.ID},
			SiteProfile: *siteProfile,
			ParentName:  parentName,
			CheckoutURL: checkoutResult.CheckoutURL,
		}

		pdfBytes, pdfErr := uc.pdfGen.Generate(ctx, pdfData)
		if pdfErr != nil {
			slog.WarnContext(ctx, "overdue_pdf_generation_failed",
				"invoice_id", inv.ID,
				"error", pdfErr,
			)
			return
		}

		s3Key := fmt.Sprintf("invoices/%s/overdue.pdf", inv.ID.String())
		if uploadErr := uc.storage.Upload(ctx, s3Key, pdfBytes, "application/pdf"); uploadErr != nil {
			slog.WarnContext(ctx, "overdue_s3_upload_failed",
				"invoice_id", inv.ID,
				"error", uploadErr,
			)
			return
		}
		inv.AttachmentS3Key = s3Key
	}
}
