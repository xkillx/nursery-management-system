package httpattendance

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/application"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	logger        *slog.Logger
	checkIn       *application.CheckInChild
	checkOut      *application.CheckOutChild
	correct       *application.CorrectAttendance
	listSessions  *application.ListCorrectionSessions
	listHistory   *application.ListCorrectionHistory
	getRegister   *application.GetRegister
	getRegSummary *application.GetRegisterSummary
	listParentAtt *application.ListParentAttendance
}

func NewHandler(
	checkIn *application.CheckInChild,
	checkOut *application.CheckOutChild,
	correct *application.CorrectAttendance,
	listSessions *application.ListCorrectionSessions,
	listHistory *application.ListCorrectionHistory,
	getRegister *application.GetRegister,
	getRegSummary *application.GetRegisterSummary,
	listParentAtt *application.ListParentAttendance,
	logger *slog.Logger,
) *Handler {
	return &Handler{logger: logger, checkIn: checkIn, checkOut: checkOut, correct: correct, listSessions: listSessions, listHistory: listHistory, getRegister: getRegister, getRegSummary: getRegSummary, listParentAtt: listParentAtt}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	g := protected.Group("")
	g.Use(httpserver.RequireRolesWithObservability(h.logger, nil, "manager", "practitioner"))
	g.POST("/attendance/check-ins", h.checkInHandler)
	g.POST("/attendance/check-outs", h.checkOutHandler)

	managerOnly := protected.Group("")
	managerOnly.Use(httpserver.RequireRolesWithObservability(h.logger, nil, "manager"))
	managerOnly.POST("/attendance/corrections", h.correctionHandler)
	managerOnly.GET("/attendance/sessions", h.listSessionsHandler)
	managerOnly.GET("/attendance/sessions/:session_id/history", h.listHistoryHandler)
	managerOnly.GET("/register", h.getRegisterHandler)
	managerOnly.GET("/register/summary", h.getRegisterSummaryHandler)
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.GET("/attendance", h.parentAttendanceHandler)
}

// parentAttendanceHandler returns attendance records for the parent's children.
//
//	@Summary		Parent attendance
//	@Description	Get attendance records for the authenticated parent's children.
//	@Tags			parent-attendance
//	@Produce		json
//	@Param			date	query		string	true	"Register date"	format(date)
//	@Success		200		{object}	object{items=[]parentAttendanceEntryResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["parent"]
//	@Router			/parent/attendance [get]
func (h *Handler) parentAttendanceHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	registerDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid date format.", nil)
		return
	}

	entries, err := h.listParentAtt.Execute(c.Request.Context(), actor, registerDate)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]parentAttendanceEntryResponse, 0, len(entries))
	for _, e := range entries {
		items = append(items, parentAttendanceEntryResponse{
			ChildID:             e.ChildID.String(),
			ChildFirstName:      e.ChildFirstName,
			ChildLastName:       e.ChildLastName,
			SessionTemplateName: e.SessionTemplateName,
			BookingType:         e.BookingType,
			AttendanceStatus:    e.AttendanceStatus,
			CheckInAt:           formatTimePtr(e.CheckInAt),
			CheckOutAt:          formatTimePtr(e.CheckOutAt),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// checkInHandler checks in a child.
//
//	@Summary		Check in child
//	@Description	Check in a child for attendance.
//	@Tags			attendance
//	@Accept			json
//	@Produce		json
//	@Param			body	body		checkInRequest	true	"Child ID"
//	@Success		201		{object}	sessionResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","practitioner"]
//	@Router			/attendance/check-ins [post]
func (h *Handler) checkInHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req checkInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	childID, err := parseChildID(req.ChildID)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	session, err := h.checkIn.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toSessionResponse(session)
	c.Header("Location", fmt.Sprintf("/api/attendance/sessions/%s", resp.ID))
	c.JSON(http.StatusCreated, resp)
}

// checkOutHandler checks out a child.
//
//	@Summary		Check out child
//	@Description	Check out a child from attendance.
//	@Tags			attendance
//	@Accept			json
//	@Produce		json
//	@Param			body	body		checkOutRequest	true	"Child ID"
//	@Success		200		{object}	sessionResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","practitioner"]
//	@Router			/attendance/check-outs [post]
func (h *Handler) checkOutHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req checkOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	childID, err := parseChildID(req.ChildID)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	session, err := h.checkOut.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionResponse(session))
}

// correctionHandler corrects attendance.
//
//	@Summary		Correct attendance
//	@Description	Correct attendance for a child.
//	@Tags			attendance
//	@Accept			json
//	@Produce		json
//	@Param			body	body		correctionRequest	true	"Correction data"
//	@Success		201		{object}	sessionResponse
//	@Success		200		{object}	sessionResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/attendance/corrections [post]
func (h *Handler) correctionHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req correctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params, err := parseCorrectionRequest(req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	result, err := h.correct.Execute(c.Request.Context(), actor, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toSessionResponse(result.Session)
	if result.Created {
		c.Header("Location", fmt.Sprintf("/api/attendance/sessions/%s", resp.ID))
		c.JSON(http.StatusCreated, resp)
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

// listSessionsHandler lists attendance sessions.
//
//	@Summary		List attendance sessions
//	@Description	Get a list of attendance sessions for a child on a date.
//	@Tags			attendance
//	@Produce		json
//	@Param			child_id	query		string	true	"Child ID"	format(uuid)
//	@Param			local_date	query		string	true	"Local date"	format(date)
//	@Success		200			{object}	object{sessions=[]sessionResponse}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/attendance/sessions [get]
func (h *Handler) listSessionsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childIDStr := c.Query("child_id")
	if childIDStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	localDateStr := c.Query("local_date")
	if localDateStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	localDate, err := time.Parse("2006-01-02", localDateStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	ctx, err := h.listSessions.Execute(c.Request.Context(), actor, childID, localDate)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCorrectionSessionContextResponse(ctx))
}

// listHistoryHandler lists correction history for a session.
//
//	@Summary		List correction history
//	@Description	Get correction history for an attendance session.
//	@Tags			attendance
//	@Produce		json
//	@Param			session_id	path		string	true	"Session ID"	format(uuid)
//	@Success		200			{object}	object{corrections=[]correctionHistoryResponse}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/attendance/sessions/{session_id}/history [get]
func (h *Handler) listHistoryHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	result, err := h.listHistory.Execute(c.Request.Context(), actor, sessionID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCorrectionHistoryResponse(result))
}

// getRegisterHandler returns the daily attendance register.
//
//	@Summary		Get daily register
//	@Description	Get the attendance register for a date, showing expected children and their attendance status.
//	@Tags			attendance
//	@Produce		json
//	@Param			date	query		string	true	"Register date"	format(date)
//	@Success		200		{object}	registerResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/register [get]
func (h *Handler) getRegisterHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	registerDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	entries, err := h.getRegister.Execute(c.Request.Context(), actor, registerDate)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRegisterResponse(dateStr, entries))
}

// getRegisterSummaryHandler returns per-room booking counts for a date range.
//
//	@Summary		Get register summary
//	@Description	Get per-room booking counts for a date range (used for date picker badges).
//	@Tags			attendance
//	@Produce		json
//	@Param			from	query		string	true	"From date"	format(date)
//	@Param			to		query		string	true	"To date"	format(date)
//	@Success		200		{object}	registerSummaryResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/register/summary [get]
func (h *Handler) getRegisterSummaryHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	fromDate, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	toDate, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	entries, err := h.getRegSummary.Execute(c.Request.Context(), actor, fromDate, toDate)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRegisterSummaryResponse(entries))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}
