package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/email/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (r *OutboxRepository) Insert(ctx context.Context, msg domain.OutboxMessage) (domain.OutboxMessage, error) {
	q := sqlc.New(r.pool)

	var attachmentsJSON []byte
	if len(msg.Attachments) > 0 {
		var err error
		attachmentsJSON, err = json.Marshal(msg.Attachments)
		if err != nil {
			return domain.OutboxMessage{}, fmt.Errorf("marshal attachments: %w", err)
		}
	}

	var entityID pgtype.Text
	if msg.EntityID != "" {
		entityID = pgtype.Text{String: msg.EntityID, Valid: true}
	}

	result, err := q.InsertEmailOutbox(ctx, sqlc.InsertEmailOutboxParams{
		ID:              pgtype.UUID{Bytes: [16]byte(msg.ID), Valid: true},
		TenantID:        pgtype.UUID{Bytes: [16]byte(msg.TenantID), Valid: true},
		BranchID:        pgtype.UUID{Bytes: [16]byte(msg.BranchID), Valid: true},
		IdempotencyKey:  msg.IdempotencyKey,
		EventType:       msg.EventType,
		Recipient:       msg.Recipient,
		RecipientName:   toPgText(msg.RecipientName),
		Subject:         msg.Subject,
		TemplateName:    msg.TemplateName,
		TemplateVersion: int32(msg.TemplateVersion),
		PayloadJson:     msg.PayloadJSON,
		Attachments:     attachmentsJSON,
		EntityID:        entityID,
		MaxAttempts:     int32(msg.MaxAttempts),
	})
	if err != nil {
		return domain.OutboxMessage{}, fmt.Errorf("insert email outbox: %w", err)
	}

	return convertEmailRow(
		result.ID, result.TenantID, result.BranchID, result.IdempotencyKey, result.EventType,
		result.Recipient, result.RecipientName, result.Subject, result.TemplateName, result.TemplateVersion,
		result.PayloadJson, result.Attachments, result.EntityID, result.Status, result.Attempts,
		result.MaxAttempts, result.NextRetryAt, result.LastError, result.ProviderMessageID,
		result.CreatedAt, result.SentAt, result.UpdatedAt,
	), nil
}

func (r *OutboxRepository) InsertTx(ctx context.Context, tx domain.Tx, msg domain.OutboxMessage) (domain.OutboxMessage, error) {
	q := sqlc.New(tx.(pgx.Tx))

	var attachmentsJSON []byte
	if len(msg.Attachments) > 0 {
		var err error
		attachmentsJSON, err = json.Marshal(msg.Attachments)
		if err != nil {
			return domain.OutboxMessage{}, fmt.Errorf("marshal attachments: %w", err)
		}
	}

	var entityID pgtype.Text
	if msg.EntityID != "" {
		entityID = pgtype.Text{String: msg.EntityID, Valid: true}
	}

	result, err := q.InsertEmailOutbox(ctx, sqlc.InsertEmailOutboxParams{
		ID:              pgtype.UUID{Bytes: [16]byte(msg.ID), Valid: true},
		TenantID:        pgtype.UUID{Bytes: [16]byte(msg.TenantID), Valid: true},
		BranchID:        pgtype.UUID{Bytes: [16]byte(msg.BranchID), Valid: true},
		IdempotencyKey:  msg.IdempotencyKey,
		EventType:       msg.EventType,
		Recipient:       msg.Recipient,
		RecipientName:   toPgText(msg.RecipientName),
		Subject:         msg.Subject,
		TemplateName:    msg.TemplateName,
		TemplateVersion: int32(msg.TemplateVersion),
		PayloadJson:     msg.PayloadJSON,
		Attachments:     attachmentsJSON,
		EntityID:        entityID,
		MaxAttempts:     int32(msg.MaxAttempts),
	})
	if err != nil {
		return domain.OutboxMessage{}, fmt.Errorf("insert email outbox: %w", err)
	}

	return convertEmailRow(
		result.ID, result.TenantID, result.BranchID, result.IdempotencyKey, result.EventType,
		result.Recipient, result.RecipientName, result.Subject, result.TemplateName, result.TemplateVersion,
		result.PayloadJson, result.Attachments, result.EntityID, result.Status, result.Attempts,
		result.MaxAttempts, result.NextRetryAt, result.LastError, result.ProviderMessageID,
		result.CreatedAt, result.SentAt, result.UpdatedAt,
	), nil
}

func (r *OutboxRepository) GetPending(ctx context.Context, batchSize int) ([]domain.OutboxMessage, error) {
	q := sqlc.New(r.pool)

	rows, err := q.GetPendingEmails(ctx, int32(batchSize))
	if err != nil {
		return nil, fmt.Errorf("get pending emails: %w", err)
	}

	result := make([]domain.OutboxMessage, len(rows))
	for i, row := range rows {
		result[i] = convertEmailRow(
			row.ID, row.TenantID, row.BranchID, row.IdempotencyKey, row.EventType,
			row.Recipient, row.RecipientName, row.Subject, row.TemplateName, row.TemplateVersion,
			row.PayloadJson, row.Attachments, row.EntityID, row.Status, row.Attempts,
			row.MaxAttempts, row.NextRetryAt, row.LastError, row.ProviderMessageID,
			row.CreatedAt, row.SentAt, row.UpdatedAt,
		)
	}
	return result, nil
}

func (r *OutboxRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, attempts int, nextRetryAt interface{}, lastError *string, providerMessageID *string) error {
	q := sqlc.New(r.pool)

	var nextRetry pgtype.Timestamptz
	if nextRetryAt != nil {
		if t, ok := nextRetryAt.(*time.Time); ok && t != nil {
			nextRetry = pgtype.Timestamptz{Time: *t, Valid: true}
		}
	}
	if !nextRetry.Valid {
		nextRetry = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	return q.UpdateEmailStatus(ctx, sqlc.UpdateEmailStatusParams{
		ID:                pgtype.UUID{Bytes: [16]byte(id), Valid: true},
		Status:            string(status),
		Attempts:          int32(attempts),
		NextRetryAt:       nextRetry,
		LastError:         toPgText(lastError),
		ProviderMessageID: toPgText(providerMessageID),
	})
}

func (r *OutboxRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.OutboxMessage, error) {
	q := sqlc.New(r.pool)

	row, err := q.GetEmailByID(ctx, sqlc.GetEmailByIDParams{
		TenantID: pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID: pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
		ID:       pgtype.UUID{Bytes: [16]byte(id), Valid: true},
	})
	if err != nil {
		return domain.OutboxMessage{}, fmt.Errorf("get email by id: %w", err)
	}

	return convertEmailRow(
		row.ID, row.TenantID, row.BranchID, row.IdempotencyKey, row.EventType,
		row.Recipient, row.RecipientName, row.Subject, row.TemplateName, row.TemplateVersion,
		row.PayloadJson, row.Attachments, row.EntityID, row.Status, row.Attempts,
		row.MaxAttempts, row.NextRetryAt, row.LastError, row.ProviderMessageID,
		row.CreatedAt, row.SentAt, row.UpdatedAt,
	), nil
}

func (r *OutboxRepository) List(ctx context.Context, tenantID, branchID uuid.UUID, status *string, limit, offset int) ([]domain.OutboxMessage, int, error) {
	q := sqlc.New(r.pool)

	var statusPg pgtype.Text
	if status != nil {
		statusPg = pgtype.Text{String: *status, Valid: true}
	}

	rows, err := q.ListEmails(ctx, sqlc.ListEmailsParams{
		TenantID: pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID: pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
		Status:   statusPg,
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list emails: %w", err)
	}

	count, err := q.CountEmails(ctx, sqlc.CountEmailsParams{
		TenantID: pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID: pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
		Status:   statusPg,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count emails: %w", err)
	}

	result := make([]domain.OutboxMessage, len(rows))
	for i, row := range rows {
		result[i] = convertEmailRow(
			row.ID, row.TenantID, row.BranchID, row.IdempotencyKey, row.EventType,
			row.Recipient, row.RecipientName, row.Subject, row.TemplateName, row.TemplateVersion,
			row.PayloadJson, row.Attachments, row.EntityID, row.Status, row.Attempts,
			row.MaxAttempts, row.NextRetryAt, row.LastError, row.ProviderMessageID,
			row.CreatedAt, row.SentAt, row.UpdatedAt,
		)
	}
	return result, int(count), nil
}

func (r *OutboxRepository) GetStats(ctx context.Context, tenantID, branchID uuid.UUID) (map[string]int, error) {
	q := sqlc.New(r.pool)

	rows, err := q.GetEmailStats(ctx, sqlc.GetEmailStatsParams{
		TenantID: pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID: pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("get email stats: %w", err)
	}

	stats := make(map[string]int)
	for _, row := range rows {
		stats[row.Status] = int(row.Count)
	}
	return stats, nil
}

func (r *OutboxRepository) InsertDelivery(ctx context.Context, record domain.DeliveryRecord) error {
	q := sqlc.New(r.pool)

	return q.InsertDeliveryEvent(ctx, sqlc.InsertDeliveryEventParams{
		ID:                pgtype.UUID{Bytes: [16]byte(record.ID), Valid: true},
		ProviderMessageID: record.ProviderMessageID,
		EmailOutboxID:     pgtype.UUID{Bytes: [16]byte(record.EmailOutboxID), Valid: true},
		Status:            record.Status,
		ResponseJson:      record.ResponseJSON,
	})
}

func (r *OutboxRepository) GetDeliveryByProviderMessageID(ctx context.Context, providerMessageID string) ([]domain.DeliveryRecord, error) {
	q := sqlc.New(r.pool)

	rows, err := q.GetDeliveryByProviderMessageID(ctx, providerMessageID)
	if err != nil {
		return nil, fmt.Errorf("get delivery by provider message id: %w", err)
	}

	result := make([]domain.DeliveryRecord, len(rows))
	for i, row := range rows {
		result[i] = domain.DeliveryRecord{
			ID:                uuid.UUID(row.ID.Bytes),
			ProviderMessageID: row.ProviderMessageID,
			EmailOutboxID:     uuid.UUID(row.EmailOutboxID.Bytes),
			Status:            row.Status,
			ResponseJSON:      row.ResponseJson,
			CreatedAt:         row.CreatedAt.Time,
		}
	}
	return result, nil
}

func (r *OutboxRepository) ResetToPending(ctx context.Context, id uuid.UUID) error {
	q := sqlc.New(r.pool)

	return q.ResetEmailToPending(ctx, pgtype.UUID{Bytes: [16]byte(id), Valid: true})
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toDomainMessage(row sqlc.EmailOutbox) domain.OutboxMessage {
	return convertEmailRow(
		row.ID, row.TenantID, row.BranchID, row.IdempotencyKey, row.EventType,
		row.Recipient, row.RecipientName, row.Subject, row.TemplateName, row.TemplateVersion,
		row.PayloadJson, row.Attachments, row.EntityID, row.Status, row.Attempts,
		row.MaxAttempts, row.NextRetryAt, row.LastError, row.ProviderMessageID,
		row.CreatedAt, row.SentAt, row.UpdatedAt,
	)
}

func convertEmailRow(
	id pgtype.UUID, tenantID pgtype.UUID, branchID pgtype.UUID,
	idempotencyKey string, eventType string, recipient string,
	recipientName pgtype.Text, subject string, templateName string,
	templateVersion int32, payloadJson []byte, attachments []byte,
	entityID pgtype.Text, status string, attempts int32, maxAttempts int32,
	nextRetryAt pgtype.Timestamptz, lastError pgtype.Text,
	providerMessageID pgtype.Text, createdAt pgtype.Timestamptz,
	sentAt pgtype.Timestamptz, updatedAt pgtype.Timestamptz,
) domain.OutboxMessage {
	var recipientNamePtr *string
	if recipientName.Valid {
		recipientNamePtr = &recipientName.String
	}

	var lastErrorPtr *string
	if lastError.Valid {
		lastErrorPtr = &lastError.String
	}

	var providerMessageIDPtr *string
	if providerMessageID.Valid {
		providerMessageIDPtr = &providerMessageID.String
	}

	var sentAtPtr *time.Time
	if sentAt.Valid {
		sentAtPtr = &sentAt.Time
	}

	var attRefs []domain.AttachmentRef
	if len(attachments) > 0 {
		_ = json.Unmarshal(attachments, &attRefs)
	}

	var eid string
	if entityID.Valid {
		eid = entityID.String
	}

	return domain.OutboxMessage{
		ID:                uuid.UUID(id.Bytes),
		TenantID:          uuid.UUID(tenantID.Bytes),
		BranchID:          uuid.UUID(branchID.Bytes),
		IdempotencyKey:    idempotencyKey,
		EventType:         eventType,
		Recipient:         recipient,
		RecipientName:     recipientNamePtr,
		Subject:           subject,
		TemplateName:      templateName,
		TemplateVersion:   int(templateVersion),
		PayloadJSON:       payloadJson,
		Attachments:       attRefs,
		EntityID:          eid,
		Status:            domain.Status(status),
		Attempts:          int(attempts),
		MaxAttempts:       int(maxAttempts),
		NextRetryAt:       nextRetryAt.Time,
		LastError:         lastErrorPtr,
		ProviderMessageID: providerMessageIDPtr,
		CreatedAt:         createdAt.Time,
		SentAt:            sentAtPtr,
		UpdatedAt:         updatedAt.Time,
	}
}
