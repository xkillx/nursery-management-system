package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type UpdateSessionTypeParams struct {
	Name      *string
	StartTime *string
	EndTime   *string
	Kind      *string
}

type UpdateSessionType struct {
	repo        domain.Repository
	siteChecker SiteExistsChecker
	txMgr       TxManager
	audit       *audit.Writer
}

func NewUpdateSessionType(repo domain.Repository, siteChecker SiteExistsChecker, txMgr TxManager, auditWriter *audit.Writer) *UpdateSessionType {
	return &UpdateSessionType{repo: repo, siteChecker: siteChecker, txMgr: txMgr, audit: auditWriter}
}

func (uc *UpdateSessionType) Execute(ctx context.Context, actor SessionTypeActor, siteID, stID uuid.UUID, params UpdateSessionTypeParams) (domain.SessionType, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.SessionType{}, err
	}

	if IsOwnerActor(actor) {
		exists, err := uc.siteChecker.SiteExists(ctx, actor.TenantID(), siteID)
		if err != nil {
			return domain.SessionType{}, internalError(err)
		}
		if !exists {
			return domain.SessionType{}, domainerrors.NotFound("site", "Site not found.")
		}
	}

	existing, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, stID)
	if err != nil {
		return domain.SessionType{}, err
	}

	newStart := existing.StartMinutes
	newEnd := existing.EndMinutes
	startSet := false
	endSet := false

	fields := make(map[string]any)

	if params.Name != nil {
		name := strings.TrimSpace(*params.Name)
		if name == "" {
			return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "name")
		}
		if len(name) > 255 {
			return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "name")
		}
		if name != existing.Name {
			exists, err := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, name, &stID)
			if err != nil {
				return domain.SessionType{}, internalError(err)
			}
			if exists {
				return domain.SessionType{}, domainerrors.Conflict("session_type_name_duplicate", "An active session type with this name already exists in this site.")
			}
			fields["name"] = name
		}
	}

	if params.StartTime != nil {
		sm, err := parseHHMM(*params.StartTime)
		if err != nil {
			return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "start_time")
		}
		newStart = sm
		startSet = true
		fields["start_time"] = sm
	}
	if params.EndTime != nil {
		em, err := parseHHMM(*params.EndTime)
		if err != nil {
			return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "end_time")
		}
		newEnd = em
		endSet = true
		fields["end_time"] = em
	}

	if startSet || endSet {
		if newStart >= newEnd {
			return domain.SessionType{}, domainerrors.New("session_type_invalid_time_order", "Invalid request payload.", "start_time")
		}
	}

	if params.Kind != nil {
		if !validSessionTypeKind(*params.Kind) {
			return domain.SessionType{}, domainerrors.Validation("Kind must be standard, wraparound_before, wraparound_after, core, or extended.", "kind")
		}
		fields["kind"] = *params.Kind
	}

	if len(fields) == 0 {
		return existing, nil
	}

	rowsAffected, err := uc.repo.Update(ctx, actor.TenantID(), siteID, stID, fields)
	if err != nil {
		return domain.SessionType{}, internalError(err)
	}
	if rowsAffected == 0 {
		return domain.SessionType{}, domainerrors.NotFound("session_type", "Session type not found.")
	}

	if uc.audit != nil {
		details := make(map[string]any, len(fields))
		for k, v := range fields {
			details[k] = v
		}
		if err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
			return uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_type_updated",
				EntityType: "session_type",
				EntityID:   stID,
				Details:    details,
			})
		}); err != nil {
			return domain.SessionType{}, internalError(err)
		}
	}

	updated, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, stID)
	if err != nil {
		return domain.SessionType{}, internalError(err)
	}

	return updated, nil
}
