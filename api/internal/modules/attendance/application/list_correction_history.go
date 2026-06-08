package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type CorrectionHistoryResult struct {
	Session domain.Session
	Events  []domain.CorrectionHistoryEvent
}

type ListCorrectionHistory struct {
	repo domain.Repository
}

func NewListCorrectionHistory(repo domain.Repository) *ListCorrectionHistory {
	return &ListCorrectionHistory{repo: repo}
}

func (uc *ListCorrectionHistory) Execute(ctx context.Context, actor tenant.ActorContext, sessionID uuid.UUID) (CorrectionHistoryResult, error) {
	session, events, err := uc.repo.ListCorrectionHistory(ctx, actor.TenantID, actor.BranchID, sessionID)
	if err != nil {
		if err == domain.ErrSessionNotFound {
			return CorrectionHistoryResult{}, domainerrors.New("attendance_session_not_found", "Attendance session not found.")
		}
		return CorrectionHistoryResult{}, domainerrors.Internal(err)
	}
	return CorrectionHistoryResult{Session: session, Events: events}, nil
}
