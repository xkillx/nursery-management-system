package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceIssued struct {
	InvoiceID uuid.UUID
	Occurred  time.Time
}

func (e InvoiceIssued) OccurredAt() time.Time { return e.Occurred }

type InvoiceMarkedOverdue struct {
	Transitioned []OverdueTransitionedInvoice
	Occurred     time.Time
}

func (e InvoiceMarkedOverdue) OccurredAt() time.Time { return e.Occurred }
