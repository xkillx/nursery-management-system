package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChildDeactivated struct {
	ChildID    uuid.UUID
	ReasonCode string
	Occurred   time.Time
}

func (e ChildDeactivated) OccurredAt() time.Time { return e.Occurred }
