package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/events"
)

type SendDueSoonReminders struct {
	repo       domain.BillingRepository
	dispatcher *events.EventDispatcher
	now        func() time.Time
}

func NewSendDueSoonReminders(repo domain.BillingRepository, dispatcher *events.EventDispatcher, now func() time.Time) *SendDueSoonReminders {
	return &SendDueSoonReminders{repo: repo, dispatcher: dispatcher, now: now}
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
			emitter.Emit(domain.InvoiceDueSoon{
				InvoiceID: inv.ID,
				TenantID:  inv.TenantID,
				BranchID:  inv.BranchID,
				DueDate:   inv.DueDate,
				Occurred:  nowUTC,
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
			emitter.Emit(domain.InvoiceDueReminder{
				InvoiceID: inv.ID,
				TenantID:  inv.TenantID,
				BranchID:  inv.BranchID,
				DueDate:   inv.DueDate,
				Occurred:  nowUTC,
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
