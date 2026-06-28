package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/lifecycle"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type TerminateTermInput struct {
	TermID         uuid.UUID
	ReasonCode     string
	ReasonNote     string
	EffectiveMonth time.Time // first day of the notice month; parent pays for this month only
}

type TerminateTermUseCase struct {
	repo  domain.Repository
	txMgr *transaction.Manager
	audit *audit.Writer
	now   func() time.Time
}

func NewTerminateTermUseCase(
	repo domain.Repository,
	txMgr *transaction.Manager,
	auditWriter *audit.Writer,
) *TerminateTermUseCase {
	return &TerminateTermUseCase{repo: repo, txMgr: txMgr, audit: auditWriter, now: func() time.Time { return time.Now().UTC() }}
}

func (uc *TerminateTermUseCase) Execute(ctx context.Context, actor tenant.ActorContext, in TerminateTermInput) (*domain.Term, error) {
	if in.TermID == uuid.Nil {
		return nil, domainerrors.Validation("Invalid request payload.", "term_id")
	}
	if !lifecycle.IsValidReasonCode(in.ReasonCode) {
		return nil, domainerrors.New("invalid_reason_code", "Invalid request payload.", "reason_code")
	}
	if len(in.ReasonNote) > lifecycle.MaxReasonNoteLen {
		return nil, domainerrors.New("reason_note_too_long", "Invalid request payload.", "reason_note")
	}

	var result *domain.Term
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		term, found, err := uc.repo.GetByIDInTx(ctx, tx, actor.TenantID, actor.BranchID, in.TermID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get term: %w", err))
		}
		if !found {
			return domainerrors.NotFound("term", "Resource not found.")
		}
		if term.Status != domain.TermStatusActive && term.Status != domain.TermStatusPendingRenewal && term.Status != domain.TermStatusPreTerm {
			return domainerrors.Conflict("term_not_active", "Term is not active.")
		}

		// 1-month notice: terminated_at is end-of-notice-month.
		today := uc.now().UTC().Truncate(24 * time.Hour)
		var noticeMonth time.Time
		if !in.EffectiveMonth.IsZero() {
			if err := domain.ValidateTermStartDate(in.EffectiveMonth); err != nil {
				return domainerrors.New("term_invalid_start_date", "Invalid request payload.", "effective_month")
			}
			noticeMonth = in.EffectiveMonth.UTC()
			earliestNoticeMonth := today.AddDate(0, 1, 0)
			earliestNoticeMonth = time.Date(earliestNoticeMonth.Year(), earliestNoticeMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
			if noticeMonth.Before(earliestNoticeMonth) {
				return domainerrors.New("termination_notice_too_short", "Invalid request payload.", "effective_month")
			}
		} else {
			noticeMonth = today.AddDate(0, 1, 0)
			noticeMonth = time.Date(noticeMonth.Year(), noticeMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		}

		// The child leaves at the end of the notice month; the parent pays that month only.
		// terminator semantics: set status='terminated' and terminated_at = end-of-notice-month.
		terminatedAt := noticeMonth.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		rows, err := uc.repo.Terminate(ctx, tx, actor.TenantID, actor.BranchID, term.ID, terminatedAt, in.ReasonCode, in.ReasonNote)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("terminate term: %w", err))
		}
		if rows == 0 {
			return domainerrors.Conflict("term_not_active", "Term is not active.")
		}

		// Clear child denormalisation since no active term remains.
		if err := uc.repo.ClearChildCurrentTermID(ctx, tx, actor.TenantID, actor.BranchID, term.ChildID); err != nil {
			return domainerrors.Internal(fmt.Errorf("clear child current term: %w", err))
		}

		note := in.ReasonNote
		termAfter := *term
		termAfter.Status = domain.TermStatusTerminated
		termAfter.TerminatedAt = &terminatedAt
		termAfter.TerminationReasonCode = &in.ReasonCode
		termAfter.TerminationReasonNote = &note

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: domain.AuditTermTerminated,
			EntityType: domain.AuditEntityTerm,
			EntityID:   term.ID,
			ReasonCode: &in.ReasonCode,
			ReasonNote: &note,
			Details: map[string]any{
				"child_id":               term.ChildID.String(),
				"previous_term_end_date": term.TermEndDate.Format("2006-01-02"),
				"notice_month":           noticeMonth.Format("2006-01"),
				"terminated_at":          terminatedAt.Format(time.RFC3339),
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit term_terminated: %w", err))
		}

		result = &termAfter
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
