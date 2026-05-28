package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
)

type UpdateChildParams struct {
	FullName            string
	DateOfBirth         string
	StartDate           string
	EndDate             string
	CoreHourlyRateMinor *int
	Notes               string
}

type UpdateChild struct {
	repo  domain.Repository
	audit *audit.Writer
	pool  *pgxpool.Pool
}

func NewUpdateChild(repo domain.Repository, auditWriter *audit.Writer, pool *pgxpool.Pool) *UpdateChild {
	return &UpdateChild{repo: repo, audit: auditWriter, pool: pool}
}

func (uc *UpdateChild) Execute(ctx context.Context, actor tenant.ActorContext, childID string, params UpdateChildParams) (domain.Child, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	fields := make(map[string]any)

	if params.FullName != "" {
		fullName := strings.TrimSpace(params.FullName)
		if fullName == "" {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "full_name")
		}
		fields["full_name"] = fullName
	}

	if params.DateOfBirth != "" {
		dob, err := parseDate(params.DateOfBirth)
		if err != nil {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "date_of_birth")
		}
		fields["date_of_birth"] = dob
	}

	if params.StartDate != "" {
		startDate, err := parseDate(params.StartDate)
		if err != nil {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "start_date")
		}
		fields["start_date"] = startDate
	}

	if params.EndDate != "" {
		endDate, err := parseDate(params.EndDate)
		if err != nil {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "end_date")
		}
		fields["end_date"] = endDate
	}

	if params.CoreHourlyRateMinor != nil {
		if *params.CoreHourlyRateMinor < 0 {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "core_hourly_rate_minor")
		}
		fields["core_hourly_rate_minor"] = *params.CoreHourlyRateMinor
	}

	if params.Notes != "" {
		notes := strings.TrimSpace(params.Notes)
		fields["notes"] = notes
	}

	if len(fields) == 0 {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "body")
	}

	rowsAffected, err := uc.repo.Update(ctx, actor.TenantID, actor.BranchID, id, fields)
	if err != nil {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("update child: %w", err))
	}
	if rowsAffected == 0 {
		return domain.Child{}, domainerrors.NotFound("child", "Resource not found.")
	}

	updated, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil || !found {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("fetch updated child: %w", err))
	}

	if err := uc.audit.Write(ctx, uc.pool, actor, audit.WriteParams{
		ActionType: "child_updated",
		EntityType: "child",
		EntityID:   id,
		Details:    map[string]any{},
	}); err != nil {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("audit child_updated: %w", err))
	}

	return updated, nil
}

// buildUpdateSetClause converts a fields map into SET clause parts and args for dynamic SQL.
// The first 3 args must be tenantID, branchID, id. Fields are appended starting at argPos.
func buildUpdateSetClause(fields map[string]any, argPos int) (string, []any) {
	setParts := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))

	for _, col := range orderedFieldColumns(fields) {
		val := fields[col]
		switch col {
		case "notes":
			setParts = append(setParts, fmt.Sprintf("notes = NULLIF($%d, '')", argPos))
		default:
			setParts = append(setParts, fmt.Sprintf("%s = $%d", col, argPos))
		}
		args = append(args, val)
		argPos++
	}

	return strings.Join(setParts, ", "), args
}

// orderedFieldColumns returns column names in a deterministic order.
func orderedFieldColumns(fields map[string]any) []string {
	order := []string{"full_name", "date_of_birth", "start_date", "end_date", "core_hourly_rate_minor", "notes"}
	result := make([]string, 0, len(fields))
	for _, col := range order {
		if _, ok := fields[col]; ok {
			result = append(result, col)
		}
	}
	return result
}

// formatDate formats a time.Time as a date string.
func formatDate(v time.Time) string {
	return v.Format("2006-01-02")
}
