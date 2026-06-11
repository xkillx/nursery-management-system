package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetOfficeChecklist struct {
	repo domain.Repository
}

func NewGetOfficeChecklist(repo domain.Repository) *GetOfficeChecklist {
	return &GetOfficeChecklist{repo: repo}
}

func (uc *GetOfficeChecklist) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string) (domain.OfficeUseChecklistWithChild, domain.OfficeUseCompleteness, error) {
	childID, err := uuid.Parse(strings.TrimSpace(childIDRaw))
	if err != nil {
		return domain.OfficeUseChecklistWithChild{}, domain.OfficeUseCompleteness{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	child, found, err := uc.repo.GetOfficeChildSummary(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.OfficeUseChecklistWithChild{}, domain.OfficeUseCompleteness{}, domainerrors.Internal(fmt.Errorf("get child summary: %w", err))
	}
	if !found {
		return domain.OfficeUseChecklistWithChild{}, domain.OfficeUseCompleteness{}, domainerrors.NotFound("child", "Child not found.")
	}

	checklist, err := uc.repo.GetOfficeChecklistByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.OfficeUseChecklistWithChild{}, domain.OfficeUseCompleteness{}, domainerrors.Internal(fmt.Errorf("get checklist by child: %w", err))
	}

	result := domain.OfficeUseChecklistWithChild{
		Child: child,
	}

	if checklist == nil {
		result.Checklist = domain.DefaultOfficeUseChecklist()
		result.ChecklistExists = false
	} else {
		result.Checklist = checklist
		result.ChecklistExists = true
	}

	completeness := domain.ComputeOfficeUseCompleteness(result.Checklist)
	return result, completeness, nil
}
