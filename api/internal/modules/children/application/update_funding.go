package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type FundingHistoryWriter interface {
	Write(ctx context.Context, tenantID, branchID, childID uuid.UUID, record *domain.ChildFundingRecord, changedByUserID uuid.UUID) error
}

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
	repo     domain.Repository
	audit    *audit.Writer
	txm      *transaction.Manager
	historyW FundingHistoryWriter
}

func NewUpdateFunding(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager, historyWriter FundingHistoryWriter) *UpdateFunding {
	return &UpdateFunding{repo: repo, audit: auditWriter, txm: txm, historyW: historyWriter}
}

func (uc *UpdateFunding) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in *ChildFundingRecordInput) (*domain.ChildFundingRecord, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in == nil {
		return nil, domainerrors.Validation("Invalid request payload.", "body")
	}

	if in.FundingEnabled {
		if fieldErrs := validateFundingInput(in, ""); len(fieldErrs) > 0 {
			return nil, domainerrors.ValidationWithFields("Some fields did not validate.", fieldErrs)
		}
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

		// Write funding history
		if herr := uc.historyW.Write(ctx, actor.TenantID, actor.BranchID, id, saved, actor.UserID); herr != nil {
			return domainerrors.Internal(fmt.Errorf("write funding history: %w", herr))
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
