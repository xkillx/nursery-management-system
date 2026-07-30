package domain

import (
	"context"

	"github.com/google/uuid"
)

type OutboxRepository interface {
	Insert(ctx context.Context, msg OutboxMessage) (OutboxMessage, error)
	InsertTx(ctx context.Context, tx Tx, msg OutboxMessage) (OutboxMessage, error)
	GetPending(ctx context.Context, batchSize int) ([]OutboxMessage, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status, attempts int, nextRetryAt interface{}, lastError *string, providerMessageID *string) error
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (OutboxMessage, error)
	List(ctx context.Context, tenantID, branchID uuid.UUID, status *string, limit, offset int) ([]OutboxMessage, int, error)
	GetStats(ctx context.Context, tenantID, branchID uuid.UUID) (map[string]int, error)
	InsertDelivery(ctx context.Context, record DeliveryRecord) error
	GetDeliveryByProviderMessageID(ctx context.Context, providerMessageID string) ([]DeliveryRecord, error)
	ResetToPending(ctx context.Context, id uuid.UUID) error
}
