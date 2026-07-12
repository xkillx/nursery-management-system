package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InvoiceNotificationSender defines the port for sending invoice notification emails.
// The interface uses primitive types (uuid, pgx.Tx) so the notifications module
// never imports the billing module directly.
type InvoiceNotificationSender interface {
	SendInvoiceIssuedEmail(ctx context.Context, tx pgx.Tx, invoiceID, tenantID, branchID uuid.UUID) error
	SendInvoiceOverdueEmail(ctx context.Context, tx pgx.Tx, invoiceID, tenantID, branchID uuid.UUID) error
}

// Audit action types for notification events.
const (
	AuditNotificationInvoiceIssuedSent    = "notification_invoice_issued_sent"
	AuditNotificationInvoiceOverdueSent   = "notification_invoice_overdue_sent"
	AuditNotificationInvoiceIssuedFailed  = "notification_invoice_issued_failed"
	AuditNotificationInvoiceOverdueFailed = "notification_invoice_overdue_failed"
)
