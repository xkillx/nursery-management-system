package httpowner

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/owner/application"
	"nursery-management-system/api/internal/modules/owner/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
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

// getSiteSummaries returns site summaries for the owner.
//
//	@Summary		Get site summaries
//	@Description	Get site summaries with billing and attendance data.
//	@Tags			owner
//	@Produce		json
//	@Param			billing_month	query		string	false	"Billing month"		format(month)
//	@Param			site_id			query		string	false	"Filter by site ID"	format(uuid)
//	@Success		200				{object}	siteSummariesResponse
//	@Failure		400				{object}	object{code=string,message=string}
//	@Failure		403				{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["owner"]
//	@Router			/owner/site-summaries [get]
func (h *Handler) getSiteSummaries(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
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
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "site_id", "message": "site_id must be a valid UUID"}})
			return
		}
		siteID = &parsed
	}

	result, err := h.summaries.Execute(c.Request.Context(), ownerActor, billingMonth, siteID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSiteSummariesResponse(result))
}

// listManagerAccess returns a paginated list of manager access records.
//
//	@Summary		List manager access
//	@Description	Get a paginated list of manager access records for a site.
//	@Tags			owner
//	@Produce		json
//	@Param			site_id		query		string	true	"Site ID"			format(uuid)
//	@Param			status		query		string	false	"Filter by status"	Enums(active, inactive, all)	default(active)
//	@Param			page		query		int		false	"Page number"		default(1)						minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)						minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]managerAccessResponse,total=int,page=int,page_size=int}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		403			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["owner"]
//	@Router			/owner/manager-access [get]
func (h *Handler) listManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
		return
	}

	siteIDStr := c.Query("site_id")
	if siteIDStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "site_id", "message": "site_id is required"}})
		return
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "site_id", "message": "site_id must be a valid UUID"}})
		return
	}

	status := c.DefaultQuery("status", "active")
	if status != "active" && status != "inactive" && status != "all" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "status", "message": "status must be active, inactive, or all"}})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	items, total, err := h.listAccess.ExecutePaginated(c.Request.Context(), ownerActor, siteID, status, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	out := make([]managerAccessResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toManagerAccessResponse(item))
	}
	c.JSON(http.StatusOK, pagination.PaginatedResponse(out, total, page, pageSize))
}

// grantManagerAccess grants manager access to a user.
//
//	@Summary		Grant manager access
//	@Description	Grant manager access to a user by email.
//	@Tags			owner
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string					true	"Site ID"	format(uuid)
//	@Param			body	body		object{email=string}	true	"Email address"
//	@Success		200		{object}	grantManagerAccessResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		403		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["owner"]
//	@Router			/owner/sites/{site_id}/manager-access [post]
func (h *Handler) grantManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "site_id", "message": "site_id must be a valid UUID"}})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", []map[string]string{{"field": "body", "message": "invalid JSON"}})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	result, err := h.grant.Execute(c.Request.Context(), ownerActor, siteID, req.Email)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toGrantResponse(result))
}

// deactivateManagerAccess deactivates manager access.
//
//	@Summary		Deactivate manager access
//	@Description	Deactivate manager access for a membership.
//	@Tags			owner
//	@Produce		json
//	@Param			site_id			path	string	true	"Site ID"		format(uuid)
//	@Param			membership_id	path	string	true	"Membership ID"	format(uuid)
//	@Success		204
//	@Failure		400	{object}	object{code=string,message=string}
//	@Failure		403	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["owner"]
//	@Router			/owner/sites/{site_id}/manager-access/{membership_id}/actions/deactivate [post]
func (h *Handler) deactivateManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "site_id", "message": "site_id must be a valid UUID"}})
		return
	}
	membershipID, err := uuid.Parse(c.Param("membership_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "membership_id", "message": "membership_id must be a valid UUID"}})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	err = h.deactivate.Execute(c.Request.Context(), ownerActor, siteID, membershipID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// reactivateManagerAccess reactivates manager access.
//
//	@Summary		Reactivate manager access
//	@Description	Reactivate manager access for a membership.
//	@Tags			owner
//	@Produce		json
//	@Param			site_id			path	string	true	"Site ID"		format(uuid)
//	@Param			membership_id	path	string	true	"Membership ID"	format(uuid)
//	@Success		204
//	@Failure		400	{object}	object{code=string,message=string}
//	@Failure		403	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["owner"]
//	@Router			/owner/sites/{site_id}/manager-access/{membership_id}/actions/activate [post]
func (h *Handler) reactivateManagerAccess(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "site_id", "message": "site_id must be a valid UUID"}})
		return
	}
	membershipID, err := uuid.Parse(c.Param("membership_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "membership_id", "message": "membership_id must be a valid UUID"}})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	err = h.reactivate.Execute(c.Request.Context(), ownerActor, siteID, membershipID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// updateSiteBillingSetup updates the site billing setup.
//
//	@Summary		Update site billing setup
//	@Description	Update the core hourly rate for a site.
//	@Tags			owner
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string								true	"Site ID"	format(uuid)
//	@Param			body	body		object{core_hourly_rate_minor=int}	true	"Billing setup data"
//	@Success		200		{object}	updateSiteBillingSetupResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		403		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["owner"]
//	@Router			/owner/sites/{site_id}/billing-setup [put]
func (h *Handler) updateSiteBillingSetup(c *gin.Context) {
	actor, ok := tenant.OwnerActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request parameters.", []map[string]string{{"field": "site_id", "message": "site_id must be a valid UUID"}})
		return
	}

	var req struct {
		CoreHourlyRateMinor int `json:"core_hourly_rate_minor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", []map[string]string{{"field": "body", "message": "invalid JSON"}})
		return
	}

	ownerActor := domain.OwnerActor{
		UserID:       actor.UserID,
		MembershipID: actor.MembershipID,
		TenantID:     actor.TenantID,
	}

	result, err := h.updateBillingSetup.Execute(c.Request.Context(), ownerActor, siteID, req.CoreHourlyRateMinor)
	if err != nil {
		h.handleError(c, err)
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

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}
