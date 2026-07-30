package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
)

type RetryEmail struct {
	repo domain.OutboxRepository
}

func NewRetryEmail(repo domain.OutboxRepository) *RetryEmail {
	return &RetryEmail{repo: repo}
}

func (uc *RetryEmail) Execute(ctx context.Context, id uuid.UUID) error {
	if err := uc.repo.ResetToPending(ctx, id); err != nil {
		slog.ErrorContext(ctx, "email_retry_failed",
			"email_id", id,
			"error", err,
		)
		return fmt.Errorf("retry email: %w", err)
	}

	slog.InfoContext(ctx, "email_retried",
		"email_id", id,
	)

	return nil
}
