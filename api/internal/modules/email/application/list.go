package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
)

type ListEmails struct {
	repo domain.OutboxRepository
}

func NewListEmails(repo domain.OutboxRepository) *ListEmails {
	return &ListEmails{repo: repo}
}

func (uc *ListEmails) Execute(ctx context.Context, tenantID, branchID uuid.UUID, status *string, limit, offset int) ([]domain.OutboxMessage, int, error) {
	emails, total, err := uc.repo.List(ctx, tenantID, branchID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list emails: %w", err)
	}
	return emails, total, nil
}
