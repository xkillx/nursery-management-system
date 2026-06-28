package application

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type CompleteInvoiceRun struct {
	repo domain.BillingRepository
}

func NewCompleteInvoiceRun(repo domain.BillingRepository) *CompleteInvoiceRun {
	return &CompleteInvoiceRun{repo: repo}
}

type CompleteInvoiceRunInput struct {
	Tx              pgx.Tx
	Actor           tenant.ActorContext
	RunID           uuid.UUID
	BillingMonthRaw string
	Generated       []domain.DraftGenerationChildResult
	Blocked         []domain.DraftGenerationBlockedChild
	IsFullMonth     bool
	ChildIDs        []uuid.UUID
}

func (uc *CompleteInvoiceRun) Execute(ctx context.Context, in CompleteInvoiceRunInput) (domain.DraftGenerationResult, error) {
	sort.Slice(in.Generated, func(i, j int) bool {
		if in.Generated[i].ChildFirstName != in.Generated[j].ChildFirstName {
			return in.Generated[i].ChildFirstName < in.Generated[j].ChildFirstName
		}
		if stringPtrValue(in.Generated[i].ChildMiddleName) != stringPtrValue(in.Generated[j].ChildMiddleName) {
			return stringPtrValue(in.Generated[i].ChildMiddleName) < stringPtrValue(in.Generated[j].ChildMiddleName)
		}
		if stringPtrValue(in.Generated[i].ChildLastName) != stringPtrValue(in.Generated[j].ChildLastName) {
			return stringPtrValue(in.Generated[i].ChildLastName) < stringPtrValue(in.Generated[j].ChildLastName)
		}
		return in.Generated[i].ChildID.String() < in.Generated[j].ChildID.String()
	})

	runStatus := domain.InvoiceRunStatusCompleted
	if len(in.Blocked) > 0 {
		runStatus = domain.InvoiceRunStatusCompletedWithExceptions
	}
	runDetails := map[string]any{
		"mode":            "full_month",
		"billing_month":   in.BillingMonthRaw,
		"generated_count": len(in.Generated),
		"blocked_count":   len(in.Blocked),
	}
	if !in.IsFullMonth {
		runDetails["mode"] = "selected_children"
		runDetails["requested_child_count"] = len(in.ChildIDs)
	}
	if len(in.Blocked) > 0 {
		blockedDetails := make([]map[string]any, 0, len(in.Blocked))
		for _, b := range in.Blocked {
			codes := make([]string, 0, len(b.Blockers))
			for _, bl := range b.Blockers {
				codes = append(codes, string(bl.Code))
			}
			blockedDetails = append(blockedDetails, map[string]any{
				"child_id":          b.ChildID.String(),
				"child_first_name":  b.ChildFirstName,
				"child_middle_name": b.ChildMiddleName,
				"child_last_name":   b.ChildLastName,
				"blockers":          codes,
			})
		}
		runDetails["blocked_children"] = blockedDetails
	}
	detailsJSON, _ := json.Marshal(runDetails)

	if compErr := uc.repo.CompleteInvoiceRun(ctx, in.Tx, domain.InvoiceRunCompleteParams{
		ID:            in.RunID,
		TenantID:      in.Actor.TenantID,
		BranchID:      in.Actor.BranchID,
		Status:        runStatus,
		EligibleCount: len(in.Generated) + len(in.Blocked),
		SuccessCount:  len(in.Generated),
		BlockedCount:  len(in.Blocked),
		Details:       detailsJSON,
	}); compErr != nil {
		return domain.DraftGenerationResult{}, fmt.Errorf("complete invoice run: %w", compErr)
	}

	totalDueSum := 0
	for _, g := range in.Generated {
		totalDueSum += g.TotalDue.Minor()
	}

	result := domain.DraftGenerationResult{
		RunID:        in.RunID,
		BillingMonth: in.BillingMonthRaw,
		RunStatus:    runStatus,
		Generated:    in.Generated,
		Blocked:      in.Blocked,
		Summary: domain.DraftGenerationSummary{
			EligibleCount: len(in.Generated) + len(in.Blocked),
			SuccessCount:  len(in.Generated),
			BlockedCount:  len(in.Blocked),
			TotalDue:      domain.MustGBP(totalDueSum),
		},
	}
	return result, nil
}
