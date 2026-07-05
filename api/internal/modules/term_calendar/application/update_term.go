package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/term_calendar/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateTermParams struct {
	Name      *string
	Kind      *string
	StartDate *time.Time
	EndDate   *time.Time
}

type UpdateTerm struct {
	repo domain.Repository
	txm  TxManager
}

func NewUpdateTerm(repo domain.Repository, txm TxManager) *UpdateTerm {
	return &UpdateTerm{repo: repo, txm: txm}
}

func (uc *UpdateTerm) Execute(ctx context.Context, actor TermCalendarActor, siteID, termID uuid.UUID, params UpdateTermParams) (domain.AcademicTerm, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.AcademicTerm{}, err
	}

	if params.Name != nil && (len(*params.Name) == 0 || len(*params.Name) > 120) {
		return domain.AcademicTerm{}, domainerrors.Validation("Name must be 1-120 characters.", "name")
	}
	if params.Kind != nil && !domain.ValidTermKind(*params.Kind) {
		return domain.AcademicTerm{}, domainerrors.Validation("Kind must be autumn, spring, or summer.", "kind")
	}

	var result domain.AcademicTerm
	err := uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		term, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, termID)
		if err != nil {
			return err
		}

		fields := make(map[string]any)
		if params.Name != nil {
			fields["name"] = *params.Name
			term.Name = *params.Name
		}
		if params.Kind != nil {
			fields["kind"] = *params.Kind
		}
		if params.StartDate != nil {
			fields["start_date"] = *params.StartDate
			term.StartDate = *params.StartDate
		}
		if params.EndDate != nil {
			fields["end_date"] = *params.EndDate
			term.EndDate = *params.EndDate
		}

		if !term.StartDate.Before(term.EndDate) {
			return domainerrors.Validation("Start date must be before end date.", "start_date")
		}

		if params.Name != nil {
			exists, err := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, term.Name, &termID)
			if err != nil {
				return internalError(err)
			}
			if exists {
				return domainerrors.Conflict("term_name_taken", "An active term with this name already exists at this site.")
			}
		}

		if len(fields) > 0 {
			if _, err := uc.repo.Update(ctx, actor.TenantID(), siteID, termID, fields); err != nil {
				return internalError(err)
			}
		}

		result, err = uc.repo.GetByID(ctx, actor.TenantID(), siteID, termID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if de, ok := err.(*domainerrors.DomainError); ok {
			return domain.AcademicTerm{}, de
		}
		return domain.AcademicTerm{}, internalError(err)
	}

	return result, nil
}
