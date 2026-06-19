package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

// nowFn is the clock used to stamp override_at. Tests can replace this.
var nowFn = func() time.Time { return time.Now() }

// OverrideAttendanceBlockUseCase records a manager's decision to clear the
// attendance block for a child + billing month, even when the invoice is
// overdue. The override is auditable and is exposed via the
// `POST /api/v1/billing/invoices/:id/override-attendance-block` route.
//
// The data model stores the override on a per-child-per-billing-month record.
// Under advance-pay the "billing month" is the month the invoice covers; under
// the 8th-of-month grace rule, the override is meaningful for any month where
// the invoice is issued and past grace.
//
// For Phase 3, the override is stored as an audit event only; the daily-list
// surface (practitioner) and the parent invoice view will read this audit log
// in later phases. The override is intended to apply for one billing month
// only; subsequent months require a fresh override if the situation recurs.
type OverrideAttendanceBlockUseCase struct {
	repo   domain.BillingRepository
	auditW *audit.Writer
	txMgr  *transaction.Manager
}

func NewOverrideAttendanceBlockUseCase(
	repo domain.BillingRepository,
	auditW *audit.Writer,
	txMgr *transaction.Manager,
) *OverrideAttendanceBlockUseCase {
	return &OverrideAttendanceBlockUseCase{repo: repo, auditW: auditW, txMgr: txMgr}
}

type OverrideAttendanceBlockInput struct {
	InvoiceID    uuid.UUID
	BillingMonth string
	Note         string
}

type OverrideAttendanceBlockResult struct {
	InvoiceID    uuid.UUID
	BillingMonth string
	OverriddenBy uuid.UUID
	OverriddenAt time.Time
}

// Note: time import referenced for the result type; ensure it's imported.
var _ = fmt.Sprintf

func (uc *OverrideAttendanceBlockUseCase) Execute(ctx context.Context, actor tenant.ActorContext, in OverrideAttendanceBlockInput) (*OverrideAttendanceBlockResult, error) {
	if in.InvoiceID == uuid.Nil {
		return nil, domainerrors.Validation("Invalid request payload.", "invoice_id")
	}
	if _, err := uuid.Parse(in.InvoiceID.String()); err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "invoice_id")
	}
	if in.BillingMonth == "" {
		return nil, domainerrors.Validation("Invalid request payload.", "billing_month")
	}

	var result OverrideAttendanceBlockResult
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		// 1. Verify the invoice exists and is in scope.
		row, found, lerr := uc.repo.GetInvoiceForIssueForUpdate(ctx, tx, actor.TenantID, actor.BranchID, in.InvoiceID)
		if lerr != nil {
			return fmt.Errorf("lookup invoice: %w", lerr)
		}
		if !found {
			return domainerrors.NotFound("invoice", "Resource not found.")
		}
		_ = row

		// 2. Audit the override.
		overriddenAt := nowFn().UTC()
		if err := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "invoice_attendance_block_overridden",
			EntityType: domain.AuditEntityInvoice,
			EntityID:   in.InvoiceID,
			Details: map[string]any{
				"billing_month": in.BillingMonth,
				"override_note": in.Note,
				"overridden_at": overriddenAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		}); err != nil {
			return fmt.Errorf("audit attendance block override: %w", err)
		}

		result = OverrideAttendanceBlockResult{
			InvoiceID:    in.InvoiceID,
			BillingMonth: in.BillingMonth,
			OverriddenBy: actor.MembershipID,
			OverriddenAt: overriddenAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}
