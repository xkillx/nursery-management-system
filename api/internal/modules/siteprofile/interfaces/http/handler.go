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
	readOnly.Use(httpserver.RequireRolesWithObservability(h.logger, nil, "manager", "owner", "practitioner"))
	readOnly.GET("/site-profile", h.getSiteProfile)

	writeOps := protected.Group("")
	writeOps.Use(httpserver.RequireRolesWithObservability(h.logger, nil, "manager", "owner"))
	writeOps.PUT("/site-profile", h.updateSiteProfile)
}

// getSiteProfile returns the site profile.
//
//	@Summary		Get site profile
//	@Description	Get the site profile for the current branch.
//	@Tags			site-profile
//	@Produce		json
//	@Success		200	{object}	getSiteProfileResponse
//	@Failure		401	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner","practitioner"]
//	@Router			/site-profile [get]
func (h *Handler) getSiteProfile(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
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

// updateSiteProfile updates the site profile.
//
//	@Summary		Update site profile
//	@Description	Update the site profile for the current branch.
//	@Tags			site-profile
//	@Accept			json
//	@Produce		json
//	@Param			body	body		updateSiteProfileRequest	true	"Profile data"
//	@Success		200		{object}	SiteProfileResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/site-profile [put]
func (h *Handler) updateSiteProfile(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req updateSiteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
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
	httpserver.WriteMappedError(c, h.logger, err)
}
