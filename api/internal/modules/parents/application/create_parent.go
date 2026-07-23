package application

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateParentUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr TxManager
}

func NewCreateParentUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager) *CreateParentUseCase {
	return &CreateParentUseCase{repo: repo, audit: auditWriter, txMgr: txMgr}
}

type CreateParentParams struct {
	FirstName               string
	LastName                *string
	Email                   *string
	Phone                   *string
	AddressLine1            *string
	AddressLine2            *string
	AddressCity             *string
	AddressPostcode         *string
	RelationshipToChild     *string
	HasParentalResponsibility bool
	CanPickUp               bool
	IsEmergencyContact      bool
	Notes                   *string
}

func (uc *CreateParentUseCase) Execute(ctx context.Context, actor ActorContext, params CreateParentParams) (domain.Parent, error) {
	if strings.TrimSpace(params.FirstName) == "" {
		return domain.Parent{}, domainerrors.Validation("Enter the parent's first name.", "first_name")
	}

	var result domain.Parent
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		parent := domain.Parent{
			ID:                      uid.NewUUID(),
			TenantID:                actor.TenantID,
			BranchID:                actor.BranchID,
			FirstName:               strings.TrimSpace(params.FirstName),
			LastName:                params.LastName,
			Email:                   params.Email,
			Phone:                   params.Phone,
			AddressLine1:            params.AddressLine1,
			AddressLine2:            params.AddressLine2,
			AddressCity:             params.AddressCity,
			AddressPostcode:         params.AddressPostcode,
			RelationshipToChild:     params.RelationshipToChild,
			HasParentalResponsibility: params.HasParentalResponsibility,
			CanPickUp:               params.CanPickUp,
			IsEmergencyContact:      params.IsEmergencyContact,
			Notes:                   params.Notes,
			IsActive:                true,
		}

		if err := uc.repo.Create(ctx, tx, parent); err != nil {
			return err
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_created",
			EntityType: "parent",
			EntityID:   parent.ID,
			Details: map[string]any{
				"first_name": parent.FirstName,
			},
		}); err != nil {
			return err
		}

		created, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, parent.ID)
		if err != nil || !found {
			return err
		}

		result = created
		return nil
	})

	return result, err
}
