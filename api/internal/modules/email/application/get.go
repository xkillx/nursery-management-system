package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/domain"
)

type GetEmail struct {
	repo domain.OutboxRepository
}

func NewGetEmail(repo domain.OutboxRepository) *GetEmail {
	return &GetEmail{repo: repo}
}

func (uc *GetEmail) Execute(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.OutboxMessage, error) {
	email, err := uc.repo.GetByID(ctx, tenantID, branchID, id)
	if err != nil {
		return domain.OutboxMessage{}, fmt.Errorf("get email: %w", err)
	}
	return email, nil
}
