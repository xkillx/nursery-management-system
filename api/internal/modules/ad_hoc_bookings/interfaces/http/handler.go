package httpadhocbookings

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/ad_hoc_bookings/application"
	"nursery-management-system/api/internal/modules/ad_hoc_bookings/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"
)

type adHocBookingResponse struct {
	ID            string `json:"id"`
	ChildID       string `json:"child_id"`
	CalendarDate  string `json:"calendar_date"`
	SessionTypeID string `json:"session_type_id"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type createAdHocBookingRequest struct {
	ChildID       string `json:"child_id" binding:"required"`
	CalendarDate  string `json:"calendar_date" binding:"required"`
	SessionTypeID string `json:"session_type_id" binding:"required"`
}

func toAdHocBookingResponse(b domain.AdHocBooking) adHocBookingResponse {
	return adHocBookingResponse{
		ID:            b.ID.String(),
		ChildID:       b.ChildID.String(),
		CalendarDate:  b.CalendarDate.UTC().Format("2006-01-02"),
		SessionTypeID: b.SessionTypeID.String(),
		Status:        b.Status,
		CreatedAt:     b.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toAdHocBookingListResponse(items []domain.AdHocBooking) []adHocBookingResponse {
	out := make([]adHocBookingResponse, 0, len(items))
	for _, b := range items {
		out = append(out, toAdHocBookingResponse(b))
	}
	return out
}

type Handler struct {
	logger *slog.Logger
	create *application.CreateAdHocBooking
	list   *application.ListAdHocBookings
	cancel *application.CancelAdHocBooking
}

func NewHandler(
	create *application.CreateAdHocBooking,
	list *application.ListAdHocBookings,
	cancel *application.CancelAdHocBooking,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger: logger,
		create: create,
		list:   list,
		cancel: cancel,
	}
}

func (h *Handler) RegisterManagerRoutes(manager *gin.RouterGroup) {
	manager.GET("/sites/:site_id/ad-hoc-bookings", h.listBookings)
	manager.POST("/sites/:site_id/ad-hoc-bookings", h.createBooking)
	manager.POST("/sites/:site_id/ad-hoc-bookings/:booking_id/cancel", h.cancelBooking)
}

func (h *Handler) resolveActor(c *gin.Context) (application.AdHocBookingActor, bool) {
	if owner, ok := tenant.OwnerActorFromGinContext(c); ok {
		return application.NewOwnerAdHocBookingActor(owner), true
	}

	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		return nil, false
	}

	if actor.BranchID == uuid.Nil {
		return nil, false
	}

	return application.NewManagerAdHocBookingActor(actor), true
}

// listBookings returns a paginated list of ad hoc bookings.
//
//	@Summary		List ad hoc bookings
//	@Description	Get a paginated list of ad hoc bookings for a site.
//	@Tags			ad-hoc-bookings
//	@Produce		json
//	@Param			site_id		path		string	true	"Site ID"	format(uuid)
//	@Param			child_id	query		string	false	"Filter by child ID"	format(uuid)
//	@Param			from		query		string	false	"Filter from date"		format(date)
//	@Param			to			query		string	false	"Filter to date"		format(date)
//	@Param			page		query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]adHocBookingResponse,total=int,page=int,page_size=int}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/ad-hoc-bookings [get]
func (h *Handler) listBookings(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	var childID *uuid.UUID
	if cid := c.Query("child_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid child_id.")
			return
		}
		childID = &id
	}

	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid from format.")
			return
		}
		from = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid to format.")
			return
		}
		to = &t
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	bookings, total, err := h.list.ExecutePaginated(c.Request.Context(), actor, siteID, childID, from, to, true, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toAdHocBookingListResponse(bookings), total, page, pageSize))
}

// createBooking creates a new ad hoc booking.
//
//	@Summary		Create ad hoc booking
//	@Description	Create a new ad hoc booking for a child.
//	@Tags			ad-hoc-bookings
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string					true	"Site ID"	format(uuid)
//	@Param			body	body		createAdHocBookingRequest	true	"Booking data"
//	@Success		201		{object}	object{ad_hoc_booking=adHocBookingResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/ad-hoc-bookings [post]
func (h *Handler) createBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	var req createAdHocBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	childID, err := uuid.Parse(req.ChildID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid child_id.")
		return
	}
	calendarDate, err := time.Parse("2006-01-02", req.CalendarDate)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid calendar_date format.")
		return
	}
	sessionTypeID, err := uuid.Parse(req.SessionTypeID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid session_type_id.")
		return
	}

	params := application.CreateAdHocBookingParams{
		ChildID:       childID,
		CalendarDate:  calendarDate,
		SessionTypeID: sessionTypeID,
	}

	booking, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ad_hoc_booking": toAdHocBookingResponse(booking)})
}

// cancelBooking cancels an ad hoc booking.
//
//	@Summary		Cancel ad hoc booking
//	@Description	Cancel an ad hoc booking.
//	@Tags			ad-hoc-bookings
//	@Produce		json
//	@Param			site_id		path	string	true	"Site ID"		format(uuid)
//	@Param			booking_id	path	string	true	"Booking ID"	format(uuid)
//	@Success		204
//	@Failure		401	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/ad-hoc-bookings/{booking_id}/cancel [post]
func (h *Handler) cancelBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	if err := h.cancel.Execute(c.Request.Context(), actor, siteID, bookingID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
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

func requireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		v, ok := c.Get(tenant.AuthContextKey)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
			return
		}

		authCtx, ok := v.(tenant.AuthorizationContext)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
			return
		}

		switch authCtx.Role {
		case "owner", "manager", "practitioner", "parent":
		default:
			writeError(c, http.StatusForbidden, "forbidden_role_unknown", "Access denied.")
			return
		}

		if _, exists := allowed[authCtx.Role]; !exists {
			writeError(c, http.StatusForbidden, "forbidden_role", "Access denied.")
			return
		}

		c.Next()
	}
}
