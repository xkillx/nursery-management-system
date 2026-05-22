package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/guardians/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/lifecycle"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

const cascadeReasonGuardianDeactNote = "guardian_deactivated_cascade"

type DeactivateGuardianParams struct {
	ReasonCode string
	ReasonNote string
}

type DeactivateGuardian struct {
	repo  domain.Repository
	txm   *transaction.Manager
	audit *audit.Writer
}

func NewDeactivateGuardian(repo domain.Repository, txm *transaction.Manager, auditWriter *audit.Writer) *DeactivateGuardian {
	return &DeactivateGuardian{repo: repo, txm: txm, audit: auditWriter}
}

func (uc *DeactivateGuardian) Execute(ctx context.Context, actor tenant.ActorContext, guardianID uuid.UUID, params DeactivateGuardianParams) (domain.Guardian, error) {
	if err := validateReason(params); err != nil {
		return domain.Guardian{}, err
	}

	var result domain.Guardian

	err := uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		guardian, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, guardianID)
		if err != nil {
			return fmt.Errorf("fetch guardian for update: %w", err)
		}

		if !guardian.IsActive {
			result = guardian
			return nil
		}

		if err := uc.repo.Deactivate(ctx, tx, actor.TenantID, actor.BranchID, guardianID, params.ReasonCode, params.ReasonNote); err != nil {
			return fmt.Errorf("deactivate guardian: %w", err)
		}

		if err := uc.repo.CascadeLinks(ctx, tx, actor.TenantID, actor.BranchID, guardianID, lifecycle.ReasonAccessRevoked, cascadeReasonGuardianDeactNote); err != nil {
			return fmt.Errorf("cascade links: %w", err)
		}

		if err := uc.repo.CascadeMappings(ctx, tx, actor.TenantID, actor.BranchID, guardianID, lifecycle.ReasonAccessRevoked, cascadeReasonGuardianDeactNote); err != nil {
			return fmt.Errorf("cascade mappings: %w", err)
		}

		reasonCode := params.ReasonCode
		reasonNotePtr := nullableReasonNote(params.ReasonNote)
		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "guardian_deactivated",
			EntityType: "guardian",
			EntityID:   guardianID,
			ReasonCode: &reasonCode,
			ReasonNote: reasonNotePtr,
			Details:    map[string]any{},
		}); err != nil {
			return fmt.Errorf("audit guardian_deactivated: %w", err)
		}

		result, err = uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, guardianID)
		if err != nil {
			return fmt.Errorf("fetch deactivated guardian: %w", err)
		}

		return nil
	})

	if err != nil {
		return domain.Guardian{}, err
	}

	return result, nil
}

func validateReason(params DeactivateGuardianParams) error {
	code := strings.TrimSpace(params.ReasonCode)
	if code == "" {
		return fmt.Errorf("reason_code is required")
	}
	if !lifecycle.IsValidReasonCode(code) {
		return fmt.Errorf("invalid reason_code")
	}
	note := strings.TrimSpace(params.ReasonNote)
	if len(note) > lifecycle.MaxReasonNoteLen {
		return fmt.Errorf("reason_note too long")
	}
	if code == lifecycle.ReasonOther && note == "" {
		return fmt.Errorf("reason_note required for other")
	}
	return nil
}

func nullableReasonNote(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}
