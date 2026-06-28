package httpterm

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/term/application"
	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
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

func NewHandler(
	createTerm *application.CreateTermUseCase,
	getTerm *application.GetTermUseCase,
	getCurrentTerm *application.GetCurrentTermForChildUseCase,
	listTerms *application.ListTermsForChildUseCase,
	listExpiring *application.ListExpiringTermsUseCase,
	requestChange *application.RequestScheduleChangeUseCase,
	approveChange *application.ApproveScheduleChangeUseCase,
	rejectChange *application.RejectScheduleChangeUseCase,
	terminate *application.TerminateTermUseCase,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:         logger,
		createTerm:     createTerm,
		getTerm:        getTerm,
		getCurrentTerm: getCurrentTerm,
		listTerms:      listTerms,
		listExpiring:   listExpiring,
		requestChange:  requestChange,
		approveChange:  approveChange,
		rejectChange:   rejectChange,
		terminate:      terminate,
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

func (h *Handler) createTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	childID, err := parseUUIDRaw(c.Param("child_id"), "child_id")
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	var req createTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	in, err := toCreateTermInput(childID, req)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	term, execErr := h.createTerm.Execute(c.Request.Context(), actor, in)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}

	c.JSON(http.StatusCreated, toTermResponse(term))
}

func (h *Handler) listTermsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	childID, err := parseUUIDRaw(c.Param("child_id"), "child_id")
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	terms, execErr := h.listTerms.Execute(c.Request.Context(), actor, childID.String())
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toTermListResponse(terms))
}

func (h *Handler) getCurrentTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	childID, err := parseUUIDRaw(c.Param("child_id"), "child_id")
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	term, execErr := h.getCurrentTerm.Execute(c.Request.Context(), actor, childID.String())
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toTermResponse(term))
}

func (h *Handler) getTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	term, execErr := h.getTerm.Execute(c.Request.Context(), actor, c.Param("term_id"))
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toTermResponse(term))
}

func (h *Handler) listExpiringHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	within := 30
	if q := c.Query("expiring_within_days"); q != "" {
		if v, perr := strconv.Atoi(q); perr == nil && v > 0 && v <= 365 {
			within = v
		}
	}
	terms, execErr := h.listExpiring.Execute(c.Request.Context(), actor, within)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toTermListResponse(terms))
}

func (h *Handler) requestScheduleChangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	termID, err := parseUUIDRaw(c.Param("term_id"), "term_id")
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	var req requestScheduleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	in, err := toRequestScheduleChangeInput(termID, req)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	change, execErr := h.requestChange.Execute(c.Request.Context(), actor, in)
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusCreated, toScheduleChangeResponse(change))
}

func (h *Handler) approveScheduleChangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	change, execErr := h.approveChange.Execute(c.Request.Context(), actor, c.Param("term_id"), c.Param("change_id"))
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toScheduleChangeResponse(change))
}

func (h *Handler) rejectScheduleChangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	change, execErr := h.rejectChange.Execute(c.Request.Context(), actor, c.Param("term_id"), c.Param("change_id"))
	if execErr != nil {
		h.handleError(c, execErr)
		return
	}
	c.JSON(http.StatusOK, toScheduleChangeResponse(change))
}

func (h *Handler) terminateTermHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		h.writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	termID, err := parseUUIDRaw(c.Param("term_id"), "term_id")
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	var req terminateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	in, err := toTerminateTermInput(termID, req)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "validation_error", err.Error())
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
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
	c.AbortWithStatusJSON(status, resp)
}

func (h *Handler) writeError(c *gin.Context, status int, code, message string) {
	requestID := httpserver.RequestIDFromContext(c)
	c.AbortWithStatusJSON(status, httpserver.ErrorResponse{
		Code: code, Message: message, RequestID: requestID,
	})
}

var _ = errors.New
var _ = domain.TermStatusActive
