package application

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetContacts struct {
	repo domain.Repository
}

func NewGetContacts(repo domain.Repository) *GetContacts {
	return &GetContacts{repo: repo}
}

func (uc *GetContacts) Execute(ctx context.Context, actor tenant.ActorContext, childID string) ([]domain.ChildContact, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}
	contacts, err := uc.repo.ListContactsByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list child contacts: %w", err))
	}
	return contacts, nil
}

type ReplaceContacts struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   TxManager
}

func NewReplaceContacts(repo domain.Repository, auditWriter *audit.Writer, txm TxManager) *ReplaceContacts {
	return &ReplaceContacts{repo: repo, audit: auditWriter, txm: txm}
}

func (uc *ReplaceContacts) Execute(ctx context.Context, actor tenant.ActorContext, childID string, inputs []ChildContactInput) ([]domain.ChildContact, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	var result []domain.ChildContact
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}

		// Bucket the inputs by type and assign sort_order within each bucket.
		bucketed := map[domain.ContactType][]ChildContactInput{
			domain.ContactTypeEmergencyContact:    {},
			domain.ContactTypeAuthorisedCollector: {},
		}
		for _, in := range inputs {
			if _, ok := bucketed[in.ContactType]; !ok {
				return domainerrors.Validation("Invalid request payload.", "contact_type")
			}
			bucketed[in.ContactType] = append(bucketed[in.ContactType], in)
		}

		types := []domain.ContactType{domain.ContactTypeEmergencyContact, domain.ContactTypeAuthorisedCollector}
		flat := make([]ChildContactInput, 0, len(inputs))
		for _, t := range types {
			flat = append(flat, bucketed[t]...)
		}
		entries := buildChildContactEntries(actor.TenantID, actor.BranchID, id, flat)
		if err := uc.repo.ReplaceContactsForTypes(ctx, tx, actor.TenantID, actor.BranchID, id, types, entries); err != nil {
			return domainerrors.Internal(fmt.Errorf("replace child contacts: %w", err))
		}

		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_contacts_replaced",
			EntityType: "child",
			EntityID:   id,
			Details: map[string]any{
				"contact_count": len(entries),
			},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_contacts_replaced: %w", aerr))
		}

		saved, eerr := uc.repo.ListContactsByChild(ctx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("reload child contacts: %w", eerr))
		}
		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
