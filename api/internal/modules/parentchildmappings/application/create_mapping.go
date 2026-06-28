package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parentchildmappings/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

var (
	ErrMembershipNotFound  = domainerrors.NotFound("membership", "Membership not found")
	ErrMembershipNotParent = domainerrors.New("membership_not_parent", "Membership is not a parent role")
	ErrMembershipNotActive = domainerrors.New("membership_not_active", "Membership is not active")
	ErrChildNotFound       = domainerrors.NotFound("child_map", "Child not found")
)

type CreateMappingUseCase struct {
	repo       domain.Repository
	audit      *audit.Writer
	txMgr      TxManager
	membership domain.MembershipChecker
	children   ChildChecker
}

type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

// ChildChecker is implemented by the children module's repository. It exists
// here as a port so this module does not import the children module directly
// (AGENTS.md: no cross-module imports).
type ChildChecker interface {
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
}

func NewCreateMappingUseCase(
	repo domain.Repository,
	auditWriter *audit.Writer,
	txMgr TxManager,
	membership domain.MembershipChecker,
	children ChildChecker,
) *CreateMappingUseCase {
	return &CreateMappingUseCase{repo: repo, audit: auditWriter, txMgr: txMgr, membership: membership, children: children}
}

type CreateMappingParams struct {
	TenantID     uuid.UUID
	BranchID     uuid.UUID
	MembershipID uuid.UUID
	ChildID      uuid.UUID
}

func (uc *CreateMappingUseCase) Execute(ctx context.Context, actor ActorContext, params CreateMappingParams) (domain.ParentChildMapping, error) {
	var result domain.ParentChildMapping

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		membership, mFound, err := uc.membership.GetForScope(ctx, tx, params.TenantID, params.BranchID, params.MembershipID)
		if err != nil {
			return err
		}
		if !mFound {
			return ErrMembershipNotFound
		}
		if membership.Role != "parent" {
			return ErrMembershipNotParent
		}
		if !membership.IsActive {
			return ErrMembershipNotActive
		}

		childExists, err := uc.children.ExistsInScope(ctx, tx, params.TenantID, params.BranchID, params.ChildID)
		if err != nil {
			return err
		}
		if !childExists {
			return ErrChildNotFound
		}

		existing, hasMapping, err := uc.repo.FindActiveByPair(ctx, tx, params.TenantID, params.BranchID, params.MembershipID, params.ChildID)
		if err != nil {
			return err
		}
		if hasMapping {
			result = existing
			return nil
		}

		mapping := domain.ParentChildMapping{
			ID:           uid.NewUUID(),
			TenantID:     params.TenantID,
			BranchID:     params.BranchID,
			MembershipID: params.MembershipID,
			ChildID:      params.ChildID,
		}

		if err := uc.repo.Create(ctx, tx, mapping); err != nil {
			return err
		}

		created, found, err := uc.repo.GetByIDForUpdate(ctx, tx, params.TenantID, params.BranchID, mapping.ID)
		if err != nil || !found {
			return err
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_mapping_created",
			EntityType: "parent_membership_child_mapping",
			EntityID:   created.ID,
			Details: map[string]any{
				"membership_id": params.MembershipID.String(),
				"child_id":      params.ChildID.String(),
			},
		}); err != nil {
			return err
		}

		result = created
		return nil
	})

	return result, err
}
