package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	"nursery-management-system/api/internal/platform/events"
)

type MarkOverdueInvoices struct {
	repo       domain.BillingRepository
	dispatcher *events.EventDispatcher
	now        func() time.Time
	london     *time.Location
	checkoutUC *paymentsapp.CreateCheckoutSession
}

func NewMarkOverdueInvoices(
	repo domain.BillingRepository,
	dispatcher *events.EventDispatcher,
	now func() time.Time,
	checkoutUC *paymentsapp.CreateCheckoutSession,
) *MarkOverdueInvoices {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(fmt.Sprintf("failed to load Europe/London timezone: %v", err))
	}
	return &MarkOverdueInvoices{
		repo:       repo,
		dispatcher: dispatcher,
		now:        now,
		london:     london,
		checkoutUC: checkoutUC,
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
			// Per-invoice pre-work: create a fresh checkout for each overdue invoice
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
}
