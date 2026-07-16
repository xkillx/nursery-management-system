package httpterm

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/term/application"
	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type (
	CoreTermUseCases struct {
		Create       *application.CreateTermUseCase
		Get          *application.GetTermUseCase
		GetCurrent   *application.GetCurrentTermForChildUseCase
		List         *application.ListTermsForChildUseCase
		ListExpiring *application.ListExpiringTermsUseCase
		Terminate    *application.TerminateTermUseCase
	}

	ScheduleChangeUseCases struct {
		Request *application.RequestScheduleChangeUseCase
		Approve *application.ApproveScheduleChangeUseCase
		Reject  *application.RejectScheduleChangeUseCase
	}

	TermHandlerConfig struct {
		Core    CoreTermUseCases
		Changes ScheduleChangeUseCases
	}
)

type Handler struct {
	logger         *slog.Logger
	createTerm     *application.CreateTermUseCase
	getTerm        *application.GetTermUseCase
	getCurrentTerm *application.GetCurrentTermForChildUseCase
	listTerms      *application.ListTermsForChildUseCase
	listExpiring   *application.ListExpiringTermsUseCase
	requestChange  *application.RequestScheduleChangeUseCase
	approveChange  *application.ApproveScheduleChangeUseCase
	rejectChange   *application.RejectScheduleChangeUseCase
	terminate      *application.TerminateTermUseCase
}

func NewHandler(cfg TermHandlerConfig, logger *slog.Logger) *Handler {
	return &Handler{
		logger:         logger,
		createTerm:     cfg.Core.Create,
		getTerm:        cfg.Core.Get,
		getCurrentTerm: cfg.Core.GetCurrent,
		listTerms:      cfg.Core.List,
		listExpiring:   cfg.Core.ListExpiring,
		terminate:      cfg.Core.Terminate,
		requestChange:  cfg.Changes.Request,
		approveChange:  cfg.Changes.Approve,
		rejectChange:   cfg.Changes.Reject,
	}
}

func (h *Handler) RegisterManagerRoutes(manager *gin.RouterGroup) {
	manager.POST("/children/:child_id/terms", h.createTermHandler)
	manager.GET("/children/:child_id/terms", h.listTermsHandler)
	manager.GET("/children/:child_id/terms/current", h.getCurrentTermHandler)
	manager.GET("/terms", h.listExpiringHandler)
	manager.GET("/terms/:term_id", h.getTermHandler)
	manager.POST("/terms/:term_id/schedule-changes", h.requestScheduleChangeHandler)
	manager.POST("/terms/:term_id/schedule-changes/:change_id/approve", h.approveScheduleChangeHandler)
	manager.POST("/terms/:term_id/schedule-changes/:change_id/reject", h.rejectScheduleChangeHandler)
	manager.POST("/terms/:term_id/actions/terminate", h.terminateTermHandler)
}

// ──────────────────────────────────────────────────────────────────────────────
// Handlers
// ──────────────────────────────────────────────────────────────────────────────

// createTermHandler creates a new term for a child.
//
//	@Summary		Create term
//	@Description	Create a new term for a child.
//	@Tags			terms
//	@Accept			json
//	@Produce		json
//	@Param			child_id	path		string				true	"Child ID"	format(uuid)
//	@Param			body		body		createTermRequest	true	"Term data"
//	@Success		201			{object}	termResponse
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children/{child_id}/terms [post]
func (h *Handler) createTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID, err := parseUUIDRaw(c.Param("child_id"), "child_id")
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}

	var req createTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	in, err := toCreateTermInput(childID, req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}

	term, execErr := h.createTerm.Execute(c.Request.Context(), actor, in)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}

	resp := toTermResponse(term)
	c.Header("Location", fmt.Sprintf("/api/children/%s/terms/%s", resp.ChildID, resp.ID))
	c.JSON(http.StatusCreated, resp)
}

// listTermsHandler returns a paginated list of terms for a child.
//
//	@Summary		List terms
//	@Description	Get a paginated list of terms for a child.
//	@Tags			terms
//	@Produce		json
//	@Param			child_id	path		string	true	"Child ID"			format(uuid)
//	@Param			page		query		int		false	"Page number"		default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]termResponse,total=int,page=int,page_size=int}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children/{child_id}/terms [get]
func (h *Handler) listTermsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	childID, err := parseUUIDRaw(c.Param("child_id"), "child_id")
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize
	result, execErr := h.listTerms.Execute(c.Request.Context(), actor, childID.String(), pageSize, offset)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, pagination.PaginatedResponse(toTermListResponse(result.Items), result.Total, page, pageSize))
}

// getCurrentTermHandler returns the current term for a child.
//
//	@Summary		Get current term
//	@Description	Get the current term for a child.
//	@Tags			terms
//	@Produce		json
//	@Param			child_id	path		string	true	"Child ID"	format(uuid)
//	@Success		200			{object}	termResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children/{child_id}/terms/current [get]
func (h *Handler) getCurrentTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	childID, err := parseUUIDRaw(c.Param("child_id"), "child_id")
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	term, execErr := h.getCurrentTerm.Execute(c.Request.Context(), actor, childID.String())
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toTermResponse(term))
}

// getTermHandler returns a single term by ID.
//
//	@Summary		Get term
//	@Description	Get a single term by ID.
//	@Tags			terms
//	@Produce		json
//	@Param			term_id	path		string	true	"Term ID"	format(uuid)
//	@Success		200		{object}	termResponse
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/terms/{term_id} [get]
func (h *Handler) getTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	term, execErr := h.getTerm.Execute(c.Request.Context(), actor, c.Param("term_id"))
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toTermResponse(term))
}

// listExpiringHandler returns a paginated list of expiring terms.
//
//	@Summary		List expiring terms
//	@Description	Get a paginated list of terms expiring within a number of days.
//	@Tags			terms
//	@Produce		json
//	@Param			expiring_within_days	query		int	false	"Days until expiry"	default(30)	minimum(1)	maximum(365)
//	@Param			page					query		int	false	"Page number"		default(1)	minimum(1)
//	@Param			page_size				query		int	false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200						{object}	object{items=[]termResponse,total=int,page=int,page_size=int}
//	@Failure		401						{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/terms [get]
func (h *Handler) listExpiringHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	within := 30
	if q := c.Query("expiring_within_days"); q != "" {
		if v, perr := strconv.Atoi(q); perr == nil && v > 0 && v <= 365 {
			within = v
		}
	}
	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize
	result, execErr := h.listExpiring.Execute(c.Request.Context(), actor, within, pageSize, offset)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, pagination.PaginatedResponse(toTermListResponse(result.Items), result.Total, page, pageSize))
}

// requestScheduleChangeHandler requests a schedule change for a term.
//
//	@Summary		Request schedule change
//	@Description	Request a schedule change for a term.
//	@Tags			terms
//	@Accept			json
//	@Produce		json
//	@Param			term_id	path		string							true	"Term ID"	format(uuid)
//	@Param			body	body		requestScheduleChangeRequest	true	"Schedule change data"
//	@Success		201		{object}	scheduleChangeResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/terms/{term_id}/schedule-changes [post]
func (h *Handler) requestScheduleChangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	termID, err := parseUUIDRaw(c.Param("term_id"), "term_id")
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	var req requestScheduleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	in, err := toRequestScheduleChangeInput(termID, req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	change, execErr := h.requestChange.Execute(c.Request.Context(), actor, in)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusCreated, toScheduleChangeResponse(change))
}

// approveScheduleChangeHandler approves a schedule change.
//
//	@Summary		Approve schedule change
//	@Description	Approve a schedule change for a term.
//	@Tags			terms
//	@Produce		json
//	@Param			term_id		path		string	true	"Term ID"	format(uuid)
//	@Param			change_id	path		string	true	"Change ID"	format(uuid)
//	@Success		200			{object}	scheduleChangeResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/terms/{term_id}/schedule-changes/{change_id}/approve [post]
func (h *Handler) approveScheduleChangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	change, execErr := h.approveChange.Execute(c.Request.Context(), actor, c.Param("term_id"), c.Param("change_id"))
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toScheduleChangeResponse(change))
}

// rejectScheduleChangeHandler rejects a schedule change.
//
//	@Summary		Reject schedule change
//	@Description	Reject a schedule change for a term.
//	@Tags			terms
//	@Produce		json
//	@Param			term_id		path		string	true	"Term ID"	format(uuid)
//	@Param			change_id	path		string	true	"Change ID"	format(uuid)
//	@Success		200			{object}	scheduleChangeResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/terms/{term_id}/schedule-changes/{change_id}/reject [post]
func (h *Handler) rejectScheduleChangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	change, execErr := h.rejectChange.Execute(c.Request.Context(), actor, c.Param("term_id"), c.Param("change_id"))
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toScheduleChangeResponse(change))
}

// terminateTermHandler terminates a term.
//
//	@Summary		Terminate term
//	@Description	Terminate a term.
//	@Tags			terms
//	@Accept			json
//	@Produce		json
//	@Param			term_id	path		string					true	"Term ID"	format(uuid)
//	@Param			body	body		terminateTermRequest	true	"Termination data"
//	@Success		200		{object}	termResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/terms/{term_id}/actions/terminate [post]
func (h *Handler) terminateTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	termID, err := parseUUIDRaw(c.Param("term_id"), "term_id")
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	var req terminateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	in, err := toTerminateTermInput(termID, req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	term, execErr := h.terminate.Execute(c.Request.Context(), actor, in)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toTermResponse(term))
}

// ──────────────────────────────────────────────────────────────────────────────
// Error handling
// ──────────────────────────────────────────────────────────────────────────────

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}

var _ = errors.New
var _ = domain.TermStatusActive
