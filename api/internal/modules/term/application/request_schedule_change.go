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
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

// RequestScheduleChangeInput is the validated input for an in-term booking pattern change.
type RequestScheduleChangeInput struct {
	TermID              uuid.UUID
	NewBookingPatternID uuid.UUID
	EffectiveFrom       time.Time
	ChangeKind          domain.ScheduleChangeKind
}

type RequestScheduleChangeUseCase struct {
	termRepo     domain.Repository
	scheduleRepo domain.ScheduleChangeRepository
	txMgr        *transaction.Manager
	audit        *audit.Writer
	now          func() time.Time
}

func NewRequestScheduleChangeUseCase(
	termRepo domain.Repository,
	scheduleRepo domain.ScheduleChangeRepository,
	txMgr *transaction.Manager,
	auditWriter *audit.Writer,
) *RequestScheduleChangeUseCase {
	return &RequestScheduleChangeUseCase{
		termRepo:     termRepo,
		scheduleRepo: scheduleRepo,
		txMgr:        txMgr,
		audit:        auditWriter,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func (uc *RequestScheduleChangeUseCase) Execute(ctx context.Context, actor tenant.ActorContext, in RequestScheduleChangeInput) (*domain.TermScheduleChange, error) {
	if in.TermID == uuid.Nil {
		return nil, domainerrors.Validation("Invalid request payload.", "term_id")
	}
	if in.NewBookingPatternID == uuid.Nil {
		return nil, domainerrors.Validation("Invalid request payload.", "new_booking_pattern_id")
	}
	if in.ChangeKind != domain.ScheduleChangeDecrease && in.ChangeKind != domain.ScheduleChangeIncrease {
		return nil, domainerrors.Validation("Invalid request payload.", "change_kind")
	}
	if err := domain.ValidateEffectiveFrom(in.EffectiveFrom); err != nil {
		return nil, domainerrors.New("term_invalid_effective_from", "Invalid request payload.", "effective_from")
	}

	var result *domain.TermScheduleChange
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		term, found, err := uc.termRepo.GetByIDInTx(ctx, tx, actor.TenantID, actor.BranchID, in.TermID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get term: %w", err))
		}
		if !found {
			return domainerrors.NotFound("term", "Resource not found.")
		}
		if term.Status != domain.TermStatusActive && term.Status != domain.TermStatusPendingRenewal {
			return domainerrors.Conflict("term_not_active", "Term is not active.")
		}
		if in.EffectiveFrom.Before(term.TermStartDate) {
			return domainerrors.New("schedule_change_before_term", "Invalid request payload.", "effective_from")
		}
		if in.NewBookingPatternID == term.BookingPatternID {
			return domainerrors.New("schedule_change_no_change", "Invalid request payload.", "new_booking_pattern_id")
		}

		// 1-month-notice rule: effective_from >= max(today + 1 month, term_start_date).
		noticeDeadline := uc.now().UTC().Truncate(24*time.Hour).AddDate(0, 1, 0)
		if in.EffectiveFrom.Before(noticeDeadline) {
			return domainerrors.New("schedule_change_notice_too_short", "Invalid request payload.", "effective_from")
		}

		// Cap at term_end_date + 1 month (a change beyond the term is invalid).
		cap := term.TermEndDate.AddDate(0, 1, 0)
		if in.EffectiveFrom.After(cap) {
			return domainerrors.New("schedule_change_after_term_end", "Invalid request payload.", "effective_from")
		}

		change := &domain.TermScheduleChange{
			ID:                       uid.NewUUID(),
			TenantID:                 actor.TenantID,
			BranchID:                 actor.BranchID,
			TermID:                   term.ID,
			PreviousBookingPatternID: term.BookingPatternID,
			NewBookingPatternID:      in.NewBookingPatternID,
			ChangeKind:               in.ChangeKind,
			EffectiveFrom:            in.EffectiveFrom,
			RequestID:                actor.RequestID,
		}
		saved, err := uc.scheduleRepo.Insert(ctx, tx, change)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("insert schedule change: %w", err))
		}

		// Auto-approve decreases (free under the plan). Increases require manager approval.
		if in.ChangeKind == domain.ScheduleChangeDecrease {
			approvedBy := actor.MembershipID
			if _, err := uc.scheduleRepo.Approve(ctx, tx, actor.TenantID, actor.BranchID, saved.ID, approvedBy); err != nil {
				return domainerrors.Internal(fmt.Errorf("auto-approve decrease: %w", err))
			}
			decision := domain.ScheduleChangeApproved
			saved.ApprovalDecision = &decision
			saved.ApprovedByMembershipID = &approvedBy
		}

		// Audit.
		auditAction := domain.AuditTermScheduleChangeRequested
		if saved.ApprovalDecision != nil {
			auditAction = domain.AuditTermScheduleChangeApproved
		}
		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: auditAction,
			EntityType: domain.AuditEntityTermScheduleChange,
			EntityID:   saved.ID,
			Details: map[string]any{
				"term_id":             saved.TermID.String(),
				"change_kind":         string(saved.ChangeKind),
				"effective_from":      saved.EffectiveFrom.Format("2006-01-02"),
				"previous_pattern_id": saved.PreviousBookingPatternID.String(),
				"new_pattern_id":      saved.NewBookingPatternID.String(),
				"auto_approved":       saved.ApprovalDecision != nil,
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit schedule change: %w", err))
		}

		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
