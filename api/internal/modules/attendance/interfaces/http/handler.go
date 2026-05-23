package httpattendance

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/attendance/application"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	checkIn  *application.CheckInChild
	checkOut *application.CheckOutChild
}

func NewHandler(checkIn *application.CheckInChild, checkOut *application.CheckOutChild) *Handler {
	return &Handler{checkIn: checkIn, checkOut: checkOut}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	g := protected.Group("")
	g.Use(requireRoles("manager", "practitioner"))
	g.POST("/attendance/check-ins", h.checkInHandler)
	g.POST("/attendance/check-outs", h.checkOutHandler)
}

func (h *Handler) checkInHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req checkInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	childID, err := parseChildID(req.ChildID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	session, err := h.checkIn.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toSessionResponse(session))
}

func (h *Handler) checkOutHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req checkOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	childID, err := parseChildID(req.ChildID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	session, err := h.checkOut.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionResponse(session))
}

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
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

		if _, exists := allowed[authCtx.Role]; !exists {
			writeError(c, http.StatusForbidden, "forbidden_role", "Access denied.")
			return
		}

		c.Next()
	}
}
