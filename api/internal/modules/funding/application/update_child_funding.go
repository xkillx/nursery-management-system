package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
)

type UpdateChildFunding struct {
	repo        domain.FundingRecordRepository
	auditW      *audit.Writer
	historyRepo domain.HistoryRepository
}

func NewUpdateChildFunding(repo domain.FundingRecordRepository, auditW *audit.Writer, historyRepo domain.HistoryRepository) *UpdateChildFunding {
	return &UpdateChildFunding{repo: repo, auditW: auditW, historyRepo: historyRepo}
}

type UpdateChildFundingParams struct {
	FundingEnabled           bool
	FundingType              domain.FundingType
	FundingModel             domain.FundingModel
	FundedHoursPerWeek       *float64
	FundingStartDate         *string
	FundingEndDate           *string
	EligibilityCode          *string
	EligibilityCodeValidated bool
	EvidenceReceived         bool
}

func (uc *UpdateChildFunding) Execute(ctx context.Context, actor tenant.ActorContext, childID string, params UpdateChildFundingParams) (domain.FundingRecord, error) {
	parsedChildID, err := uuid.Parse(childID)
	if err != nil {
		return domain.FundingRecord{}, fmt.Errorf("invalid child ID: %w", err)
	}

	record := domain.FundingRecord{
		ID:                       uuid.New(),
		TenantID:                 actor.TenantID,
		BranchID:                 actor.BranchID,
		ChildID:                  parsedChildID,
		FundingEnabled:           params.FundingEnabled,
		FundingType:              params.FundingType,
		FundingModel:             params.FundingModel,
		FundedHoursPerWeek:       params.FundedHoursPerWeek,
		EligibilityCode:          params.EligibilityCode,
		EligibilityCodeValidated: params.EligibilityCodeValidated,
		EvidenceReceived:         params.EvidenceReceived,
	}

	saved, err := uc.repo.UpsertFundingRecord(ctx, nil, record)
	if err != nil {
		return domain.FundingRecord{}, fmt.Errorf("upsert funding record: %w", err)
	}

	if uc.auditW != nil {
		_ = uc.auditW.WriteWithTx(ctx, nil, actor, audit.WriteParams{
			ActionType: "funding.record.updated",
			EntityType: "child",
			EntityID:   parsedChildID,
			Details: map[string]any{
				"funding_type":    string(params.FundingType),
				"funding_model":   string(params.FundingModel),
				"funding_enabled": params.FundingEnabled,
			},
		})
	}

	return saved, nil
}
