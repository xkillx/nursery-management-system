package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/guardians/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateGuardianParams struct {
	FullName string
	Email    string
	Phone    string
	Notes    string
}

type CreateGuardian struct {
	repo  domain.Repository
	audit *audit.Writer
	pool  *pgxpool.Pool
}

func NewCreateGuardian(repo domain.Repository, auditWriter *audit.Writer, pool *pgxpool.Pool) *CreateGuardian {
	return &CreateGuardian{repo: repo, audit: auditWriter, pool: pool}
}

func (uc *CreateGuardian) Execute(ctx context.Context, actor tenant.ActorContext, params CreateGuardianParams) (domain.Guardian, error) {
	params.FullName = strings.TrimSpace(params.FullName)
	if params.FullName == "" {
		return domain.Guardian{}, fmt.Errorf("full_name is required")
	}

	guardian := domain.Guardian{
		ID:       uid.NewUUID(),
		TenantID: actor.TenantID,
		BranchID: actor.BranchID,
		FullName: params.FullName,
	}

	email := strings.TrimSpace(params.Email)
	if email != "" {
		guardian.Email = &email
	}
	phone := strings.TrimSpace(params.Phone)
	if phone != "" {
		guardian.Phone = &phone
	}
	notes := strings.TrimSpace(params.Notes)
	if notes != "" {
		guardian.Notes = &notes
	}

	if err := uc.repo.Create(ctx, guardian); err != nil {
		return domain.Guardian{}, fmt.Errorf("create guardian: %w", err)
	}

	created, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, guardian.ID)
	if err != nil {
		return domain.Guardian{}, fmt.Errorf("fetch created guardian: %w", err)
	}

	if err := uc.audit.Write(ctx, uc.pool, actor, audit.WriteParams{
		ActionType: "guardian_created",
		EntityType: "guardian",
		EntityID:   guardian.ID,
		Details:    map[string]any{},
	}); err != nil {
		return domain.Guardian{}, fmt.Errorf("audit guardian_created: %w", err)
	}

	return created, nil
}
