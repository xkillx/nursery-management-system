package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Tx = pgx.Tx

type BillingRepository interface {
	// Preflight (API-16) — read-only, no transaction required.
	ListPreflightChildren(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth, nextBillingMonth time.Time) ([]PreflightChildRow, error)
	ListPreflightAttendanceSessions(ctx context.Context, tenantID, branchID uuid.UUID, periodStartLocalDate, periodEndExclusiveLocalDate time.Time) ([]PreflightAttendanceSessionRow, error)

	// Generation (API-17) — transactional methods using Tx.
	ListCandidateChildrenForUpdate(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, billingMonth, nextBillingMonth time.Time) ([]PreflightChildRow, error)
	ListSelectedChildrenForUpdate(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, childIDs []uuid.UUID) ([]PreflightChildRow, error)
	ListAttendanceSessions(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, periodStart, periodEndExclusive time.Time) ([]PreflightAttendanceSessionRow, error)
	CreateInvoiceRun(ctx context.Context, tx Tx, params InvoiceRunCreateParams) error
	CompleteInvoiceRun(ctx context.Context, tx Tx, params InvoiceRunCompleteParams) error
	GetMonthlyInvoiceForUpdate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (InvoiceRow, bool, error)
	CreateDraftInvoice(ctx context.Context, tx Tx, params DraftInvoiceCreateParams) error
	UpdateDraftInvoice(ctx context.Context, tx Tx, params DraftInvoiceUpdateParams) error
	DeleteDraftSystemInvoiceLines(ctx context.Context, tx Tx, tenantID, branchID, invoiceID uuid.UUID) error
	ListDraftExtraLines(ctx context.Context, tx Tx, tenantID, branchID, invoiceID uuid.UUID) ([]ExtraLineRow, error)
	InsertInvoiceLine(ctx context.Context, tx Tx, params InvoiceLineCreateParams) error

	// Manager Invoice Review (API-18) — read-only, no transaction required.
	ListInvoicesForManagerReview(ctx context.Context, tenantID, branchID uuid.UUID, filters InvoiceReviewFilters) ([]InvoiceReviewRow, error)
	GetInvoiceForManagerReview(ctx context.Context, tenantID, branchID, invoiceID uuid.UUID) (InvoiceReviewRow, bool, error)
	ListInvoiceLinesForManagerReview(ctx context.Context, tenantID, branchID, invoiceID uuid.UUID) ([]InvoiceReviewLineRow, error)

	// Invoice Issue (API-19) — transactional methods using Tx.
	GetInvoiceForIssueForUpdate(ctx context.Context, tx Tx, tenantID, branchID, invoiceID uuid.UUID) (InvoiceIssueCandidateRow, bool, error)
	ListDraftInvoicesForIssueForUpdate(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]InvoiceIssueCandidateRow, error)
	ListSelectedInvoicesForIssueForUpdate(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, invoiceIDs []uuid.UUID) ([]InvoiceIssueCandidateRow, error)
	AllocateInvoiceNumberSequence(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, year, month int) (int, error)
	MarkInvoiceIssued(ctx context.Context, tx Tx, params IssueInvoiceUpdateParams) error
}

// InvoiceRow maps a row from the invoices table.
type InvoiceRow struct {
	ID                   uuid.UUID
	Status               string
	InvoiceKind          string
	SubtotalMinor        int
	FundedDeductionMinor int
	TotalDueMinor        int
	CalculationDetails   json.RawMessage
}

// ExtraLineRow maps an extra invoice line row.
type ExtraLineRow struct {
	ID              uuid.UUID
	LineKind        string
	LineAmountMinor int
	Details         json.RawMessage
}
