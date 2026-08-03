package smtp

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
	platformemail "nursery-management-system/api/internal/platform/email"
	"nursery-management-system/api/internal/platform/storage"
)

func TestProvider_Send_UsesRenderedBodies(t *testing.T) {
	sender := platformemail.NewFakeSender()
	p := NewProvider(sender, nil)

	msg := domain.OutboxMessage{
		ID:              uuid.New(),
		Recipient:       "parent@example.com",
		Subject:         "New Invoice",
		PayloadJSON:     []byte(`{"TotalDue":"£10.00"}`),
		RenderedHTML:    "<html><body><h1>New Invoice</h1></body></html>",
		RenderedText:    "A new invoice is ready.",
		TemplateName:    "issued",
		TemplateVersion: 2,
	}

	if _, err := p.Send(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.Messages))
	}
	got := sender.Messages[0]
	if got.HTML != msg.RenderedHTML {
		t.Errorf("HTML = %q, want %q", got.HTML, msg.RenderedHTML)
	}
	if got.Text != msg.RenderedText {
		t.Errorf("Text = %q, want %q", got.Text, msg.RenderedText)
	}
}

func TestProvider_Send_FallsBackToPayloadJSON(t *testing.T) {
	sender := platformemail.NewFakeSender()
	p := NewProvider(sender, nil)

	msg := domain.OutboxMessage{
		ID:          uuid.New(),
		Recipient:   "parent@example.com",
		Subject:     "Plain",
		PayloadJSON: []byte(`{"message":"hello"}`),
	}

	if _, err := p.Send(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.Messages))
	}
	got := sender.Messages[0]
	if got.Text != string(msg.PayloadJSON) {
		t.Errorf("Text = %q, want %q", got.Text, string(msg.PayloadJSON))
	}
	if got.HTML != "" {
		t.Errorf("HTML = %q, want empty", got.HTML)
	}
}

func TestProvider_Send_AttachesAndDeletesS3Object(t *testing.T) {
	sender := platformemail.NewFakeSender()
	fakeStorage := storage.NewFakeService()
	key := "invoices/abc/invoice.pdf"
	fakeStorage.Objects[key] = []byte("%PDF-1.4 test")

	p := NewProvider(sender, fakeStorage)

	msg := domain.OutboxMessage{
		ID:           uuid.New(),
		Recipient:    "parent@example.com",
		Subject:      "Invoice",
		RenderedHTML: "<html><body>Invoice</body></html>",
		RenderedText: "Invoice",
		Attachments: []domain.AttachmentRef{
			{Filename: "invoice.pdf", ContentType: "application/pdf", S3Key: key},
		},
	}

	if _, err := p.Send(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.Messages))
	}
	got := sender.Messages[0]
	if len(got.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(got.Attachments))
	}
	if string(got.Attachments[0].Data) != "%PDF-1.4 test" {
		t.Errorf("attachment data = %q, want %q", got.Attachments[0].Data, "%PDF-1.4 test")
	}

	if _, exists := fakeStorage.Objects[key]; exists {
		t.Error("expected S3 object to be deleted after successful send")
	}
}

func TestProvider_Send_DownloadFailureRetries(t *testing.T) {
	sender := platformemail.NewFakeSender()
	fakeStorage := storage.NewFakeService()
	key := "invoices/missing/invoice.pdf"

	p := NewProvider(sender, fakeStorage)

	msg := domain.OutboxMessage{
		ID:           uuid.New(),
		Recipient:    "parent@example.com",
		Subject:      "Invoice",
		RenderedHTML: "<html>Invoice</html>",
		Attachments: []domain.AttachmentRef{
			{Filename: "invoice.pdf", ContentType: "application/pdf", S3Key: key},
		},
	}

	if _, err := p.Send(context.Background(), msg); err == nil {
		t.Fatal("expected error for missing attachment object")
	}
	if len(sender.Messages) != 0 {
		t.Errorf("expected no message sent on download failure, got %d", len(sender.Messages))
	}
}
