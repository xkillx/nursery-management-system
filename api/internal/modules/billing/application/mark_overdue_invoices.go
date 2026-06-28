package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
)

type txManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type MarkOverdueInvoices struct {
	repo   domain.BillingRepository
	txMgr  txManager
	now    func() time.Time
	london *time.Location
}

func NewMarkOverdueInvoices(repo domain.BillingRepository, txMgr txManager, now func() time.Time) *MarkOverdueInvoices {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(fmt.Sprintf("failed to load Europe/London timezone: %v", err))
	}
	return &MarkOverdueInvoices{repo: repo, txMgr: txMgr, now: now, london: london}
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

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
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
		return nil
	})

	if txErr != nil {
		return domain.OverdueTransitionResult{}, txErr
	}

	return result, nil
}
