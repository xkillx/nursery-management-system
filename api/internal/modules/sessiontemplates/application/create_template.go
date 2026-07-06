package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

type SessionTemplateEntryInput struct {
	DayOfWeek     int
	SessionTypeID uuid.UUID
}

type CreateSessionTemplateParams struct {
	Name        string
	Description *string
	Entries     []SessionTemplateEntryInput
}

type CreateSessionTemplate struct {
	repo          domain.Repository
	siteChecker   SiteExistsChecker
	sessionLookup SessionTypeLookup
	txMgr         TxManager
	audit         *audit.Writer
}

func NewCreateSessionTemplate(repo domain.Repository, siteChecker SiteExistsChecker, lookup SessionTypeLookup, txMgr TxManager, auditWriter *audit.Writer) *CreateSessionTemplate {
	return &CreateSessionTemplate{repo: repo, siteChecker: siteChecker, sessionLookup: lookup, txMgr: txMgr, audit: auditWriter}
}

func (uc *CreateSessionTemplate) Execute(ctx context.Context, actor SessionTemplateActor, siteID uuid.UUID, params CreateSessionTemplateParams) (domain.SessionTemplate, error) {
	name, err := parseName(params.Name)
	if err != nil {
		return domain.SessionTemplate{}, err
	}
	description, err := parseDescription(params.Description)
	if err != nil {
		return domain.SessionTemplate{}, err
	}
	if len(params.Entries) == 0 {
		return domain.SessionTemplate{}, domainerrors.Validation("Invalid request payload.", "entries")
	}

	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.SessionTemplate{}, err
	}

	if IsOwnerActor(actor) {
		exists, err := uc.siteChecker.SiteExists(ctx, actor.TenantID(), siteID)
		if err != nil {
			return domain.SessionTemplate{}, internalError(err)
		}
		if !exists {
			return domain.SessionTemplate{}, domainerrors.NotFound("site", "Site not found.")
		}
	}

	exists, err := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, name, nil)
	if err != nil {
		return domain.SessionTemplate{}, internalError(err)
	}
	if exists {
		return domain.SessionTemplate{}, domainerrors.Conflict("session_template_name_duplicate", "An active session template with this name already exists in this site.")
	}

	resolved, err := uc.resolveEntries(ctx, actor, siteID, params.Entries)
	if err != nil {
		return domain.SessionTemplate{}, err
	}

	tpl := domain.SessionTemplate{
		ID:          uid.NewUUID(),
		TenantID:    actor.TenantID(),
		BranchID:    siteID,
		Name:        name,
		Description: description,
		IsActive:    true,
	}

	if err := uc.repo.Create(ctx, tpl); err != nil {
		return domain.SessionTemplate{}, internalError(err)
	}

	if err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		for i := range resolved {
			e := &resolved[i]
			e.TemplateID = tpl.ID
			e.TenantID = tpl.TenantID
			e.BranchID = tpl.BranchID
			if err := uc.repo.InsertEntry(ctx, tx, *e); err != nil {
				return internalError(err)
			}
		}
		if uc.audit != nil {
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_template_created",
				EntityType: "session_template",
				EntityID:   tpl.ID,
				Details: map[string]any{
					"name":        name,
					"entry_count": len(resolved),
				},
			}); err != nil {
				return internalError(err)
			}
		}
		return nil
	}); err != nil {
		return domain.SessionTemplate{}, err
	}

	created, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, tpl.ID)
	if err != nil {
		return domain.SessionTemplate{}, err
	}
	entries, eerr := uc.repo.EntriesListByTemplate(ctx, actor.TenantID(), siteID, tpl.ID)
	if eerr != nil {
		return domain.SessionTemplate{}, internalError(eerr)
	}
	created.Entries = entries
	return created, nil
}

func (uc *CreateSessionTemplate) resolveEntries(ctx context.Context, actor SessionTemplateActor, siteID uuid.UUID, entries []SessionTemplateEntryInput) ([]domain.SessionTemplateEntry, error) {
	seen := make(map[SessionTemplateEntryInput]struct{}, len(entries))
	resolved := make([]domain.SessionTemplateEntry, 0, len(entries))
	for _, e := range entries {
		if e.DayOfWeek < 1 || e.DayOfWeek > 5 {
			return nil, domainerrors.Validation("Invalid request payload.", "day_of_week")
		}
		if e.SessionTypeID == uuid.Nil {
			return nil, domainerrors.Validation("Invalid request payload.", "session_type_id")
		}
		key := SessionTemplateEntryInput{DayOfWeek: e.DayOfWeek, SessionTypeID: e.SessionTypeID}
		if _, dup := seen[key]; dup {
			return nil, domainerrors.New("session_template_duplicate_entry", "Invalid request payload.", "entries")
		}
		seen[key] = struct{}{}

		info, found, err := uc.sessionLookup.GetActiveInScope(ctx, actor.TenantID(), siteID, e.SessionTypeID)
		if err != nil {
			return nil, internalError(fmt.Errorf("lookup session type: %w", err))
		}
		if !found {
			return nil, domainerrors.Forbidden("session_type_not_in_branch", "Invalid request payload.")
		}
		if !info.IsActive {
			return nil, domainerrors.New("session_type_archived", "Invalid request payload.", "session_type_id")
		}
		resolved = append(resolved, domain.SessionTemplateEntry{
			ID:            uid.NewUUID(),
			DayOfWeek:     e.DayOfWeek,
			SessionTypeID: e.SessionTypeID,
			SessionType: &domain.EntrySessionType{
				ID:           info.ID,
				Name:         info.Name,
				StartMinutes: info.StartMinutes,
				EndMinutes:   info.EndMinutes,
				IsActive:     info.IsActive,
			},
		})
	}
	return resolved, nil
}
