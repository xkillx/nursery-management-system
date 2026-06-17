package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type GetBillingProfile struct {
	repo domain.Repository
}

func NewGetBillingProfile(repo domain.Repository) *GetBillingProfile {
	return &GetBillingProfile{repo: repo}
}

func (uc *GetBillingProfile) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildBillingProfile, error) {
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
	bp, found, err := uc.repo.GetBillingProfileByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child billing profile: %w", err))
	}
	if !found {
		return nil, nil
	}
	return bp, nil
}

type UpdateBillingProfile struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewUpdateBillingProfile(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *UpdateBillingProfile {
	return &UpdateBillingProfile{repo: repo, audit: auditWriter, txm: txm}
}

type UpdateBillingProfileInput struct {
	BillingBasis   string
	CustomRateMinor *int
	EffectiveFrom   *string
}

func (uc *UpdateBillingProfile) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in UpdateBillingProfileInput) (*domain.ChildBillingProfile, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in.BillingBasis == "" {
		return nil, domainerrors.Validation("Invalid request payload.", "body")
	}
	basis := domain.BillingBasis(in.BillingBasis)
	if basis != domain.BillingBasisSiteRate && basis != domain.BillingBasisCustom {
		return nil, domainerrors.Validation("Invalid request payload.", "billing_basis")
	}
	if basis == domain.BillingBasisCustom {
		if in.CustomRateMinor == nil || *in.CustomRateMinor <= 0 {
			return nil, domainerrors.Validation("Invalid request payload.", "custom_rate_minor")
		}
	} else if in.CustomRateMinor != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "custom_rate_minor")
	}

	var effectiveFrom time.Time
	if in.EffectiveFrom != nil && *in.EffectiveFrom != "" {
		t, err := time.Parse("2006-01-02", *in.EffectiveFrom)
		if err != nil {
			return nil, domainerrors.Validation("Invalid request payload.", "effective_from")
		}
		effectiveFrom = t
	} else {
		effectiveFrom = time.Now().UTC()
	}

	var result *domain.ChildBillingProfile
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}
		p := &domain.ChildBillingProfile{
			ID:             uid.NewUUID(),
			TenantID:       actor.TenantID,
			BranchID:       actor.BranchID,
			ChildID:        id,
			BillingBasis:   basis,
			CustomRateMinor: in.CustomRateMinor,
			EffectiveFrom:  effectiveFrom,
		}
		saved, eerr := uc.repo.UpsertBillingProfile(ctx, tx, p)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("upsert child billing profile: %w", eerr))
		}
		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_billing_profile_updated",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_billing_profile_updated: %w", aerr))
		}
		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
