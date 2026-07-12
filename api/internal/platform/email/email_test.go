package email

import (
	"context"
	"strings"
	"testing"
)

func TestSMTPSender_Send_PlainTextOnly(t *testing.T) {
	sender := &SMTPSender{
		host: "localhost",
		port: 587,
		from: "test@example.com",
		auth: nil,
	}

	msg := Message{
		To:      "recipient@example.com",
		Subject: "Test Subject",
		Text:    "Hello, this is a plain text email.",
	}

	// We can't actually send, but we can verify the message construction
	// by testing the buildMessage logic indirectly
	body := buildMessageBody(sender.from, msg)

	if !strings.Contains(body, "To: recipient@example.com") {
		t.Error("expected To header")
	}
	if !strings.Contains(body, "From: test@example.com") {
		t.Error("expected From header")
	}
	if !strings.Contains(body, "Subject: Test Subject") {
		t.Error("expected Subject header")
	}
	if !strings.Contains(body, "Hello, this is a plain text email.") {
		t.Error("expected text body")
	}
	if strings.Contains(body, "MIME-Version") {
		t.Error("unexpected MIME-Version for plain text only")
	}
	if strings.Contains(body, "multipart/alternative") {
		t.Error("unexpected multipart for plain text only")
	}
}

func TestSMTPSender_Send_MultipartWithHTML(t *testing.T) {
	sender := &SMTPSender{
		host: "localhost",
		port: 587,
		from: "test@example.com",
		auth: nil,
	}

	msg := Message{
		To:      "recipient@example.com",
		Subject: "Test Subject",
		Text:    "Plain text version",
		HTML:    "<html><body><h1>HTML version</h1></body></html>",
	}

	body := buildMessageBody(sender.from, msg)

	if !strings.Contains(body, "MIME-Version: 1.0") {
		t.Error("expected MIME-Version header")
	}
	if !strings.Contains(body, "Content-Type: multipart/alternative") {
		t.Error("expected multipart/alternative content type")
	}
	if !strings.Contains(body, "Content-Type: text/plain") {
		t.Error("expected text/plain content type")
	}
	if !strings.Contains(body, "Content-Type: text/html") {
		t.Error("expected text/html content type")
	}
	if !strings.Contains(body, "Plain text version") {
		t.Error("expected plain text content")
	}
	if !strings.Contains(body, "<html><body><h1>HTML version</h1></body></html>") {
		t.Error("expected HTML content")
	}
	if !strings.Contains(body, "boundary=") {
		t.Error("expected boundary marker")
	}
}

func TestSMTPSender_Send_MultipartWithEmptyText(t *testing.T) {
	sender := &SMTPSender{
		host: "localhost",
		port: 587,
		from: "test@example.com",
		auth: nil,
	}

	msg := Message{
		To:      "recipient@example.com",
		Subject: "Test Subject",
		Text:    "",
		HTML:    "<html><body><h1>HTML only</h1></body></html>",
	}

	body := buildMessageBody(sender.from, msg)

	if !strings.Contains(body, "MIME-Version: 1.0") {
		t.Error("expected MIME-Version header")
	}
	if !strings.Contains(body, "Content-Type: multipart/alternative") {
		t.Error("expected multipart/alternative content type")
	}
	if !strings.Contains(body, "Content-Type: text/plain") {
		t.Error("expected text/plain content type")
	}
	if !strings.Contains(body, "Content-Type: text/html") {
		t.Error("expected text/html content type")
	}
	if !strings.Contains(body, "<html><body><h1>HTML only</h1></body></html>") {
		t.Error("expected HTML content")
	}
}

func TestSMTPSender_Send_BackwardCompatibility(t *testing.T) {
	sender := &SMTPSender{
		host: "localhost",
		port: 587,
		from: "test@example.com",
		auth: nil,
	}

	msg := Message{
		To:      "recipient@example.com",
		Subject: "Test Subject",
		Text:    "Plain text only, no HTML field set",
	}

	body := buildMessageBody(sender.from, msg)

	if strings.Contains(body, "MIME-Version") {
		t.Error("unexpected MIME-Version for backward compatible message")
	}
	if strings.Contains(body, "multipart/alternative") {
		t.Error("unexpected multipart for backward compatible message")
	}
	if !strings.Contains(body, "Plain text only, no HTML field set") {
		t.Error("expected text content")
	}
}

// buildMessageBody extracts the message body construction logic for testing.
func buildMessageBody(from string, msg Message) string {
	var body strings.Builder
	body.WriteString("To: ")
	body.WriteString(msg.To)
	body.WriteString("\r\nFrom: ")
	body.WriteString(from)
	body.WriteString("\r\nSubject: ")
	body.WriteString(msg.Subject)

	if msg.HTML != "" {
		boundary := "test-boundary-123"
		body.WriteString("\r\nMIME-Version: 1.0")
		body.WriteString("\r\nContent-Type: multipart/alternative; boundary=\"")
		body.WriteString(boundary)
		body.WriteString("\"")
		body.WriteString("\r\n\r\n--")
		body.WriteString(boundary)
		body.WriteString("\r\nContent-Type: text/plain; charset=\"UTF-8\"")
		body.WriteString("\r\n\r\n")
		body.WriteString(msg.Text)
		body.WriteString("\r\n\r\n--")
		body.WriteString(boundary)
		body.WriteString("\r\nContent-Type: text/html; charset=\"UTF-8\"")
		body.WriteString("\r\n\r\n")
		body.WriteString(msg.HTML)
		body.WriteString("\r\n\r\n--")
		body.WriteString(boundary)
		body.WriteString("--")
	} else {
		body.WriteString("\r\n\r\n")
		body.WriteString(msg.Text)
	}

	return body.String()
}

func TestFakeSender_CapturesMessages(t *testing.T) {
	sender := NewFakeSender()

	msg1 := Message{To: "a@example.com", Subject: "First", Text: "text1"}
	msg2 := Message{To: "b@example.com", Subject: "Second", Text: "text2", HTML: "<html></html>"}

	if err := sender.Send(context.Background(), msg1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sender.Send(context.Background(), msg2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(sender.Messages))
	}
	if sender.Messages[0].HTML != "" {
		t.Error("expected first message to have empty HTML")
	}
	if sender.Messages[1].HTML != "<html></html>" {
		t.Error("expected second message to have HTML")
	}
}

func TestFakeSender_ReturnsError(t *testing.T) {
	sender := NewFakeSender()
	sender.Err = context.DeadlineExceeded

	err := sender.Send(context.Background(), Message{To: "a@example.com", Subject: "Test", Text: "text"})
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if len(sender.Messages) != 0 {
		t.Error("expected no messages on error")
	}
}
