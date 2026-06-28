package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type UpdateChildParams struct {
	FirstName   string
	MiddleName  *string
	LastName    *string
	DateOfBirth string
	StartDate   string
	EndDate     string
	Notes       string
}

type UpdateChild struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr *transaction.Manager
}

func NewUpdateChild(repo domain.Repository, auditWriter *audit.Writer, txMgr *transaction.Manager) *UpdateChild {
	return &UpdateChild{repo: repo, audit: auditWriter, txMgr: txMgr}
}

func (uc *UpdateChild) Execute(ctx context.Context, actor tenant.ActorContext, childID string, params UpdateChildParams) (domain.Child, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	fields := make(map[string]any)

	if params.FirstName != "" {
		firstName := strings.TrimSpace(params.FirstName)
		if firstName == "" {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "first_name")
		}
		fields["first_name"] = firstName
	}

	if params.MiddleName != nil {
		fields["middle_name"] = strings.TrimSpace(*params.MiddleName)
	}

	if params.LastName != nil {
		fields["last_name"] = strings.TrimSpace(*params.LastName)
	}

	if params.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", strings.TrimSpace(params.DateOfBirth))
		if err != nil {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "date_of_birth")
		}
		fields["date_of_birth"] = dob
	}

	if params.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", strings.TrimSpace(params.StartDate))
		if err != nil {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "start_date")
		}
		fields["start_date"] = startDate
	}

	if params.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", strings.TrimSpace(params.EndDate))
		if err != nil {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "end_date")
		}
		fields["end_date"] = endDate
	}

	if params.Notes != "" {
		notes := strings.TrimSpace(params.Notes)
		fields["notes"] = notes
	}

	if len(fields) == 0 {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "body")
	}

	var updated domain.Child

	err = uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		rowsAffected, err := uc.repo.UpdateWithTx(ctx, tx, actor.TenantID, actor.BranchID, id, fields)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("update child: %w", err))
		}
		if rowsAffected == 0 {
			return domainerrors.NotFound("child", "Resource not found.")
		}

		if err := uc.audit.Write(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_updated",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_updated: %w", err))
		}

		return nil
	})
	if err != nil {
		return domain.Child{}, err
	}

	updated, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil || !found {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("fetch updated child: %w", err))
	}

	return updated, nil
}
