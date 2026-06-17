package application

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type GetFunding struct {
	repo domain.Repository
}

func NewGetFunding(repo domain.Repository) *GetFunding {
	return &GetFunding{repo: repo}
}

func (uc *GetFunding) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildFundingRecord, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}
	f, found, err := uc.repo.GetFundingByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child funding: %w", err))
	}
	if !found {
		return nil, nil
	}
	return f, nil
}

type UpdateFunding struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewUpdateFunding(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *UpdateFunding {
	return &UpdateFunding{repo: repo, audit: auditWriter, txm: txm}
}

func (uc *UpdateFunding) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in *ChildFundingRecordInput) (*domain.ChildFundingRecord, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in == nil {
		return nil, domainerrors.Validation("Invalid request payload.", "body")
	}

	var result *domain.ChildFundingRecord
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}
		p := buildChildFundingFromInput(actor.TenantID, actor.BranchID, id, in)
		saved, eerr := uc.repo.UpsertFunding(ctx, tx, p)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("upsert child funding: %w", eerr))
		}
		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_funding_updated",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_funding_updated: %w", aerr))
		}
		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
