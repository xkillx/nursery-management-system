package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
)

type UpdateChildFunding struct {
	repo        domain.FundingRecordRepository
	auditW      *audit.Writer
	historyRepo domain.HistoryRepository
	txMgr       TxManager
}

func NewUpdateChildFunding(repo domain.FundingRecordRepository, auditW *audit.Writer, historyRepo domain.HistoryRepository, txMgr TxManager) *UpdateChildFunding {
	return &UpdateChildFunding{repo: repo, auditW: auditW, historyRepo: historyRepo, txMgr: txMgr}
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
		FundingStartDate:         parseDatePtr(params.FundingStartDate),
		FundingEndDate:           parseDatePtr(params.FundingEndDate),
		EligibilityCode:          params.EligibilityCode,
		EligibilityCodeValidated: params.EligibilityCodeValidated,
		EvidenceReceived:         params.EvidenceReceived,
	}

	var saved domain.FundingRecord
	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		var lerr error
		saved, lerr = uc.repo.UpsertFundingRecord(ctx, tx, record)
		if lerr != nil {
			return fmt.Errorf("upsert funding record: %w", lerr)
		}

		if uc.auditW != nil {
			if aerr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "funding.record.updated",
				EntityType: "child",
				EntityID:   parsedChildID,
				Details: map[string]any{
					"funding_type":    string(params.FundingType),
					"funding_model":   string(params.FundingModel),
					"funding_enabled": params.FundingEnabled,
				},
			}); aerr != nil {
				return fmt.Errorf("audit funding update: %w", aerr)
			}
		}

		return nil
	})
	if txErr != nil {
		return domain.FundingRecord{}, txErr
	}

	return saved, nil
}

func parseDatePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}
