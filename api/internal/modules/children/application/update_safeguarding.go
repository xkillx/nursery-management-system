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

type GetSafeguarding struct {
	repo domain.Repository
}

func NewGetSafeguarding(repo domain.Repository) *GetSafeguarding {
	return &GetSafeguarding{repo: repo}
}

func (uc *GetSafeguarding) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildSafeguardingProfile, error) {
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
	s, err := uc.repo.GetSafeguardingByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child safeguarding: %w", err))
	}
	return s, nil
}

type UpdateSafeguarding struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewUpdateSafeguarding(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *UpdateSafeguarding {
	return &UpdateSafeguarding{repo: repo, audit: auditWriter, txm: txm}
}

func (uc *UpdateSafeguarding) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in *ChildSafeguardingProfileInput) (*domain.ChildSafeguardingProfile, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in == nil {
		return nil, domainerrors.Validation("Invalid request payload.", "body")
	}

	var result *domain.ChildSafeguardingProfile
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}
		p := buildChildSafeguardingFromInput(actor.TenantID, actor.BranchID, id, in)
		saved, eerr := uc.repo.UpsertSafeguarding(ctx, tx, p)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("upsert child safeguarding: %w", eerr))
		}
		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_safeguarding_updated",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_safeguarding_updated: %w", aerr))
		}
		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
