package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ParentInvoiceFilters holds optional list filters for parent invoice view.
type ParentInvoiceFilters struct {
	BillingMonth *time.Time
	Status       *string
	ChildID      *uuid.UUID
	Limit        int
	Offset       int
}

// ParentInvoiceRow maps a joined invoice + child row for parent view.
// Excludes manager-only fields: generated run, lock, adjustment, created/updated.
type ParentInvoiceRow struct {
	ID                     uuid.UUID
	InvoiceKind            string
	InvoiceNumber          *string
	Status                 string
	ChildID                uuid.UUID
	ChildFirstName         string
	ChildMiddleName        *string
	ChildLastName          *string
	BillingMonth           time.Time
	PeriodStartDate        time.Time
	PeriodEndDate          time.Time
	CurrencyCode           string
	Subtotal               Money
	FundedDeduction        Money
	TotalDue               Money
	AmountPaid             Money
	DueAt                  *time.Time
	IssuedAt               *time.Time
	PaidAt                 *time.Time
	PaymentFailedAt        *time.Time
	PaymentStatusUpdatedAt *time.Time
	CalculationDetails     json.RawMessage
}

// ParentInvoiceLineRow maps an invoice line for parent view.
// Excludes line ID and line-level calculation snapshots.
type ParentInvoiceLineRow struct {
	LineKind        string
	Description     string
	SortOrder       int
	QuantityMinutes *int
	UnitAmount      *Money
	LineAmount      Money
}

// ParentInvoiceDetail is the full detail response for a parent invoice.
type ParentInvoiceDetail struct {
	Invoice     ParentInvoiceRow
	Lines       []ParentInvoiceLineRow
	Calculation InvoiceReviewCalculation
}
