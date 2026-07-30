package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryRecord struct {
	ID                uuid.UUID
	ProviderMessageID string
	EmailOutboxID     uuid.UUID
	Status            string
	ResponseJSON      []byte
	CreatedAt         time.Time
}
