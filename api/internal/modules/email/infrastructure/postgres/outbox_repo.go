package postgres

import (
	"context"
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
		MaxAttempts:     int32(msg.MaxAttempts),
	})
	if err != nil {
		return domain.OutboxMessage{}, fmt.Errorf("insert email outbox: %w", err)
	}

	return toDomainMessage(result), nil
}

func (r *OutboxRepository) InsertTx(ctx context.Context, tx domain.Tx, msg domain.OutboxMessage) (domain.OutboxMessage, error) {
	q := sqlc.New(tx.(pgx.Tx))

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
		MaxAttempts:     int32(msg.MaxAttempts),
	})
	if err != nil {
		return domain.OutboxMessage{}, fmt.Errorf("insert email outbox: %w", err)
	}

	return toDomainMessage(result), nil
}

func (r *OutboxRepository) GetPending(ctx context.Context, batchSize int) ([]domain.OutboxMessage, error) {
	q := sqlc.New(r.pool)

	rows, err := q.GetPendingEmails(ctx, int32(batchSize))
	if err != nil {
		return nil, fmt.Errorf("get pending emails: %w", err)
	}

	result := make([]domain.OutboxMessage, len(rows))
	for i, row := range rows {
		result[i] = toDomainMessage(row)
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

	return toDomainMessage(row), nil
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
		result[i] = toDomainMessage(row)
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
	var recipientName *string
	if row.RecipientName.Valid {
		recipientName = &row.RecipientName.String
	}

	var lastError *string
	if row.LastError.Valid {
		lastError = &row.LastError.String
	}

	var providerMessageID *string
	if row.ProviderMessageID.Valid {
		providerMessageID = &row.ProviderMessageID.String
	}

	var sentAt *time.Time
	if row.SentAt.Valid {
		sentAt = &row.SentAt.Time
	}

	return domain.OutboxMessage{
		ID:                uuid.UUID(row.ID.Bytes),
		TenantID:          uuid.UUID(row.TenantID.Bytes),
		BranchID:          uuid.UUID(row.BranchID.Bytes),
		IdempotencyKey:    row.IdempotencyKey,
		EventType:         row.EventType,
		Recipient:         row.Recipient,
		RecipientName:     recipientName,
		Subject:           row.Subject,
		TemplateName:      row.TemplateName,
		TemplateVersion:   int(row.TemplateVersion),
		PayloadJSON:       row.PayloadJson,
		Status:            domain.Status(row.Status),
		Attempts:          int(row.Attempts),
		MaxAttempts:       int(row.MaxAttempts),
		NextRetryAt:       row.NextRetryAt.Time,
		LastError:         lastError,
		ProviderMessageID: providerMessageID,
		CreatedAt:         row.CreatedAt.Time,
		SentAt:            sentAt,
	}
}
