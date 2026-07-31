package application

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	paymentsdomain "nursery-management-system/api/internal/modules/payments/domain"
	siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"
	"nursery-management-system/api/internal/platform/storage"
)

type fakeCheckoutProvider struct {
	result paymentsdomain.CheckoutSessionResult
	err    error
}

func (f *fakeCheckoutProvider) CreateCheckoutSession(_ context.Context, _ paymentsdomain.CheckoutSessionCreateParams) (paymentsdomain.CheckoutSessionResult, error) {
	return f.result, f.err
}

type fakeTxMgr struct{}

func (f *fakeTxMgr) ExecTx(_ context.Context, fn func(tx paymentsdomain.Tx) error) error {
	return fn(nil)
}

type fakeParentContact struct {
	email string
}

func (f *fakeParentContact) GetForInvoice(_ context.Context, _, _, _ uuid.UUID) (*domain.ParentContact, error) {
	if f.email == "" {
		return nil, nil
	}
	return &domain.ParentContact{FullName: "Test Parent", Email: f.email}, nil
}

type fakeSiteProfileLookup struct{}

func (f *fakeSiteProfileLookup) GetForInvoice(_ context.Context, _, _ uuid.UUID) (*siteprofiledomain.SiteProfile, error) {
	return nil, nil
}

func TestIssueInvoiceWithCheckout_Orchestrator(t *testing.T) {
	t.Run("creates orchestrator with dependencies", func(t *testing.T) {
		fakeStorage := storage.NewFakeService()
		fakeCheckout := &fakeCheckoutProvider{
			result: paymentsdomain.CheckoutSessionResult{
				CheckoutSessionID: "cs_test",
				CheckoutURL:       "https://checkout.stripe.com/test",
			},
		}

		checkoutUC := paymentsapp.NewCreateCheckoutSession(
			nil,
			&fakeTxMgr{},
			fakeCheckout,
			"http://localhost:4200",
			true,
		)

		orch := NewIssueInvoiceWithCheckout(
			nil,
			checkoutUC,
			nil,
			fakeStorage,
			&fakeParentContact{email: "test@example.com"},
			&fakeSiteProfileLookup{},
		)

		if orch == nil {
			t.Fatal("expected non-nil orchestrator")
		}
	})
}
