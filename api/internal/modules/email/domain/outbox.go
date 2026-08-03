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
	ID              uuid.UUID
	TenantID        uuid.UUID
	BranchID        uuid.UUID
	IdempotencyKey  string
	EventType       string
	Recipient       string
	RecipientName   *string
	Subject         string
	TemplateName    string
	TemplateVersion int
	PayloadJSON     []byte
	Attachments     []AttachmentRef
	// RenderedHTML and RenderedText carry the rendered template output through
	// the send step only. They are never persisted: a DB round-trip leaves them
	// empty, so they are populated in memory right before delivery (KTD-1).
	RenderedHTML      string
	RenderedText      string
	EntityID          string
	Status            Status
	Attempts          int
	MaxAttempts       int
	NextRetryAt       time.Time
	LastError         *string
	ProviderMessageID *string
	CreatedAt         time.Time
	SentAt            *time.Time
	UpdatedAt         time.Time
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type AttachmentRef struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	S3Key       string `json:"s3_key"`
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
	AttachmentRefs  []AttachmentRef
}

type EmailEnqueuer interface {
	Enqueue(ctx context.Context, tenantID, branchID uuid.UUID, params EnqueueParams) (uuid.UUID, error)
	EnqueueWithTx(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, params EnqueueParams) (uuid.UUID, error)
}
