package domain

import (
	"time"

	"github.com/google/uuid"
)

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
	Enqueue(params EnqueueParams) OutboxMessage
}
