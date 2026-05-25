package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/platform/email"
)

type EmailAdapter struct {
	sender email.Sender
	from   string
}

func NewEmailAdapter(sender email.Sender) *EmailAdapter {
	return &EmailAdapter{sender: sender}
}

func (a *EmailAdapter) SendPasswordReset(ctx context.Context, toEmail string, resetURL string) error {
	msg := email.Message{
		To:      toEmail,
		Subject: "Reset your password",
		Text:    fmt.Sprintf("You requested a password reset. Click the link below to set a new password:\n\n%s\n\nIf you did not request this, ignore this email.", resetURL),
	}
	return a.sender.Send(ctx, msg)
}
