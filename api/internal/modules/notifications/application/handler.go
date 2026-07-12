package application

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	billingdomain "nursery-management-system/api/internal/modules/billing/domain"
)

// InvoiceIssuedHandler handles InvoiceIssued domain events by sending
// notification emails to parents.
type InvoiceIssuedHandler struct {
	sender InvoiceNotificationSender
}

// NewInvoiceIssuedHandler creates a new handler for InvoiceIssued events.
func NewInvoiceIssuedHandler(sender InvoiceNotificationSender) *InvoiceIssuedHandler {
	return &InvoiceIssuedHandler{sender: sender}
}

// Handle implements events.TypedHandler[billingdomain.InvoiceIssued].
func (h *InvoiceIssuedHandler) Handle(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceIssued) error {
	if err := h.sender.SendInvoiceIssuedEmail(ctx, tx, event.InvoiceID, event.TenantID, event.BranchID); err != nil {
		return fmt.Errorf("send invoice issued email: %w", err)
	}
	return nil
}

// InvoiceOverdueHandler handles InvoiceMarkedOverdue domain events by sending
// notification emails to parents for each overdue invoice.
type InvoiceOverdueHandler struct {
	sender InvoiceNotificationSender
}

// NewInvoiceOverdueHandler creates a new handler for InvoiceMarkedOverdue events.
func NewInvoiceOverdueHandler(sender InvoiceNotificationSender) *InvoiceOverdueHandler {
	return &InvoiceOverdueHandler{sender: sender}
}

// Handle implements events.TypedHandler[billingdomain.InvoiceMarkedOverdue].
// It iterates over transitioned invoices and sends one email per invoice (KTD-6).
func (h *InvoiceOverdueHandler) Handle(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceMarkedOverdue) error {
	for _, inv := range event.Transitioned {
		if err := h.sender.SendInvoiceOverdueEmail(ctx, tx, inv.ID, inv.TenantID, inv.BranchID); err != nil {
			return fmt.Errorf("send invoice overdue email for %s: %w", inv.ID, err)
		}
	}
	return nil
}

// InvoiceDueSoonHandler handles InvoiceDueSoon domain events by sending
// notification emails to parents.
type InvoiceDueSoonHandler struct {
	sender InvoiceNotificationSender
}

// NewInvoiceDueSoonHandler creates a new handler for InvoiceDueSoon events.
func NewInvoiceDueSoonHandler(sender InvoiceNotificationSender) *InvoiceDueSoonHandler {
	return &InvoiceDueSoonHandler{sender: sender}
}

// Handle implements events.TypedHandler[billingdomain.InvoiceDueSoon].
func (h *InvoiceDueSoonHandler) Handle(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceDueSoon) error {
	if err := h.sender.SendInvoiceDueSoonEmail(ctx, tx, event.InvoiceID, event.TenantID, event.BranchID); err != nil {
		return fmt.Errorf("send invoice due soon email: %w", err)
	}
	return nil
}

// InvoiceDueReminderHandler handles InvoiceDueReminder domain events by sending
// notification emails to parents.
type InvoiceDueReminderHandler struct {
	sender InvoiceNotificationSender
}

// NewInvoiceDueReminderHandler creates a new handler for InvoiceDueReminder events.
func NewInvoiceDueReminderHandler(sender InvoiceNotificationSender) *InvoiceDueReminderHandler {
	return &InvoiceDueReminderHandler{sender: sender}
}

// Handle implements events.TypedHandler[billingdomain.InvoiceDueReminder].
func (h *InvoiceDueReminderHandler) Handle(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceDueReminder) error {
	if err := h.sender.SendInvoiceDueReminderEmail(ctx, tx, event.InvoiceID, event.TenantID, event.BranchID); err != nil {
		return fmt.Errorf("send invoice due reminder email: %w", err)
	}
	return nil
}
