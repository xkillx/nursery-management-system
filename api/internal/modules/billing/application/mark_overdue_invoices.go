package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/events"
)

type MarkOverdueInvoices struct {
	repo       domain.BillingRepository
	dispatcher *events.EventDispatcher
	now        func() time.Time
	london     *time.Location
}

func NewMarkOverdueInvoices(repo domain.BillingRepository, dispatcher *events.EventDispatcher, now func() time.Time) *MarkOverdueInvoices {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(fmt.Sprintf("failed to load Europe/London timezone: %v", err))
	}
	return &MarkOverdueInvoices{repo: repo, dispatcher: dispatcher, now: now, london: london}
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
