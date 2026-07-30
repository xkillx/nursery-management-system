package smtp

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/modules/email/domain"
	platformemail "nursery-management-system/api/internal/platform/email"
)

type Provider struct {
	sender platformemail.Sender
}

func NewProvider(sender platformemail.Sender) *Provider {
	return &Provider{sender: sender}
}

func (p *Provider) Send(ctx context.Context, msg domain.OutboxMessage) (domain.SendResult, error) {
	platformMsg := platformemail.Message{
		To:      msg.Recipient,
		Subject: msg.Subject,
		Text:    string(msg.PayloadJSON),
		HTML:    "",
	}

	if err := p.sender.Send(ctx, platformMsg); err != nil {
		return domain.SendResult{}, fmt.Errorf("smtp send: %w", err)
	}

	return domain.SendResult{ProviderMessageID: ""}, nil
}
