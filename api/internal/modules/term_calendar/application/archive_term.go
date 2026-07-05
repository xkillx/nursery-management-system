package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/term_calendar/domain"
)

type ArchiveTerm struct {
	repo domain.Repository
	txm  TxManager
}

func NewArchiveTerm(repo domain.Repository, txm TxManager) *ArchiveTerm {
	return &ArchiveTerm{repo: repo, txm: txm}
}

func (uc *ArchiveTerm) Execute(ctx context.Context, actor TermCalendarActor, siteID, termID uuid.UUID) error {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return err
	}

	return uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		return uc.repo.Archive(ctx, tx, actor.TenantID(), siteID, termID)
	})
}
