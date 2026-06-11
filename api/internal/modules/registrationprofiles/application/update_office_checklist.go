package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type UpdateOfficeChecklist struct {
	repo        domain.Repository
	auditWriter *audit.Writer
	txManager   *transaction.Manager
}

func NewUpdateOfficeChecklist(repo domain.Repository, auditWriter *audit.Writer, txManager *transaction.Manager) *UpdateOfficeChecklist {
	return &UpdateOfficeChecklist{repo: repo, auditWriter: auditWriter, txManager: txManager}
}

type UpdateOfficeChecklistResult struct {
	ChecklistWithChild domain.OfficeUseChecklistWithChild
	Completeness       domain.OfficeUseCompleteness
	Created            bool
}

func (uc *UpdateOfficeChecklist) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string, patch OfficeUseChecklistPatch) (*UpdateOfficeChecklistResult, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	if !hasAnyOfficeField(patch) {
		return nil, domainerrors.Validation("Invalid request payload.", "patch")
	}

	var result *UpdateOfficeChecklistResult

	err = uc.txManager.ExecTx(ctx, func(tx domain.Tx) error {
		child, found, err := uc.repo.GetOfficeChildSummary(ctx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get child summary: %w", err))
		}
		if !found {
			return domainerrors.NotFound("child", "Child not found.")
		}

		existingChecklist, err := uc.repo.GetOfficeChecklistForUpdateByChild(ctx, tx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get checklist for update: %w", err))
		}

		var checklist *domain.OfficeUseChecklist
		created := false

		if existingChecklist == nil {
			checklist = domain.DefaultOfficeUseChecklist()
			checklist.ID = uid.NewUUID()
			checklist.TenantID = actor.TenantID
			checklist.BranchID = actor.BranchID
			checklist.ChildID = childID
			created = true
		} else {
			checklist = existingChecklist
		}

		completenessBefore := domain.ComputeOfficeUseCompleteness(checklist)

		changedFields, err := MergeOfficeChecklistPatch(checklist, patch)
		if err != nil {
			return err
		}

		if len(changedFields) == 0 {
			completeness := domain.ComputeOfficeUseCompleteness(checklist)
			result = &UpdateOfficeChecklistResult{
				ChecklistWithChild: domain.OfficeUseChecklistWithChild{
					Checklist:       checklist,
					Child:           child,
					ChecklistExists: !created,
				},
				Completeness: completeness,
				Created:      false,
			}
			return nil
		}

		if created {
			savedChecklist, createErr := uc.repo.CreateOfficeChecklist(ctx, tx, checklist)
			if createErr != nil {
				return domainerrors.Internal(fmt.Errorf("create checklist: %w", createErr))
			}
			checklist = savedChecklist
		} else {
			savedChecklist, updateErr := uc.repo.UpdateOfficeChecklist(ctx, tx, checklist)
			if updateErr != nil {
				return domainerrors.Internal(fmt.Errorf("update checklist: %w", updateErr))
			}
			checklist = savedChecklist
		}

		completenessAfter := domain.ComputeOfficeUseCompleteness(checklist)

		auditAction := "registration_office_checklist_updated"
		if created {
			auditAction = "registration_office_checklist_created"
		}

		auditDetails := map[string]any{
			"child_id":               childID.String(),
			"checklist_id":           checklist.ID.String(),
			"changed_fields":         changedFields,
			"checklist_created":      created,
			"completeness_before":    officeCompletenessToStrings(completenessBefore),
			"completeness_after":     officeCompletenessToStrings(completenessAfter),
		}

		if auditErr := uc.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: auditAction,
			EntityType: "child_registration_office_checklist",
			EntityID:   checklist.ID,
			Details:    auditDetails,
		}); auditErr != nil {
			return domainerrors.Internal(fmt.Errorf("audit checklist update: %w", auditErr))
		}

		result = &UpdateOfficeChecklistResult{
			ChecklistWithChild: domain.OfficeUseChecklistWithChild{
				Checklist:       checklist,
				Child:           child,
				ChecklistExists: true,
			},
			Completeness: completenessAfter,
			Created:      created,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func hasAnyOfficeField(p OfficeUseChecklistPatch) bool {
	return p.DepositStatus != nil ||
		p.DepositPaidDate != nil ||
		p.ApplicationDateStatus != nil ||
		p.ApplicationDate != nil ||
		p.StartDateStatus != nil ||
		p.DateLeft != nil ||
		p.SessionsDaysRequestedStatus != nil ||
		p.SessionsDaysRequested != nil ||
		p.TermTimeOnlySpaceStatus != nil ||
		p.ContractStatus != nil ||
		p.ContractDate != nil ||
		p.HandbookStatus != nil ||
		p.HandbookDate != nil ||
		p.RedBookStatus != nil ||
		p.RedBookCheckedDate != nil ||
		p.BirthCertificatePassportStatus != nil ||
		p.BirthCertificatePassportCheckedDate != nil ||
		p.ProofOfAddressStatus != nil ||
		p.ProofOfAddressCheckedDate != nil ||
		p.Notes != nil
}

func officeCompletenessToStrings(c domain.OfficeUseCompleteness) []string {
	s := make([]string, len(c.MissingFields))
	for i, mf := range c.MissingFields {
		s[i] = string(mf)
	}
	return s
}
