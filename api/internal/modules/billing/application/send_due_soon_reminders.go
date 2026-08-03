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
)

type SendDueSoonReminders struct {
	repo       domain.BillingRepository
	dispatcher *events.EventDispatcher
	now        func() time.Time
	checkoutUC *paymentsapp.CreateCheckoutSession
}

func NewSendDueSoonReminders(
	repo domain.BillingRepository,
	dispatcher *events.EventDispatcher,
	now func() time.Time,
	checkoutUC *paymentsapp.CreateCheckoutSession,
) *SendDueSoonReminders {
	return &SendDueSoonReminders{
		repo:       repo,
		dispatcher: dispatcher,
		now:        now,
		checkoutUC: checkoutUC,
	}
}

func (uc *SendDueSoonReminders) Execute(ctx context.Context) (domain.ReminderJobResult, error) {
	nowUTC := uc.now().UTC()

	var result domain.ReminderJobResult

	txErr := uc.dispatcher.DispatchInTx(ctx, func(tx pgx.Tx, emitter events.Emitter) error {
		acquired, lockErr := uc.repo.TryAcquireReminderJobLock(ctx, tx)
		if lockErr != nil {
			return fmt.Errorf("acquire reminder job lock: %w", lockErr)
		}
		if !acquired {
			result.LockAcquired = false
			return nil
		}
		result.LockAcquired = true

		dueSoon, err := uc.repo.ListInvoicesDueSoon(ctx, tx)
		if err != nil {
			return fmt.Errorf("list invoices due soon: %w", err)
		}
		result.DueSoon = dueSoon

		for _, inv := range dueSoon {
			checkoutURL := uc.enrichReminder(ctx, inv.ID, inv.TenantID, inv.BranchID)
			emitter.Emit(domain.InvoiceDueReminder{
				InvoiceID:   inv.ID,
				TenantID:    inv.TenantID,
				BranchID:    inv.BranchID,
				DueDate:     inv.DueDate,
				DaysBefore:  7,
				CheckoutURL: checkoutURL,
				Occurred:    nowUTC,
			})
			if logErr := uc.repo.InsertInvoiceReminderLog(ctx, tx, inv.TenantID, inv.BranchID, inv.ID, "due_soon"); logErr != nil {
				return fmt.Errorf("insert reminder log for %s: %w", inv.ID, logErr)
			}
		}

		dueToday, err := uc.repo.ListInvoicesDueToday(ctx, tx)
		if err != nil {
			return fmt.Errorf("list invoices due today: %w", err)
		}
		result.DueToday = dueToday

		for _, inv := range dueToday {
			checkoutURL := uc.enrichReminder(ctx, inv.ID, inv.TenantID, inv.BranchID)
			emitter.Emit(domain.InvoiceDueReminder{
				InvoiceID:   inv.ID,
				TenantID:    inv.TenantID,
				BranchID:    inv.BranchID,
				DueDate:     inv.DueDate,
				DaysBefore:  0,
				CheckoutURL: checkoutURL,
				Occurred:    nowUTC,
			})
			if logErr := uc.repo.InsertInvoiceReminderLog(ctx, tx, inv.TenantID, inv.BranchID, inv.ID, "due_today"); logErr != nil {
				return fmt.Errorf("insert reminder log for %s: %w", inv.ID, logErr)
			}
		}

		return nil
	})

	if txErr != nil {
		return domain.ReminderJobResult{}, txErr
	}

	return result, nil
}

func (uc *SendDueSoonReminders) enrichReminder(ctx context.Context, invoiceID, tenantID, branchID uuid.UUID) (checkoutURL string) {
	if uc.checkoutUC == nil {
		return ""
	}

	checkoutResult, checkoutErr := uc.checkoutUC.Execute(ctx,
		tenantID.String(),
		branchID.String(),
		"", // membershipID — system-initiated
		"", // userID — system-initiated
		invoiceID.String(),
		"reminder-dispatch",
	)
	if checkoutErr != nil {
		slog.WarnContext(ctx, "reminder_checkout_creation_failed",
			"invoice_id", invoiceID,
			"error", checkoutErr,
		)
		return ""
	}

	return checkoutResult.CheckoutURL
}
