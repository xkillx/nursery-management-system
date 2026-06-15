package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type AuditWriter interface {
	WriteWithTx(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, params audit.WriteParams) error
}

type UpsertProfileParams struct {
	BillingMonth           string
	FundedAllowanceMinutes int
}

type UpsertResult struct {
	Profile domain.FundingProfile
	Created bool
}

type UpsertProfile struct {
	repo  domain.Repository
	txm   *transaction.Manager
	audit AuditWriter
}

func NewUpsertProfile(repo domain.Repository, txm *transaction.Manager, auditWriter AuditWriter) *UpsertProfile {
	return &UpsertProfile{repo: repo, txm: txm, audit: auditWriter}
}

func (uc *UpsertProfile) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string, params UpsertProfileParams) (UpsertResult, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return UpsertResult{}, domainerrors.Validation("Invalid child ID.", "child_id")
	}

	billingMonth, err := ParseBillingMonth(params.BillingMonth)
	if err != nil {
		return UpsertResult{}, domainerrors.Validation("Invalid billing month. Must be YYYY-MM.", "billing_month")
	}

	if params.FundedAllowanceMinutes < 0 || params.FundedAllowanceMinutes > 44640 {
		return UpsertResult{}, domainerrors.Validation("Funded allowance must be between 0 and 44640 minutes.", "funded_allowance_minutes")
	}

	var result UpsertResult

	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		enrollment, found, err := uc.repo.GetChildEnrollmentForUpdate(ctx, tx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(err)
		}
		if !found {
			return domainerrors.NotFound("child", "Child not found.")
		}

		if !validateMonthOverlap(billingMonth, enrollment) {
			return domainerrors.Conflict("funding_month_outside_enrollment_window", "Billing month is outside the child's enrollment window.")
		}

		existing, found, err := uc.repo.GetForUpdate(ctx, tx, actor.TenantID, actor.BranchID, childID, billingMonth)
		if err != nil {
			return domainerrors.Internal(err)
		}

		if !found {
			profile := domain.FundingProfile{
				ID:                     uid.NewUUID(),
				TenantID:               actor.TenantID,
				BranchID:               actor.BranchID,
				ChildID:                childID,
				BillingMonth:           billingMonth,
				FundedAllowanceMinutes: params.FundedAllowanceMinutes,
			}
			created, err := uc.repo.Create(ctx, tx, profile)
			if err != nil {
				return domainerrors.Internal(err)
			}

			if auditErr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "funding_profile_created",
				EntityType: "funding_profile",
				EntityID:   created.ID,
				Details: map[string]any{
					"child_id":                     childID.String(),
					"billing_month":                billingMonth.Format("2006-01"),
					"new_funded_allowance_minutes": params.FundedAllowanceMinutes,
				},
			}); auditErr != nil {
				return domainerrors.Internal(auditErr)
			}

			result = UpsertResult{Profile: created, Created: true}
			return nil
		}

		if existing.FundedAllowanceMinutes == params.FundedAllowanceMinutes {
			result = UpsertResult{Profile: existing, Created: false}
			return nil
		}

		previous := existing.FundedAllowanceMinutes
		updated, err := uc.repo.UpdateAllowance(ctx, tx, actor.TenantID, actor.BranchID, childID, billingMonth, params.FundedAllowanceMinutes)
		if err != nil {
			return domainerrors.Internal(err)
		}

		if auditErr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "funding_profile_updated",
			EntityType: "funding_profile",
			EntityID:   updated.ID,
			Details: map[string]any{
				"child_id":                          childID.String(),
				"billing_month":                     billingMonth.Format("2006-01"),
				"previous_funded_allowance_minutes": previous,
				"new_funded_allowance_minutes":      params.FundedAllowanceMinutes,
			},
		}); auditErr != nil {
			return domainerrors.Internal(auditErr)
		}

		result = UpsertResult{Profile: updated, Created: false}
		return nil
	})

	return result, err
}
