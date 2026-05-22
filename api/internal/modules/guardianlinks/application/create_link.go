package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/guardianlinks/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/uid"
)

type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type GuardianActiveChecker interface {
	IsActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID) (bool, bool, error)
}

type ChildExistsChecker interface {
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
}

type CreateLinkUseCase struct {
	repo            domain.Repository
	audit           *audit.Writer
	txMgr           TxManager
	guardianChecker GuardianActiveChecker
	childChecker    ChildExistsChecker
}

type CreateLinkParams struct {
	TenantID   uuid.UUID
	BranchID   uuid.UUID
	GuardianID uuid.UUID
	ChildID    uuid.UUID
}

func NewCreateLinkUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager, guardianChecker GuardianActiveChecker, childChecker ChildExistsChecker) *CreateLinkUseCase {
	return &CreateLinkUseCase{repo: repo, audit: auditWriter, txMgr: txMgr, guardianChecker: guardianChecker, childChecker: childChecker}
}

var ErrGuardianNotFound = errors.New("guardian not found")
var ErrGuardianNotActive = errors.New("guardian not active")
var ErrChildNotFound = errors.New("child not found")

func (uc *CreateLinkUseCase) Execute(ctx context.Context, actor ActorContext, params CreateLinkParams) (domain.GuardianChildLink, error) {
	var result domain.GuardianChildLink

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		guardianActive, exists, err := uc.guardianChecker.IsActive(ctx, tx, params.TenantID, params.BranchID, params.GuardianID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrGuardianNotFound
		}
		if !guardianActive {
			return ErrGuardianNotActive
		}

		childExists, err := uc.childChecker.ExistsInScope(ctx, tx, params.TenantID, params.BranchID, params.ChildID)
		if err != nil {
			return err
		}
		if !childExists {
			return ErrChildNotFound
		}

		existing, exists, err := uc.repo.FindActiveByPair(ctx, tx, params.TenantID, params.BranchID, params.GuardianID, params.ChildID)
		if err != nil {
			return err
		}
		if exists {
			result = existing
			return nil
		}

		link := domain.GuardianChildLink{
			ID:         uid.NewUUID(),
			TenantID:   params.TenantID,
			BranchID:   params.BranchID,
			GuardianID: params.GuardianID,
			ChildID:    params.ChildID,
		}

		if err := uc.repo.Create(ctx, tx, link); err != nil {
			return err
		}

		created, found, err := uc.repo.GetByIDForUpdate(ctx, tx, params.TenantID, params.BranchID, link.ID)
		if err != nil || !found {
			return err
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "guardian_child_link_created",
			EntityType: "guardian_child_link",
			EntityID:   created.ID,
			Details: map[string]any{
				"guardian_id": params.GuardianID.String(),
				"child_id":    params.ChildID.String(),
			},
		}); err != nil {
			return err
		}

		result = created
		return nil
	})

	return result, err
}
