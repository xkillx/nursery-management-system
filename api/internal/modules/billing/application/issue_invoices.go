package application

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/events"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type IssueInvoice struct {
	repo       domain.BillingRepository
	txMgr      *transaction.Manager
	auditW     *audit.Writer
	dispatcher *events.EventDispatcher
}

func NewIssueInvoice(
	repo domain.BillingRepository,
	txMgr *transaction.Manager,
	auditW *audit.Writer,
	dispatcher *events.EventDispatcher,
) *IssueInvoice {
	return &IssueInvoice{repo: repo, txMgr: txMgr, auditW: auditW, dispatcher: dispatcher}
}

func (uc *IssueInvoice) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string, confirm bool) (domain.IssueInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return domain.IssueInvoiceResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	if !confirm {
		return domain.IssueInvoiceResult{}, domainerrors.Validation("Confirmation required.", "confirm")
	}

	runID := uid.NewUUID()

	var result domain.IssueInvoiceResult

	txErr := uc.dispatcher.DispatchInTx(ctx, func(tx pgx.Tx, emitter events.Emitter) error {
		candidate, found, lockErr := uc.repo.GetInvoiceForIssueForUpdate(ctx, tx, actor.TenantID, actor.BranchID, invoiceID)
		if lockErr != nil {
			return fmt.Errorf("lock invoice for issue: %w", lockErr)
		}
		if !found {
			return domainerrors.NotFound("invoice", "Invoice not found.")
		}

		if candidate.InvoiceKind != domain.InvoiceKindMonthly {
			return domainerrors.Conflict(domain.IssueBlockerInvoiceNotMonthly, "Invoice is not a monthly invoice.")
		}
		if candidate.Status != domain.InvoiceStatusDraft {
			return domainerrors.Conflict(domain.IssueBlockerInvoiceNotDraft, "Invoice is not a draft.")
		}

		if runErr := uc.repo.CreateInvoiceRun(ctx, tx, domain.InvoiceRunCreateParams{
			ID:                      runID,
			TenantID:                actor.TenantID,
			BranchID:                actor.BranchID,
			BillingMonth:            candidate.BillingMonth,
			RunType:                 domain.InvoiceRunTypeIssue,
			Status:                  domain.InvoiceRunStatusStarted,
			RequestedByUserID:       actor.UserID,
			RequestedByMembershipID: actor.MembershipID,
			RequestID:               actor.RequestID,
		}); runErr != nil {
			return fmt.Errorf("create issue run: %w", runErr)
		}

		issueTime := time.Now().UTC()

		year := candidate.BillingMonth.Year()
		month := candidate.BillingMonth.Month()

		seq, seqErr := uc.repo.AllocateInvoiceNumberSequence(ctx, tx, actor.TenantID, actor.BranchID, year, int(month))
		if seqErr != nil {
			return fmt.Errorf("allocate invoice number sequence: %w", seqErr)
		}

		invoiceNumber := fmt.Sprintf("INV-%04d%02d-%04d", year, month, seq)

		// Advance-pay: due_at = first day of billing month at 00:00 UTC.
		dueAt := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

		if _, markErr := uc.repo.MarkInvoiceIssued(ctx, tx, domain.IssueInvoiceUpdateParams{
			ID:                   invoiceID,
			TenantID:             actor.TenantID,
			BranchID:             actor.BranchID,
			InvoiceNumber:        invoiceNumber,
			IssuedSequence:       seq,
			IssuedRunID:          runID,
			IssuedAt:             issueTime,
			IssuedByUserID:       actor.UserID,
			IssuedByMembershipID: actor.MembershipID,
			DueAt:                dueAt,
		}); markErr != nil {
			return fmt.Errorf("mark invoice issued: %w", markErr)
		}

		if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: domain.AuditInvoiceIssued,
			EntityType: domain.AuditEntityInvoice,
			EntityID:   invoiceID,
			Details: map[string]any{
				"invoice_number":  invoiceNumber,
				"billing_month":   candidate.BillingMonth.Format("2006-01"),
				"issued_run_id":   runID.String(),
				"issue_mode":      "single",
				"total_due_minor": candidate.TotalDue.Minor(),
				"due_at":          issueTime.Format(time.RFC3339),
			},
		}); auditErr != nil {
			return fmt.Errorf("write audit: %w", auditErr)
		}

		runDetails, _ := json.Marshal(map[string]any{
			"mode":          "single_invoice",
			"invoice_id":    invoiceID.String(),
			"issued_count":  1,
			"blocked_count": 0,
		})
		if compErr := uc.repo.CompleteInvoiceRun(ctx, tx, domain.InvoiceRunCompleteParams{
			ID:            runID,
			TenantID:      actor.TenantID,
			BranchID:      actor.BranchID,
			Status:        domain.InvoiceRunStatusCompleted,
			EligibleCount: 1,
			SuccessCount:  1,
			BlockedCount:  0,
			Details:       runDetails,
		}); compErr != nil {
			return fmt.Errorf("complete issue run: %w", compErr)
		}

		emitter.Emit(domain.InvoiceIssued{
			InvoiceID: invoiceID,
			Occurred:  issueTime,
		})

		result = domain.IssueInvoiceResult{
			InvoiceID:     invoiceID,
			InvoiceNumber: invoiceNumber,
			Status:        domain.InvoiceStatusIssued,
			IssuedAt:      issueTime,
			LockedAt:      issueTime,
			DueAt:         issueTime,
			IssuedRunID:   runID,
			TotalDue:      candidate.TotalDue,
		}

		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return domain.IssueInvoiceResult{}, txErr
		}
		return domain.IssueInvoiceResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

type BulkIssueInvoices struct {
	repo   domain.BillingRepository
	txMgr  *transaction.Manager
	auditW *audit.Writer
}

func NewBulkIssueInvoices(
	repo domain.BillingRepository,
	txMgr *transaction.Manager,
	auditW *audit.Writer,
) *BulkIssueInvoices {
	return &BulkIssueInvoices{repo: repo, txMgr: txMgr, auditW: auditW}
}

func (uc *BulkIssueInvoices) Execute(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string, rawInvoiceIDs []string, invoiceIDsProvided bool, confirm bool) (domain.BulkIssueInvoicesResult, error) {
	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.BulkIssueInvoicesResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	if !confirm {
		return domain.BulkIssueInvoicesResult{}, domainerrors.Validation("Confirmation required.", "confirm")
	}

	dedupedIDs, err := parseAndDedupeInvoiceIDs(rawInvoiceIDs)
	if err != nil {
		return domain.BulkIssueInvoicesResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_ids")
	}

	runID := uid.NewUUID()
	billingMonthStr := billingMonth.Format("2006-01")

	var result domain.BulkIssueInvoicesResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if runErr := uc.repo.CreateInvoiceRun(ctx, tx, domain.InvoiceRunCreateParams{
			ID:                      runID,
			TenantID:                actor.TenantID,
			BranchID:                actor.BranchID,
			BillingMonth:            billingMonth,
			RunType:                 domain.InvoiceRunTypeIssue,
			Status:                  domain.InvoiceRunStatusStarted,
			RequestedByUserID:       actor.UserID,
			RequestedByMembershipID: actor.MembershipID,
			RequestID:               actor.RequestID,
		}); runErr != nil {
			return fmt.Errorf("create issue run: %w", runErr)
		}

		var eligible []domain.InvoiceIssueCandidateRow
		var blocked []domain.InvoiceIssueBlocked

		if !invoiceIDsProvided {
			// All draft monthly invoices for the billing month.
			rows, listErr := uc.repo.ListDraftInvoicesForIssueForUpdate(ctx, tx, actor.TenantID, actor.BranchID, billingMonth)
			if listErr != nil {
				return fmt.Errorf("list draft invoices: %w", listErr)
			}
			eligible = rows
		} else {
			// Selected invoice IDs.
			if len(dedupedIDs) > 0 {
				rows, listErr := uc.repo.ListSelectedInvoicesForIssueForUpdate(ctx, tx, actor.TenantID, actor.BranchID, dedupedIDs)
				if listErr != nil {
					return fmt.Errorf("list selected invoices: %w", listErr)
				}

				foundSet := make(map[uuid.UUID]domain.InvoiceIssueCandidateRow, len(rows))
				for _, r := range rows {
					foundSet[r.ID] = r
				}

				for _, id := range dedupedIDs {
					row, found := foundSet[id]
					if !found {
						blocked = append(blocked, domain.InvoiceIssueBlocked{
							InvoiceID: id,
							Blockers: []domain.InvoiceIssueBlocker{
								{Code: domain.IssueBlockerInvoiceNotFound, Message: "Invoice not found or not in scope."},
							},
						})
						continue
					}
					if row.BillingMonth.Format("2006-01") != billingMonthStr {
						blocked = append(blocked, domain.InvoiceIssueBlocked{
							InvoiceID:       id,
							ChildID:         &row.ChildID,
							ChildFirstName:  row.ChildFirstName,
							ChildMiddleName: row.ChildMiddleName,
							ChildLastName:   row.ChildLastName,
							Blockers: []domain.InvoiceIssueBlocker{
								{Code: domain.IssueBlockerInvoiceNotInBillingMonth, Message: "Invoice does not match the requested billing month."},
							},
						})
						continue
					}
					if row.InvoiceKind != domain.InvoiceKindMonthly {
						blocked = append(blocked, domain.InvoiceIssueBlocked{
							InvoiceID:       id,
							ChildID:         &row.ChildID,
							ChildFirstName:  row.ChildFirstName,
							ChildMiddleName: row.ChildMiddleName,
							ChildLastName:   row.ChildLastName,
							Blockers: []domain.InvoiceIssueBlocker{
								{Code: domain.IssueBlockerInvoiceNotMonthly, Message: "Invoice is not a monthly invoice."},
							},
						})
						continue
					}
					if row.Status != domain.InvoiceStatusDraft {
						blocked = append(blocked, domain.InvoiceIssueBlocked{
							InvoiceID:       id,
							ChildID:         &row.ChildID,
							ChildFirstName:  row.ChildFirstName,
							ChildMiddleName: row.ChildMiddleName,
							ChildLastName:   row.ChildLastName,
							Blockers: []domain.InvoiceIssueBlocker{
								{Code: domain.IssueBlockerInvoiceNotDraft, Message: "Invoice is not a draft."},
							},
						})
						continue
					}
					eligible = append(eligible, row)
				}
			}
		}

		// Sort eligible by structured child name then invoice ID.
		sort.Slice(eligible, func(i, j int) bool {
			if eligible[i].ChildFirstName != eligible[j].ChildFirstName {
				return eligible[i].ChildFirstName < eligible[j].ChildFirstName
			}
			if stringPtrValue(eligible[i].ChildMiddleName) != stringPtrValue(eligible[j].ChildMiddleName) {
				return stringPtrValue(eligible[i].ChildMiddleName) < stringPtrValue(eligible[j].ChildMiddleName)
			}
			if stringPtrValue(eligible[i].ChildLastName) != stringPtrValue(eligible[j].ChildLastName) {
				return stringPtrValue(eligible[i].ChildLastName) < stringPtrValue(eligible[j].ChildLastName)
			}
			return eligible[i].ID.String() < eligible[j].ID.String()
		})

		issueTime := time.Now().UTC()
		year := billingMonth.Year()
		month := billingMonth.Month()

		var issued []domain.IssuedInvoiceResult
		var totalDueSum int

		for _, inv := range eligible {
			seq, seqErr := uc.repo.AllocateInvoiceNumberSequence(ctx, tx, actor.TenantID, actor.BranchID, year, int(month))
			if seqErr != nil {
				return fmt.Errorf("allocate sequence for invoice %s: %w", inv.ID, seqErr)
			}

			invoiceNumber := fmt.Sprintf("INV-%04d%02d-%04d", year, month, seq)

			// Advance-pay: due_at = first day of billing month at 00:00 UTC.
			dueAt := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

			if _, markErr := uc.repo.MarkInvoiceIssued(ctx, tx, domain.IssueInvoiceUpdateParams{
				ID:                   inv.ID,
				TenantID:             actor.TenantID,
				BranchID:             actor.BranchID,
				InvoiceNumber:        invoiceNumber,
				IssuedSequence:       seq,
				IssuedRunID:          runID,
				IssuedAt:             issueTime,
				IssuedByUserID:       actor.UserID,
				IssuedByMembershipID: actor.MembershipID,
				DueAt:                dueAt,
			}); markErr != nil {
				return fmt.Errorf("mark invoice issued %s: %w", inv.ID, markErr)
			}

			if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: domain.AuditInvoiceIssued,
				EntityType: domain.AuditEntityInvoice,
				EntityID:   inv.ID,
				Details: map[string]any{
					"invoice_number":  invoiceNumber,
					"billing_month":   billingMonthStr,
					"issued_run_id":   runID.String(),
					"issue_mode":      "bulk",
					"total_due_minor": inv.TotalDue.Minor(),
					"due_at":          issueTime.Format(time.RFC3339),
				},
			}); auditErr != nil {
				return fmt.Errorf("write audit for invoice %s: %w", inv.ID, auditErr)
			}

			issued = append(issued, domain.IssuedInvoiceResult{
				InvoiceID:       inv.ID,
				ChildID:         inv.ChildID,
				ChildFirstName:  inv.ChildFirstName,
				ChildMiddleName: inv.ChildMiddleName,
				ChildLastName:   inv.ChildLastName,
				InvoiceNumber:   invoiceNumber,
				IssuedAt:        issueTime,
				DueAt:           issueTime,
				TotalDue:        inv.TotalDue,
			})
			totalDueSum += inv.TotalDue.Minor()
		}

		runStatus := domain.InvoiceRunStatusCompleted
		if len(blocked) > 0 {
			runStatus = domain.InvoiceRunStatusCompletedWithExceptions
		}

		runDetails := map[string]any{
			"mode":          "bulk_issue",
			"billing_month": billingMonthStr,
			"issued_count":  len(issued),
			"blocked_count": len(blocked),
		}
		if invoiceIDsProvided {
			runDetails["mode"] = "bulk_issue_selected"
			runDetails["requested_count"] = len(dedupedIDs)
		}
		if len(blocked) > 0 {
			blockedDetails := make([]map[string]any, 0, len(blocked))
			for _, b := range blocked {
				codes := make([]string, 0, len(b.Blockers))
				for _, bl := range b.Blockers {
					codes = append(codes, bl.Code)
				}
				entry := map[string]any{
					"invoice_id": b.InvoiceID.String(),
					"blockers":   codes,
				}
				if b.ChildID != nil {
					entry["child_id"] = b.ChildID.String()
					entry["child_first_name"] = b.ChildFirstName
					entry["child_middle_name"] = b.ChildMiddleName
					entry["child_last_name"] = b.ChildLastName
				}
				blockedDetails = append(blockedDetails, entry)
			}
			runDetails["blocked_invoices"] = blockedDetails
		}
		detailsJSON, _ := json.Marshal(runDetails)

		if compErr := uc.repo.CompleteInvoiceRun(ctx, tx, domain.InvoiceRunCompleteParams{
			ID:            runID,
			TenantID:      actor.TenantID,
			BranchID:      actor.BranchID,
			Status:        runStatus,
			EligibleCount: len(issued) + len(blocked),
			SuccessCount:  len(issued),
			BlockedCount:  len(blocked),
			Details:       detailsJSON,
		}); compErr != nil {
			return fmt.Errorf("complete issue run: %w", compErr)
		}

		result = domain.BulkIssueInvoicesResult{
			RunID:        runID,
			BillingMonth: billingMonthStr,
			Status:       runStatus,
			Summary: domain.InvoiceIssueSummary{
				EligibleCount: len(issued) + len(blocked),
				SuccessCount:  len(issued),
				BlockedCount:  len(blocked),
				TotalDue:      domain.MustGBP(totalDueSum),
			},
			Issued:  issued,
			Blocked: blocked,
		}

		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return domain.BulkIssueInvoicesResult{}, txErr
		}
		return domain.BulkIssueInvoicesResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func parseAndDedupeInvoiceIDs(raw []string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	seen := make(map[uuid.UUID]struct{}, len(raw))
	result := make([]uuid.UUID, 0, len(raw))
	for _, r := range raw {
		id, err := uuid.Parse(r)
		if err != nil {
			return nil, fmt.Errorf("invalid invoice_id %q: %w", r, err)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func stringPtrValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
