package application

import (
	"context"
	"fmt"
	"strings"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type MarkInactiveParams struct {
	ReasonCode string
	ReasonNote string
}

type MarkInactive struct {
	repo domain.Repository
	txm  *transaction.Manager
	audit *audit.Writer
}

func NewMarkInactive(repo domain.Repository, txm *transaction.Manager, auditWriter *audit.Writer) *MarkInactive {
	return &MarkInactive{repo: repo, txm: txm, audit: auditWriter}
}

func (uc *MarkInactive) Execute(ctx context.Context, actor tenant.ActorContext, childID string, params MarkInactiveParams) (domain.Child, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	if err := ValidateReasonCode(params.ReasonCode, params.ReasonNote); err != nil {
		return domain.Child{}, err
	}

	reasonCode := strings.TrimSpace(params.ReasonCode)
	reasonNote := strings.TrimSpace(params.ReasonNote)

	var result domain.Child

	err = uc.txm.ExecTx(ctx, func(tx domain.Tx) error {
		child, found, fetchErr := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, id)
		if fetchErr != nil {
			return domainerrors.Internal(fmt.Errorf("fetch child for update: %w", fetchErr))
		}
		if !found {
			return domainerrors.NotFound("child", "Resource not found.")
		}

		if child.IsActive {
			if markErr := uc.repo.MarkInactive(ctx, tx, actor.TenantID, actor.BranchID, id, reasonCode, reasonNote); markErr != nil {
				return domainerrors.Internal(fmt.Errorf("mark child inactive: %w", markErr))
			}

			reasonCodePtr := &reasonCode
			var reasonNotePtr *string
			if note := strings.TrimSpace(reasonNote); note != "" {
				reasonNotePtr = &note
			}

			if auditErr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "child_marked_inactive",
				EntityType: "child",
				EntityID:   id,
				ReasonCode: reasonCodePtr,
				ReasonNote: reasonNotePtr,
				Details:    map[string]any{},
			}); auditErr != nil {
				return domainerrors.Internal(fmt.Errorf("audit child_marked_inactive: %w", auditErr))
			}
		}

		updated, found, fetchErr := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, id)
		if fetchErr != nil || !found {
			return domainerrors.Internal(fmt.Errorf("fetch updated child: %w", fetchErr))
		}
		result = updated
		return nil
	})

	if err != nil {
		return domain.Child{}, err
	}

	return result, nil
}
