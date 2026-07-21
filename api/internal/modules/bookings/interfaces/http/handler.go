package httpbookings

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/application"
	"nursery-management-system/api/internal/modules/bookings/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger              *slog.Logger
	create              *application.CreateBooking
	get                 *application.GetBooking
	list                *application.ListBookings
	update              *application.UpdateBooking
	cancel              *application.CancelBooking
	pause               *application.PauseBooking
	clone               *application.CloneBooking
	listCapacity        *application.ListCapacity
	listParentBookings  *application.ListParentBookings
	createBookingReq    *application.CreateBookingRequest
	cancelParentBooking *application.CancelParentBooking
}

func NewHandler(
	create *application.CreateBooking,
	get *application.GetBooking,
	list *application.ListBookings,
	update *application.UpdateBooking,
	cancel *application.CancelBooking,
	pause *application.PauseBooking,
	clone *application.CloneBooking,
	listCapacity *application.ListCapacity,
	listParentBookings *application.ListParentBookings,
	createBookingReq *application.CreateBookingRequest,
	cancelParentBooking *application.CancelParentBooking,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:              logger,
		create:              create,
		get:                 get,
		list:                list,
		update:              update,
		cancel:              cancel,
		pause:               pause,
		clone:               clone,
		listCapacity:        listCapacity,
		listParentBookings:  listParentBookings,
		createBookingReq:    createBookingReq,
		cancelParentBooking: cancelParentBooking,
	}
}

func (h *Handler) RegisterManagerRoutes(manager *gin.RouterGroup) {
	manager.GET("/sites/:site_id/bookings", h.listBookings)
	manager.GET("/sites/:site_id/bookings/capacity", h.listCapacityHandler)
	manager.GET("/sites/:site_id/bookings/:booking_id", h.getBooking)
	manager.POST("/sites/:site_id/bookings", h.createBooking)
	manager.PATCH("/sites/:site_id/bookings/:booking_id", h.updateBooking)
	manager.POST("/sites/:site_id/bookings/:booking_id/clone", h.cloneBooking)
	manager.POST("/sites/:site_id/bookings/:booking_id/cancel", h.cancelBooking)
	manager.POST("/sites/:site_id/bookings/:booking_id/pause", h.pauseBooking)
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.GET("/bookings", h.listParentBookingsHandler)
	parent.GET("/bookings/recurring", h.listParentRecurringHandler)
	parent.POST("/bookings/requests", h.createBookingRequestHandler)
	parent.POST("/bookings/:booking_id/cancel", h.cancelParentBookingHandler)
}

// listParentBookingsHandler returns upcoming bookings for the parent's children.
//
//	@Summary		List parent bookings
//	@Description	Get upcoming bookings for the authenticated parent's children.
//	@Tags			parent-bookings
//	@Produce		json
//	@Param			from	query		string	false	"From date"	format(date)
//	@Param			to		query		string	false	"To date"	format(date)
//	@Success		200		{object}	object{items=[]unifiedBookingResponse}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["parent"]
//	@Router			/parent/bookings [get]
func (h *Handler) listParentBookingsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID := actor.BranchID
	parentActor := application.NewParentBookingActor(actor)

	now := time.Now()
	from := now
	to := now.AddDate(0, 0, 7)

	if v := c.Query("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			to = t
		}
	}

	rows, err := h.listParentBookings.Execute(c.Request.Context(), parentActor, siteID, from, to)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": toUnifiedBookingListResponse(rows)})
}

// listParentRecurringHandler returns recurring booking patterns for the parent's children.
//
//	@Summary		List parent recurring bookings
//	@Description	Get recurring booking patterns for the authenticated parent's children.
//	@Tags			parent-bookings
//	@Produce		json
//	@Success		200	{object}	object{items=[]bookingResponse}
//	@Failure		401	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["parent"]
//	@Router			/parent/bookings/recurring [get]
func (h *Handler) listParentRecurringHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID := actor.BranchID
	parentActor := application.NewParentBookingActor(actor)

	childLook := h.listParentBookings
	_ = childLook

	filters := domain.ListFilters{
		ActiveOnly: true,
	}
	bookings, _, err := h.list.ExecutePaginated(c.Request.Context(), parentActor, siteID, filters, 200, 0)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]bookingResponse, 0, len(bookings))
	for _, b := range bookings {
		items = append(items, toBookingResponse(b))
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// createBookingRequestHandler creates a new booking request from a parent.
//
//	@Summary		Create booking request
//	@Description	Create a new booking request from a parent for their child.
//	@Tags			parent-bookings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createBookingRequest	true	"Booking request data"
//	@Success		201		{object}	object{booking=bookingResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["parent"]
//	@Router			/parent/bookings/requests [post]
func (h *Handler) createBookingRequestHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID := actor.BranchID
	parentActor := application.NewParentBookingActor(actor)

	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params, err := parseCreateRequest(req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	booking, err := h.createBookingReq.Execute(c.Request.Context(), parentActor, siteID, application.CreateBookingRequestParams{
		ChildID:             params.ChildID,
		SessionTemplateID:   *params.SessionTemplateID,
		DaysOfWeek:          params.DaysOfWeek,
		EffectiveStartDate:  params.EffectiveStartDate,
		EffectiveEndDate:    params.EffectiveEndDate,
		FundingType:         params.FundingType,
		FundingHoursPerWeek: params.FundingHoursPerWeek,
		LaReference:         params.LaReference,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toBookingResponse(booking)
	c.Header("Location", fmt.Sprintf("/api/parent/bookings/%s", resp.ID))
	c.JSON(http.StatusCreated, gin.H{"booking": resp})
}

// cancelParentBookingHandler cancels a booking for a parent.
//
//	@Summary		Cancel parent booking
//	@Description	Cancel an eligible booking for the authenticated parent's child.
//	@Tags			parent-bookings
//	@Produce		json
//	@Param			booking_id	path	string	true	"Booking ID"	format(uuid)
//	@Success		204
//	@Failure		401	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Failure		409	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["parent"]
//	@Router			/parent/bookings/{booking_id}/cancel [post]
func (h *Handler) cancelParentBookingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID := actor.BranchID
	parentActor := application.NewParentBookingActor(actor)

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid booking_id.", nil)
		return
	}

	if err := h.cancelParentBooking.Execute(c.Request.Context(), parentActor, siteID, bookingID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) resolveActor(c *gin.Context) (application.BookingActor, bool) {
	if owner, ok := tenant.OwnerActorFromGinContext(c); ok {
		return application.NewOwnerBookingActor(owner), true
	}

	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		return nil, false
	}

	if actor.BranchID == uuid.Nil {
		return nil, false
	}

	return application.NewManagerBookingActor(actor), true
}

// listBookings returns a paginated list of bookings (unified or recurring-only).
//
//	@Summary		List bookings
//	@Description	Get a paginated list of bookings for a site. Supports unified view (recurring + ad-hoc + hourly) or recurring-only.
//	@Tags			bookings
//	@Produce		json
//	@Param			site_id			path		string	true	"Site ID"					format(uuid)
//	@Param			child_id		query		string	false	"Filter by child ID"		format(uuid)
//	@Param			session_type_id	query		string	false	"Filter by session type"	format(uuid)
//	@Param			status			query		string	false	"Filter by status"			Enums(active, paused, cancelled)
//	@Param			funding_type	query		string	false	"Filter by funding type"
//	@Param			search			query		string	false	"Search by child name"
//	@Param			from			query		string	false	"Filter from date"	format(date)
//	@Param			to				query		string	false	"Filter to date"	format(date)
//	@Param			view			query		string	false	"View mode"			Enums(list, calendar)	default(list)
//	@Param			page			query		int		false	"Page number"		default(1)				minimum(1)
//	@Param			page_size		query		int		false	"Items per page"	default(50)				minimum(1)	maximum(200)
//	@Success		200				{object}	object{items=[]unifiedBookingResponse,total=int,page=int,page_size=int}
//	@Failure		400				{object}	object{code=string,message=string}
//	@Failure		401				{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings [get]
func (h *Handler) listBookings(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	filters, err := parseListFilters(c)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	// Use unified list (includes recurring + ad-hoc + hourly)
	bookings, err := h.list.ExecuteUnified(c.Request.Context(), actor, siteID, filters, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toUnifiedBookingListResponse(bookings), len(bookings), page, pageSize))
}

// getBooking returns a single booking.
//
//	@Summary		Get booking
//	@Description	Get a single booking by ID.
//	@Tags			bookings
//	@Produce		json
//	@Param			site_id		path		string	true	"Site ID"		format(uuid)
//	@Param			booking_id	path		string	true	"Booking ID"	format(uuid)
//	@Success		200			{object}	object{booking=bookingResponse}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings/{booking_id} [get]
func (h *Handler) getBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid booking_id.", nil)
		return
	}

	booking, err := h.get.Execute(c.Request.Context(), actor, siteID, bookingID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": toBookingResponse(booking)})
}

// createBooking creates a new recurring booking.
//
//	@Summary		Create recurring booking
//	@Description	Create a new recurring booking for a child.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string					true	"Site ID"	format(uuid)
//	@Param			body	body		createBookingRequest	true	"Booking data"
//	@Success		201		{object}	object{booking=bookingResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		422		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings [post]
func (h *Handler) createBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params, err := parseCreateRequest(req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	booking, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toBookingResponse(booking)
	c.Header("Location", fmt.Sprintf("/api/sites/%s/bookings/%s", siteID, resp.ID))
	c.JSON(http.StatusCreated, gin.H{"booking": resp})
}

// updateBooking updates an existing booking.
//
//	@Summary		Update booking
//	@Description	Update an existing booking's fields.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Param			site_id		path		string					true	"Site ID"		format(uuid)
//	@Param			booking_id	path		string					true	"Booking ID"	format(uuid)
//	@Param			body		body		updateBookingRequest	true	"Fields to update"
//	@Success		200			{object}	object{booking=bookingResponse}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Failure		409			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings/{booking_id} [patch]
func (h *Handler) updateBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid booking_id.", nil)
		return
	}

	var req updateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params, err := parseUpdateRequest(req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	booking, err := h.update.Execute(c.Request.Context(), actor, siteID, bookingID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": toBookingResponse(booking)})
}

// cloneBooking clones an existing booking.
//
//	@Summary		Clone booking
//	@Description	Clone an existing booking, optionally for a different child.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Param			site_id		path		string				true	"Site ID"		format(uuid)
//	@Param			booking_id	path		string				true	"Booking ID"	format(uuid)
//	@Param			body		body		cloneBookingRequest	false	"Optional child override"
//	@Success		201			{object}	object{booking=bookingResponse}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Failure		409			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings/{booking_id}/clone [post]
func (h *Handler) cloneBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid booking_id.", nil)
		return
	}

	var req cloneBookingRequest
	// Body is optional for clone
	_ = c.ShouldBindJSON(&req)

	var params application.CloneBookingParams
	if req.ChildID != nil {
		childID, err := uuid.Parse(*req.ChildID)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid child_id.", nil)
			return
		}
		params.ChildID = &childID
	}

	booking, err := h.clone.Execute(c.Request.Context(), actor, siteID, bookingID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toBookingResponse(booking)
	c.Header("Location", fmt.Sprintf("/api/sites/%s/bookings/%s", siteID, resp.ID))
	c.JSON(http.StatusCreated, gin.H{"booking": resp})
}

// cancelBooking cancels a booking.
//
//	@Summary		Cancel booking
//	@Description	Cancel a recurring booking.
//	@Tags			bookings
//	@Produce		json
//	@Param			site_id		path	string	true	"Site ID"		format(uuid)
//	@Param			booking_id	path	string	true	"Booking ID"	format(uuid)
//	@Success		204
//	@Failure		401	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Failure		409	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings/{booking_id}/cancel [post]
func (h *Handler) cancelBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid booking_id.", nil)
		return
	}

	if err := h.cancel.Execute(c.Request.Context(), actor, siteID, bookingID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// pauseBooking pauses a booking.
//
//	@Summary		Pause booking
//	@Description	Pause a recurring booking (temporarily inactive).
//	@Tags			bookings
//	@Produce		json
//	@Param			site_id		path		string	true	"Site ID"		format(uuid)
//	@Param			booking_id	path		string	true	"Booking ID"	format(uuid)
//	@Success		200			{object}	object{booking=bookingResponse}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Failure		409			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings/{booking_id}/pause [post]
func (h *Handler) pauseBooking(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	bookingID, err := uuid.Parse(c.Param("booking_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid booking_id.", nil)
		return
	}

	booking, err := h.pause.Execute(c.Request.Context(), actor, siteID, bookingID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": toBookingResponse(booking)})
}

// listCapacityHandler returns room capacity snapshot for a date range.
//
//	@Summary		Room capacity snapshot
//	@Description	Get room capacity and booking counts for a date range.
//	@Tags			bookings
//	@Produce		json
//	@Param			site_id	path		string	true	"Site ID"	format(uuid)
//	@Param			from	query		string	true	"From date"	format(date)
//	@Param			to		query		string	true	"To date"	format(date)
//	@Success		200		{object}	object{items=[]roomCapacityEntryResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/bookings/capacity [get]
func (h *Handler) listCapacityHandler(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site_id.", nil)
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "from and to query parameters are required.", nil)
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid from date format.", nil)
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid to date format.", nil)
		return
	}

	entries, err := h.listCapacity.Execute(c.Request.Context(), actor, siteID, from, to)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": toRoomCapacityListResponse(entries)})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}

func parseListFilters(c *gin.Context) (domain.ListFilters, error) {
	var filters domain.ListFilters

	if v := c.Query("child_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return domain.ListFilters{}, fmt.Errorf("invalid child_id")
		}
		filters.ChildID = &id
	}
	if v := c.Query("session_type_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return domain.ListFilters{}, fmt.Errorf("invalid session_type_id")
		}
		filters.SessionTypeID = &id
	}
	if v := c.Query("status"); v != "" {
		filters.Status = &v
	}
	if v := c.Query("funding_type"); v != "" {
		filters.FundingType = &v
	}
	if v := c.Query("search"); v != "" {
		filters.Search = &v
	}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return domain.ListFilters{}, fmt.Errorf("invalid from date")
		}
		filters.From = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return domain.ListFilters{}, fmt.Errorf("invalid to date")
		}
		filters.To = &t
	}

	return filters, nil
}
