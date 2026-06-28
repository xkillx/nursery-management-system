package httpbilling

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/application"
	"nursery-management-system/api/internal/modules/billing/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger                *slog.Logger
	preflight             *application.PreflightDraftInvoices
	generation            *application.GenerateDraftInvoicesUseCase
	listInvoices          *application.ListInvoices
	getInvoice            *application.GetInvoice
	issueInvoice          *application.IssueInvoice
	bulkIssueInvoices     *application.BulkIssueInvoices
	overrideAttendanceBlk *application.OverrideAttendanceBlockUseCase
	listParentInvoices    *application.ListParentInvoices
	getParentInvoice      *application.GetParentInvoice
}

func NewHandler(
	preflight *application.PreflightDraftInvoices,
	generation *application.GenerateDraftInvoicesUseCase,
	listInvoices *application.ListInvoices,
	getInvoice *application.GetInvoice,
	issueInvoice *application.IssueInvoice,
	bulkIssueInvoices *application.BulkIssueInvoices,
	overrideAttendanceBlk *application.OverrideAttendanceBlockUseCase,
	listParentInvoices *application.ListParentInvoices,
	getParentInvoice *application.GetParentInvoice,
) *Handler {
	return &Handler{
		preflight:             preflight,
		generation:            generation,
		listInvoices:          listInvoices,
		getInvoice:            getInvoice,
		issueInvoice:          issueInvoice,
		bulkIssueInvoices:     bulkIssueInvoices,
		overrideAttendanceBlk: overrideAttendanceBlk,
		listParentInvoices:    listParentInvoices,
		getParentInvoice:      getParentInvoice,
	}
}

func (h *Handler) WithObservability(logger *slog.Logger) *Handler {
	return &Handler{
		preflight:             h.preflight,
		generation:            h.generation,
		listInvoices:          h.listInvoices,
		getInvoice:            h.getInvoice,
		issueInvoice:          h.issueInvoice,
		bulkIssueInvoices:     h.bulkIssueInvoices,
		overrideAttendanceBlk: h.overrideAttendanceBlk,
		listParentInvoices:    h.listParentInvoices,
		getParentInvoice:      h.getParentInvoice,
		logger:                logger,
	}
}

func (h *Handler) RegisterRoutes(manager *gin.RouterGroup) {
	manager.GET("/invoices/drafts/preflight", h.preflightHandler)
	manager.GET("/invoices", h.listInvoicesHandler)
	manager.GET("/invoices/:invoice_id", h.getInvoiceHandler)
	manager.POST("/invoice-runs/drafts", h.generateDraftsHandler)
	manager.POST("/invoices/bulk-issue", h.bulkIssueInvoicesHandler)
	manager.POST("/invoices/:invoice_id/issue", h.issueInvoiceHandler)
	manager.POST("/invoices/:invoice_id/override-attendance-block", h.overrideAttendanceBlockHandler)
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.GET("/invoices", h.listParentInvoicesHandler)
	parent.GET("/invoices/:invoice_id", h.getParentInvoiceHandler)
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
		h.handleError(c, err)
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
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toGenerateDraftsResponse(result))
}

func (h *Handler) listInvoicesHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	result, err := h.listInvoices.Execute(c.Request.Context(), actor, application.ListInvoicesParams{
		BillingMonth: queryParamPtr(c, "billing_month"),
		Status:       queryParamPtr(c, "status"),
		ChildID:      queryParamPtr(c, "child_id"),
		Limit:        queryParamPtr(c, "limit"),
		Offset:       queryParamPtr(c, "offset"),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toInvoiceListResponse(result))
}

func (h *Handler) getInvoiceHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	result, err := h.getInvoice.Execute(c.Request.Context(), actor, c.Param("invoice_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toInvoiceDetailResponse(result))
}

func (h *Handler) overrideAttendanceBlockHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	invoiceID, err := uuid.Parse(strings.TrimSpace(c.Param("invoice_id")))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid invoice_id.")
		return
	}

	var req overrideAttendanceBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	result, err := h.overrideAttendanceBlk.Execute(c.Request.Context(), actor, application.OverrideAttendanceBlockInput{
		InvoiceID:    invoiceID,
		BillingMonth: strings.TrimSpace(req.BillingMonth),
		Note:         strings.TrimSpace(req.Note),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toOverrideAttendanceBlockResponse(result))
}

func (h *Handler) issueInvoiceHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req issueInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request body.")
		return
	}

	result, err := h.issueInvoice.Execute(c.Request.Context(), actor, c.Param("invoice_id"), req.Confirm)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toIssueInvoiceResponse(result))
}

func (h *Handler) bulkIssueInvoicesHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req bulkIssueInvoicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request body.")
		return
	}

	invoiceIDsProvided := false
	var rawIDs []string
	if c.Request.ContentLength > 0 {
		rawIDs = req.InvoiceIDs
		invoiceIDsProvided = req.InvoiceIDs != nil
	}

	result, err := h.bulkIssueInvoices.Execute(c.Request.Context(), actor, req.BillingMonth, rawIDs, invoiceIDsProvided, req.Confirm)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBulkIssueResponse(result))
}

func (h *Handler) listParentInvoicesHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	result, err := h.listParentInvoices.Execute(c.Request.Context(), actor, application.ListParentInvoicesParams{
		BillingMonth: queryParamPtr(c, "billing_month"),
		Status:       queryParamPtr(c, "status"),
		ChildID:      queryParamPtr(c, "child_id"),
		Limit:        queryParamPtr(c, "limit"),
		Offset:       queryParamPtr(c, "offset"),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toParentInvoiceListResponse(result))
}

func (h *Handler) getParentInvoiceHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	result, err := h.getParentInvoice.Execute(c.Request.Context(), actor, c.Param("invoice_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toParentInvoiceDetailResponse(result))
}

func queryParamPtr(c *gin.Context, key string) *string {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	return &v
}

func (h *Handler) handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
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
			ChildFirstName:         ec.ChildFirstName,
			ChildMiddleName:        ec.ChildMiddleName,
			ChildLastName:          ec.ChildLastName,
			CoreHourlyRateMinor:    ec.CoreHourlyRate.Minor(),
			FundingProfileID:       fundingProfileID,
			FundedAllowanceMinutes: ec.FundedAllowanceMinutes,
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
			ChildID:         bc.ChildID.String(),
			ChildFirstName:  bc.ChildFirstName,
			ChildMiddleName: bc.ChildMiddleName,
			ChildLastName:   bc.ChildLastName,
			Blockers:        blockers,
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
			FundedAllowanceMinutes: r.Summary.FundedAllowanceMinutes,
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
			ChildFirstName:       g.ChildFirstName,
			ChildMiddleName:      g.ChildMiddleName,
			ChildLastName:        g.ChildLastName,
			Action:               string(g.Action),
			InvoiceID:            g.InvoiceID.String(),
			SubtotalMinor:        g.Subtotal.Minor(),
			FundedDeductionMinor: g.FundedDeduction.Minor(),
			TotalDueMinor:        g.TotalDue.Minor(),
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
			ChildID:         b.ChildID.String(),
			ChildFirstName:  b.ChildFirstName,
			ChildMiddleName: b.ChildMiddleName,
			ChildLastName:   b.ChildLastName,
			Blockers:        blockers,
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
			TotalDueMinor: r.Summary.TotalDue.Minor(),
		},
		Generated: generated,
		Blocked:   blocked,
	}
}

func strPtr(s string) *string { return &s }

func moneyPtrToIntPtr(m *domain.Money) *int {
	if m == nil {
		return nil
	}
	v := m.Minor()
	return &v
}

func invoiceNumberDisplay(status string, invoiceNumber *string) string {
	if invoiceNumber == nil || *invoiceNumber == "" {
		if status == domain.InvoiceStatusDraft {
			return "Draft"
		}
	}
	if invoiceNumber != nil {
		return *invoiceNumber
	}
	return ""
}

func dueStatus(status string) string {
	switch status {
	case domain.InvoiceStatusDraft:
		return "not_due"
	case domain.InvoiceStatusPaid:
		return "paid"
	case domain.InvoiceStatusOverdue:
		return "overdue"
	case domain.InvoiceStatusIssued, domain.InvoiceStatusPaymentFailed:
		return "due"
	default:
		return "not_due"
	}
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format("2006-01-02T15:04:05Z")
	return &s
}

func formatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatBillingMonth(t time.Time) string {
	return t.Format("2006-01")
}

func formatUUIDPtr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

func toInvoiceListResponse(r application.ListInvoicesResult) invoiceListResponse {
	items := make([]invoiceListItemResponse, 0, len(r.Items))
	for _, inv := range r.Items {
		exceptionCount := countRunExceptions(inv.GeneratedRunDetails)
		items = append(items, invoiceListItemResponse{
			InvoiceID:                  inv.ID.String(),
			InvoiceKind:                inv.InvoiceKind,
			InvoiceNumber:              inv.InvoiceNumber,
			InvoiceNumberDisplay:       invoiceNumberDisplay(inv.Status, inv.InvoiceNumber),
			ChildID:                    inv.ChildID.String(),
			ChildFirstName:             inv.ChildFirstName,
			ChildMiddleName:            inv.ChildMiddleName,
			ChildLastName:              inv.ChildLastName,
			BillingMonth:               formatBillingMonth(inv.BillingMonth),
			Status:                     inv.Status,
			DueStatus:                  dueStatus(inv.Status),
			CurrencyCode:               inv.CurrencyCode,
			SubtotalMinor:              inv.Subtotal.Minor(),
			FundedDeductionMinor:       inv.FundedDeduction.Minor(),
			TotalDueMinor:              inv.TotalDue.Minor(),
			AmountPaidMinor:            inv.AmountPaid.Minor(),
			DueAt:                      formatTimePtr(inv.DueAt),
			IssuedAt:                   formatTimePtr(inv.IssuedAt),
			PaidAt:                     formatTimePtr(inv.PaidAt),
			PaymentFailedAt:            formatTimePtr(inv.PaymentFailedAt),
			PaymentStatusUpdatedAt:     formatTimePtr(inv.PaymentStatusUpdatedAt),
			GeneratedRunID:             formatUUIDPtr(inv.GeneratedRunID),
			GeneratedRunStatus:         inv.GeneratedRunStatus,
			GeneratedRunStartedAt:      formatTimePtr(inv.GeneratedRunStartedAt),
			GeneratedRunCompletedAt:    formatTimePtr(inv.GeneratedRunCompletedAt),
			GeneratedRunExceptionCount: exceptionCount,
			CreatedAt:                  formatTime(inv.CreatedAt),
			UpdatedAt:                  formatTime(inv.UpdatedAt),
		})
		item := &items[len(items)-1]
		item.Period.StartDate = formatDate(inv.PeriodStartDate)
		item.Period.EndDate = formatDate(inv.PeriodEndDate)
	}
	return invoiceListResponse{
		Items:  items,
		Limit:  r.Limit,
		Offset: r.Offset,
	}
}

func toInvoiceDetailResponse(r application.GetInvoiceResult) invoiceDetailResponse {
	inv := r.Invoice
	resp := invoiceDetailResponse{
		InvoiceID:                  inv.ID.String(),
		InvoiceKind:                inv.InvoiceKind,
		InvoiceNumber:              inv.InvoiceNumber,
		InvoiceNumberDisplay:       invoiceNumberDisplay(inv.Status, inv.InvoiceNumber),
		ChildID:                    inv.ChildID.String(),
		ChildFirstName:             inv.ChildFirstName,
		ChildMiddleName:            inv.ChildMiddleName,
		ChildLastName:              inv.ChildLastName,
		BillingMonth:               formatBillingMonth(inv.BillingMonth),
		Status:                     inv.Status,
		DueStatus:                  dueStatus(inv.Status),
		CurrencyCode:               inv.CurrencyCode,
		SubtotalMinor:              inv.Subtotal.Minor(),
		FundedDeductionMinor:       inv.FundedDeduction.Minor(),
		TotalDueMinor:              inv.TotalDue.Minor(),
		AmountPaidMinor:            inv.AmountPaid.Minor(),
		IssuedAt:                   formatTimePtr(inv.IssuedAt),
		LockedAt:                   formatTimePtr(inv.LockedAt),
		DueAt:                      formatTimePtr(inv.DueAt),
		PaidAt:                     formatTimePtr(inv.PaidAt),
		PaymentFailedAt:            formatTimePtr(inv.PaymentFailedAt),
		PaymentStatusUpdatedAt:     formatTimePtr(inv.PaymentStatusUpdatedAt),
		AdjustsInvoiceID:           formatUUIDPtr(inv.AdjustsInvoiceID),
		AdjustmentReasonCode:       inv.AdjustmentReasonCode,
		AdjustmentReasonNote:       inv.AdjustmentReasonNote,
		GeneratedRunID:             formatUUIDPtr(inv.GeneratedRunID),
		GeneratedRunStatus:         inv.GeneratedRunStatus,
		GeneratedRunStartedAt:      formatTimePtr(inv.GeneratedRunStartedAt),
		GeneratedRunCompletedAt:    formatTimePtr(inv.GeneratedRunCompletedAt),
		GeneratedRunExceptionCount: r.GeneratedRunExceptionCount,
		Calculation:                toCalculationResponse(r.Calculation),
		CreatedAt:                  formatTime(inv.CreatedAt),
		UpdatedAt:                  formatTime(inv.UpdatedAt),
	}
	resp.Period.StartDate = formatDate(inv.PeriodStartDate)
	resp.Period.EndDate = formatDate(inv.PeriodEndDate)

	resp.GeneratedRunExceptions = make([]invoiceRunExceptionResponse, 0, len(r.GeneratedRunExceptions))
	for _, ex := range r.GeneratedRunExceptions {
		resp.GeneratedRunExceptions = append(resp.GeneratedRunExceptions, invoiceRunExceptionResponse{
			ChildID:         ex.ChildID,
			ChildFirstName:  ex.ChildFirstName,
			ChildMiddleName: ex.ChildMiddleName,
			ChildLastName:   ex.ChildLastName,
			BlockerCodes:    ex.BlockerCodes,
		})
	}

	resp.Lines = make([]invoiceLineResponse, 0, len(r.Lines))
	for _, line := range r.Lines {
		resp.Lines = append(resp.Lines, invoiceLineResponse{
			LineID:                 line.ID.String(),
			LineKind:               line.LineKind,
			Description:            line.Description,
			SortOrder:              line.SortOrder,
			QuantityMinutes:        line.QuantityMinutes,
			UnitAmountMinor:        moneyPtrToIntPtr(line.UnitAmount),
			LineAmountMinor:        line.LineAmount.Minor(),
			FundedAllowanceMinutes: line.FundedAllowanceMinutes,
			FundedDeductionMinutes: line.FundedDeductionMinutes,
			CoreBillableMinutes:    line.CoreBillableMinutes,
			SessionCount:           line.SessionCount,
		})
	}

	return resp
}

func toCalculationResponse(calc domain.InvoiceReviewCalculation) invoiceCalculationResponse {
	sessions := make([]bookedSessionResponse, 0, len(calc.BookedSessions))
	for _, s := range calc.BookedSessions {
		sessions = append(sessions, bookedSessionResponse{
			DayOfWeek:       s.DayOfWeek,
			OccurrenceDate:  s.OccurrenceDate.Format("2006-01-02"),
			DurationMinutes: s.DurationMinutes,
			SessionTypeID:   s.SessionTypeID,
			SessionTypeName: s.SessionTypeName,
		})
	}
	perEntry := make([]bookedEntryResponse, 0, len(calc.BookedPerEntry))
	for _, e := range calc.BookedPerEntry {
		perEntry = append(perEntry, bookedEntryResponse{
			DayOfWeek:          e.DayOfWeek,
			SessionTypeID:      e.SessionTypeID,
			SessionTypeName:    e.SessionTypeName,
			DurationMinutes:    e.DurationMinutes,
			OccurrencesInMonth: e.OccurrencesInMonth,
			TotalMinutes:       e.TotalMinutes,
		})
	}
	return invoiceCalculationResponse{
		CoreHourlyRateMinor:    calc.CoreHourlyRate.Minor(),
		BookedCoreMinutes:      calc.BookedCoreMinutes,
		BookedSessionCount:     calc.BookedSessionCount,
		FundedAllowanceMinutes: calc.FundedAllowanceMinutes,
		FundedDeductionMinutes: calc.FundedDeductionMinutes,
		CoreBillableMinutes:    calc.CoreBillableMinutes,
		CoreSubtotalMinor:      calc.CoreSubtotal.Minor(),
		ExtrasTotalMinor:       calc.ExtrasTotal.Minor(),
		TermID:                 calc.TermID.String(),
		BookingPatternID:       calc.BookingPatternID.String(),
		BookedSessions:         sessions,
		BookedPerEntry:         perEntry,
	}
}

func toOverrideAttendanceBlockResponse(r *application.OverrideAttendanceBlockResult) overrideAttendanceBlockResponse {
	return overrideAttendanceBlockResponse{
		InvoiceID:    r.InvoiceID.String(),
		BillingMonth: r.BillingMonth,
		OverriddenBy: r.OverriddenBy.String(),
		OverriddenAt: r.OverriddenAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toIssueInvoiceResponse(r domain.IssueInvoiceResult) issueInvoiceResponse {
	return issueInvoiceResponse{
		InvoiceID:     r.InvoiceID.String(),
		InvoiceNumber: r.InvoiceNumber,
		Status:        r.Status,
		IssuedAt:      formatTime(r.IssuedAt),
		LockedAt:      formatTime(r.LockedAt),
		DueAt:         formatTime(r.DueAt),
		IssuedRunID:   r.IssuedRunID.String(),
		TotalDueMinor: r.TotalDue.Minor(),
	}
}

func toBulkIssueResponse(r domain.BulkIssueInvoicesResult) bulkIssueInvoicesResponse {
	issued := make([]issuedInvoiceResponse, 0, len(r.Issued))
	for _, inv := range r.Issued {
		issued = append(issued, issuedInvoiceResponse{
			InvoiceID:       inv.InvoiceID.String(),
			ChildID:         inv.ChildID.String(),
			ChildFirstName:  inv.ChildFirstName,
			ChildMiddleName: inv.ChildMiddleName,
			ChildLastName:   inv.ChildLastName,
			InvoiceNumber:   inv.InvoiceNumber,
			IssuedAt:        formatTime(inv.IssuedAt),
			DueAt:           formatTime(inv.DueAt),
			TotalDueMinor:   inv.TotalDue.Minor(),
		})
	}

	blocked := make([]blockedInvoiceResponse, 0, len(r.Blocked))
	for _, b := range r.Blocked {
		blockers := make([]issueBlockerResponse, 0, len(b.Blockers))
		for _, bl := range b.Blockers {
			blockers = append(blockers, issueBlockerResponse{
				Code:    bl.Code,
				Message: bl.Message,
			})
		}
		resp := blockedInvoiceResponse{
			InvoiceID: b.InvoiceID.String(),
			Blockers:  blockers,
		}
		if b.ChildID != nil {
			resp.ChildID = strPtr(b.ChildID.String())
			resp.ChildFirstName = b.ChildFirstName
			resp.ChildMiddleName = b.ChildMiddleName
			resp.ChildLastName = b.ChildLastName
		}
		blocked = append(blocked, resp)
	}

	return bulkIssueInvoicesResponse{
		RunID:        r.RunID.String(),
		BillingMonth: r.BillingMonth,
		Status:       r.Status,
		Summary: bulkIssueSummary{
			EligibleCount: r.Summary.EligibleCount,
			SuccessCount:  r.Summary.SuccessCount,
			BlockedCount:  r.Summary.BlockedCount,
			TotalDueMinor: r.Summary.TotalDue.Minor(),
		},
		Issued:  issued,
		Blocked: blocked,
	}
}

func toParentInvoiceListResponse(r application.ListParentInvoicesResult) parentInvoiceListResponse {
	items := make([]parentInvoiceListItemResponse, 0, len(r.Items))
	for _, inv := range r.Items {
		item := parentInvoiceListItemResponse{
			InvoiceID:              inv.ID.String(),
			InvoiceKind:            inv.InvoiceKind,
			InvoiceNumber:          inv.InvoiceNumber,
			InvoiceNumberDisplay:   invoiceNumberDisplay(inv.Status, inv.InvoiceNumber),
			ChildID:                inv.ChildID.String(),
			ChildFirstName:         inv.ChildFirstName,
			ChildMiddleName:        inv.ChildMiddleName,
			ChildLastName:          inv.ChildLastName,
			BillingMonth:           formatBillingMonth(inv.BillingMonth),
			Status:                 inv.Status,
			DueStatus:              dueStatus(inv.Status),
			CurrencyCode:           inv.CurrencyCode,
			SubtotalMinor:          inv.Subtotal.Minor(),
			FundedDeductionMinor:   inv.FundedDeduction.Minor(),
			TotalDueMinor:          inv.TotalDue.Minor(),
			AmountPaidMinor:        inv.AmountPaid.Minor(),
			IssuedAt:               formatTimePtr(inv.IssuedAt),
			DueAt:                  formatTimePtr(inv.DueAt),
			PaidAt:                 formatTimePtr(inv.PaidAt),
			PaymentFailedAt:        formatTimePtr(inv.PaymentFailedAt),
			PaymentStatusUpdatedAt: formatTimePtr(inv.PaymentStatusUpdatedAt),
		}
		item.Period.StartDate = formatDate(inv.PeriodStartDate)
		item.Period.EndDate = formatDate(inv.PeriodEndDate)
		items = append(items, item)
	}
	return parentInvoiceListResponse{
		Items:  items,
		Limit:  r.Limit,
		Offset: r.Offset,
	}
}

func toParentInvoiceDetailResponse(r application.GetParentInvoiceResult) parentInvoiceDetailResponse {
	inv := r.Invoice
	resp := parentInvoiceDetailResponse{
		InvoiceID:              inv.ID.String(),
		InvoiceKind:            inv.InvoiceKind,
		InvoiceNumber:          inv.InvoiceNumber,
		InvoiceNumberDisplay:   invoiceNumberDisplay(inv.Status, inv.InvoiceNumber),
		ChildID:                inv.ChildID.String(),
		ChildFirstName:         inv.ChildFirstName,
		ChildMiddleName:        inv.ChildMiddleName,
		ChildLastName:          inv.ChildLastName,
		BillingMonth:           formatBillingMonth(inv.BillingMonth),
		Status:                 inv.Status,
		DueStatus:              dueStatus(inv.Status),
		CurrencyCode:           inv.CurrencyCode,
		SubtotalMinor:          inv.Subtotal.Minor(),
		FundedDeductionMinor:   inv.FundedDeduction.Minor(),
		TotalDueMinor:          inv.TotalDue.Minor(),
		AmountPaidMinor:        inv.AmountPaid.Minor(),
		IssuedAt:               formatTimePtr(inv.IssuedAt),
		DueAt:                  formatTimePtr(inv.DueAt),
		PaidAt:                 formatTimePtr(inv.PaidAt),
		PaymentFailedAt:        formatTimePtr(inv.PaymentFailedAt),
		PaymentStatusUpdatedAt: formatTimePtr(inv.PaymentStatusUpdatedAt),
		Calculation:            toParentCalculationResponse(r.Calculation),
	}
	resp.Period.StartDate = formatDate(inv.PeriodStartDate)
	resp.Period.EndDate = formatDate(inv.PeriodEndDate)

	resp.Lines = make([]parentInvoiceLineResponse, 0, len(r.Lines))
	for _, line := range r.Lines {
		resp.Lines = append(resp.Lines, parentInvoiceLineResponse{
			LineKind:        line.LineKind,
			Description:     line.Description,
			SortOrder:       line.SortOrder,
			QuantityMinutes: line.QuantityMinutes,
			UnitAmountMinor: moneyPtrToIntPtr(line.UnitAmount),
			LineAmountMinor: line.LineAmount.Minor(),
		})
	}

	return resp
}

func toParentCalculationResponse(calc domain.InvoiceReviewCalculation) parentInvoiceCalculationResponse {
	return parentInvoiceCalculationResponse{
		CoreHourlyRateMinor:    calc.CoreHourlyRate.Minor(),
		BookedCoreMinutes:      calc.BookedCoreMinutes,
		BookedSessionCount:     calc.BookedSessionCount,
		FundedAllowanceMinutes: calc.FundedAllowanceMinutes,
		FundedDeductionMinutes: calc.FundedDeductionMinutes,
		CoreBillableMinutes:    calc.CoreBillableMinutes,
		CoreSubtotalMinor:      calc.CoreSubtotal.Minor(),
		ExtrasTotalMinor:       calc.ExtrasTotal.Minor(),
	}
}

func countRunExceptions(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}
	var details struct {
		BlockedChildren []struct{} `json:"blocked_children"`
	}
	if err := json.Unmarshal(raw, &details); err != nil {
		return 0
	}
	return len(details.BlockedChildren)
}
