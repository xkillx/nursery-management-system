package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type fakeRepo struct {
	candidate      domain.CheckoutInvoiceCandidate
	candidateFound bool
	paymentState   domain.InvoicePaymentState
	stateFound     bool
	createdAttempt domain.PaymentAttemptCreateParams
	markedCreated  domain.PaymentAttemptCheckoutCreatedParams
	markedFailed   domain.PaymentAttemptCheckoutCreationFailedParams
	markCreatedErr error
	markFailedErr  error
}

func (f *fakeRepo) GetParentInvoiceForCheckoutForUpdate(_ context.Context, _ domain.Tx, _, _, _, _ string) (domain.CheckoutInvoiceCandidate, bool, error) {
	return f.candidate, f.candidateFound, nil
}

func (f *fakeRepo) CreatePaymentAttempt(_ context.Context, _ domain.Tx, params domain.PaymentAttemptCreateParams) error {
	f.createdAttempt = params
	return nil
}

func (f *fakeRepo) GetInvoicePaymentState(_ context.Context, _, _, _ string) (domain.InvoicePaymentState, bool, error) {
	return f.paymentState, f.stateFound, nil
}

func (f *fakeRepo) MarkPaymentAttemptCheckoutCreated(_ context.Context, params domain.PaymentAttemptCheckoutCreatedParams) error {
	f.markedCreated = params
	return f.markCreatedErr
}

func (f *fakeRepo) MarkPaymentAttemptCheckoutCreationFailed(_ context.Context, params domain.PaymentAttemptCheckoutCreationFailedParams) error {
	f.markedFailed = params
	return f.markFailedErr
}

type fakeProvider struct {
	result domain.CheckoutSessionResult
	err    error
}

func (f *fakeProvider) CreateCheckoutSession(_ context.Context, _ domain.CheckoutSessionCreateParams) (domain.CheckoutSessionResult, error) {
	return f.result, f.err
}

type fakeTxManager struct{}

func (f *fakeTxManager) ExecTx(_ context.Context, fn func(tx domain.Tx) error) error {
	return fn(nil)
}

func payableCandidate() domain.CheckoutInvoiceCandidate {
	return domain.CheckoutInvoiceCandidate{
		ID:              uuid.New().String(),
		InvoiceKind:     "monthly",
		InvoiceNumber:   "INV-2026-001",
		Status:          "issued",
		CurrencyCode:    "GBP",
		TotalDueMinor:   5000,
		AmountPaidMinor: 0,
		ChildID:         uuid.New().String(),
	}
}

func payableState() domain.InvoicePaymentState {
	return domain.InvoicePaymentState{
		InvoiceKind:     "monthly",
		Status:          "issued",
		CurrencyCode:    "GBP",
		TotalDueMinor:   5000,
		AmountPaidMinor: 0,
	}
}

func newUC(repo *fakeRepo, provider *fakeProvider) *CreateCheckoutSession {
	return &CreateCheckoutSession{
		repo:             repo,
		txMgr:            &fakeTxManager{},
		provider:         provider,
		webBaseURL:       "http://localhost:4200",
		stripeConfigured: true,
		newUUID:          func() uuid.UUID { return uuid.MustParse("00000000-0000-0000-0000-000000000001") },
	}
}

func assertCode(t *testing.T, err error, want string) {
	t.Helper()
	d, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T: %v", err, err)
	}
	if d.Code != want {
		t.Fatalf("expected code %q, got %q", want, d.Code)
	}
}

func TestCreateCheckoutSession_MalformedInvoiceID(t *testing.T) {
	uc := newUC(&fakeRepo{}, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", "not-a-uuid", "req")
	assertCode(t, err, "validation_error")
}

func TestCreateCheckoutSession_StripeUnconfigured(t *testing.T) {
	uc := newUC(&fakeRepo{}, &fakeProvider{})
	uc.stripeConfigured = false
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "payment_provider_unconfigured")
}

func TestCreateCheckoutSession_ParentInvisibleInvoice(t *testing.T) {
	uc := newUC(&fakeRepo{}, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "invoice_not_found")
}

func TestCreateCheckoutSession_DraftInvoiceNotPayable(t *testing.T) {
	repo := &fakeRepo{candidateFound: true}
	repo.candidate = payableCandidate()
	repo.candidate.Status = "draft"
	uc := newUC(repo, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "invoice_not_payable")
}

func TestCreateCheckoutSession_PaidInvoiceNotPayable(t *testing.T) {
	repo := &fakeRepo{candidateFound: true}
	repo.candidate = payableCandidate()
	repo.candidate.Status = "paid"
	uc := newUC(repo, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "invoice_not_payable")
}

func TestCreateCheckoutSession_ZeroTotalNotPayable(t *testing.T) {
	repo := &fakeRepo{candidateFound: true}
	repo.candidate = payableCandidate()
	repo.candidate.TotalDueMinor = 0
	uc := newUC(repo, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "invoice_not_payable")
}

func TestCreateCheckoutSession_NonGBPNotPayable(t *testing.T) {
	repo := &fakeRepo{candidateFound: true}
	repo.candidate = payableCandidate()
	repo.candidate.CurrencyCode = "USD"
	uc := newUC(repo, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "invoice_not_payable")
}

func TestCreateCheckoutSession_NonMonthlyNotPayable(t *testing.T) {
	repo := &fakeRepo{candidateFound: true}
	repo.candidate = payableCandidate()
	repo.candidate.InvoiceKind = "adjustment"
	uc := newUC(repo, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "invoice_not_payable")
}

func TestCreateCheckoutSession_NonzeroAmountPaidNotPayable(t *testing.T) {
	repo := &fakeRepo{candidateFound: true}
	repo.candidate = payableCandidate()
	repo.candidate.AmountPaidMinor = 100
	uc := newUC(repo, &fakeProvider{})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", uuid.New().String(), "req")
	assertCode(t, err, "invoice_not_payable")
}

func TestCreateCheckoutSession_IssuedInvoiceSuccess(t *testing.T) {
	repo := &fakeRepo{
		candidateFound: true,
		candidate:      payableCandidate(),
		stateFound:     true,
		paymentState:   payableState(),
	}
	provider := &fakeProvider{
		result: domain.CheckoutSessionResult{
			CheckoutSessionID: "cs_test_123",
			CheckoutURL:       "https://checkout.stripe.com/test",
			PaymentIntentID:   "pi_test_456",
		},
	}
	uc := newUC(repo, provider)
	result, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CheckoutSessionID != "cs_test_123" {
		t.Fatalf("expected cs_test_123, got %s", result.CheckoutSessionID)
	}
	if result.CheckoutURL != "https://checkout.stripe.com/test" {
		t.Fatalf("expected checkout URL, got %s", result.CheckoutURL)
	}
	if result.PaymentAttemptID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("expected fixed attempt ID, got %s", result.PaymentAttemptID)
	}
	if repo.createdAttempt.Status != domain.AttemptStatusCheckoutCreationStarted {
		t.Fatalf("expected checkout_creation_started, got %s", repo.createdAttempt.Status)
	}
	if repo.markedCreated.StripeCheckoutSessionID != "cs_test_123" {
		t.Fatalf("expected session ID in mark created, got %s", repo.markedCreated.StripeCheckoutSessionID)
	}
}

func TestCreateCheckoutSession_PaymentFailedInvoiceCreatesAttempt(t *testing.T) {
	repo := &fakeRepo{
		candidateFound: true,
		stateFound:     true,
	}
	repo.candidate = payableCandidate()
	repo.candidate.Status = "payment_failed"
	repo.paymentState = payableState()
	repo.paymentState.Status = "payment_failed"

	uc := newUC(repo, &fakeProvider{result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"}})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCheckoutSession_OverdueInvoiceCreatesAttempt(t *testing.T) {
	repo := &fakeRepo{
		candidateFound: true,
		stateFound:     true,
	}
	repo.candidate = payableCandidate()
	repo.candidate.Status = "overdue"
	repo.paymentState = payableState()
	repo.paymentState.Status = "overdue"

	uc := newUC(repo, &fakeProvider{result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"}})
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCheckoutSession_ProviderError(t *testing.T) {
	repo := &fakeRepo{
		candidateFound: true,
		candidate:      payableCandidate(),
	}
	provider := &fakeProvider{err: fmt.Errorf("stripe API error")}
	uc := newUC(repo, provider)
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req")
	assertCode(t, err, "payment_provider_error")
	if repo.markedFailed.FailureReason != domain.FailureReasonStripeError {
		t.Fatalf("expected stripe_error, got %s", repo.markedFailed.FailureReason)
	}
}

func TestCreateCheckoutSession_InvoicePaidBetweenProviderAndResponse(t *testing.T) {
	repo := &fakeRepo{
		candidateFound: true,
		candidate:      payableCandidate(),
		stateFound:     true,
	}
	repo.paymentState = payableState()
	repo.paymentState.Status = "paid"

	provider := &fakeProvider{result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"}}
	uc := newUC(repo, provider)
	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req")
	assertCode(t, err, "invoice_not_payable")
	if repo.markedFailed.FailureReason != domain.FailureReasonInvoiceNoLongerPayable {
		t.Fatalf("expected invoice_no_longer_payable, got %s", repo.markedFailed.FailureReason)
	}
}

func TestCreateCheckoutSession_SuccessBuildsCorrectURLs(t *testing.T) {
	repo := &fakeRepo{
		candidateFound: true,
		candidate:      payableCandidate(),
		stateFound:     true,
		paymentState:   payableState(),
	}
	provider := &fakeProvider{result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"}}
	uc := newUC(repo, provider)

	var capturedParams domain.CheckoutSessionCreateParams
	uc.provider = &capturingProvider{inner: provider, captured: &capturedParams}

	_, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSuccess := fmt.Sprintf("http://localhost:4200/parent/invoices/%s?checkout=success&session_id={CHECKOUT_SESSION_ID}", repo.candidate.ID)
	if capturedParams.SuccessURL != expectedSuccess {
		t.Fatalf("expected success URL %s, got %s", expectedSuccess, capturedParams.SuccessURL)
	}
	expectedCancel := fmt.Sprintf("http://localhost:4200/parent/invoices/%s?checkout=cancelled", repo.candidate.ID)
	if capturedParams.CancelURL != expectedCancel {
		t.Fatalf("expected cancel URL %s, got %s", expectedCancel, capturedParams.CancelURL)
	}
	if capturedParams.AmountMinor != 5000 {
		t.Fatalf("expected 5000, got %d", capturedParams.AmountMinor)
	}
	if capturedParams.Currency != "gbp" {
		t.Fatalf("expected gbp, got %s", capturedParams.Currency)
	}
	if capturedParams.ProductName != "Nursery invoice payment" {
		t.Fatalf("expected Nursery invoice payment, got %s", capturedParams.ProductName)
	}
	if capturedParams.ProductDesc != "Invoice INV-2026-001" {
		t.Fatalf("expected Invoice INV-2026-001, got %s", capturedParams.ProductDesc)
	}
}

func TestCreateCheckoutSession_SeparateCallsUseSeparateAttemptIDs(t *testing.T) {
	repo := &fakeRepo{
		candidateFound: true,
		candidate:      payableCandidate(),
		stateFound:     true,
		paymentState:   payableState(),
	}
	provider := &fakeProvider{result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"}}

	count := 0
	uc := &CreateCheckoutSession{
		repo:             repo,
		txMgr:            &fakeTxManager{},
		provider:         provider,
		webBaseURL:       "http://localhost:4200",
		stripeConfigured: true,
		newUUID: func() uuid.UUID {
			count++
			return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", count))
		},
	}

	r1, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := uc.Execute(context.Background(), "t", "b", "m", "u", repo.candidate.ID, "req2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r1.PaymentAttemptID == r2.PaymentAttemptID {
		t.Fatal("expected different attempt IDs for separate calls")
	}
}

type capturingProvider struct {
	inner    domain.CheckoutProvider
	captured *domain.CheckoutSessionCreateParams
}

func (c *capturingProvider) CreateCheckoutSession(ctx context.Context, params domain.CheckoutSessionCreateParams) (domain.CheckoutSessionResult, error) {
	*c.captured = params
	return c.inner.CreateCheckoutSession(ctx, params)
}
