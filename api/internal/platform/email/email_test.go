package email

import (
	"bufio"
	"context"
	"net"
	"strconv"
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

	body := sender.buildMessage(msg)

	if !strings.Contains(body, "To: recipient@example.com") {
		t.Error("expected To header")
	}
	if !strings.Contains(body, "From: NurseryPro <test@example.com>") {
		t.Errorf("expected From header with display name, got body:\n%s", body)
	}
	if !strings.Contains(body, "Subject: Test Subject") {
		t.Error("expected Subject header")
	}
	if !strings.Contains(body, "Hello, this is a plain text email.") {
		t.Error("expected text body")
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

	body := sender.buildMessage(msg)

	if !strings.Contains(body, "MIME-Version: 1.0") {
		t.Error("expected MIME-Version header")
	}
	if !strings.Contains(body, "Content-Type: multipart/alternative") {
		t.Error("expected multipart/alternative content type")
	}
	if !strings.Contains(body, "From: NurseryPro <test@example.com>") {
		t.Errorf("expected From header with display name, got body:\n%s", body)
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

	body := sender.buildMessage(msg)

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

	body := sender.buildMessage(msg)

	if !strings.Contains(body, "From: NurseryPro <test@example.com>") {
		t.Errorf("expected From header with display name, got body:\n%s", body)
	}
	if strings.Contains(body, "multipart/alternative") {
		t.Error("unexpected multipart for backward compatible message")
	}
	if !strings.Contains(body, "Plain text only, no HTML field set") {
		t.Error("expected text content")
	}
}

func TestSMTPSender_Send_EnvelopeStaysBareWhileFromHeaderHasDisplayName(t *testing.T) {
	addr, envelopeFrom, data := startFakeSMTPServer(t)
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}

	sender := NewSMTPSender(host, port, "", "", "no-reply@example.local")

	msg := Message{
		To:      "recipient@example.com",
		Subject: "Test Subject",
		Text:    "Plain text version",
		HTML:    "<html><body><h1>HTML version</h1></body></html>",
	}

	if err := sender.Send(context.Background(), msg); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if *envelopeFrom != "no-reply@example.local" {
		t.Errorf("envelope sender = %q, want bare address %q", *envelopeFrom, "no-reply@example.local")
	}
	if !strings.Contains(*data, "From: NurseryPro <no-reply@example.local>") {
		t.Errorf("message From header missing display name, got:\n%s", *data)
	}
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

// startFakeSMTPServer runs a minimal SMTP server on a local port that accepts a
// single message, recording the MAIL FROM envelope and the DATA payload.
func startFakeSMTPServer(t *testing.T) (addr string, envelopeFrom *string, data *string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	env := new(string)
	payload := new(string)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		write := func(s string) { _, _ = conn.Write([]byte(s)) }
		write("220 localhost ESMTP fake\r\n")

		reader := bufio.NewReader(conn)
		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			trimmed := strings.ToUpper(strings.TrimSpace(line))

			if inData {
				*payload += line
				if line == ".\r\n" {
					write("250 OK: queued\r\n")
					inData = false
				}
				continue
			}

			switch {
			case strings.HasPrefix(trimmed, "EHLO"):
				write("250-localhost\r\n250 SIZE 1000000\r\n")
			case strings.HasPrefix(trimmed, "MAIL FROM"):
				*env = strings.Trim(strings.TrimPrefix(strings.TrimSpace(line), "MAIL FROM:"), "<>")
				write("250 OK\r\n")
			case strings.HasPrefix(trimmed, "RCPT TO"):
				write("250 OK\r\n")
			case trimmed == "DATA":
				write("354 End data with <CR><LF>.<CR><LF>\r\n")
				inData = true
			case trimmed == "QUIT":
				write("221 Bye\r\n")
				return
			default:
				write("250 OK\r\n")
			}
		}
	}()

	return ln.Addr().String(), env, payload
}
