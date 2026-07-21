package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
)

type GetChildFunding struct {
	repo domain.FundingRecordRepository
}

func NewGetChildFunding(repo domain.FundingRecordRepository) *GetChildFunding {
	return &GetChildFunding{repo: repo}
}

type GetChildFundingResult struct {
	Record domain.FundingRecord
	Found  bool
}

func (uc *GetChildFunding) Execute(ctx context.Context, tenantID, branchID, childID uuid.UUID) (GetChildFundingResult, error) {
	record, found, err := uc.repo.GetFundingRecord(ctx, tenantID, branchID, childID)
	if err != nil {
		return GetChildFundingResult{}, fmt.Errorf("get funding record: %w", err)
	}
	return GetChildFundingResult{Record: record, Found: found}, nil
}
