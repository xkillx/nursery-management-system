package httphourlybookings

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/hourly_bookings/application"
	"nursery-management-system/api/internal/modules/hourly_bookings/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"
)

type hourlyBookingResponse struct {
	ID               string  `json:"id"`
	ChildID          string  `json:"child_id"`
	CalendarDate     string  `json:"calendar_date"`
	StartTimeMinutes int     `json:"start_time_minutes"`
	DurationMinutes  int     `json:"duration_minutes"`
	SessionTypeID    *string `json:"session_type_id,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type createHourlyBookingRequest struct {
	ChildID          string  `json:"child_id" binding:"required"`
	CalendarDate     string  `json:"calendar_date" binding:"required"`
	StartTimeMinutes int     `json:"start_time_minutes" binding:"required"`
	DurationMinutes  int     `json:"duration_minutes" binding:"required"`
	SessionTypeID    *string `json:"session_type_id,omitempty"`
}

func toHourlyBookingResponse(b domain.HourlyBooking) hourlyBookingResponse {
	resp := hourlyBookingResponse{
		ID:               b.ID.String(),
		ChildID:          b.ChildID.String(),
		CalendarDate:     b.CalendarDate.UTC().Format("2006-01-02"),
		StartTimeMinutes: b.StartTimeMinutes,
		DurationMinutes:  b.DurationMinutes,
		Status:           b.Status,
		CreatedAt:        b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        b.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if b.SessionTypeID != nil {
		s := b.SessionTypeID.String()
		resp.SessionTypeID = &s
	}
	return resp
}

func toHourlyBookingListResponse(items []domain.HourlyBooking) []hourlyBookingResponse {
	out := make([]hourlyBookingResponse, 0, len(items))
	for _, b := range items {
		out = append(out, toHourlyBookingResponse(b))
	}
	return out
}

type Handler struct {
	logger *slog.Logger
	create *application.CreateHourlyBooking
	list   *application.ListHourlyBookings
	cancel *application.CancelHourlyBooking
}

func NewHandler(
	create *application.CreateHourlyBooking,
	list *application.ListHourlyBookings,
	cancel *application.CancelHourlyBooking,
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
	manager.GET("/sites/:site_id/hourly-bookings", h.listBookings)
	manager.POST("/sites/:site_id/hourly-bookings", h.createBooking)
	manager.POST("/sites/:site_id/hourly-bookings/:booking_id/cancel", h.cancelBooking)
}

func (h *Handler) resolveActor(c *gin.Context) (application.HourlyBookingActor, bool) {
	if owner, ok := tenant.OwnerActorFromGinContext(c); ok {
		return application.NewOwnerHourlyBookingActor(owner), true
	}

	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		return nil, false
	}

	if actor.BranchID == uuid.Nil {
		return nil, false
	}

	return application.NewManagerHourlyBookingActor(actor), true
}

// listBookings returns a paginated list of hourly bookings.
//
//	@Summary		List hourly bookings
//	@Description	Get a paginated list of hourly bookings for a site.
//	@Tags			hourly-bookings
//	@Produce		json
//	@Param			site_id		path		string	true	"Site ID"	format(uuid)
//	@Param			child_id	query		string	false	"Filter by child ID"	format(uuid)
//	@Param			from		query		string	false	"Filter from date"		format(date)
//	@Param			to			query		string	false	"Filter to date"		format(date)
//	@Param			page		query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]hourlyBookingResponse,total=int,page=int,page_size=int}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/hourly-bookings [get]
func (h *Handler) listBookings(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	var childID *uuid.UUID
	if cid := c.Query("child_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid child_id.", nil)
			return
		}
		childID = &id
	}

	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid from format.", nil)
			return
		}
		from = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid to format.", nil)
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

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toHourlyBookingListResponse(bookings), total, page, pageSize))
}

// createBooking creates a new hourly booking.
//
//	@Summary		Create hourly booking
//	@Description	Create a new hourly booking for a child.
//	@Tags			hourly-bookings
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string					true	"Site ID"	format(uuid)
//	@Param			body	body		createHourlyBookingRequest	true	"Booking data"
//	@Success		201		{object}	object{hourly_booking=hourlyBookingResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/hourly-bookings [post]
func (h *Handler) createBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	var req createHourlyBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	childID, err := uuid.Parse(req.ChildID)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid child_id.", nil)
		return
	}
	calendarDate, err := time.Parse("2006-01-02", req.CalendarDate)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid calendar_date format.", nil)
		return
	}

	var sessionTypeID *uuid.UUID
	if req.SessionTypeID != nil {
		id, err := uuid.Parse(*req.SessionTypeID)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid session_type_id.", nil)
			return
		}
		sessionTypeID = &id
	}

	params := application.CreateHourlyBookingParams{
		ChildID:          childID,
		CalendarDate:     calendarDate,
		StartTimeMinutes: req.StartTimeMinutes,
		DurationMinutes:  req.DurationMinutes,
		SessionTypeID:    sessionTypeID,
	}

	booking, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toHourlyBookingResponse(booking)
	c.Header("Location", fmt.Sprintf("/api/sites/%s/hourly-bookings/%s", siteID, resp.ID))
	c.JSON(http.StatusCreated, gin.H{"hourly_booking": resp})
}

// cancelBooking cancels an hourly booking.
//
//	@Summary		Cancel hourly booking
//	@Description	Cancel an hourly booking.
//	@Tags			hourly-bookings
//	@Produce		json
//	@Param			site_id		path	string	true	"Site ID"		format(uuid)
//	@Param			booking_id	path	string	true	"Booking ID"	format(uuid)
//	@Success		204
//	@Failure		401	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/hourly-bookings/{booking_id}/cancel [post]
func (h *Handler) cancelBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	if err := h.cancel.Execute(c.Request.Context(), actor, siteID, bookingID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}
