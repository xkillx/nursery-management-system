package application

import (
	"context"

	"nursery-management-system/api/internal/modules/payments/domain"
)

type TestPayLookupRepository interface {
	GetAttemptForTest(ctx context.Context, attemptID string) (*domain.AttemptTestInfo, bool, error)
}

type TestPayLookupResult struct {
	ID                      string
	TenantID                string
	BranchID                string
	InvoiceID               string
	AmountMinor             int
	CurrencyCode            string
	StripeCheckoutSessionID string
	Status                  string
}

type TestPayLookup struct {
	repo TestPayLookupRepository
}

func NewTestPayLookup(repo TestPayLookupRepository) *TestPayLookup {
	return &TestPayLookup{repo: repo}
}

func (uc *TestPayLookup) Execute(ctx context.Context, attemptID string) (*TestPayLookupResult, error) {
	info, found, err := uc.repo.GetAttemptForTest(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &TestPayLookupResult{
		ID:                      info.ID,
		TenantID:                info.TenantID,
		BranchID:                info.BranchID,
		InvoiceID:               info.InvoiceID,
		AmountMinor:             info.AmountMinor,
		CurrencyCode:            info.CurrencyCode,
		StripeCheckoutSessionID: info.StripeCheckoutSessionID,
		Status:                  info.Status,
	}, nil
}
