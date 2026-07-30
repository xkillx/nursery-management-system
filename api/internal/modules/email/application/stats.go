package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
)

type GetEmailStats struct {
	repo domain.OutboxRepository
}

func NewGetEmailStats(repo domain.OutboxRepository) *GetEmailStats {
	return &GetEmailStats{repo: repo}
}

func (uc *GetEmailStats) Execute(ctx context.Context, tenantID, branchID uuid.UUID) (map[string]int, error) {
	stats, err := uc.repo.GetStats(ctx, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("get email stats: %w", err)
	}
	return stats, nil
}
