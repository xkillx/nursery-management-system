package httpchild

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/children/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger                  *slog.Logger
	listChildren            *application.ListChildren
	getChild                *application.GetChild
	createChildWithFull     *application.CreateChildWithFullProfile
	updateChild             *application.UpdateChild
	markInactive            *application.MarkInactive
	listAttendance          *application.ListAttendance

	getProfile              *application.GetProfile
	updateProfile           *application.UpdateProfile

	getContacts             *application.GetContacts
	replaceContacts         *application.ReplaceContacts

	getHealth               *application.GetHealth
	updateHealth            *application.UpdateHealth

	getSafeguarding         *application.GetSafeguarding
	updateSafeguarding      *application.UpdateSafeguarding

	getConsent              *application.GetConsent
	updateConsent           *application.UpdateConsent

	getFunding              *application.GetFunding
	updateFunding           *application.UpdateFunding

	getCollectionSetting    *application.GetCollectionSetting
	setCollectionPassword   *application.SetCollectionPassword

	listRoomAssignments     *application.ListRoomAssignments
	createRoomAssignment    *application.CreateRoomAssignment
	closeRoomAssignment     *application.CloseRoomAssignment

	getBillingProfile       *application.GetBillingProfile
	updateBillingProfile    *application.UpdateBillingProfile

	getLeavingRecord        *application.GetLeavingRecord

	listBookingPatterns     *application.ListBookingPatterns
	getBookingPattern       *application.GetBookingPattern
	getCurrentBookingPattern *application.GetCurrentBookingPattern
	createBookingPattern    *application.CreateBookingPattern
	updateBookingPattern    *application.UpdateBookingPattern
}

func NewHandler(
	listChildren *application.ListChildren,
	getChild *application.GetChild,
	createChildWithFull *application.CreateChildWithFullProfile,
	updateChild *application.UpdateChild,
	markInactive *application.MarkInactive,
	listAttendance *application.ListAttendance,
	getProfile *application.GetProfile,
	updateProfile *application.UpdateProfile,
	getContacts *application.GetContacts,
	replaceContacts *application.ReplaceContacts,
	getHealth *application.GetHealth,
	updateHealth *application.UpdateHealth,
	getSafeguarding *application.GetSafeguarding,
	updateSafeguarding *application.UpdateSafeguarding,
	getConsent *application.GetConsent,
	updateConsent *application.UpdateConsent,
	getFunding *application.GetFunding,
	updateFunding *application.UpdateFunding,
	getCollectionSetting *application.GetCollectionSetting,
	setCollectionPassword *application.SetCollectionPassword,
	listRoomAssignments *application.ListRoomAssignments,
	createRoomAssignment *application.CreateRoomAssignment,
	closeRoomAssignment *application.CloseRoomAssignment,
	getBillingProfile *application.GetBillingProfile,
	updateBillingProfile *application.UpdateBillingProfile,
	getLeavingRecord *application.GetLeavingRecord,
	listBookingPatterns *application.ListBookingPatterns,
	getBookingPattern *application.GetBookingPattern,
	getCurrentBookingPattern *application.GetCurrentBookingPattern,
	createBookingPattern *application.CreateBookingPattern,
	updateBookingPattern *application.UpdateBookingPattern,
) *Handler {
	return &Handler{
		listChildren: listChildren, getChild: getChild, createChildWithFull: createChildWithFull,
		updateChild: updateChild, markInactive: markInactive, listAttendance: listAttendance,
		getProfile: getProfile, updateProfile: updateProfile,
		getContacts: getContacts, replaceContacts: replaceContacts,
		getHealth: getHealth, updateHealth: updateHealth,
		getSafeguarding: getSafeguarding, updateSafeguarding: updateSafeguarding,
		getConsent: getConsent, updateConsent: updateConsent,
		getFunding: getFunding, updateFunding: updateFunding,
		getCollectionSetting: getCollectionSetting, setCollectionPassword: setCollectionPassword,
		listRoomAssignments: listRoomAssignments, createRoomAssignment: createRoomAssignment,
		closeRoomAssignment: closeRoomAssignment,
		getBillingProfile: getBillingProfile, updateBillingProfile: updateBillingProfile,
		getLeavingRecord: getLeavingRecord,
		listBookingPatterns: listBookingPatterns, getBookingPattern: getBookingPattern,
		getCurrentBookingPattern: getCurrentBookingPattern,
		createBookingPattern: createBookingPattern, updateBookingPattern: updateBookingPattern,
	}
}

func (h *Handler) WithObservability(logger *slog.Logger) *Handler {
	return &Handler{
		logger:                  logger,
		listChildren:            h.listChildren,
		getChild:                h.getChild,
		createChildWithFull:     h.createChildWithFull,
		updateChild:             h.updateChild,
		markInactive:            h.markInactive,
		listAttendance:          h.listAttendance,
		getProfile:              h.getProfile,
		updateProfile:           h.updateProfile,
		getContacts:             h.getContacts,
		replaceContacts:         h.replaceContacts,
		getHealth:               h.getHealth,
		updateHealth:            h.updateHealth,
		getSafeguarding:         h.getSafeguarding,
		updateSafeguarding:      h.updateSafeguarding,
		getConsent:              h.getConsent,
		updateConsent:           h.updateConsent,
		getFunding:              h.getFunding,
		updateFunding:           h.updateFunding,
		getCollectionSetting:    h.getCollectionSetting,
		setCollectionPassword:   h.setCollectionPassword,
		listRoomAssignments:     h.listRoomAssignments,
		createRoomAssignment:    h.createRoomAssignment,
		closeRoomAssignment:     h.closeRoomAssignment,
		getBillingProfile:       h.getBillingProfile,
		updateBillingProfile:    h.updateBillingProfile,
		getLeavingRecord:        h.getLeavingRecord,
		listBookingPatterns:     h.listBookingPatterns,
		getBookingPattern:       h.getBookingPattern,
		getCurrentBookingPattern: h.getCurrentBookingPattern,
		createBookingPattern:    h.createBookingPattern,
		updateBookingPattern:    h.updateBookingPattern,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.GET("/children/attendance", requireRoles("manager", "practitioner"), h.listAttendanceHandler)

	bookingRead := protected.Group("")
	bookingRead.Use(requireRoles("manager", "practitioner"))
	bookingRead.GET("/children/:child_id/booking-patterns", h.listBookingPatternsHandler)
	bookingRead.GET("/children/:child_id/booking-patterns/current", h.getCurrentBookingPatternHandler)
	bookingRead.GET("/children/:child_id/booking-patterns/:pattern_id", h.getBookingPatternHandler)

	manager := protected.Group("")
	manager.Use(requireRoles("manager"))

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

	manager.GET("/children/:child_id/funding", h.getFundingHandler)
	manager.PATCH("/children/:child_id/funding", h.updateFundingHandler)

	manager.GET("/children/:child_id/collection-settings", h.getCollectionSettingHandler)
	manager.PUT("/children/:child_id/collection-settings", h.setCollectionSettingHandler)

	manager.GET("/children/:child_id/room-assignments", h.listRoomAssignmentsHandler)
	manager.POST("/children/:child_id/room-assignments", h.createRoomAssignmentHandler)
	manager.DELETE("/children/:child_id/room-assignments/:assignment_id", h.closeRoomAssignmentHandler)

	manager.GET("/children/:child_id/billing-profile", h.getBillingProfileHandler)
	manager.PATCH("/children/:child_id/billing-profile", h.updateBillingProfileHandler)

	manager.GET("/children/:child_id/leaving-record", h.getLeavingRecordHandler)

	manager.POST("/children/:child_id/booking-patterns", h.createBookingPatternHandler)
	manager.PATCH("/children/:child_id/booking-patterns/:pattern_id", h.updateBookingPatternHandler)
}

func (h *Handler) listChildrenHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	children, err := h.listChildren.Execute(c.Request.Context(), actor, c.Query("status"), parseIntQuery(c, "limit", 50), parseIntQuery(c, "offset", 0))
	if err != nil {
		h.handleError(c, err)
		return
	}

	out := make([]childResponse, 0, len(children))
	for _, child := range children {
		out = append(out, toChildResponse(child))
	}

	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) getChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	child, err := h.getChild.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildResponse(child))
}

func (h *Handler) createChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req createChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	input, err := mapCreateChildRequest(req)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	result, err := h.createChildWithFull.Execute(c.Request.Context(), actor, input)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toChildCreationResponse(result))
}

func (h *Handler) updateChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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

func (h *Handler) markInactiveHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req reasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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

func (h *Handler) listAttendanceHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childHealthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childSafeguardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	p, err := h.updateConsent.Execute(c.Request.Context(), actor, c.Param("child_id"), mapChildConsentRequest(req))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildConsentResponse(p))
}

func (h *Handler) getFundingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	p, err := h.getFunding.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	if p == nil {
		c.JSON(http.StatusOK, gin.H{"funding": nil})
		return
	}
	c.JSON(http.StatusOK, toChildFundingResponse(p))
}

func (h *Handler) updateFundingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childFundingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	p, err := h.updateFunding.Execute(c.Request.Context(), actor, c.Param("child_id"), mapChildFundingRequest(req))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildFundingResponse(p))
}

func (h *Handler) getCollectionSettingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childCollectionSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	p, err := h.setCollectionPassword.Execute(c.Request.Context(), actor, c.Param("child_id"), application.SetCollectionPasswordInput{
		Over18CollectionAcknowledged: req.Over18CollectionAcknowledged,
		Password:                     req.Password,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChildCollectionSettingsResponse(p))
}

func (h *Handler) listRoomAssignmentsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	items, err := h.listRoomAssignments.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	out := make([]roomAssignmentResponse, 0, len(items))
	for _, it := range items {
		out = append(out, toRoomAssignmentResponse(it))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) createRoomAssignmentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req roomAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
	c.JSON(http.StatusCreated, toRoomAssignmentResponse(*a))
}

func (h *Handler) closeRoomAssignmentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req childBillingProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
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

// --- Booking Pattern handlers ---

func (h *Handler) listBookingPatternsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	items, err := h.listBookingPatterns.Execute(c.Request.Context(), actor, c.Param("child_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": toBookingPatternListResponse(items)})
}

func (h *Handler) getBookingPatternHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	p, err := h.getBookingPattern.Execute(c.Request.Context(), actor, c.Param("child_id"), c.Param("pattern_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBookingPatternResponse(*p))
}

func (h *Handler) getCurrentBookingPatternHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	p, err := h.getCurrentBookingPattern.Execute(c.Request.Context(), actor, c.Param("child_id"), c.Query("date"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBookingPatternResponse(*p))
}

func (h *Handler) createBookingPatternHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req bookingPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	effectiveFrom, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	entries := make([]application.BookingPatternEntryInput, 0, len(req.Entries))
	for _, e := range req.Entries {
		stID, perr := uuid.Parse(e.SessionTypeID)
		if perr != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
			return
		}
		entries = append(entries, application.BookingPatternEntryInput{
			DayOfWeek:     e.DayOfWeek,
			SessionTypeID: stID,
		})
	}
	result, err := h.createBookingPattern.Execute(c.Request.Context(), actor, c.Param("child_id"), application.CreateBookingPatternInput{
		EffectiveFrom: effectiveFrom,
		Entries:       entries,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toBookingPatternResponse(*result))
}

func (h *Handler) updateBookingPatternHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}
	var req bookingPatternUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	in := application.UpdateBookingPatternInput{}
	if req.EffectiveFrom != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveFrom)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
			return
		}
		in.EffectiveFrom = &t
	}
	if req.Entries != nil {
		entries := make([]application.BookingPatternEntryInput, 0, len(*req.Entries))
		for _, e := range *req.Entries {
			stID, perr := uuid.Parse(e.SessionTypeID)
			if perr != nil {
				writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
				return
			}
			entries = append(entries, application.BookingPatternEntryInput{
				DayOfWeek:     e.DayOfWeek,
				SessionTypeID: stID,
			})
		}
		in.Entries = &entries
	}
	result, err := h.updateBookingPattern.Execute(c.Request.Context(), actor, c.Param("child_id"), c.Param("pattern_id"), in)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBookingPatternResponse(*result))
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

func parseIntQuery(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	var n int
	for _, r := range v {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

// requireRoles checks that the authenticated user has one of the allowed roles.
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
