package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
)

var errRenderFailed = errors.New("render failed")

type stubRenderer struct {
	html            string
	text            string
	err             error
	receivedPayLink bool
}

func (r *stubRenderer) Render(_ string, _ int, data map[string]interface{}) (string, string, error) {
	if _, ok := data["PayLink"]; ok {
		r.receivedPayLink = true
	}
	return r.html, r.text, r.err
}

type stubOutboxRepo struct {
	domain.OutboxRepository
	pending      []domain.OutboxMessage
	sentID       uuid.UUID
	lastErr      string
	lastStatus   domain.Status
	updateCalled bool
}

func (s *stubOutboxRepo) GetPending(_ context.Context, _ int) ([]domain.OutboxMessage, error) {
	return s.pending, nil
}

func (s *stubOutboxRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.Status, _ int, _ interface{}, lastError *string, _ *string) error {
	s.updateCalled = true
	s.sentID = id
	s.lastStatus = status
	if lastError != nil {
		s.lastErr = *lastError
	}
	return nil
}

type stubProvider struct {
	got domain.OutboxMessage
	err error
}

func (p *stubProvider) Send(_ context.Context, msg domain.OutboxMessage) (domain.SendResult, error) {
	p.got = msg
	return domain.SendResult{}, p.err
}

type stubPayLinkProvider struct {
	calls []payLinkCall
	url   string
	ok    bool
	err   error
}

type payLinkCall struct {
	tenantID  uuid.UUID
	branchID  uuid.UUID
	invoiceID string
	requestID string
}

func (p *stubPayLinkProvider) CreateEmailCheckoutSession(_ context.Context, tenantID, branchID uuid.UUID, invoiceID, requestID string) (string, bool, error) {
	p.calls = append(p.calls, payLinkCall{tenantID: tenantID, branchID: branchID, invoiceID: invoiceID, requestID: requestID})
	return p.url, p.ok, p.err
}

func newInvoiceMessage() domain.OutboxMessage {
	return domain.OutboxMessage{
		ID:              uuid.New(),
		TenantID:        uuid.New(),
		BranchID:        uuid.New(),
		Recipient:       "parent@example.com",
		Subject:         "New Invoice",
		TemplateName:    "issued",
		TemplateVersion: 2,
		PayloadJSON:     []byte(`{"ChildName":"Leo","PortalLink":"https://app/parent/invoices/x"}`),
		EntityID:        "invoice-123",
		MaxAttempts:     8,
	}
}

func TestSendPendingEmails_PayLinkInjectedIntoInvoicePayload(t *testing.T) {
	msg := newInvoiceMessage()
	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{html: "<html>Pay</html>", text: "Pay now"}
	provider := &stubProvider{}
	payLinks := &stubPayLinkProvider{url: "https://checkout.stripe.com/email-link", ok: true}

	uc := NewSendPendingEmails(repo, provider, renderer, payLinks, 0, 10)
	if _, _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(payLinks.calls) != 1 {
		t.Fatalf("expected 1 pay-link provider call, got %d", len(payLinks.calls))
	}
	call := payLinks.calls[0]
	if call.tenantID != msg.TenantID || call.branchID != msg.BranchID || call.invoiceID != "invoice-123" {
		t.Fatalf("provider called with wrong ids: %+v", call)
	}
	if !strings.HasPrefix(call.requestID, "email_send:") {
		t.Fatalf("expected email_send: request-id marker, got %q", call.requestID)
	}

	if !renderer.receivedPayLink {
		t.Fatal("expected PayLink to be injected into the render payload")
	}
}

func TestSendPendingEmails_ReceiptNeverCallsPayLinkProvider(t *testing.T) {
	msg := newInvoiceMessage()
	msg.TemplateName = "receipt"
	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{html: "<html>Receipt</html>", text: "Receipt"}
	provider := &stubProvider{}
	payLinks := &stubPayLinkProvider{url: "https://checkout.stripe.com/email-link", ok: true}

	uc := NewSendPendingEmails(repo, provider, renderer, payLinks, 0, 10)
	if _, _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payLinks.calls) != 0 {
		t.Fatalf("expected no pay-link provider calls for receipt, got %d", len(payLinks.calls))
	}
	if renderer.receivedPayLink {
		t.Fatal("receipt payload must not contain PayLink")
	}
}

func TestSendPendingEmails_PayLinkUnavailableStillSends(t *testing.T) {
	msg := newInvoiceMessage()
	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{html: "<html>Pay</html>", text: "Pay now"}
	provider := &stubProvider{}
	payLinks := &stubPayLinkProvider{ok: false}

	uc := NewSendPendingEmails(repo, provider, renderer, payLinks, 0, 10)
	sent, failed, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 1 || failed != 0 {
		t.Fatalf("sent = %d, failed = %d, want 1, 0", sent, failed)
	}
	if renderer.receivedPayLink {
		t.Fatal("ok=false must not inject PayLink")
	}
	if repo.lastStatus != domain.StatusSent {
		t.Fatalf("status = %q, want %q", repo.lastStatus, domain.StatusSent)
	}
}

func TestSendPendingEmails_PayLinkProviderErrorStillSends(t *testing.T) {
	msg := newInvoiceMessage()
	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{html: "<html>Pay</html>", text: "Pay now"}
	provider := &stubProvider{}
	payLinks := &stubPayLinkProvider{err: errors.New("stripe down")}

	uc := NewSendPendingEmails(repo, provider, renderer, payLinks, 0, 10)
	sent, failed, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 1 || failed != 0 {
		t.Fatalf("sent = %d, failed = %d, want 1, 0", sent, failed)
	}
	if renderer.receivedPayLink {
		t.Fatal("provider error must not inject PayLink")
	}
	if repo.lastStatus != domain.StatusSent {
		t.Fatalf("status = %q, want %q (email must be marked sent, not failed)", repo.lastStatus, domain.StatusSent)
	}
}

func TestSendPendingEmails_PayLinkProviderCalledAgainOnRetry(t *testing.T) {
	msg := newInvoiceMessage()
	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{html: "<html>Pay</html>", text: "Pay now"}
	provider := &stubProvider{}
	payLinks := &stubPayLinkProvider{url: "https://checkout.stripe.com/email-link", ok: true}

	uc := NewSendPendingEmails(repo, provider, renderer, payLinks, 0, 10)
	if _, _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payLinks.calls) != 2 {
		t.Fatalf("expected the provider to be called again on retry, got %d calls", len(payLinks.calls))
	}
}

func TestSendPendingEmails_CarriesRenderedBodiesToProvider(t *testing.T) {
	msg := domain.OutboxMessage{
		ID:              uuid.New(),
		Recipient:       "parent@example.com",
		Subject:         "New Invoice",
		TemplateName:    "issued",
		TemplateVersion: 2,
		PayloadJSON:     []byte(`{"ChildName":"Leo"}`),
		MaxAttempts:     8,
	}

	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{html: "<html>Hi</html>", text: "Hi Leo"}
	provider := &stubProvider{}

	uc := NewSendPendingEmails(repo, provider, renderer, nil, 0, 10)
	sent, failed, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 1 || failed != 0 {
		t.Fatalf("sent = %d, failed = %d, want 1, 0", sent, failed)
	}

	if provider.got.RenderedHTML != "<html>Hi</html>" {
		t.Errorf("provider RenderedHTML = %q, want %q", provider.got.RenderedHTML, "<html>Hi</html>")
	}
	if provider.got.RenderedText != "Hi Leo" {
		t.Errorf("provider RenderedText = %q, want %q", provider.got.RenderedText, "Hi Leo")
	}
	if repo.lastStatus != domain.StatusSent {
		t.Errorf("status = %q, want %q", repo.lastStatus, domain.StatusSent)
	}
}

func TestSendPendingEmails_RenderErrorMarksFailedAndSkipsProvider(t *testing.T) {
	msg := domain.OutboxMessage{
		ID:              uuid.New(),
		Recipient:       "parent@example.com",
		Subject:         "New Invoice",
		TemplateName:    "issued",
		TemplateVersion: 2,
		PayloadJSON:     []byte(`{}`),
		MaxAttempts:     8,
	}

	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{err: errRenderFailed}
	provider := &stubProvider{}

	uc := NewSendPendingEmails(repo, provider, renderer, nil, 0, 10)
	sent, failed, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 0 || failed != 1 {
		t.Fatalf("sent = %d, failed = %d, want 0, 1", sent, failed)
	}
	if !repo.updateCalled {
		t.Fatal("expected UpdateStatus to be called on render failure")
	}
	if repo.lastStatus != domain.StatusFailed {
		t.Errorf("status = %q, want %q", repo.lastStatus, domain.StatusFailed)
	}
	if provider.got.ID != uuid.Nil {
		t.Errorf("provider was called with message %v, want no call", provider.got.ID)
	}
}

func TestSendPendingEmails_EmptyRenderedTextPassedThrough(t *testing.T) {
	msg := domain.OutboxMessage{
		ID:              uuid.New(),
		Recipient:       "parent@example.com",
		Subject:         "Plain",
		TemplateName:    "issued",
		TemplateVersion: 2,
		PayloadJSON:     []byte(`{"ChildName":"Leo"}`),
		MaxAttempts:     8,
	}

	repo := &stubOutboxRepo{pending: []domain.OutboxMessage{msg}}
	renderer := &stubRenderer{html: "", text: ""}
	provider := &stubProvider{}

	uc := NewSendPendingEmails(repo, provider, renderer, nil, 0, 10)
	if _, _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.got.RenderedText != "" {
		t.Errorf("provider RenderedText = %q, want empty (fallback handled by provider)", provider.got.RenderedText)
	}
}
