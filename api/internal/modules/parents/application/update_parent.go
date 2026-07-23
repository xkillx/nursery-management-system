package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateParentUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr TxManager
}

func NewUpdateParentUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager) *UpdateParentUseCase {
	return &UpdateParentUseCase{repo: repo, audit: auditWriter, txMgr: txMgr}
}

type UpdateParentParams struct {
	FirstName               *string
	LastName                *string
	Email                   *string
	Phone                   *string
	AddressLine1            *string
	AddressLine2            *string
	AddressCity             *string
	AddressPostcode         *string
	RelationshipToChild     *string
	HasParentalResponsibility *bool
	CanPickUp               *bool
	IsEmergencyContact      *bool
	Notes                   *string
	IsActive                *bool
}

func (uc *UpdateParentUseCase) Execute(ctx context.Context, actor ActorContext, parentID uuid.UUID, params UpdateParentParams) (domain.Parent, error) {
	var result domain.Parent

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		existing, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, parentID)
		if err != nil {
			return err
		}
		if !found {
			return domain.ErrParentNotFound
		}

		if params.FirstName != nil {
			if strings.TrimSpace(*params.FirstName) == "" {
				return domainerrors.Validation("Enter the parent's first name.", "first_name")
			}
			existing.FirstName = strings.TrimSpace(*params.FirstName)
		}
		if params.LastName != nil {
			existing.LastName = params.LastName
		}
		if params.Email != nil {
			existing.Email = params.Email
		}
		if params.Phone != nil {
			existing.Phone = params.Phone
		}
		if params.AddressLine1 != nil {
			existing.AddressLine1 = params.AddressLine1
		}
		if params.AddressLine2 != nil {
			existing.AddressLine2 = params.AddressLine2
		}
		if params.AddressCity != nil {
			existing.AddressCity = params.AddressCity
		}
		if params.AddressPostcode != nil {
			existing.AddressPostcode = params.AddressPostcode
		}
		if params.RelationshipToChild != nil {
			existing.RelationshipToChild = params.RelationshipToChild
		}
		if params.HasParentalResponsibility != nil {
			existing.HasParentalResponsibility = *params.HasParentalResponsibility
		}
		if params.CanPickUp != nil {
			existing.CanPickUp = *params.CanPickUp
		}
		if params.IsEmergencyContact != nil {
			existing.IsEmergencyContact = *params.IsEmergencyContact
		}
		if params.Notes != nil {
			existing.Notes = params.Notes
		}
		if params.IsActive != nil {
			existing.IsActive = *params.IsActive
		}

		if err := uc.repo.Update(ctx, tx, existing); err != nil {
			return err
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_updated",
			EntityType: "parent",
			EntityID:   parentID,
			Details:    map[string]any{},
		}); err != nil {
			return err
		}

		updated, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, parentID)
		if err != nil || !found {
			return err
		}

		result = updated
		return nil
	})

	return result, err
}
