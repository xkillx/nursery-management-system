package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateAttestation struct {
	profileRepo    domain.Repository
	consentRepo    domain.ConsentRepository
	attestRepo     domain.AttestationRepository
	getWorkflow    *GetWorkflowStatus
	audit          *audit.Writer
	txMgr          *transaction.Manager
}

func NewCreateAttestation(
	profileRepo domain.Repository,
	consentRepo domain.ConsentRepository,
	attestRepo domain.AttestationRepository,
	getWorkflow *GetWorkflowStatus,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *CreateAttestation {
	return &CreateAttestation{
		profileRepo: profileRepo,
		consentRepo: consentRepo,
		attestRepo:  attestRepo,
		getWorkflow: getWorkflow,
		audit:       auditWriter,
		txMgr:       txMgr,
	}
}

func (uc *CreateAttestation) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string) (*domain.CompletionAttestation, error) {
	childID, err := uuid.Parse(strings.TrimSpace(childIDRaw))
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	status, err := uc.getWorkflow.Execute(ctx, actor, childIDRaw)
	if err != nil {
		return nil, err
	}

	if !status.CanMarkComplete {
		return nil, domainerrors.Validation("Registration cannot be marked complete. Missing: "+strings.Join(status.MissingGroups, ", "), "registration_status")
	}

	profile, err := uc.profileRepo.GetByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get profile: %w", err))
	}

	currentConsent, err := uc.consentRepo.GetLatestByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get latest consent: %w", err))
	}

	if profile == nil || currentConsent == nil {
		return nil, domainerrors.Validation("Cannot attest: profile or consent record is missing.", "registration_status")
	}

	var consentRecordID *uuid.UUID
	consentID := currentConsent.ID
	consentRecordID = &consentID

	attestation := &domain.CompletionAttestation{
		ID:                     uid.NewUUID(),
		TenantID:               actor.TenantID,
		BranchID:               actor.BranchID,
		ChildID:                childID,
		ConsentRecordID:        consentRecordID,
		ProfileUpdatedAt:       profile.UpdatedAt,
		AttestedByUserID:       actor.UserID,
		AttestedByMembershipID: actor.MembershipID,
		RequestID:              nil,
	}

	if actor.RequestID != "" {
		attestation.RequestID = &actor.RequestID
	}

	err = uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if err := uc.attestRepo.CreateAttestation(ctx, tx, attestation); err != nil {
			return fmt.Errorf("create attestation: %w", err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "registration_completion_attested",
			EntityType: "child",
			EntityID:   childID,
			Details: map[string]any{
				"consent_record_id":  consentID.String(),
				"completeness_state": "complete",
			},
		}); err != nil {
			return fmt.Errorf("audit attestation: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, domainerrors.Internal(err)
	}

	return attestation, nil
}
