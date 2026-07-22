package httpchild

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/children/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/http/queryparams"
	"nursery-management-system/api/internal/platform/tenant"
)

type (
	CoreUseCases struct {
		List           *application.ListChildren
		Get            *application.GetChild
		Create         *application.CreateChildWithFullProfile
		Update         *application.UpdateChild
		MarkInactive   *application.MarkInactive
		ListAttendance *application.ListAttendance
	}

	ProfileUseCases struct {
		Get    *application.GetProfile
		Update *application.UpdateProfile
	}

	ContactsUseCases struct {
		Get     *application.GetContacts
		Replace *application.ReplaceContacts
	}

	HealthUseCases struct {
		Get    *application.GetHealth
		Update *application.UpdateHealth
	}

	SafeguardingUseCases struct {
		Get    *application.GetSafeguarding
		Update *application.UpdateSafeguarding
	}

	ConsentUseCases struct {
		Get    *application.GetConsent
		Update *application.UpdateConsent
	}

	CollectionUseCases struct {
		GetSetting  *application.GetCollectionSetting
		SetPassword *application.SetCollectionPassword
	}

	RoomAssignmentUseCases struct {
		List   *application.ListRoomAssignments
		Create *application.CreateRoomAssignment
		Close  *application.CloseRoomAssignment
	}

	BillingProfileUseCases struct {
		Get    *application.GetBillingProfile
		Update *application.UpdateBillingProfile
	}

	ChildrenHandlerConfig struct {
		Core            CoreUseCases
		Profile         ProfileUseCases
		Contacts        ContactsUseCases
		Health          HealthUseCases
		Safeguarding    SafeguardingUseCases
		Consent         ConsentUseCases
		Collection      CollectionUseCases
		RoomAssignments RoomAssignmentUseCases
		BillingProfile  BillingProfileUseCases
		LeavingRecord   *application.GetLeavingRecord
		Photo           PhotoUseCases
	}
)

type Handler struct {
	logger              *slog.Logger
	listChildren        *application.ListChildren
	getChild            *application.GetChild
	createChildWithFull *application.CreateChildWithFullProfile
	updateChild         *application.UpdateChild
	markInactive        *application.MarkInactive
	listAttendance      *application.ListAttendance

	getProfile    *application.GetProfile
	updateProfile *application.UpdateProfile

	getContacts     *application.GetContacts
	replaceContacts *application.ReplaceContacts

	getHealth    *application.GetHealth
	updateHealth *application.UpdateHealth

	getSafeguarding    *application.GetSafeguarding
	updateSafeguarding *application.UpdateSafeguarding

	getConsent    *application.GetConsent
	updateConsent *application.UpdateConsent

	getCollectionSetting  *application.GetCollectionSetting
	setCollectionPassword *application.SetCollectionPassword

	listRoomAssignments  *application.ListRoomAssignments
	createRoomAssignment *application.CreateRoomAssignment
	closeRoomAssignment  *application.CloseRoomAssignment

	getBillingProfile    *application.GetBillingProfile
	updateBillingProfile *application.UpdateBillingProfile

	getLeavingRecord *application.GetLeavingRecord

	uploadPhoto *application.UploadPhoto
	removePhoto *application.RemovePhoto
}

func NewHandler(cfg ChildrenHandlerConfig, logger *slog.Logger) *Handler {
	return &Handler{
		logger:                logger,
		listChildren:          cfg.Core.List,
		getChild:              cfg.Core.Get,
		createChildWithFull:   cfg.Core.Create,
		updateChild:           cfg.Core.Update,
		markInactive:          cfg.Core.MarkInactive,
		listAttendance:        cfg.Core.ListAttendance,
		getProfile:            cfg.Profile.Get,
		updateProfile:         cfg.Profile.Update,
		getContacts:           cfg.Contacts.Get,
		replaceContacts:       cfg.Contacts.Replace,
		getHealth:             cfg.Health.Get,
		updateHealth:          cfg.Health.Update,
		getSafeguarding:       cfg.Safeguarding.Get,
		updateSafeguarding:    cfg.Safeguarding.Update,
		getConsent:            cfg.Consent.Get,
		updateConsent:         cfg.Consent.Update,
		getCollectionSetting:  cfg.Collection.GetSetting,
		setCollectionPassword: cfg.Collection.SetPassword,
		listRoomAssignments:   cfg.RoomAssignments.List,
		createRoomAssignment:  cfg.RoomAssignments.Create,
		closeRoomAssignment:   cfg.RoomAssignments.Close,
		getBillingProfile:     cfg.BillingProfile.Get,
		updateBillingProfile:  cfg.BillingProfile.Update,
		getLeavingRecord:      cfg.LeavingRecord,
		uploadPhoto:           cfg.Photo.Upload,
		removePhoto:           cfg.Photo.Remove,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.GET("/children/attendance", httpserver.RequireRolesWithObservability(h.logger, nil, "manager", "practitioner"), h.listAttendanceHandler)

	manager := protected.Group("")
	manager.Use(httpserver.RequireRolesWithObservability(h.logger, nil, "manager"))

	manager.GET("/children", h.listChildrenHandler)
	manager.POST("/children", h.createChildHandler)
	manager.GET("/children/:child_id", h.getChildHandler)
	manager.PATCH("/children/:child_id", h.updateChildHandler)
	manager.POST("/children/:child_id/actions/mark-inactive", h.markInactiveHandler)

	manager.GET("/children/:child_id/profile", h.getProfileHandler)
	manager.PATCH("/children/:child_id/profile", h.updateProfileHandler)

	manager.GET("/children/:child_id/contacts", h.getContactsHandler)
	manager.PUT("/children/:child_id/contacts", h.replaceContactsHandler)

	manager.GET("/children/:child_id/health", h.getHealthHandler)
	manager.PATCH("/children/:child_id/health", h.updateHealthHandler)

	manager.GET("/children/:child_id/safeguarding", h.getSafeguardingHandler)
	manager.PATCH("/children/:child_id/safeguarding", h.updateSafeguardingHandler)

	manager.GET("/children/:child_id/consent", h.getConsentHandler)
	manager.PUT("/children/:child_id/consent", h.updateConsentHandler)

	manager.GET("/children/:child_id/collection-settings", h.getCollectionSettingHandler)
	manager.PUT("/children/:child_id/collection-settings", h.setCollectionSettingHandler)

	manager.GET("/children/:child_id/room-assignments", h.listRoomAssignmentsHandler)
	manager.POST("/children/:child_id/room-assignments", h.createRoomAssignmentHandler)
	manager.DELETE("/children/:child_id/room-assignments/:assignment_id", h.closeRoomAssignmentHandler)

	manager.GET("/children/:child_id/billing-profile", h.getBillingProfileHandler)
	manager.PATCH("/children/:child_id/billing-profile", h.updateBillingProfileHandler)

	manager.GET("/children/:child_id/leaving-record", h.getLeavingRecordHandler)

	manager.PUT("/children/:child_id/photo", h.uploadPhotoHandler)
	manager.DELETE("/children/:child_id/photo", h.removePhotoHandler)
	manager.GET("/children/:child_id/photo", h.getPhotoHandler)
}

// listChildrenHandler returns a paginated list of children.
//
//	@Summary		List children
//	@Description	Get a paginated list of children for the current branch.
//	@Tags			children
//	@Produce		json
//	@Param			status		query		string	false	"Filter by status"	Enums(active, inactive, all)
//	@Param			page		query		int		false	"Page number"		default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]childResponse,total=int,page=int,page_size=int}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children [get]
func (h *Handler) listChildrenHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	filters := queryparams.ParseFilterParams(c, map[string]string{
		"status":  "string",
		"room_id": "uuid",
	})

	statusFilter := filters["status"]
	if statusFilter == "" {
		statusFilter = "all"
	}

	var roomID *uuid.UUID
	if rid, ok := filters["room_id"]; ok {
		parsed, err := uuid.Parse(rid)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Validation failed.", []map[string]string{{"field": "room_id", "message": "must be a valid UUID"}})
			return
		}
		roomID = &parsed
	}

	allowedSorts := map[string][]string{
		"name":       {"asc", "desc"},
		"created_at": {"asc", "desc"},
	}
	sortExpr, err := queryparams.ParseSortParams(c, allowedSorts)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	children, err := h.listChildren.Execute(c.Request.Context(), actor, statusFilter, pageSize, offset, roomID, sortExpr.Field, sortExpr.Direction)
	if err != nil {
		h.handleError(c, err)
		return
	}

	total, err := h.listChildren.Count(c.Request.Context(), actor, statusFilter, roomID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	out := make([]childResponse, 0, len(children))
	for _, child := range children {
		out = append(out, toChildResponse(child))
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(out, total, page, pageSize))
}

// getChildHandler returns a single child by ID.
//
//	@Summary		Get child
//	@Description	Get a single child by ID.
//	@Tags			children
//	@Produce		json
//	@Param			child_id	path		string	true	"Child ID"	format(uuid)
//	@Success		200			{object}	childResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children/{child_id} [get]
func (h *Handler) getChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	child, err := h.getChild.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildResponse(child))
}

// createChildHandler creates a new child with full profile.
//
//	@Summary		Create child
//	@Description	Create a new child with full profile (contacts, health, etc.).
//	@Tags			children
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createChildRequest	true	"Child data"
//	@Success		201		{object}	childCreationResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children [post]
func (h *Handler) createChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req createChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	input, err := mapCreateChildRequest(req)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	result, err := h.createChildWithFull.Execute(c.Request.Context(), actor, input)
	if err != nil {
		h.handleError(c, err)
		return
	}
	resp := toChildCreationResponse(result)
	c.Header("Location", fmt.Sprintf("/api/children/%s", resp.ID))
	c.JSON(http.StatusCreated, resp)
}

// updateChildHandler updates an existing child.
//
//	@Summary		Update child
//	@Description	Update an existing child's basic information.
//	@Tags			children
//	@Accept			json
//	@Produce		json
//	@Param			child_id	path		string				true	"Child ID"	format(uuid)
//	@Param			body		body		childWriteRequest	true	"Child data"
//	@Success		200			{object}	childResponse
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children/{child_id} [patch]
func (h *Handler) updateChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	params := application.UpdateChildParams{
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Notes:       req.Notes,
	}
	child, err := h.updateChild.Execute(c.Request.Context(), actor, c.Param("child_id"), params)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildResponse(child))
}

// markInactiveHandler marks a child as inactive.
//
//	@Summary		Mark child inactive
//	@Description	Mark a child as inactive with a reason.
//	@Tags			children
//	@Accept			json
//	@Produce		json
//	@Param			child_id	path		string			true	"Child ID"	format(uuid)
//	@Param			body		body		reasonRequest	true	"Reason for marking inactive"
//	@Success		200			{object}	childResponse
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children/{child_id}/actions/mark-inactive [post]
func (h *Handler) markInactiveHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req reasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	child, err := h.markInactive.Execute(c.Request.Context(), actor, c.Param("child_id"), application.MarkInactiveParams{
		ReasonCode: req.ReasonCode,
		ReasonNote: req.ReasonNote,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildResponse(child))
}

// listAttendanceHandler returns children with attendance state for today.
//
//	@Summary		List attendance children
//	@Description	Get children with their attendance state for today.
//	@Tags			children
//	@Produce		json
//	@Success		200	{object}	object{items=[]attendanceChildResponse}
//	@Failure		401	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","practitioner"]
//	@Router			/children/attendance [get]
func (h *Handler) listAttendanceHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	children, err := h.listAttendance.Execute(c.Request.Context(), actor)
	if err != nil {
		h.handleError(c, err)
		return
	}
	out := make([]attendanceChildResponse, 0, len(children))
	for _, child := range children {
		out = append(out, toAttendanceResponse(child))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

// --- Sub-resource handlers ---

func (h *Handler) getProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	p, err := h.getProfile.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildProfileResponse(p))
}

func (h *Handler) updateProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	p, err := h.updateProfile.Execute(c.Request.Context(), actor, c.Param("child_id"), mapChildProfileRequest(req))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildProfileResponse(p))
}

func (h *Handler) getContactsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	contacts, err := h.getContacts.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildContactsResponse(contacts))
}

func (h *Handler) replaceContactsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	contacts, err := h.replaceContacts.Execute(c.Request.Context(), actor, c.Param("child_id"), mapChildContactsRequest(req))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildContactsResponse(contacts))
}

func (h *Handler) getHealthHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	p, err := h.getHealth.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildHealthResponse(p))
}

func (h *Handler) updateHealthHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childHealthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	p, err := h.updateHealth.Execute(c.Request.Context(), actor, c.Param("child_id"), mapChildHealthRequest(req))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildHealthResponse(p))
}

func (h *Handler) getSafeguardingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	p, err := h.getSafeguarding.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildSafeguardingResponse(p))
}

func (h *Handler) updateSafeguardingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childSafeguardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	p, err := h.updateSafeguarding.Execute(c.Request.Context(), actor, c.Param("child_id"), mapChildSafeguardingRequest(req))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildSafeguardingResponse(p))
}

func (h *Handler) getConsentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	p, err := h.getConsent.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	if p == nil {
		c.JSON(http.StatusOK, gin.H{"consent": nil})
		return
	}
	c.JSON(http.StatusOK, toChildConsentResponse(p))
}

func (h *Handler) updateConsentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	p, err := h.updateConsent.Execute(c.Request.Context(), actor, c.Param("child_id"), mapChildConsentRequest(req))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildConsentResponse(p))
}

func (h *Handler) getCollectionSettingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	p, err := h.getCollectionSetting.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	if p == nil {
		c.JSON(http.StatusOK, gin.H{"collection_settings": nil})
		return
	}
	c.JSON(http.StatusOK, toChildCollectionSettingsResponse(p))
}

func (h *Handler) setCollectionSettingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childCollectionSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	p, err := h.setCollectionPassword.Execute(c.Request.Context(), actor, c.Param("child_id"), application.SetCollectionPasswordInput{
		Over18CollectionAcknowledged: req.Over18CollectionAcknowledged,
		Password:                     req.Password,
		PasswordHint:                 req.PasswordHint,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildCollectionSettingsResponse(p))
}

// listRoomAssignmentsHandler returns a paginated list of room assignments for a child.
//
//	@Summary		List room assignments
//	@Description	Get a paginated list of room assignments for a child.
//	@Tags			children
//	@Produce		json
//	@Param			child_id	path		string	true	"Child ID"			format(uuid)
//	@Param			page		query		int		false	"Page number"		default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]roomAssignmentResponse,total=int,page=int,page_size=int}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/children/{child_id}/room-assignments [get]
func (h *Handler) listRoomAssignmentsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	items, total, err := h.listRoomAssignments.ExecutePaginated(c.Request.Context(), actor, c.Param("child_id"), pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}
	out := make([]roomAssignmentResponse, 0, len(items))
	for _, it := range items {
		out = append(out, toRoomAssignmentResponse(it))
	}
	c.JSON(http.StatusOK, pagination.PaginatedResponse(out, total, page, pageSize))
}

func (h *Handler) createRoomAssignmentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req roomAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	a, err := h.createRoomAssignment.Execute(c.Request.Context(), actor, c.Param("child_id"), application.CreateRoomAssignmentInput{
		RoomID:    req.RoomID,
		StartDate: req.StartDate,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	resp := toRoomAssignmentResponse(*a)
	c.Header("Location", fmt.Sprintf("/api/children/%s/room-assignments/%s", resp.ChildID, resp.ID))
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) closeRoomAssignmentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	if err := h.closeRoomAssignment.Execute(c.Request.Context(), actor, c.Param("child_id"), c.Param("assignment_id")); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) getBillingProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	p, err := h.getBillingProfile.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	if p == nil {
		c.JSON(http.StatusOK, gin.H{"billing_profile": nil})
		return
	}
	c.JSON(http.StatusOK, toChildBillingProfileResponse(p))
}

func (h *Handler) updateBillingProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	var req childBillingProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}
	p, err := h.updateBillingProfile.Execute(c.Request.Context(), actor, c.Param("child_id"), application.UpdateBillingProfileInput{
		BillingBasis:    req.BillingBasis,
		CustomRateMinor: req.CustomRateMinor,
		EffectiveFrom:   req.EffectiveFrom,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildBillingProfileResponse(p))
}

func (h *Handler) getLeavingRecordHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}
	p, err := h.getLeavingRecord.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	if p == nil {
		c.JSON(http.StatusOK, gin.H{"leaving_record": nil})
		return
	}
	c.JSON(http.StatusOK, toChildLeavingRecordResponse(p))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}
