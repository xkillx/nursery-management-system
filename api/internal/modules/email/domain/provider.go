package domain

import "context"

type SendResult struct {
	ProviderMessageID string
}

type EmailProvider interface {
	Send(ctx context.Context, msg OutboxMessage) (SendResult, error)
}
