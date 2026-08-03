package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"

	"nursery-management-system/api/internal/platform/uid"
)

// senderDisplayName is shown in the From header while the SMTP envelope sender
// remains the bare address (KTD-6: net/smtp rejects a display name in MAIL FROM).
const senderDisplayName = "NurseryPro"

type SMTPSender struct {
	host string
	port int
	from string
	auth smtp.Auth
}

func NewSMTPSender(host string, port int, user, pass, from string) *SMTPSender {
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return &SMTPSender{
		host: host,
		port: port,
		from: from,
		auth: auth,
	}
}

func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	body := s.buildMessage(msg)

	// The envelope sender stays the bare address; the display name appears only
	// in the From header (KTD-6).
	return smtp.SendMail(addr, s.auth, s.from, []string{msg.To}, []byte(body))
}

func (s *SMTPSender) buildMessage(msg Message) string {
	var body strings.Builder
	body.WriteString("To: ")
	body.WriteString(msg.To)
	body.WriteString("\r\nFrom: ")
	body.WriteString(s.fromHeader())
	body.WriteString("\r\nSubject: ")
	body.WriteString(msg.Subject)
	body.WriteString("\r\nMIME-Version: 1.0")

	if len(msg.Attachments) > 0 {
		s.buildMultipartMixed(&body, msg)
	} else if msg.HTML != "" {
		s.buildMultipartAlternative(&body, msg)
	} else {
		body.WriteString("\r\nContent-Type: text/plain; charset=\"UTF-8\"")
		body.WriteString("\r\n\r\n")
		body.WriteString(msg.Text)
	}

	return body.String()
}

func (s *SMTPSender) fromHeader() string {
	return fmt.Sprintf("%s <%s>", senderDisplayName, s.from)
}

func (s *SMTPSender) buildMultipartAlternative(body *strings.Builder, msg Message) {
	boundary := uid.NewUUID().String()
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
}

func (s *SMTPSender) buildMultipartMixed(body *strings.Builder, msg Message) {
	mixedBoundary := uid.NewUUID().String()
	body.WriteString("\r\nContent-Type: multipart/mixed; boundary=\"")
	body.WriteString(mixedBoundary)
	body.WriteString("\"")

	// Body part
	body.WriteString("\r\n\r\n--")
	body.WriteString(mixedBoundary)
	if msg.HTML != "" {
		altBoundary := uid.NewUUID().String()
		body.WriteString("\r\nContent-Type: multipart/alternative; boundary=\"")
		body.WriteString(altBoundary)
		body.WriteString("\"")
		body.WriteString("\r\n\r\n--")
		body.WriteString(altBoundary)
		body.WriteString("\r\nContent-Type: text/plain; charset=\"UTF-8\"")
		body.WriteString("\r\n\r\n")
		body.WriteString(msg.Text)
		body.WriteString("\r\n\r\n--")
		body.WriteString(altBoundary)
		body.WriteString("\r\nContent-Type: text/html; charset=\"UTF-8\"")
		body.WriteString("\r\n\r\n")
		body.WriteString(msg.HTML)
		body.WriteString("\r\n\r\n--")
		body.WriteString(altBoundary)
		body.WriteString("--")
	} else {
		body.WriteString("\r\nContent-Type: text/plain; charset=\"UTF-8\"")
		body.WriteString("\r\n\r\n")
		body.WriteString(msg.Text)
	}

	// Attachment parts
	for _, att := range msg.Attachments {
		body.WriteString("\r\n\r\n--")
		body.WriteString(mixedBoundary)
		body.WriteString("\r\nContent-Type: ")
		body.WriteString(att.ContentType)
		body.WriteString("; name=\"")
		body.WriteString(att.Filename)
		body.WriteString("\"")
		body.WriteString("\r\nContent-Disposition: attachment; filename=\"")
		body.WriteString(att.Filename)
		body.WriteString("\"")
		body.WriteString("\r\nContent-Transfer-Encoding: base64")
		body.WriteString("\r\n\r\n")

		encoded := base64.StdEncoding.EncodeToString(att.Data)
		// Wrap at 76 characters per RFC 2045
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			body.WriteString(encoded[i:end])
			if end < len(encoded) {
				body.WriteString("\r\n")
			}
		}
	}

	body.WriteString("\r\n\r\n--")
	body.WriteString(mixedBoundary)
	body.WriteString("--")
}
