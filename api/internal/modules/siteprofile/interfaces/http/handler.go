package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/siteprofile/application"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	logger   *slog.Logger
	getUC    *application.GetSiteProfileUseCase
	updateUC *application.UpdateSiteProfileUseCase
}

func NewHandler(
	getUC *application.GetSiteProfileUseCase,
	updateUC *application.UpdateSiteProfileUseCase,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:   logger,
		getUC:    getUC,
		updateUC: updateUC,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	readOnly := protected.Group("")
	readOnly.Use(requireRoles("manager", "owner", "practitioner"))
	readOnly.GET("/site-profile", h.getSiteProfile)

	writeOps := protected.Group("")
	writeOps.Use(requireRoles("manager", "owner"))
	writeOps.PUT("/site-profile", h.updateSiteProfile)
}

func (h *Handler) getSiteProfile(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	profile, err := h.getUC.Execute(c.Request.Context(), actor)
	if err != nil {
		var de *domainerrors.DomainError
		if errors.As(err, &de) && de.Code == "site_profile_not_found" {
			c.JSON(http.StatusOK, getSiteProfileResponse{
				SiteProfile: nil,
			})
			return
		}
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, getSiteProfileResponse{
		SiteProfile: toSiteProfileResponse(profile),
	})
}

func (h *Handler) updateSiteProfile(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req updateSiteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	input := application.UpdateSiteProfileInput{
		NurseryName:     req.NurseryName,
		Description:     req.Description,
		Phone:           req.Phone,
		Email:           req.Email,
		Website:         req.Website,
		AddressStreet:   req.AddressStreet,
		AddressCity:     req.AddressCity,
		AddressPostcode: req.AddressPostcode,
	}

	saved, err := h.updateUC.Execute(c.Request.Context(), actor, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSiteProfileResponse(saved))
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
