package httpowner

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/owner/application"
	"nursery-management-system/api/internal/modules/owner/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	summaries          *application.GetSiteSummariesUseCase
	listAccess         *application.ListManagerAccessUseCase
	grant              *application.GrantManagerAccessUseCase
	deactivate         *application.DeactivateManagerAccessUseCase
	reactivate         *application.ReactivateManagerAccessUseCase
	updateBillingSetup *application.UpdateSiteBillingSetupUseCase
	logger             *slog.Logger
	recorder           *metrics.Recorder
}

func NewHandler(
	summaries *application.GetSiteSummariesUseCase,
	listAccess *application.ListManagerAccessUseCase,
	grant *application.GrantManagerAccessUseCase,
	deactivate *application.DeactivateManagerAccessUseCase,
	reactivate *application.ReactivateManagerAccessUseCase,
	recorder *metrics.Recorder,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		summaries:  summaries,
		listAccess: listAccess,
		grant:      grant,
		deactivate: deactivate,
		reactivate: reactivate,
		recorder:   recorder,
		logger:     logger,
	}
}

func (h *Handler) WithUpdateBillingSetup(uc *application.UpdateSiteBillingSetupUseCase) *Handler {
	return &Handler{
		summaries:          h.summaries,
		listAccess:         h.listAccess,
		grant:              h.grant,
		deactivate:         h.deactivate,
		reactivate:         h.reactivate,
		updateBillingSetup: uc,
		logger:             h.logger,
		recorder:           h.recorder,
	}
}

func (h *Handler) RegisterRoutes(ownerGroup *gin.RouterGroup) {
	ownerGroup.GET("/site-summaries", h.getSiteSummaries)
	ownerGroup.GET("/manager-access", h.listManagerAccess)
	ownerGroup.POST("/sites/:site_id/manager-access", h.grantManagerAccess)
	ownerGroup.POST("/sites/:site_id/manager-access/:membership_id/actions/deactivate", h.deactivateManagerAccess)
	ownerGroup.POST("/sites/:site_id/manager-access/:membership_id/actions/activate", h.reactivateManagerAccess)
	ownerGroup.PUT("/sites/:site_id/billing-setup", h.updateSiteBillingSetup)
}

func (h *Handler) getSiteSummaries(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":       "forbidden_role",
			"message":    "Access denied.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	billingMonth := c.Query("billing_month")

	var siteID *uuid.UUID
	if raw := c.Query("site_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":       "validation_error",
				"message":    "Invalid request parameters.",
				"request_id": httpserver.RequestIDFromContext(c),
				"details":    map[string]string{"field": "site_id", "message": "site_id must be a valid UUID"},
			})
			return
		}
		siteID = &parsed
	}

	result, err := h.summaries.Execute(c.Request.Context(), ownerActor, billingMonth, siteID)
	if err != nil {
		var valErr *domain.ValidationError
		switch {
		case errors.As(err, &valErr):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":       "validation_error",
				"message":    "Invalid request parameters.",
				"request_id": httpserver.RequestIDFromContext(c),
				"details":    map[string]string{"field": valErr.Field, "message": valErr.Message},
			})
		case errors.Is(err, domain.ErrSiteNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":       "site_not_found",
				"message":    "Site not found.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":       "internal_error",
				"message":    "Something went wrong.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		}
		return
	}

	c.JSON(http.StatusOK, toSiteSummariesResponse(result))
}

func (h *Handler) listManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":       "forbidden_role",
			"message":    "Access denied.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	siteIDStr := c.Query("site_id")
	if siteIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
			"details":    map[string]string{"field": "site_id", "message": "site_id is required"},
		})
		return
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
			"details":    map[string]string{"field": "site_id", "message": "site_id must be a valid UUID"},
		})
		return
	}

	status := c.DefaultQuery("status", "active")
	if status != "active" && status != "inactive" && status != "all" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
			"details":    map[string]string{"field": "status", "message": "status must be active, inactive, or all"},
		})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	items, err := h.listAccess.Execute(c.Request.Context(), ownerActor, siteID, status)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSiteNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":       "site_not_found",
				"message":    "Site not found.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":       "internal_error",
				"message":    "Something went wrong.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		}
		return
	}

	out := make([]managerAccessResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toManagerAccessResponse(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) grantManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":       "forbidden_role",
			"message":    "Access denied.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
			"details":    map[string]string{"field": "site_id", "message": "site_id must be a valid UUID"},
		})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request payload.",
			"request_id": httpserver.RequestIDFromContext(c),
			"details":    err.Error(),
		})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	result, err := h.grant.Execute(c.Request.Context(), ownerActor, siteID, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSiteNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":       "site_not_found",
				"message":    "Site not found.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":       "internal_error",
				"message":    "Something went wrong.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		}
		return
	}

	c.JSON(http.StatusOK, toGrantResponse(result))
}

func (h *Handler) deactivateManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":       "forbidden_role",
			"message":    "Access denied.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}
	membershipID, err := uuid.Parse(c.Param("membership_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	err = h.deactivate.Execute(c.Request.Context(), ownerActor, siteID, membershipID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSiteNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":       "site_not_found",
				"message":    "Site not found.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		case errors.Is(err, domain.ErrMembershipNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":       "manager_membership_not_found",
				"message":    "Manager membership not found.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":       "internal_error",
				"message":    "Something went wrong.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) reactivateManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":       "forbidden_role",
			"message":    "Access denied.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}
	membershipID, err := uuid.Parse(c.Param("membership_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	err = h.reactivate.Execute(c.Request.Context(), ownerActor, siteID, membershipID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSiteNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":       "site_not_found",
				"message":    "Site not found.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":       "internal_error",
				"message":    "Something went wrong.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) updateSiteBillingSetup(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":       "forbidden_role",
			"message":    "Access denied.",
			"request_id": httpserver.RequestIDFromContext(c),
		})
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request parameters.",
			"request_id": httpserver.RequestIDFromContext(c),
			"details":    map[string]string{"field": "site_id", "message": "site_id must be a valid UUID"},
		})
		return
	}

	var req struct {
		CoreHourlyRateMinor int `json:"core_hourly_rate_minor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request payload.",
			"request_id": httpserver.RequestIDFromContext(c),
			"details":    err.Error(),
		})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	result, err := h.updateBillingSetup.Execute(c.Request.Context(), ownerActor, siteID, req.CoreHourlyRateMinor)
	if err != nil {
		var valErr *domain.ValidationError
		switch {
		case errors.As(err, &valErr):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":       "validation_error",
				"message":    "Invalid request parameters.",
				"request_id": httpserver.RequestIDFromContext(c),
				"details":    map[string]string{"field": valErr.Field, "message": valErr.Message},
			})
		case errors.Is(err, domain.ErrSiteNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":       "site_not_found",
				"message":    "Site not found.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":       "internal_error",
				"message":    "Something went wrong.",
				"request_id": httpserver.RequestIDFromContext(c),
			})
		}
		return
	}

	c.JSON(http.StatusOK, updateSiteBillingSetupResponse{
		SiteID:              result.SiteID,
		CoreHourlyRateMinor: result.SiteCoreHourlyRateMinor,
	})
}

// ── Response DTOs ────────────────────────────────────────────────────────────

type siteSummariesResponse struct {
	BillingMonth        string                    `json:"billing_month"`
	AttendanceLocalDate string                    `json:"attendance_local_date"`
	CurrencyCode        string                    `json:"currency_code"`
	Totals              siteSummaryTotalsResponse `json:"totals"`
	Sites               []siteSummaryResponse     `json:"sites"`
}

type siteSummaryTotalsResponse struct {
	ActiveManagerCount        int   `json:"active_manager_count"`
	PendingManagerInviteCount int   `json:"pending_manager_invite_count"`
	ActiveChildrenCount       int   `json:"active_children_count"`
	CheckedInTodayCount       int   `json:"checked_in_today_count"`
	IncompleteAttendanceCount int   `json:"incomplete_attendance_count"`
	DraftCount                int   `json:"draft_count"`
	IssuedCount               int   `json:"issued_count"`
	OverdueCount              int   `json:"overdue_count"`
	PaymentFailedCount        int   `json:"payment_failed_count"`
	PaidCount                 int   `json:"paid_count"`
	TotalIssuedMinor          int64 `json:"total_issued_minor"`
	TotalPaidMinor            int64 `json:"total_paid_minor"`
	OutstandingMinor          int64 `json:"outstanding_minor"`
	OverdueOutstandingMinor   int64 `json:"overdue_outstanding_minor"`
}

type siteSummaryResponse struct {
	SiteID                    uuid.UUID                    `json:"site_id"`
	SiteName                  string                       `json:"site_name"`
	SetupStatus               string                       `json:"setup_status"`
	SetupIssues               []string                     `json:"setup_issues"`
	SiteCoreHourlyRateMinor   *int                         `json:"site_core_hourly_rate_minor"`
	ActiveManagerCount        int                          `json:"active_manager_count"`
	PendingManagerInviteCount int                          `json:"pending_manager_invite_count"`
	ActiveChildrenCount       int                          `json:"active_children_count"`
	Attendance                attendanceSummaryResponse    `json:"attendance"`
	FundingReadiness          fundingReadinessResponse     `json:"funding_readiness"`
	InvoicePaymentHealth      invoicePaymentHealthResponse `json:"invoice_payment_health"`
}

type attendanceSummaryResponse struct {
	CheckedInTodayCount       int `json:"checked_in_today_count"`
	IncompleteAttendanceCount int `json:"incomplete_attendance_count"`
}

type fundingReadinessResponse struct {
	IncludedChildCount  int `json:"included_child_count"`
	FlaggedChildCount   int `json:"flagged_child_count"`
	MissingProfileCount int `json:"missing_profile_count"`
	ExplicitZeroCount   int `json:"explicit_zero_count"`
	UnderOneHourCount   int `json:"under_one_hour_count"`
	Above160HoursCount  int `json:"above_160_hours_count"`
}

type invoicePaymentHealthResponse struct {
	DraftCount              int   `json:"draft_count"`
	IssuedCount             int   `json:"issued_count"`
	OverdueCount            int   `json:"overdue_count"`
	PaymentFailedCount      int   `json:"payment_failed_count"`
	PaidCount               int   `json:"paid_count"`
	TotalIssuedMinor        int64 `json:"total_issued_minor"`
	TotalPaidMinor          int64 `json:"total_paid_minor"`
	OutstandingMinor        int64 `json:"outstanding_minor"`
	OverdueOutstandingMinor int64 `json:"overdue_outstanding_minor"`
	FailedPaymentCount      int   `json:"failed_payment_count"`
}

type managerAccessResponse struct {
	MembershipID string `json:"membership_id"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	IsActive     bool   `json:"is_active"`
}

type grantManagerAccessResponse struct {
	Outcome      string                      `json:"outcome"`
	MembershipID *string                     `json:"membership_id,omitempty"`
	Invite       *grantInviteDetailsResponse `json:"invite,omitempty"`
}

type grantInviteDetailsResponse struct {
	Email     string `json:"email"`
	ExpiresAt string `json:"expires_at"`
}

type updateSiteBillingSetupResponse struct {
	SiteID              uuid.UUID `json:"site_id"`
	CoreHourlyRateMinor int       `json:"core_hourly_rate_minor"`
}

func toManagerAccessResponse(item application.ManagerAccessItem) managerAccessResponse {
	return managerAccessResponse{
		MembershipID: item.MembershipID.String(),
		UserID:       item.UserID.String(),
		Email:        item.Email,
		IsActive:     item.IsActive,
	}
}

func toGrantResponse(result application.GrantManagerAccessResult) grantManagerAccessResponse {
	resp := grantManagerAccessResponse{
		Outcome: string(result.Outcome),
	}
	if result.MembershipID != nil {
		s := result.MembershipID.String()
		resp.MembershipID = &s
	}
	if result.InviteDetails != nil {
		resp.Invite = &grantInviteDetailsResponse{
			Email:     result.InviteDetails.Email,
			ExpiresAt: result.InviteDetails.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return resp
}

func toSiteSummariesResponse(r domain.SiteSummariesResult) siteSummariesResponse {
	sites := make([]siteSummaryResponse, 0, len(r.Sites))
	for _, s := range r.Sites {
		sites = append(sites, toSiteSummaryResponse(s))
	}
	return siteSummariesResponse{
		BillingMonth:        r.BillingMonth,
		AttendanceLocalDate: r.AttendanceLocalDate,
		CurrencyCode:        r.CurrencyCode,
		Totals:              toTotalsResponse(r.Totals),
		Sites:               sites,
	}
}

func toSiteSummaryResponse(s domain.SiteSummary) siteSummaryResponse {
	return siteSummaryResponse{
		SiteID:                    s.SiteID,
		SiteName:                  s.SiteName,
		SetupStatus:               s.SetupStatus,
		SetupIssues:               s.SetupIssues,
		SiteCoreHourlyRateMinor:   s.SiteCoreHourlyRateMinor,
		ActiveManagerCount:        s.ActiveManagerCount,
		PendingManagerInviteCount: s.PendingManagerInviteCount,
		ActiveChildrenCount:       s.ActiveChildrenCount,
		Attendance: attendanceSummaryResponse{
			CheckedInTodayCount:       s.Attendance.CheckedInTodayCount,
			IncompleteAttendanceCount: s.Attendance.IncompleteAttendanceCount,
		},
		FundingReadiness: fundingReadinessResponse{
			IncludedChildCount:  s.FundingReadiness.IncludedChildCount,
			FlaggedChildCount:   s.FundingReadiness.FlaggedCount(),
			MissingProfileCount: s.FundingReadiness.MissingProfileCount,
			ExplicitZeroCount:   s.FundingReadiness.ExplicitZeroCount,
			UnderOneHourCount:   s.FundingReadiness.UnderOneHourCount,
			Above160HoursCount:  s.FundingReadiness.Above160HoursCount,
		},
		InvoicePaymentHealth: invoicePaymentHealthResponse{
			DraftCount:              s.InvoicePaymentHealth.DraftCount,
			IssuedCount:             s.InvoicePaymentHealth.IssuedCount,
			OverdueCount:            s.InvoicePaymentHealth.OverdueCount,
			PaymentFailedCount:      s.InvoicePaymentHealth.PaymentFailedCount,
			PaidCount:               s.InvoicePaymentHealth.PaidCount,
			TotalIssuedMinor:        s.InvoicePaymentHealth.TotalIssuedMinor,
			TotalPaidMinor:          s.InvoicePaymentHealth.TotalPaidMinor,
			OutstandingMinor:        s.InvoicePaymentHealth.OutstandingMinor,
			OverdueOutstandingMinor: s.InvoicePaymentHealth.OverdueOutstandingMinor,
			FailedPaymentCount:      s.InvoicePaymentHealth.PaymentFailedCount,
		},
	}
}

func toTotalsResponse(t domain.SiteSummaryTotals) siteSummaryTotalsResponse {
	return siteSummaryTotalsResponse{
		ActiveManagerCount:        t.ActiveManagerCount,
		PendingManagerInviteCount: t.PendingManagerInviteCount,
		ActiveChildrenCount:       t.ActiveChildrenCount,
		CheckedInTodayCount:       t.CheckedInTodayCount,
		IncompleteAttendanceCount: t.IncompleteAttendanceCount,
		DraftCount:                t.DraftCount,
		IssuedCount:               t.IssuedCount,
		OverdueCount:              t.OverdueCount,
		PaymentFailedCount:        t.PaymentFailedCount,
		PaidCount:                 t.PaidCount,
		TotalIssuedMinor:          t.TotalIssuedMinor,
		TotalPaidMinor:            t.TotalPaidMinor,
		OutstandingMinor:          t.OutstandingMinor,
		OverdueOutstandingMinor:   t.OverdueOutstandingMinor,
	}
}
