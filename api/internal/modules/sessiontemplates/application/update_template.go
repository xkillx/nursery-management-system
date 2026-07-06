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

type UpdateSessionTemplateParams struct {
	Name        *string
	Description *string
	Entries     *[]SessionTemplateEntryInput
}

type UpdateSessionTemplate struct {
	repo          domain.Repository
	siteChecker   SiteExistsChecker
	sessionLookup SessionTypeLookup
	txMgr         TxManager
	audit         *audit.Writer
}

func NewUpdateSessionTemplate(repo domain.Repository, siteChecker SiteExistsChecker, lookup SessionTypeLookup, txMgr TxManager, auditWriter *audit.Writer) *UpdateSessionTemplate {
	return &UpdateSessionTemplate{repo: repo, siteChecker: siteChecker, sessionLookup: lookup, txMgr: txMgr, audit: auditWriter}
}

func (uc *UpdateSessionTemplate) Execute(ctx context.Context, actor SessionTemplateActor, siteID, templateID uuid.UUID, params UpdateSessionTemplateParams) (domain.SessionTemplate, error) {
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

	existing, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, templateID)
	if err != nil {
		return domain.SessionTemplate{}, err
	}

	fields := make(map[string]any)

	if params.Name != nil {
		name, err := parseName(*params.Name)
		if err != nil {
			return domain.SessionTemplate{}, err
		}
		if name != existing.Name {
			conflict, cerr := uc.repo.ActiveNameExists(ctx, actor.TenantID(), siteID, name, &templateID)
			if cerr != nil {
				return domain.SessionTemplate{}, internalError(cerr)
			}
			if conflict {
				return domain.SessionTemplate{}, domainerrors.Conflict("session_template_name_duplicate", "An active session template with this name already exists in this site.")
			}
			fields["name"] = name
		}
	}

	if params.Description != nil {
		desc, err := parseDescription(params.Description)
		if err != nil {
			return domain.SessionTemplate{}, err
		}
		fields["description"] = desc
	}

	var resolved []domain.SessionTemplateEntry
	if params.Entries != nil {
		if len(*params.Entries) == 0 {
			return domain.SessionTemplate{}, domainerrors.Validation("Invalid request payload.", "entries")
		}
		seen := make(map[SessionTemplateEntryInput]struct{}, len(*params.Entries))
		resolved = make([]domain.SessionTemplateEntry, 0, len(*params.Entries))
		for _, e := range *params.Entries {
			if e.DayOfWeek < 1 || e.DayOfWeek > 5 {
				return domain.SessionTemplate{}, domainerrors.Validation("Invalid request payload.", "day_of_week")
			}
			if e.SessionTypeID == uuid.Nil {
				return domain.SessionTemplate{}, domainerrors.Validation("Invalid request payload.", "session_type_id")
			}
			key := SessionTemplateEntryInput{DayOfWeek: e.DayOfWeek, SessionTypeID: e.SessionTypeID}
			if _, dup := seen[key]; dup {
				return domain.SessionTemplate{}, domainerrors.New("session_template_duplicate_entry", "Invalid request payload.", "entries")
			}
			seen[key] = struct{}{}
			info, found, lerr := uc.sessionLookup.GetActiveInScope(ctx, actor.TenantID(), siteID, e.SessionTypeID)
			if lerr != nil {
				return domain.SessionTemplate{}, internalError(fmt.Errorf("lookup session type: %w", lerr))
			}
			if !found {
				return domain.SessionTemplate{}, domainerrors.Forbidden("session_type_not_in_branch", "Invalid request payload.")
			}
			if !info.IsActive {
				return domain.SessionTemplate{}, domainerrors.New("session_type_archived", "Invalid request payload.", "session_type_id")
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
	}

	if len(fields) == 0 && params.Entries == nil {
		return existing, nil
	}

	if err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if len(fields) > 0 {
			if _, uerr := uc.repo.Update(ctx, actor.TenantID(), siteID, templateID, fields); uerr != nil {
				return internalError(uerr)
			}
		}
		if resolved != nil {
			if derr := uc.repo.DeleteEntriesByTemplate(ctx, tx, actor.TenantID(), siteID, templateID); derr != nil {
				return internalError(derr)
			}
			for i := range resolved {
				e := &resolved[i]
				e.TemplateID = templateID
				e.TenantID = actor.TenantID()
				e.BranchID = siteID
				if ierr := uc.repo.InsertEntry(ctx, tx, *e); ierr != nil {
					return internalError(ierr)
				}
			}
		}
		if uc.audit != nil {
			details := map[string]any{}
			if v, ok := fields["name"]; ok {
				details["name"] = v
			}
			if v, ok := fields["description"]; ok {
				if d, ok := v.(*string); ok {
					if d == nil {
						details["description"] = nil
					} else {
						details["description"] = *d
					}
				}
			}
			if resolved != nil {
				details["entry_count"] = len(resolved)
			}
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_template_updated",
				EntityType: "session_template",
				EntityID:   templateID,
				Details:    details,
			}); err != nil {
				return internalError(err)
			}
		}
		return nil
	}); err != nil {
		return domain.SessionTemplate{}, err
	}

	updated, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, templateID)
	if err != nil {
		return domain.SessionTemplate{}, err
	}
	entries, eerr := uc.repo.EntriesListByTemplate(ctx, actor.TenantID(), siteID, templateID)
	if eerr != nil {
		return domain.SessionTemplate{}, internalError(eerr)
	}
	updated.Entries = entries
	return updated, nil
}
