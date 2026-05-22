package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parentmappings/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/uid"
)

var (
	ErrMembershipNotFound = errors.New("membership not found")
	ErrMembershipNotParent = errors.New("membership not parent role")
	ErrMembershipNotActive = errors.New("membership not active")
	ErrGuardianNotFound   = errors.New("guardian not found")
	ErrGuardianNotActive  = errors.New("guardian not active")
	ErrActiveConflict     = errors.New("parent mapping active conflict")
)

type CreateMappingUseCase struct {
	repo       domain.Repository
	audit      *audit.Writer
	txMgr      TxManager
	membership domain.MembershipChecker
	guardian   domain.GuardianChecker
}

type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

func NewCreateMappingUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager, membership domain.MembershipChecker, guardian domain.GuardianChecker) *CreateMappingUseCase {
	return &CreateMappingUseCase{repo: repo, audit: auditWriter, txMgr: txMgr, membership: membership, guardian: guardian}
}

type CreateMappingParams struct {
	TenantID     uuid.UUID
	BranchID     uuid.UUID
	MembershipID uuid.UUID
	GuardianID   uuid.UUID
}

func (uc *CreateMappingUseCase) Execute(ctx context.Context, actor ActorContext, params CreateMappingParams) (domain.ParentMapping, error) {
	var result domain.ParentMapping

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

		guardianActive, gFound, err := uc.guardian.IsActive(ctx, tx, params.TenantID, params.BranchID, params.GuardianID)
		if err != nil {
			return err
		}
		if !gFound {
			return ErrGuardianNotFound
		}
		if !guardianActive {
			return ErrGuardianNotActive
		}

		existing, hasMapping, err := uc.repo.FindActiveByMembership(ctx, tx, params.TenantID, params.BranchID, params.MembershipID)
		if err != nil {
			return err
		}
		if hasMapping {
			if existing.GuardianID == params.GuardianID {
				result = existing
				return nil
			}
			return ErrActiveConflict
		}

		mapping := domain.ParentMapping{
			ID:           uid.NewUUID(),
			TenantID:     params.TenantID,
			BranchID:     params.BranchID,
			MembershipID: params.MembershipID,
			GuardianID:   params.GuardianID,
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
			EntityType: "parent_membership_guardian_mapping",
			EntityID:   created.ID,
			Details: map[string]any{
				"membership_id": params.MembershipID.String(),
				"guardian_id":   params.GuardianID.String(),
			},
		}); err != nil {
			return err
		}

		result = created
		return nil
	})

	return result, err
}
