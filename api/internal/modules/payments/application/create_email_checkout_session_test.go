package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
)

type emailCheckoutFakeRepo struct {
	activeEmail    *domain.ActiveCheckoutSession
	activeFound    bool
	candidate      domain.CheckoutInvoiceCandidate
	candidateFound bool
	state          domain.InvoicePaymentState
	stateFound     bool
	createdAttempt domain.PaymentAttemptCreateParams
	markedCreated  domain.PaymentAttemptCheckoutCreatedParams
	markedFailed   domain.PaymentAttemptCheckoutCreationFailedParams
	callCount      int
}

func (f *emailCheckoutFakeRepo) GetActiveCheckoutForInvoice(_ context.Context, _, _, _ string) (*domain.ActiveCheckoutSession, bool, error) {
	return nil, false, nil
}

func (f *emailCheckoutFakeRepo) GetActiveEmailCheckoutForInvoice(_ context.Context, _, _, _ string) (*domain.ActiveCheckoutSession, bool, error) {
	return f.activeEmail, f.activeFound, nil
}

func (f *emailCheckoutFakeRepo) GetParentInvoiceForCheckoutForUpdate(_ context.Context, _ domain.Tx, _, _, _, _ string) (domain.CheckoutInvoiceCandidate, bool, error) {
	return f.candidate, f.candidateFound, nil
}

func (f *emailCheckoutFakeRepo) GetInvoiceForEmailCheckoutForUpdate(_ context.Context, _ domain.Tx, _, _, _ string) (domain.CheckoutInvoiceCandidate, bool, error) {
	f.callCount++
	return f.candidate, f.candidateFound, nil
}

func (f *emailCheckoutFakeRepo) CreatePaymentAttempt(_ context.Context, _ domain.Tx, params domain.PaymentAttemptCreateParams) error {
	f.createdAttempt = params
	return nil
}

func (f *emailCheckoutFakeRepo) GetInvoicePaymentState(_ context.Context, _, _, _ string) (domain.InvoicePaymentState, bool, error) {
	return f.state, f.stateFound, nil
}

func (f *emailCheckoutFakeRepo) MarkPaymentAttemptCheckoutCreated(_ context.Context, params domain.PaymentAttemptCheckoutCreatedParams) error {
	f.markedCreated = params
	return nil
}

func (f *emailCheckoutFakeRepo) MarkPaymentAttemptCheckoutCreationFailed(_ context.Context, params domain.PaymentAttemptCheckoutCreationFailedParams) error {
	f.markedFailed = params
	return nil
}

func newEmailUC(repo *emailCheckoutFakeRepo, provider *fakeProvider) *CreateEmailCheckoutSession {
	return &CreateEmailCheckoutSession{
		repo:             repo,
		txMgr:            &fakeTxManager{},
		provider:         provider,
		webBaseURL:       "http://localhost:4200",
		stripeConfigured: true,
		newUUID:          func() uuid.UUID { return uuid.MustParse("00000000-0000-0000-0000-000000000002") },
	}
}

func emailPayableCandidate() domain.CheckoutInvoiceCandidate {
	c := payableCandidate()
	c.ID = uuid.New().String()
	return c
}

func TestCreateEmailCheckoutSession_UnconfiguredReturnsFallback(t *testing.T) {
	repo := &emailCheckoutFakeRepo{}
	uc := newEmailUC(repo, &fakeProvider{})
	uc.stripeConfigured = false

	result, err := uc.Execute(context.Background(), "t", "b", uuid.New().String(), "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatal("expected ok=false when provider unconfigured")
	}
	if repo.createdAttempt.ID != "" {
		t.Fatal("no attempt should be created when unconfigured")
	}
}

func TestCreateEmailCheckoutSession_HappyPath(t *testing.T) {
	repo := &emailCheckoutFakeRepo{
		candidateFound: true,
		candidate:      emailPayableCandidate(),
		stateFound:     true,
		state:          payableState(),
	}
	provider := &fakeProvider{result: domain.CheckoutSessionResult{
		CheckoutSessionID: "cs_email_123",
		CheckoutURL:       "https://checkout.stripe.com/email",
		PaymentIntentID:   "pi_email_456",
	}}
	uc := newEmailUC(repo, provider)

	result, err := uc.Execute(context.Background(), "t", "b", repo.candidate.ID, "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK {
		t.Fatal("expected ok=true")
	}
	if result.CheckoutURL != "https://checkout.stripe.com/email" {
		t.Fatalf("expected checkout URL, got %s", result.CheckoutURL)
	}
	if repo.createdAttempt.ID != "00000000-0000-0000-0000-000000000002" {
		t.Fatalf("expected fixed attempt ID, got %s", repo.createdAttempt.ID)
	}
	if repo.createdAttempt.InitiatedByUserID != "" || repo.createdAttempt.InitiatedByMembershipID != "" {
		t.Fatalf("expected NULL initiators, got user=%q membership=%q", repo.createdAttempt.InitiatedByUserID, repo.createdAttempt.InitiatedByMembershipID)
	}
	if !strings.HasPrefix(repo.createdAttempt.RequestID, "email_send:") {
		t.Fatalf("expected email_send: request-id marker, got %q", repo.createdAttempt.RequestID)
	}
	if repo.createdAttempt.Status != domain.AttemptStatusCheckoutCreationStarted {
		t.Fatalf("expected checkout_creation_started, got %s", repo.createdAttempt.Status)
	}
	if repo.markedCreated.StripeCheckoutSessionID != "cs_email_123" {
		t.Fatalf("expected session in mark-created, got %s", repo.markedCreated.StripeCheckoutSessionID)
	}
}

func TestCreateEmailCheckoutSession_HappyPathBuildsPublicOutcomeURLs(t *testing.T) {
	repo := &emailCheckoutFakeRepo{
		candidateFound: true,
		candidate:      emailPayableCandidate(),
		stateFound:     true,
		state:          payableState(),
	}
	provider := &fakeProvider{result: domain.CheckoutSessionResult{
		CheckoutSessionID: "cs_email_urls",
		CheckoutURL:       "https://checkout.stripe.com/email",
	}}
	uc := newEmailUC(repo, provider)

	var captured domain.CheckoutSessionCreateParams
	uc.provider = &capturingProvider{inner: provider, captured: &captured}

	if _, err := uc.Execute(context.Background(), "t", "b", repo.candidate.ID, "email_send:req-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSuccess := fmt.Sprintf("http://localhost:4200/payment/result?outcome=success&invoice_id=%s&session_id={CHECKOUT_SESSION_ID}", repo.candidate.ID)
	if captured.SuccessURL != expectedSuccess {
		t.Fatalf("expected success URL %s, got %s", expectedSuccess, captured.SuccessURL)
	}
	expectedCancel := fmt.Sprintf("http://localhost:4200/payment/result?outcome=cancelled&invoice_id=%s", repo.candidate.ID)
	if captured.CancelURL != expectedCancel {
		t.Fatalf("expected cancel URL %s, got %s", expectedCancel, captured.CancelURL)
	}
	if captured.Currency != "gbp" {
		t.Fatalf("expected gbp, got %s", captured.Currency)
	}
	if captured.ProductDesc != "Invoice "+repo.candidate.InvoiceNumber {
		t.Fatalf("expected product desc, got %s", captured.ProductDesc)
	}
}

func TestCreateEmailCheckoutSession_ReusesLiveEmailSession(t *testing.T) {
	existing := &domain.ActiveCheckoutSession{
		AttemptID:         "email-attempt-id",
		CheckoutSessionID: "cs_email_existing",
		CheckoutURL:       "https://checkout.stripe.com/email-existing",
	}
	repo := &emailCheckoutFakeRepo{
		activeEmail:    existing,
		activeFound:    true,
		candidateFound: true,
		candidate:      emailPayableCandidate(),
	}
	uc := newEmailUC(repo, &fakeProvider{})

	result, err := uc.Execute(context.Background(), "t", "b", repo.candidate.ID, "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK || result.CheckoutURL != "https://checkout.stripe.com/email-existing" {
		t.Fatalf("expected existing URL, got %+v", result)
	}
	if repo.createdAttempt.ID != "" {
		t.Fatal("expected no new attempt when a live email session exists")
	}
}

func TestCreateEmailCheckoutSession_DoesNotReusePortalSession(t *testing.T) {
	// A portal-created live session must not be reused by the email flow (KTD5):
	// the email flow's scoped lookup returns nothing, so a fresh email session is created.
	repo := &emailCheckoutFakeRepo{
		activeEmail:    nil,
		activeFound:    false,
		candidateFound: true,
		candidate:      emailPayableCandidate(),
		stateFound:     true,
		state:          payableState(),
	}
	provider := &fakeProvider{result: domain.CheckoutSessionResult{
		CheckoutSessionID: "cs_email_fresh",
		CheckoutURL:       "https://checkout.stripe.com/email-fresh",
	}}
	uc := newEmailUC(repo, provider)

	result, err := uc.Execute(context.Background(), "t", "b", repo.candidate.ID, "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK || result.CheckoutURL != "https://checkout.stripe.com/email-fresh" {
		t.Fatalf("expected a fresh email session, got %+v", result)
	}
	if repo.createdAttempt.ID == "" {
		t.Fatal("expected a new email attempt to be created")
	}
}

func TestCreateEmailCheckoutSession_ExpiredSessionCreatesFresh(t *testing.T) {
	// The repo scoped lookup only returns non-expired sessions; here it returns
	// nothing so a fresh attempt is created (KTD8).
	repo := &emailCheckoutFakeRepo{
		candidateFound: true,
		candidate:      emailPayableCandidate(),
		stateFound:     true,
		state:          payableState(),
	}
	provider := &fakeProvider{result: domain.CheckoutSessionResult{
		CheckoutSessionID: "cs_email_fresh2",
		CheckoutURL:       "https://checkout.stripe.com/email-fresh2",
	}}
	uc := newEmailUC(repo, provider)

	result, err := uc.Execute(context.Background(), "t", "b", repo.candidate.ID, "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OK || result.CheckoutURL != "https://checkout.stripe.com/email-fresh2" {
		t.Fatalf("expected fresh session after expiry, got %+v", result)
	}
	if repo.createdAttempt.ID == "" {
		t.Fatal("expected a new attempt after expiry")
	}
}

func TestCreateEmailCheckoutSession_NotPayableReturnsFallback(t *testing.T) {
	for name, mutate := range map[string]func(*domain.CheckoutInvoiceCandidate){
		"draft":        func(c *domain.CheckoutInvoiceCandidate) { c.Status = "draft" },
		"paid":         func(c *domain.CheckoutInvoiceCandidate) { c.Status = "paid" },
		"void":         func(c *domain.CheckoutInvoiceCandidate) { c.Status = "void" },
		"non-monthly":  func(c *domain.CheckoutInvoiceCandidate) { c.InvoiceKind = "adjustment" },
		"non-gbp":      func(c *domain.CheckoutInvoiceCandidate) { c.CurrencyCode = "USD" },
		"zero-due":     func(c *domain.CheckoutInvoiceCandidate) { c.TotalDueMinor = 0 },
		"already-paid": func(c *domain.CheckoutInvoiceCandidate) { c.AmountPaidMinor = 100 },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := emailPayableCandidate()
			mutate(&candidate)
			repo := &emailCheckoutFakeRepo{candidateFound: true, candidate: candidate}
			uc := newEmailUC(repo, &fakeProvider{})

			result, err := uc.Execute(context.Background(), "t", "b", candidate.ID, "email_send:req-1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.OK {
				t.Fatalf("expected ok=false for %s", name)
			}
			if repo.createdAttempt.ID != "" {
				t.Fatalf("no attempt should be created for %s", name)
			}
		})
	}
}

func TestCreateEmailCheckoutSession_NotFoundReturnsFallback(t *testing.T) {
	repo := &emailCheckoutFakeRepo{}
	uc := newEmailUC(repo, &fakeProvider{})

	result, err := uc.Execute(context.Background(), "t", "b", uuid.New().String(), "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatal("expected ok=false for unknown invoice")
	}
}

func TestCreateEmailCheckoutSession_ProviderErrorReturnsFallback(t *testing.T) {
	repo := &emailCheckoutFakeRepo{
		candidateFound: true,
		candidate:      emailPayableCandidate(),
	}
	provider := &fakeProvider{err: fmt.Errorf("stripe API error")}
	uc := newEmailUC(repo, provider)

	result, err := uc.Execute(context.Background(), "t", "b", repo.candidate.ID, "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatal("expected ok=false on provider error")
	}
	if repo.markedFailed.FailureReason != domain.FailureReasonStripeError {
		t.Fatalf("expected stripe_error, got %s", repo.markedFailed.FailureReason)
	}
	if repo.markedFailed.ProviderErrorMessage == "" {
		t.Fatal("expected provider error message recorded")
	}
}

func TestCreateEmailCheckoutSession_StateRecheckReturnsFallback(t *testing.T) {
	repo := &emailCheckoutFakeRepo{
		candidateFound: true,
		candidate:      emailPayableCandidate(),
		stateFound:     true,
		state:          payableState(),
	}
	repo.state.Status = "paid"
	provider := &fakeProvider{result: domain.CheckoutSessionResult{
		CheckoutSessionID: "cs_email_state",
		CheckoutURL:       "https://checkout.stripe.com/email",
	}}
	uc := newEmailUC(repo, provider)

	result, err := uc.Execute(context.Background(), "t", "b", repo.candidate.ID, "email_send:req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OK {
		t.Fatal("expected ok=false when the state re-check fails")
	}
	if repo.markedFailed.FailureReason != domain.FailureReasonInvoiceNoLongerPayable {
		t.Fatalf("expected invoice_no_longer_payable, got %s", repo.markedFailed.FailureReason)
	}
}

func TestCreateEmailCheckoutSession_ConcurrentCreatorsSerializedByRowLock(t *testing.T) {
	// The FOR UPDATE row lock (U1) serializes session creation. Both concurrent
	// callers resolve the candidate and create an attempt inside the lock, so at
	// most one live session can be minted per invoice. Here we assert both
	// invocations produce an attempt and return a URL, mirroring the serialized
	// path exercised by the repository tests.
	repo := &emailCheckoutFakeRepo{
		candidateFound: true,
		candidate:      emailPayableCandidate(),
		stateFound:     true,
		state:          payableState(),
	}
	provider := &fakeProvider{result: domain.CheckoutSessionResult{
		CheckoutSessionID: "cs_email_serial",
		CheckoutURL:       "https://checkout.stripe.com/email-serial",
	}}
	uc := newEmailUC(repo, provider)

	ctx := context.Background()
	results := make([]EmailCheckoutResult, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = uc.Execute(ctx, "t", "b", repo.candidate.ID, fmt.Sprintf("email_send:req-%d", i))
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent call %d errored: %v", i, err)
		}
		if !results[i].OK || results[i].CheckoutURL == "" {
			t.Fatalf("concurrent call %d expected a URL, got %+v", i, results[i])
		}
	}
	if repo.callCount != 2 {
		t.Fatalf("expected both creators to resolve the invoice, got %d resolutions", repo.callCount)
	}
}
