package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term_calendar/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CreateTermParams struct {
	Name      string
	Kind      string
	StartDate time.Time
	EndDate   time.Time
}

type CreateTerm struct {
	repo domain.Repository
}

func NewCreateTerm(repo domain.Repository) *CreateTerm {
	return &CreateTerm{repo: repo}
}

func (uc *CreateTerm) Execute(ctx context.Context, actor TermCalendarActor, siteID uuid.UUID, params CreateTermParams) (domain.AcademicTerm, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.AcademicTerm{}, err
	}

	if !IsOwnerActor(actor) && actor.TenantID() == uuid.Nil {
		return domain.AcademicTerm{}, domainerrors.Unauthorized("Access denied.")
	}

	if params.Name == "" || len(params.Name) > 120 {
		return domain.AcademicTerm{}, domainerrors.Validation("Name is required and must be 120 characters or fewer.", "name")
	}

	if !domain.ValidTermKind(params.Kind) {
		return domain.AcademicTerm{}, domainerrors.Validation("Kind must be autumn, spring, or summer.", "kind")
	}

	if !params.StartDate.Before(params.EndDate) {
		return domain.AcademicTerm{}, domainerrors.Validation("Start date must be before end date.", "start_date")
	}

	exists, err := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, params.Name, nil)
	if err != nil {
		return domain.AcademicTerm{}, internalError(err)
	}
	if exists {
		return domain.AcademicTerm{}, domainerrors.Conflict("term_name_taken", "An active term with this name already exists at this site.")
	}

	term := domain.AcademicTerm{
		ID:        uuid.New(),
		TenantID:  actor.TenantID(),
		BranchID:  siteID,
		Name:      params.Name,
		Kind:      params.Kind,
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
		IsActive:  true,
	}

	if err := uc.repo.Create(ctx, term); err != nil {
		return domain.AcademicTerm{}, internalError(err)
	}

	return term, nil
}
