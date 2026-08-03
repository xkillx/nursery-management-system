package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
)

var errRenderFailed = errors.New("render failed")

type stubRenderer struct {
	html string
	text string
	err  error
}

func (r *stubRenderer) Render(_ string, _ int, _ map[string]interface{}) (string, string, error) {
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

	uc := NewSendPendingEmails(repo, provider, renderer, 0, 10)
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

	uc := NewSendPendingEmails(repo, provider, renderer, 0, 10)
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

	uc := NewSendPendingEmails(repo, provider, renderer, 0, 10)
	if _, _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.got.RenderedText != "" {
		t.Errorf("provider RenderedText = %q, want empty (fallback handled by provider)", provider.got.RenderedText)
	}
}
