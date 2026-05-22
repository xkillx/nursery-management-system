package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/guardians/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
)

type UpdateGuardianParams struct {
	FullName string
	Email    string
	Phone    string
	Notes    string
}

type UpdateGuardian struct {
	repo  domain.Repository
	audit *audit.Writer
	pool  *pgxpool.Pool
}

func NewUpdateGuardian(repo domain.Repository, auditWriter *audit.Writer, pool *pgxpool.Pool) *UpdateGuardian {
	return &UpdateGuardian{repo: repo, audit: auditWriter, pool: pool}
}

func (uc *UpdateGuardian) Execute(ctx context.Context, actor tenant.ActorContext, guardianID uuid.UUID, params UpdateGuardianParams) (domain.Guardian, error) {
	fields := buildUpdateFields(params)
	if len(fields) == 0 {
		return domain.Guardian{}, fmt.Errorf("no fields to update")
	}

	rows, err := uc.repo.Update(ctx, actor.TenantID, actor.BranchID, guardianID, fields)
	if err != nil {
		return domain.Guardian{}, fmt.Errorf("update guardian: %w", err)
	}
	if rows == 0 {
		return domain.Guardian{}, fmt.Errorf("guardian not found")
	}

	updated, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, guardianID)
	if err != nil {
		return domain.Guardian{}, fmt.Errorf("fetch updated guardian: %w", err)
	}

	if err := uc.audit.Write(ctx, uc.pool, actor, audit.WriteParams{
		ActionType: "guardian_updated",
		EntityType: "guardian",
		EntityID:   guardianID,
		Details:    map[string]any{},
	}); err != nil {
		return domain.Guardian{}, fmt.Errorf("audit guardian_updated: %w", err)
	}

	return updated, nil
}

func buildUpdateFields(params UpdateGuardianParams) map[string]any {
	fields := make(map[string]any)

	if params.FullName != "" {
		fullName := strings.TrimSpace(params.FullName)
		if fullName != "" {
			fields["full_name"] = fullName
		}
	}
	if params.Email != "" {
		fields["email"] = strings.TrimSpace(params.Email)
	}
	if params.Phone != "" {
		fields["phone"] = strings.TrimSpace(params.Phone)
	}
	if params.Notes != "" {
		fields["notes"] = strings.TrimSpace(params.Notes)
	}

	return fields
}
