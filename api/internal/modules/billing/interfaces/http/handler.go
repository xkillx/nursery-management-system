package httpbilling

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/billing/application"
	"nursery-management-system/api/internal/modules/billing/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	preflight  *application.PreflightDraftInvoices
	generation *application.GenerateDraftInvoices
}

func NewHandler(preflight *application.PreflightDraftInvoices, generation *application.GenerateDraftInvoices) *Handler {
	return &Handler{preflight: preflight, generation: generation}
}

func (h *Handler) RegisterRoutes(manager *gin.RouterGroup) {
	manager.GET("/invoices/drafts/preflight", h.preflightHandler)
	manager.POST("/invoice-runs/drafts", h.generateDraftsHandler)
}

func (h *Handler) preflightHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	billingMonth := strings.TrimSpace(c.Query("billing_month"))
	if billingMonth == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "Missing billing_month query parameter.")
		return
	}

	result, err := h.preflight.Execute(c.Request.Context(), actor, billingMonth)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPreflightResponse(result))
}

func (h *Handler) generateDraftsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req generateDraftsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request body.")
		return
	}

	req.BillingMonth = strings.TrimSpace(req.BillingMonth)
	if req.BillingMonth == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "Missing billing_month.")
		return
	}

	result, err := h.generation.Execute(c.Request.Context(), actor, req.BillingMonth, req.ChildIDs)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toGenerateDraftsResponse(result))
}

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	c.AbortWithStatusJSON(status, resp)
}

func writeError(c *gin.Context, status int, code, message string) {
	requestID := httpserver.RequestIDFromContext(c)
	c.AbortWithStatusJSON(status, httpserver.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}

func toPreflightResponse(r domain.PreflightResult) preflightResponse {
	eligible := make([]eligibleChildResponse, 0, len(r.EligibleChildren))
	for _, ec := range r.EligibleChildren {
		var existingInvoice *existingInvoiceRef
		if ec.ExistingInvoice != nil {
			existingInvoice = &existingInvoiceRef{
				ID:     ec.ExistingInvoice.ID.String(),
				Status: ec.ExistingInvoice.Status,
			}
		}
		var fundingProfileID *string
		if ec.FundingProfileID != nil {
			s := ec.FundingProfileID.String()
			fundingProfileID = &s
		}
		eligible = append(eligible, eligibleChildResponse{
			ChildID:                ec.ChildID.String(),
			ChildName:              ec.ChildName,
			CoreHourlyRateMinor:    ec.CoreHourlyRateMinor,
			FundingProfileID:       fundingProfileID,
			FundedAllowanceMinutes: ec.FundedAllowanceMinutes,
			RawAttendedMinutes:     ec.RawAttendedMinutes,
			RoundedAttendedMinutes: ec.RoundedAttendedMinutes,
			IncludedSessionCount:   ec.IncludedSessionCount,
			FundedDeductionMinutes: ec.FundedDeductionMinutes,
			CoreBillableMinutes:    ec.CoreBillableMinutes,
			SubtotalMinor:          ec.SubtotalMinor,
			FundedDeductionMinor:   ec.FundedDeductionMinor,
			TotalDueMinor:          ec.TotalDueMinor,
			ExistingInvoice:        existingInvoice,
		})
	}

	blocked := make([]blockedChildResponse, 0, len(r.BlockedChildren))
	for _, bc := range r.BlockedChildren {
		blockers := make([]blockerResponse, 0, len(bc.Blockers))
		for _, b := range bc.Blockers {
			br := blockerResponse{
				Code:    string(b.Code),
				Message: b.Message,
			}
			if b.SessionID != nil {
				br.SessionID = strPtr(b.SessionID.String())
			}
			if b.CheckInAt != nil {
				br.CheckInAt = strPtr(b.CheckInAt.UTC().Format("2006-01-02T15:04:05Z"))
			}
			if b.CheckInLocalDate != nil {
				br.CheckInLocalDate = b.CheckInLocalDate
			}
			if b.InvoiceID != nil {
				br.InvoiceID = strPtr(b.InvoiceID.String())
			}
			if b.InvoiceStatus != nil {
				br.InvoiceStatus = b.InvoiceStatus
			}
			if b.Field != nil {
				br.Field = b.Field
			}
			blockers = append(blockers, br)
		}
		blocked = append(blocked, blockedChildResponse{
			ChildID:   bc.ChildID.String(),
			ChildName: bc.ChildName,
			Blockers:  blockers,
		})
	}

	blockerCounts := make([]blockerCountResponse, 0, len(r.Summary.BlockerCounts))
	for _, bc := range r.Summary.BlockerCounts {
		blockerCounts = append(blockerCounts, blockerCountResponse{
			Code:          string(bc.Code),
			ChildrenCount: bc.ChildrenCount,
		})
	}

	return preflightResponse{
		BillingMonth: r.BillingMonth,
		CurrencyCode: r.CurrencyCode,
		Period: periodResponse{
			StartDate:        r.Period.StartDate,
			EndDate:          r.Period.EndDate,
			EndExclusiveDate: r.Period.EndExclusiveDate,
		},
		Summary: summaryResponse{
			TotalChildrenCount:     r.Summary.TotalChildrenCount,
			EligibleChildrenCount:  r.Summary.EligibleChildrenCount,
			BlockedChildrenCount:   r.Summary.BlockedChildrenCount,
			IncludedSessionCount:   r.Summary.IncludedSessionCount,
			RawAttendedMinutes:     r.Summary.RawAttendedMinutes,
			RoundedAttendedMinutes: r.Summary.RoundedAttendedMinutes,
			FundedAllowanceMinutes: r.Summary.FundedAllowanceMinutes,
			FundedDeductionMinutes: r.Summary.FundedDeductionMinutes,
			CoreBillableMinutes:    r.Summary.CoreBillableMinutes,
			SubtotalMinor:          r.Summary.SubtotalMinor,
			FundedDeductionMinor:   r.Summary.FundedDeductionMinor,
			TotalDueMinor:          r.Summary.TotalDueMinor,
			BlockerCounts:          blockerCounts,
		},
		EligibleChildren: eligible,
		BlockedChildren:  blocked,
	}
}

func toGenerateDraftsResponse(r domain.DraftGenerationResult) generateDraftsResponse {
	generated := make([]generatedDraftResponse, 0, len(r.Generated))
	for _, g := range r.Generated {
		generated = append(generated, generatedDraftResponse{
			ChildID:              g.ChildID.String(),
			ChildName:            g.ChildName,
			Action:               string(g.Action),
			InvoiceID:            g.InvoiceID.String(),
			SubtotalMinor:        g.SubtotalMinor,
			FundedDeductionMinor: g.FundedDeductionMinor,
			TotalDueMinor:        g.TotalDueMinor,
		})
	}

	blocked := make([]generateBlockedChildResponse, 0, len(r.Blocked))
	for _, b := range r.Blocked {
		blockers := make([]generateBlockerResponse, 0, len(b.Blockers))
		for _, bl := range b.Blockers {
			blockers = append(blockers, generateBlockerResponse{
				Code:    string(bl.Code),
				Message: bl.Message,
			})
		}
		blocked = append(blocked, generateBlockedChildResponse{
			ChildID:   b.ChildID.String(),
			ChildName: b.ChildName,
			Blockers:  blockers,
		})
	}

	return generateDraftsResponse{
		RunID:        r.RunID.String(),
		BillingMonth: r.BillingMonth,
		Status:       r.RunStatus,
		Summary: generateDraftsSummary{
			EligibleCount: r.Summary.EligibleCount,
			SuccessCount:  r.Summary.SuccessCount,
			BlockedCount:  r.Summary.BlockedCount,
			TotalDueMinor: r.Summary.TotalDueMinor,
		},
		Generated: generated,
		Blocked:   blocked,
	}
}

func strPtr(s string) *string { return &s }
