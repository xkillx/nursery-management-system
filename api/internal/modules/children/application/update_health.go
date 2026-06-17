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

type GetHealth struct {
	repo domain.Repository
}

func NewGetHealth(repo domain.Repository) *GetHealth {
	return &GetHealth{repo: repo}
}

func (uc *GetHealth) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildHealthProfile, error) {
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
	h, err := uc.repo.GetHealthByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child health: %w", err))
	}
	return h, nil
}

type UpdateHealth struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewUpdateHealth(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *UpdateHealth {
	return &UpdateHealth{repo: repo, audit: auditWriter, txm: txm}
}

func (uc *UpdateHealth) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in *ChildHealthProfileInput) (*domain.ChildHealthProfile, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in == nil {
		return nil, domainerrors.Validation("Invalid request payload.", "body")
	}

	var result *domain.ChildHealthProfile
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}
		p := buildChildHealthFromInput(actor.TenantID, actor.BranchID, id, in)
		saved, eerr := uc.repo.UpsertHealth(ctx, tx, p)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("upsert child health: %w", eerr))
		}
		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_health_updated",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_health_updated: %w", aerr))
		}
		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
