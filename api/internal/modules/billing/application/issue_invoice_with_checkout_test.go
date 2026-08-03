package application

import (
	"context"
	"testing"

	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	paymentsdomain "nursery-management-system/api/internal/modules/payments/domain"
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

func TestIssueInvoiceWithCheckout_Orchestrator(t *testing.T) {
	t.Run("creates orchestrator with dependencies", func(t *testing.T) {
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
		)

		if orch == nil {
			t.Fatal("expected non-nil orchestrator")
		}
	})
}
