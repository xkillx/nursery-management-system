package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Tx is an opaque transaction handle. The infrastructure layer casts it
// to the concrete driver type (pgx.Tx). Defined as a type alias for any
// so the domain layer never imports a driver package.
type Tx = any

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusSent       Status = "sent"
	StatusFailed     Status = "failed"
	StatusDeadLetter Status = "dead_letter"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusSent, StatusFailed, StatusDeadLetter:
		return true
	}
	return false
}

type OutboxMessage struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	BranchID          uuid.UUID
	IdempotencyKey    string
	EventType         string
	Recipient         string
	RecipientName     *string
	Subject           string
	TemplateName      string
	TemplateVersion   int
	PayloadJSON       []byte
	Status            Status
	Attempts          int
	MaxAttempts       int
	NextRetryAt       time.Time
	LastError         *string
	ProviderMessageID *string
	CreatedAt         time.Time
	SentAt            *time.Time
}

type EnqueueParams struct {
	EventType       string
	Recipient       string
	RecipientName   *string
	Subject         string
	TemplateName    string
	TemplateVersion int
	PayloadJSON     []byte
	EntityID        string
}

type EmailEnqueuer interface {
	Enqueue(ctx context.Context, tenantID, branchID uuid.UUID, params EnqueueParams) (uuid.UUID, error)
	EnqueueWithTx(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, params EnqueueParams) (uuid.UUID, error)
}
