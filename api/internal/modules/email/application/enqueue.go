package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
)

type EnqueueEmail struct {
	repo domain.OutboxRepository
}

func NewEnqueueEmail(repo domain.OutboxRepository) *EnqueueEmail {
	return &EnqueueEmail{repo: repo}
}

func enqueueIdempotencyKey(params domain.EnqueueParams) string {
	if params.IdempotencyKey != "" {
		return params.IdempotencyKey
	}
	return fmt.Sprintf("%s_%s_1", params.EventType, params.EntityID)
}

func (uc *EnqueueEmail) Enqueue(ctx context.Context, tenantID, branchID uuid.UUID, params domain.EnqueueParams) (uuid.UUID, error) {
	idempotencyKey := enqueueIdempotencyKey(params)

	var attachmentsJSON []byte
	if len(params.AttachmentRefs) > 0 {
		var err error
		attachmentsJSON, err = json.Marshal(params.AttachmentRefs)
		if err != nil {
			return uuid.Nil, fmt.Errorf("marshal attachment refs: %w", err)
		}
	}

	msg := domain.OutboxMessage{
		ID:              uuid.New(),
		TenantID:        tenantID,
		BranchID:        branchID,
		IdempotencyKey:  idempotencyKey,
		EventType:       params.EventType,
		Recipient:       params.Recipient,
		RecipientName:   params.RecipientName,
		Subject:         params.Subject,
		TemplateName:    params.TemplateName,
		TemplateVersion: params.TemplateVersion,
		PayloadJSON:     params.PayloadJSON,
		Attachments:     params.AttachmentRefs,
		EntityID:        params.EntityID,
		Status:          domain.StatusPending,
		MaxAttempts:     8,
	}
	_ = attachmentsJSON

	inserted, err := uc.repo.Insert(ctx, msg)
	if err != nil {
		slog.ErrorContext(ctx, "email_enqueue_failed",
			"tenant_id", tenantID,
			"branch_id", branchID,
			"event_type", params.EventType,
			"error", err,
		)
		return uuid.Nil, fmt.Errorf("enqueue email: %w", err)
	}

	slog.InfoContext(ctx, "email_enqueued",
		"email_id", inserted.ID,
		"event_type", params.EventType,
		"recipient", params.Recipient,
	)

	return inserted.ID, nil
}

func (uc *EnqueueEmail) EnqueueWithTx(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, params domain.EnqueueParams) (uuid.UUID, error) {
	idempotencyKey := enqueueIdempotencyKey(params)

	msg := domain.OutboxMessage{
		ID:              uuid.New(),
		TenantID:        tenantID,
		BranchID:        branchID,
		IdempotencyKey:  idempotencyKey,
		EventType:       params.EventType,
		Recipient:       params.Recipient,
		RecipientName:   params.RecipientName,
		Subject:         params.Subject,
		TemplateName:    params.TemplateName,
		TemplateVersion: params.TemplateVersion,
		PayloadJSON:     params.PayloadJSON,
		Attachments:     params.AttachmentRefs,
		EntityID:        params.EntityID,
		Status:          domain.StatusPending,
		MaxAttempts:     8,
	}

	inserted, err := uc.repo.InsertTx(ctx, tx, msg)
	if err != nil {
		slog.ErrorContext(ctx, "email_enqueue_failed",
			"tenant_id", tenantID,
			"branch_id", branchID,
			"event_type", params.EventType,
			"error", err,
		)
		return uuid.Nil, fmt.Errorf("enqueue email: %w", err)
	}

	slog.InfoContext(ctx, "email_enqueued",
		"email_id", inserted.ID,
		"event_type", params.EventType,
		"recipient", params.Recipient,
	)

	return inserted.ID, nil
}
