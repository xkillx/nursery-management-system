package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

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

	var body strings.Builder
	body.WriteString("To: ")
	body.WriteString(msg.To)
	body.WriteString("\r\nFrom: ")
	body.WriteString(s.from)
	body.WriteString("\r\nSubject: ")
	body.WriteString(msg.Subject)
	body.WriteString("\r\n\r\n")
	body.WriteString(msg.Text)

	return smtp.SendMail(addr, s.auth, s.from, []string{msg.To}, []byte(body.String()))
}
