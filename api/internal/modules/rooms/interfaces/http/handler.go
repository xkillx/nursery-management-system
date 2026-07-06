package httprooms

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/rooms/application"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
)

type Handler struct {
	logger     *slog.Logger
	create     *application.CreateRoom
	update     *application.UpdateRoom
	list       *application.ListRooms
	get        *application.GetRoom
	archive    *application.ArchiveRoom
	reactivate *application.ReactivateRoom
}

func NewHandler(
	create *application.CreateRoom,
	update *application.UpdateRoom,
	list *application.ListRooms,
	get *application.GetRoom,
	archive *application.ArchiveRoom,
	reactivate *application.ReactivateRoom,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:     logger,
		create:     create,
		update:     update,
		list:       list,
		get:        get,
		archive:    archive,
		reactivate: reactivate,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	readOnly := protected.Group("")
	readOnly.Use(requireRoles("manager", "owner", "practitioner"))
	readOnly.GET("/sites/:site_id/rooms", h.listRooms)
	readOnly.GET("/sites/:site_id/rooms/:room_id", h.getRoom)

	writeOps := protected.Group("")
	writeOps.Use(requireRoles("manager", "owner"))
	writeOps.POST("/sites/:site_id/rooms", h.createRoom)
	writeOps.PATCH("/sites/:site_id/rooms/:room_id", h.updateRoom)
	writeOps.POST("/sites/:site_id/rooms/:room_id/actions/archive", h.archiveRoom)
	writeOps.POST("/sites/:site_id/rooms/:room_id/actions/activate", h.reactivateRoom)
}

func (h *Handler) resolveActor(c *gin.Context) (application.RoomActor, bool) {
	if owner, ok := tenant.OwnerActorFromGinContext(c); ok {
		return application.NewOwnerRoomActor(owner), true
	}

	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		return nil, false
	}

	switch actor.BranchID {
	case uuid.Nil:
		return nil, false
	}

	role := ""
	if v, authOk := c.Get(tenant.AuthContextKey); authOk {
		if authCtx, authOk := v.(tenant.AuthorizationContext); authOk {
			role = authCtx.Role
		}
	}

	switch role {
	case "manager":
		return application.NewManagerRoomActor(actor), true
	case "practitioner":
		return application.NewPractitionerRoomActor(actor), true
	}

	return nil, false
}

// listRooms returns a paginated list of rooms for a site.
//
//	@Summary		List rooms
//	@Description	Get a paginated list of rooms for a site.
//	@Tags			rooms
//	@Produce		json
//	@Param			site_id			path		string	true	"Site ID"	format(uuid)
//	@Param			include_archived	query		bool	false	"Include archived rooms"
//	@Param			include			query		string	false	"Include additional data"	Enums(occupancy)
//	@Param			page			query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size		query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200				{object}	object{items=[]roomResponse,total=int,page=int,page_size=int}
//	@Failure		401				{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner","practitioner"]
//	@Router			/sites/{site_id}/rooms [get]
func (h *Handler) listRooms(c *gin.Context) {
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

	includeArchived := c.Query("include_archived") == "true"
	includeOccupancy := c.Query("include") == "occupancy"

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	rooms, counts, total, err := h.list.ExecutePaginated(c.Request.Context(), actor, siteID, includeArchived, includeOccupancy, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toRoomListResponse(rooms, counts), total, page, pageSize))
}

// createRoom creates a new room for a site.
//
//	@Summary		Create room
//	@Description	Create a new room for a site.
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string			true	"Site ID"	format(uuid)
//	@Param			body	body		createRoomRequest	true	"Room data"
//	@Success		201		{object}	roomResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/rooms [post]
func (h *Handler) createRoom(c *gin.Context) {
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

	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	params := application.CreateRoomParams{
		Name:        req.Name,
		AgeGroup:    req.AgeGroup,
		Capacity:    req.Capacity,
		Description: req.Description,
	}

	room, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toRoomResponse(room))
}

// getRoom returns a single room by ID.
//
//	@Summary		Get room
//	@Description	Get a single room by ID.
//	@Tags			rooms
//	@Produce		json
//	@Param			site_id	path		string	true	"Site ID"	format(uuid)
//	@Param			room_id	path		string	true	"Room ID"	format(uuid)
//	@Success		200		{object}	roomResponse
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner","practitioner"]
//	@Router			/sites/{site_id}/rooms/{room_id} [get]
func (h *Handler) getRoom(c *gin.Context) {
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

	roomID, err := uuid.Parse(c.Param("room_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	room, err := h.get.Execute(c.Request.Context(), actor, siteID, roomID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRoomResponse(room))
}

// updateRoom updates an existing room.
//
//	@Summary		Update room
//	@Description	Update an existing room.
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string			true	"Site ID"	format(uuid)
//	@Param			room_id	path		string			true	"Room ID"	format(uuid)
//	@Param			body	body		updateRoomRequest	true	"Room data"
//	@Success		200		{object}	roomResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/rooms/{room_id} [patch]
func (h *Handler) updateRoom(c *gin.Context) {
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

	roomID, err := uuid.Parse(c.Param("room_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	var req updateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	params := application.UpdateRoomParams{
		Name:        req.Name,
		AgeGroup:    req.AgeGroup,
		Capacity:    req.Capacity,
		Description: req.Description,
	}

	room, err := h.update.Execute(c.Request.Context(), actor, siteID, roomID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRoomResponse(room))
}

// archiveRoom archives a room.
//
//	@Summary		Archive room
//	@Description	Archive a room. Children must be reassigned first.
//	@Tags			rooms
//	@Produce		json
//	@Param			site_id	path	string	true	"Site ID"	format(uuid)
//	@Param			room_id	path	string	true	"Room ID"	format(uuid)
//	@Success		200		{object}	object
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/rooms/{room_id}/actions/archive [post]
func (h *Handler) archiveRoom(c *gin.Context) {
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

	roomID, err := uuid.Parse(c.Param("room_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	if err := h.archive.Execute(c.Request.Context(), actor, siteID, roomID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// reactivateRoom reactivates an archived room.
//
//	@Summary		Reactivate room
//	@Description	Reactivate an archived room.
//	@Tags			rooms
//	@Produce		json
//	@Param			site_id	path	string	true	"Site ID"	format(uuid)
//	@Param			room_id	path	string	true	"Room ID"	format(uuid)
//	@Success		200		{object}	roomResponse
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/rooms/{room_id}/actions/activate [post]
func (h *Handler) reactivateRoom(c *gin.Context) {
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

	roomID, err := uuid.Parse(c.Param("room_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	room, err := h.reactivate.Execute(c.Request.Context(), actor, siteID, roomID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRoomResponse(room))
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
