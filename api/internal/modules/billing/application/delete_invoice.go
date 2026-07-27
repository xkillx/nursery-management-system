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
	"nursery-management-system/api/internal/platform/events"
	"nursery-management-system/api/internal/platform/tenant"
)

type DeleteInvoice struct {
	repo       domain.BillingRepository
	auditW     *audit.Writer
	dispatcher *events.EventDispatcher
}

func NewDeleteInvoice(
	repo domain.BillingRepository,
	auditW *audit.Writer,
	dispatcher *events.EventDispatcher,
) *DeleteInvoice {
	return &DeleteInvoice{repo: repo, auditW: auditW, dispatcher: dispatcher}
}

type DeleteInvoiceResult struct {
	InvoiceID uuid.UUID
	Status    string
	DeletedAt time.Time
}

func (uc *DeleteInvoice) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string) (DeleteInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return DeleteInvoiceResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	var result DeleteInvoiceResult

	txErr := uc.dispatcher.DispatchInTx(ctx, func(tx pgx.Tx, emitter events.Emitter) error {
		deletedAt := time.Now().UTC()

		n, delErr := uc.repo.DeleteInvoice(ctx, tx, actor.TenantID, actor.BranchID, invoiceID)
		if delErr != nil {
			return fmt.Errorf("delete invoice: %w", delErr)
		}
		if n == 0 {
			// Disambiguate: invoice not found vs ineligible status.
			_, found, getErr := uc.repo.GetInvoiceForManagerReview(ctx, actor.TenantID, actor.BranchID, invoiceID)
			if getErr != nil {
				return fmt.Errorf("check invoice existence: %w", getErr)
			}
			if !found {
				return domainerrors.NotFound("invoice", "Invoice not found.")
			}
			return domainerrors.Conflict("invoice_not_deletable", "Only draft or void invoices can be deleted.")
		}

		if uc.auditW != nil {
			if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: domain.AuditInvoiceDeleted,
				EntityType: domain.AuditEntityInvoice,
				EntityID:   invoiceID,
				Details:    nil,
			}); auditErr != nil {
				return fmt.Errorf("write audit: %w", auditErr)
			}
		}

		emitter.Emit(domain.InvoiceDeleted{
			InvoiceID: invoiceID,
			TenantID:  actor.TenantID,
			BranchID:  actor.BranchID,
			Occurred:  deletedAt,
		})

		result = DeleteInvoiceResult{
			InvoiceID: invoiceID,
			Status:    "deleted",
			DeletedAt: deletedAt,
		}
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return DeleteInvoiceResult{}, txErr
		}
		return DeleteInvoiceResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

// --- BulkDeleteInvoices ---

type BulkDeleteInvoices struct {
	repo       domain.BillingRepository
	auditW     *audit.Writer
	dispatcher *events.EventDispatcher
}

func NewBulkDeleteInvoices(
	repo domain.BillingRepository,
	auditW *audit.Writer,
	dispatcher *events.EventDispatcher,
) *BulkDeleteInvoices {
	return &BulkDeleteInvoices{repo: repo, auditW: auditW, dispatcher: dispatcher}
}

type BulkDeleteInvoicesResult struct {
	Deleted []DeleteInvoiceResult
	Blocked []BulkDeleteBlockedInvoice
}

type BulkDeleteBlockedInvoice struct {
	InvoiceID uuid.UUID
	ErrorCode string
	Message   string
}

func (uc *BulkDeleteInvoices) Execute(ctx context.Context, actor tenant.ActorContext, rawInvoiceIDs []string) (BulkDeleteInvoicesResult, error) {
	if len(rawInvoiceIDs) == 0 {
		return BulkDeleteInvoicesResult{}, domainerrors.Validation("Invoice IDs list must not be empty.", "invoice_ids")
	}

	// Parse and deduplicate IDs.
	seen := make(map[uuid.UUID]bool)
	invoiceIDs := make([]uuid.UUID, 0, len(rawInvoiceIDs))
	for _, raw := range rawInvoiceIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return BulkDeleteInvoicesResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_ids")
		}
		if !seen[id] {
			seen[id] = true
			invoiceIDs = append(invoiceIDs, id)
		}
	}

	var result BulkDeleteInvoicesResult

	txErr := uc.dispatcher.DispatchInTx(ctx, func(tx pgx.Tx, emitter events.Emitter) error {
		deletedAt := time.Now().UTC()

		for _, invoiceID := range invoiceIDs {
			n, delErr := uc.repo.DeleteInvoice(ctx, tx, actor.TenantID, actor.BranchID, invoiceID)
			if delErr != nil {
				return fmt.Errorf("delete invoice %s: %w", invoiceID, delErr)
			}
			if n == 0 {
				// Disambiguate: invoice not found vs ineligible status.
				_, found, getErr := uc.repo.GetInvoiceForManagerReview(ctx, actor.TenantID, actor.BranchID, invoiceID)
				if getErr != nil {
					return fmt.Errorf("check invoice existence: %w", getErr)
				}
				if !found {
					result.Blocked = append(result.Blocked, BulkDeleteBlockedInvoice{
						InvoiceID: invoiceID,
						ErrorCode: "invoice_not_found",
						Message:   "Invoice not found.",
					})
				} else {
					result.Blocked = append(result.Blocked, BulkDeleteBlockedInvoice{
						InvoiceID: invoiceID,
						ErrorCode: "invoice_not_deletable",
						Message:   "Only draft or void invoices can be deleted.",
					})
				}
				continue
			}

			if uc.auditW != nil {
				if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
					ActionType: domain.AuditInvoiceDeleted,
					EntityType: domain.AuditEntityInvoice,
					EntityID:   invoiceID,
					Details:    nil,
				}); auditErr != nil {
					return fmt.Errorf("write audit for %s: %w", invoiceID, auditErr)
				}
			}

			emitter.Emit(domain.InvoiceDeleted{
				InvoiceID: invoiceID,
				TenantID:  actor.TenantID,
				BranchID:  actor.BranchID,
				Occurred:  deletedAt,
			})

			result.Deleted = append(result.Deleted, DeleteInvoiceResult{
				InvoiceID: invoiceID,
				Status:    "deleted",
				DeletedAt: deletedAt,
			})
		}
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return BulkDeleteInvoicesResult{}, txErr
		}
		return BulkDeleteInvoicesResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}
