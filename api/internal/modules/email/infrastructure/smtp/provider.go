package smtp

import (
	"context"
	"fmt"
	"log/slog"

	"nursery-management-system/api/internal/modules/email/domain"
	platformemail "nursery-management-system/api/internal/platform/email"
	"nursery-management-system/api/internal/platform/storage"
)

type Provider struct {
	sender  platformemail.Sender
	storage storage.Service
}

func NewProvider(sender platformemail.Sender, storage storage.Service) *Provider {
	return &Provider{sender: sender, storage: storage}
}

func (p *Provider) Send(ctx context.Context, msg domain.OutboxMessage) (domain.SendResult, error) {
	// Prefer the rendered template bodies. When no rendered text exists (e.g. a
	// non-template message), fall back to the payload JSON as the plain-text body.
	textBody := msg.RenderedText
	if textBody == "" {
		textBody = string(msg.PayloadJSON)
	}

	platformMsg := platformemail.Message{
		To:      msg.Recipient,
		Subject: msg.Subject,
		Text:    textBody,
		HTML:    msg.RenderedHTML,
	}

	// Fetch attachment data from S3 if present
	if len(msg.Attachments) > 0 && p.storage != nil {
		for _, ref := range msg.Attachments {
			data, err := p.storage.Download(ctx, ref.S3Key)
			if err != nil {
				return domain.SendResult{}, fmt.Errorf("download attachment %s from S3: %w", ref.S3Key, err)
			}
			platformMsg.Attachments = append(platformMsg.Attachments, platformemail.Attachment{
				Filename:    ref.Filename,
				ContentType: ref.ContentType,
				Data:        data,
			})
		}
	}

	if err := p.sender.Send(ctx, platformMsg); err != nil {
		return domain.SendResult{}, fmt.Errorf("smtp send: %w", err)
	}

	// Clean up S3 objects after successful send
	if len(msg.Attachments) > 0 && p.storage != nil {
		for _, ref := range msg.Attachments {
			if delErr := p.storage.Delete(ctx, ref.S3Key); delErr != nil {
				slog.WarnContext(ctx, "s3_cleanup_failed",
					"s3_key", ref.S3Key,
					"error", delErr,
				)
			}
		}
	}

	return domain.SendResult{ProviderMessageID: ""}, nil
}
