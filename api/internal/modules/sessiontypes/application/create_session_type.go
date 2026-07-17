package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateSessionTypeParams struct {
	Name      string
	StartTime string
	EndTime   string
}

type CreateSessionType struct {
	repo        domain.Repository
	siteChecker SiteExistsChecker
	txMgr       TxManager
	audit       *audit.Writer
}

func NewCreateSessionType(repo domain.Repository, siteChecker SiteExistsChecker, txMgr TxManager, auditWriter *audit.Writer) *CreateSessionType {
	return &CreateSessionType{repo: repo, siteChecker: siteChecker, txMgr: txMgr, audit: auditWriter}
}

func (uc *CreateSessionType) Execute(ctx context.Context, actor SessionTypeActor, siteID uuid.UUID, params CreateSessionTypeParams) (domain.SessionType, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "name")
	}
	if len(name) > 255 {
		return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "name")
	}

	startMinutes, err := parseHHMM(params.StartTime)
	if err != nil {
		return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "start_time")
	}
	endMinutes, err := parseHHMM(params.EndTime)
	if err != nil {
		return domain.SessionType{}, domainerrors.Validation("Invalid request payload.", "end_time")
	}
	if startMinutes >= endMinutes {
		return domain.SessionType{}, domainerrors.New("session_type_invalid_time_order", "Invalid request payload.", "start_time")
	}

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

	exists, err := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, name, nil)
	if err != nil {
		return domain.SessionType{}, internalError(err)
	}
	if exists {
		return domain.SessionType{}, domainerrors.Conflict("session_type_name_duplicate", "An active session type with this name already exists in this site.")
	}

	st := domain.SessionType{
		ID:           uid.NewUUID(),
		TenantID:     actor.TenantID(),
		BranchID:     siteID,
		Name:         name,
		StartMinutes: startMinutes,
		EndMinutes:   endMinutes,
		IsActive:     true,
	}

	if err := uc.repo.Create(ctx, st); err != nil {
		return domain.SessionType{}, internalError(err)
	}

	if uc.audit != nil {
		if err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
			return uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_type_created",
				EntityType: "session_type",
				EntityID:   st.ID,
				Details: map[string]any{
					"name":       name,
					"start_time": startMinutes,
					"end_time":   endMinutes,
				},
			})
		}); err != nil {
			return domain.SessionType{}, internalError(err)
		}
	}

	created, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, st.ID)
	if err != nil {
		return domain.SessionType{}, internalError(err)
	}

	return created, nil
}
