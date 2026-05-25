package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/platform/email"
)

type InviteEmailAdapter struct {
	sender email.Sender
}

func NewInviteEmailAdapter(sender email.Sender) *InviteEmailAdapter {
	return &InviteEmailAdapter{sender: sender}
}

func (a *InviteEmailAdapter) SendInvite(ctx context.Context, toEmail, role, acceptURL string) error {
	msg := email.Message{
		To:      toEmail,
		Subject: fmt.Sprintf("You're invited to join as %s", role),
		Text: fmt.Sprintf(
			"You have been invited to join as a %s.\n\nClick the link below to accept:\n%s\n\nThis invitation expires in 7 days.",
			role, acceptURL,
		),
	}
	return a.sender.Send(ctx, msg)
}
