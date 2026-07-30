package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"nursery-management-system/api/internal/modules/email/domain"
)

type TemplateRenderer interface {
	Render(templateName string, version int, data map[string]interface{}) (htmlBody, textBody string, err error)
}

type SendPendingEmails struct {
	repo      domain.OutboxRepository
	provider  domain.EmailProvider
	renderer  TemplateRenderer
	rateLimit float64
	batchSize int
}

func NewSendPendingEmails(
	repo domain.OutboxRepository,
	provider domain.EmailProvider,
	renderer TemplateRenderer,
	rateLimit float64,
	batchSize int,
) *SendPendingEmails {
	return &SendPendingEmails{
		repo:      repo,
		provider:  provider,
		renderer:  renderer,
		rateLimit: rateLimit,
		batchSize: batchSize,
	}
}

func (uc *SendPendingEmails) Execute(ctx context.Context) (int, int, error) {
	messages, err := uc.repo.GetPending(ctx, uc.batchSize)
	if err != nil {
		return 0, 0, fmt.Errorf("get pending emails: %w", err)
	}

	sent := 0
	failed := 0

	for _, msg := range messages {
		if err := ctx.Err(); err != nil {
			break
		}

		if err := uc.processMessage(ctx, msg); err != nil {
			failed++
		} else {
			sent++
		}

		if uc.rateLimit > 0 {
			time.Sleep(time.Duration(float64(time.Second) / uc.rateLimit))
		}
	}

	return sent, failed, nil
}

func (uc *SendPendingEmails) processMessage(ctx context.Context, msg domain.OutboxMessage) error {
	var payload map[string]interface{}
	if len(msg.PayloadJSON) > 0 {
		payload = make(map[string]interface{})
		if err := json.Unmarshal(msg.PayloadJSON, &payload); err != nil {
			uc.markFailed(ctx, msg, fmt.Sprintf("unmarshal payload: %v", err))
			return err
		}
	}

	htmlBody, textBody, err := uc.renderer.Render(msg.TemplateName, msg.TemplateVersion, payload)
	if err != nil {
		uc.markFailed(ctx, msg, fmt.Sprintf("render template: %v", err))
		return err
	}

	_ = htmlBody
	_ = textBody

	result, err := uc.provider.Send(ctx, msg)
	if err != nil {
		uc.markFailed(ctx, msg, err.Error())
		return err
	}

	providerMsgID := result.ProviderMessageID
	if err := uc.repo.UpdateStatus(ctx, msg.ID, domain.StatusSent, msg.Attempts+1, nil, nil, &providerMsgID); err != nil {
		slog.ErrorContext(ctx, "email_status_update_failed",
			"email_id", msg.ID,
			"error", err,
		)
		return err
	}

	slog.InfoContext(ctx, "email_sent",
		"email_id", msg.ID,
		"recipient", msg.Recipient,
	)

	return nil
}

func (uc *SendPendingEmails) markFailed(ctx context.Context, msg domain.OutboxMessage, errMsg string) {
	attempts := msg.Attempts + 1
	var nextRetryAt *time.Time

	if attempts >= msg.MaxAttempts {
		if err := uc.repo.UpdateStatus(ctx, msg.ID, domain.StatusDeadLetter, attempts, nil, &errMsg, nil); err != nil {
			slog.ErrorContext(ctx, "email_dead_letter_failed",
				"email_id", msg.ID,
				"error", err,
			)
		}
		slog.WarnContext(ctx, "email_dead_letter",
			"email_id", msg.ID,
			"attempts", attempts,
		)
		return
	}

	backoff := computeBackoff(attempts)
	next := time.Now().Add(time.Duration(backoff) * time.Second)
	nextRetryAt = &next

	if err := uc.repo.UpdateStatus(ctx, msg.ID, domain.StatusFailed, attempts, nextRetryAt, &errMsg, nil); err != nil {
		slog.ErrorContext(ctx, "email_status_update_failed",
			"email_id", msg.ID,
			"error", err,
		)
	}

	slog.WarnContext(ctx, "email_failed",
		"email_id", msg.ID,
		"attempts", attempts,
		"next_retry_at", nextRetryAt,
		"error", errMsg,
	)
}

func computeBackoff(attempts int) int {
	backoffs := []int{5, 30, 120, 600, 3600, 21600}
	if attempts <= 0 {
		return backoffs[0]
	}
	if attempts > len(backoffs) {
		return backoffs[len(backoffs)-1]
	}
	return backoffs[attempts-1]
}
