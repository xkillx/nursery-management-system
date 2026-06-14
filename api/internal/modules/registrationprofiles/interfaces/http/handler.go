package httpregistrationprofile

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/registrationprofiles/application"
	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	getProfile             *application.GetProfile
	updateProfile          *application.UpdateProfile
	setCollectionPassword  *application.SetCollectionPassword
	getConsents            *application.GetConsents
	createConsent          *application.CreateConsent
	getWorkflowStatus      *application.GetWorkflowStatus
	createAttestation      *application.CreateAttestation
}

func NewHandler(
	getProfile *application.GetProfile,
	updateProfile *application.UpdateProfile,
	setCollectionPassword *application.SetCollectionPassword,
	getConsents *application.GetConsents,
	createConsent *application.CreateConsent,
	getWorkflowStatus *application.GetWorkflowStatus,
	createAttestation *application.CreateAttestation,
) *Handler {
	return &Handler{
		getProfile:            getProfile,
		updateProfile:         updateProfile,
		setCollectionPassword: setCollectionPassword,
		getConsents:           getConsents,
		createConsent:         createConsent,
		getWorkflowStatus:     getWorkflowStatus,
		createAttestation:     createAttestation,
	}
}

func (h *Handler) RegisterRoutes(manager *gin.RouterGroup) {
	manager.GET("/children/:child_id/registration-profile", h.getProfileHandler)
	manager.PATCH("/children/:child_id/registration-profile", h.updateProfileHandler)
	manager.PUT("/children/:child_id/registration-profile/collection-password", h.setCollectionPasswordHandler)
	manager.GET("/children/:child_id/registration-consents", h.getConsentsHandler)
	manager.POST("/children/:child_id/registration-consents", h.createConsentHandler)
	manager.GET("/children/:child_id/registration-workflow-status", h.getWorkflowStatusHandler)
	manager.POST("/children/:child_id/registration-completion-attestations", h.createCompletionAttestationHandler)
}

func (h *Handler) getProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID := c.Param("child_id")

	pwc, err := h.getProfile.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		handleError(c, err)
		return
	}

	passwordIsSet := false
	if pwc.Profile != nil && pwc.Profile.CollectionPasswordHash != nil {
		passwordIsSet = true
	}

	completeness := domain.ComputeCompleteness(pwc.Profile, pwc.Contacts, passwordIsSet)

	c.JSON(http.StatusOK, toRegistrationProfileResponse(pwc, completeness))
}

func (h *Handler) updateProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID := c.Param("child_id")

	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	patch, err := parsePatch(raw)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	result, err := h.updateProfile.Execute(c.Request.Context(), actor, childID, patch)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRegistrationProfileResponse(result.ProfileWithChild, result.Completeness))
}

func (h *Handler) setCollectionPasswordHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID := c.Param("child_id")

	var req collectionPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	result, err := h.setCollectionPassword.Execute(c.Request.Context(), actor, childID, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRegistrationProfileResponse(result.ProfileWithChild, result.Completeness))
}

func (h *Handler) getConsentsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID := c.Param("child_id")

	cwc, err := h.getConsents.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		handleError(c, err)
		return
	}

	childSummary, err := h.getProfile.ExecuteGetChildSummary(c.Request.Context(), actor, childID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toConsentsResponse(childSummary, cwc))
}

func (h *Handler) getWorkflowStatusHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID := c.Param("child_id")

	status, err := h.getWorkflowStatus.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toWorkflowStatusResponse(status))
}

func (h *Handler) createCompletionAttestationHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID := c.Param("child_id")

	attestation, err := h.createAttestation.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toAttestationResponse(*attestation))
}

func (h *Handler) createConsentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID := c.Param("child_id")

	var req application.CreateConsentParams
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	record, err := h.createConsent.Execute(c.Request.Context(), actor, childID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toConsentRecordResponse(record))
}

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	c.AbortWithStatusJSON(status, resp)
}

func parsePatch(raw map[string]json.RawMessage) (application.PatchSection, error) {
	var patch application.PatchSection

	if v, ok := raw["demographics_home"]; ok && v != nil {
		var p application.DemographicsHomePatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.DemographicsHome = &p
	}

	if v, ok := raw["medical_dietary"]; ok && v != nil {
		var p application.MedicalDietaryPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.MedicalDietary = &p
	}

	if v, ok := raw["health_contacts"]; ok && v != nil {
		var p application.HealthContactsPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.HealthContacts = &p
	}

	if v, ok := raw["social_development"]; ok && v != nil {
		var p application.SocialDevelopmentPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.SocialDevelopment = &p
	}

	if v, ok := raw["parent_carers"]; ok && v != nil {
		var p []application.ContactEntryPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.ParentCarers = &p
	}

	if v, ok := raw["emergency_contacts"]; ok && v != nil {
		var p []application.ContactEntryPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.EmergencyContacts = &p
	}

	if v, ok := raw["authorised_collectors"]; ok && v != nil {
		var p []application.ContactEntryPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.AuthorisedCollectors = &p
	}

	if v, ok := raw["collection"]; ok && v != nil {
		var p application.CollectionPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.Collection = &p
	}

	if v, ok := raw["funding_support"]; ok && v != nil {
		var p application.FundingSupportPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.FundingSupport = &p
	}

	if v, ok := raw["routine_care"]; ok && v != nil {
		var p application.RoutineCarePatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.RoutineCare = &p
	}

	if v, ok := raw["gdpr_declaration"]; ok && v != nil {
		var p application.GDPRDeclarationPatch
		if err := json.Unmarshal(v, &p); err != nil {
			return patch, err
		}
		patch.GDPRDeclaration = &p
	}

	return patch, nil
}


