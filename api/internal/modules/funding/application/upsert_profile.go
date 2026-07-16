package application

import (
	"context"
	"time"

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

type ChildFundingRecordReader interface {
	GetByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildFundingRecordData, error)
}

type ChildFundingRecordData struct {
	FundingType        *string
	FundingModel       *string
	FundedHoursPerWeek *float64
	FundingStartDate   *string
	FundingEndDate     *string
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
	repo     domain.Repository
	txm      *transaction.Manager
	audit    AuditWriter
	fundingR ChildFundingRecordReader
	historyR domain.HistoryRepository
}

func NewUpsertProfile(repo domain.Repository, txm *transaction.Manager, auditWriter AuditWriter, fundingReader ChildFundingRecordReader, historyRepo domain.HistoryRepository) *UpsertProfile {
	return &UpsertProfile{repo: repo, txm: txm, audit: auditWriter, fundingR: fundingReader, historyR: historyRepo}
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

		// Read child's funding record for denormalization
		fundingRecord, err := uc.fundingR.GetByChild(ctx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(err)
		}

		existing, found, err := uc.repo.GetForUpdate(ctx, tx, actor.TenantID, actor.BranchID, childID, billingMonth)
		if err != nil {
			return domainerrors.Internal(err)
		}

		if !found {
			// Determine allowance: auto-calculate for stretched if not explicitly provided
			allowance := params.FundedAllowanceMinutes
			if allowance == 0 && fundingRecord != nil && fundingRecord.FundingModel != nil && *fundingRecord.FundingModel == "stretched" && fundingRecord.FundedHoursPerWeek != nil {
				allowance = domain.CalculateStretchedFundedAllowanceMinutes(*fundingRecord.FundedHoursPerWeek)
			}

			profile := domain.FundingProfile{
				ID:                     uid.NewUUID(),
				TenantID:               actor.TenantID,
				BranchID:               actor.BranchID,
				ChildID:                childID,
				BillingMonth:           billingMonth,
				FundedAllowanceMinutes: allowance,
				FundingType:            fundingRecord.FundingType,
				FundingModel:           fundingRecord.FundingModel,
				FundedHoursPerWeek:     fundingRecord.FundedHoursPerWeek,
			}
			created, err := uc.repo.Create(ctx, tx, profile)
			if err != nil {
				return domainerrors.Internal(err)
			}

			// Write funding history
			if err := uc.writeHistory(ctx, actor, childID, fundingRecord, created.FundedAllowanceMinutes); err != nil {
				return err
			}

			if auditErr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "funding_profile_created",
				EntityType: "funding_profile",
				EntityID:   created.ID,
				Details: map[string]any{
					"child_id":                     childID.String(),
					"billing_month":                billingMonth.Format("2006-01"),
					"new_funded_allowance_minutes": created.FundedAllowanceMinutes,
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

		// Write funding history for update
		if err := uc.writeHistory(ctx, actor, childID, fundingRecord, updated.FundedAllowanceMinutes); err != nil {
			return err
		}

		if auditErr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "funding_profile_updated",
			EntityType: "funding_profile",
			EntityID:   updated.ID,
			Details: map[string]any{
				"child_id":                          childID.String(),
				"billing_month":                     billingMonth.Format("2006-01"),
				"previous_funded_allowance_minutes": previous,
				"new_funded_allowance_minutes":      updated.FundedAllowanceMinutes,
			},
		}); auditErr != nil {
			return domainerrors.Internal(auditErr)
		}

		result = UpsertResult{Profile: updated, Created: false}
		return nil
	})

	return result, err
}

func (uc *UpsertProfile) writeHistory(ctx context.Context, actor tenant.ActorContext, childID uuid.UUID, fundingRecord *ChildFundingRecordData, allowanceMinutes int) error {
	history := domain.FundingHistory{
		ID:              uid.NewUUID(),
		TenantID:        actor.TenantID,
		BranchID:        actor.BranchID,
		ChildID:         childID,
		ChangedAt:       time.Now(),
		ChangedByUserID: actor.UserID,
	}

	if fundingRecord != nil {
		history.FundingType = fundingRecord.FundingType
		history.FundingModel = fundingRecord.FundingModel
		history.FundedHoursPerWeek = fundingRecord.FundedHoursPerWeek
		if fundingRecord.FundingStartDate != nil {
			t, _ := time.Parse("2006-01-02", *fundingRecord.FundingStartDate)
			history.FundingStartDate = &t
		}
		if fundingRecord.FundingEndDate != nil {
			t, _ := time.Parse("2006-01-02", *fundingRecord.FundingEndDate)
			history.FundingEndDate = &t
		}
	}

	if err := uc.historyR.Create(ctx, history); err != nil {
		return domainerrors.Internal(err)
	}
	return nil
}
