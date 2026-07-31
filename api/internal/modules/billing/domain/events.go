package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceIssued struct {
	InvoiceID       uuid.UUID
	TenantID        uuid.UUID
	BranchID        uuid.UUID
	UserID          uuid.UUID
	MembershipID    uuid.UUID
	CheckoutURL     string
	AttachmentS3Key string
	Occurred        time.Time
}

func (e InvoiceIssued) OccurredAt() time.Time { return e.Occurred }

type InvoiceMarkedOverdue struct {
	Transitioned []OverdueTransitionedInvoice
	Occurred     time.Time
}

func (e InvoiceMarkedOverdue) OccurredAt() time.Time { return e.Occurred }

type InvoiceVoided struct {
	InvoiceID uuid.UUID
	Reason    string
	Occurred  time.Time
}

func (e InvoiceVoided) OccurredAt() time.Time { return e.Occurred }

type PaymentCompleted struct {
	InvoiceID   uuid.UUID
	TenantID    uuid.UUID
	BranchID    uuid.UUID
	AmountPaid  int
	PaymentDate time.Time
	Occurred    time.Time
}

func (e PaymentCompleted) OccurredAt() time.Time { return e.Occurred }

type InvoiceDeleted struct {
	InvoiceID uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	Occurred  time.Time
}

func (e InvoiceDeleted) OccurredAt() time.Time { return e.Occurred }

type InvoiceDueSoon struct {
	InvoiceID uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	DueDate   time.Time
	Occurred  time.Time
}

func (e InvoiceDueSoon) OccurredAt() time.Time { return e.Occurred }

type InvoiceDueReminder struct {
	InvoiceID       uuid.UUID
	TenantID        uuid.UUID
	BranchID        uuid.UUID
	DueDate         time.Time
	DaysBefore      int
	CheckoutURL     string
	AttachmentS3Key string
	Occurred        time.Time
}

func (e InvoiceDueReminder) OccurredAt() time.Time { return e.Occurred }
