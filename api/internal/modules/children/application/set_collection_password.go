package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type GetCollectionSetting struct {
	repo domain.Repository
}

func NewGetCollectionSetting(repo domain.Repository) *GetCollectionSetting {
	return &GetCollectionSetting{repo: repo}
}

func (uc *GetCollectionSetting) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildCollectionSetting, error) {
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
	cs, err := uc.repo.GetCollectionSettingByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child collection setting: %w", err))
	}
	return cs, nil
}

type SetCollectionPassword struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewSetCollectionPassword(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *SetCollectionPassword {
	return &SetCollectionPassword{repo: repo, audit: auditWriter, txm: txm}
}

type SetCollectionPasswordInput struct {
	Over18CollectionAcknowledged *bool
	Password                     *string
}

func (uc *SetCollectionPassword) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in SetCollectionPasswordInput) (*domain.ChildCollectionSetting, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	var result *domain.ChildCollectionSetting
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}

		existing, eerr := uc.repo.GetCollectionSettingByChild(ctx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("get child collection setting: %w", eerr))
		}

		var cs *domain.ChildCollectionSetting
		if existing == nil {
			cs = &domain.ChildCollectionSetting{
				ID:       uid.NewUUID(),
				TenantID: actor.TenantID,
				BranchID: actor.BranchID,
				ChildID:  id,
			}
			if in.Over18CollectionAcknowledged != nil {
				cs.Over18CollectionAcknowledged = *in.Over18CollectionAcknowledged
			}
			cs, eerr = uc.repo.UpsertCollectionSetting(ctx, tx, cs)
			if eerr != nil {
				return domainerrors.Internal(fmt.Errorf("upsert child collection setting: %w", eerr))
			}
		} else if in.Over18CollectionAcknowledged != nil {
			existing.Over18CollectionAcknowledged = *in.Over18CollectionAcknowledged
			cs, eerr = uc.repo.UpsertCollectionSetting(ctx, tx, existing)
			if eerr != nil {
				return domainerrors.Internal(fmt.Errorf("update child collection setting: %w", eerr))
			}
		} else {
			cs = existing
		}

		if in.Password != nil && *in.Password != "" {
			if perr := uc.repo.SetCollectionPassword(ctx, tx, actor.TenantID, actor.BranchID, id, cs.ID, *in.Password, time.Now().UTC(), actor.UserID, actor.MembershipID); perr != nil {
				return domainerrors.Internal(fmt.Errorf("set collection password: %w", perr))
			}
		}

		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_collection_password_set",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_collection_password_set: %w", aerr))
		}

		fresh, eerr := uc.repo.GetCollectionSettingByChild(ctx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("reload child collection setting: %w", eerr))
		}
		result = fresh
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
