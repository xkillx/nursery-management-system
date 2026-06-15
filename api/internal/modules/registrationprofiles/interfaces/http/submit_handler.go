package httpregistrationprofile

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	app "nursery-management-system/api/internal/modules/registrationprofiles/application"
	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type SubmitHandler struct {
	submitUC *app.SubmitCompleteRegistration
	logger   *slog.Logger
}

func NewSubmitHandler(submitUC *app.SubmitCompleteRegistration) *SubmitHandler {
	return &SubmitHandler{submitUC: submitUC}
}

func (h *SubmitHandler) WithObservability(logger *slog.Logger) *SubmitHandler {
	return &SubmitHandler{
		submitUC: h.submitUC,
		logger:   logger,
	}
}

func (h *SubmitHandler) RegisterRoutes(manager *gin.RouterGroup) {
	manager.POST("/children/with-registration", h.handleSubmit)
}

func (h *SubmitHandler) handleSubmit(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req submitCompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	input := mapToDomainInput(req)

	result, err := h.submitUC.Execute(c.Request.Context(), actor, input)
	if err != nil {
		requestID := httpserver.RequestIDFromContext(c)
		status, resp := httpserver.MapDomainError(err, requestID)
		httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
		c.AbortWithStatusJSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, childRecordResponse{
		ID:        result.ChildID.String(),
		FullName:  result.FullName,
		StartDate: result.StartDate,
	})
}

func mapToDomainInput(req submitCompleteRegistrationRequest) domain.CompleteRegistrationInput {
	input := domain.CompleteRegistrationInput{
		Child: domain.ChildRegistrationInfo{
			FullName:    req.Child.FullName,
			DateOfBirth: req.Child.DateOfBirth,
			StartDate:   req.Child.StartDate,
			Notes:       req.Child.Notes,
		},
	}

	if req.RegistrationProfile != nil {
		psi := domain.ProfileSectionsInput{}

		if rp := req.RegistrationProfile.DemographicsHome; rp != nil {
			psi.DemographicsHome = &domain.DemographicsHomeInput{
				Sex:                      rp.Sex,
				Religion:                 rp.Religion,
				EthnicOrigin:             rp.EthnicOrigin,
				FirstLanguage:            rp.FirstLanguage,
				OtherLanguages:           rp.OtherLanguages,
				HomeAddress:              toStringMap(rp.HomeAddress),
				HomePostcode:             rp.HomePostcode,
				HomeTelephone:            rp.HomeTelephone,
				DisabilityStatus:         rp.DisabilityStatus,
				DisabilityNotes:          rp.DisabilityNotes,
				AccessRequirements:       rp.AccessRequirements,
				DemographicsHomeReviewed: rp.DemographicsHomeReviewed,
			}
		}

		if rp := req.RegistrationProfile.MedicalDietary; rp != nil {
			psi.MedicalDietary = &domain.MedicalDietaryInput{
				MedicalConditionsStatus:    rp.MedicalConditionsStatus,
				MedicalConditionsNotes:     rp.MedicalConditionsNotes,
				PrescribedMedicationStatus: rp.PrescribedMedicationStatus,
				MedicationNotes:            rp.MedicationNotes,
				ImmunisationStatus:         rp.ImmunisationStatus,
				ImmunisationCountry:        rp.ImmunisationCountry,
				IllnessDiagnosisHistory:    rp.IllnessDiagnosisHistory,
				DietaryRequirementsStatus:  rp.DietaryRequirementsStatus,
				DietaryRequirementsNotes:   rp.DietaryRequirementsNotes,
				DietarySideEffects:         rp.DietarySideEffects,
				MedicalDietaryReviewed:     rp.MedicalDietaryReviewed,
			}
		}

		if rp := req.RegistrationProfile.HealthContacts; rp != nil {
			psi.HealthContacts = &domain.HealthContactsInput{
				DoctorName:             rp.DoctorName,
				DoctorAddress:          rp.DoctorAddress,
				DoctorPhone:            rp.DoctorPhone,
				HealthVisitorName:      rp.HealthVisitorName,
				HealthVisitorAddress:   rp.HealthVisitorAddress,
				HealthVisitorPhone:     rp.HealthVisitorPhone,
				HealthContactsReviewed: rp.HealthContactsReviewed,
			}
		}

		if rp := req.RegistrationProfile.SocialDevelopment; rp != nil {
			psi.SocialDevelopment = &domain.SocialDevelopmentInput{
				SocialServicesStatus:       rp.SocialServicesStatus,
				SocialServicesNotes:        rp.SocialServicesNotes,
				SocialWorkerName:           rp.SocialWorkerName,
			SocialWorkerPhone:          rp.SocialWorkerPhone,
			SocialWorkerEmail:          rp.SocialWorkerEmail,
				ConcernWalking:             rp.ConcernWalking,
				ConcernSpeechLanguage:      rp.ConcernSpeechLanguage,
				ConcernHearing:             rp.ConcernHearing,
				ConcernSight:               rp.ConcernSight,
				ConcernEmotionalWellbeing:  rp.ConcernEmotionalWellbeing,
				ConcernBehaviour:           rp.ConcernBehaviour,
				SocialDevelopmentReviewed:  rp.SocialDevelopmentReviewed,
			}
		}

		for _, src := range req.RegistrationProfile.ParentCarers {
			psi.ParentCarers = append(psi.ParentCarers, mapContactEntryInput(src))
		}
		for _, src := range req.RegistrationProfile.EmergencyContacts {
			psi.EmergencyContacts = append(psi.EmergencyContacts, mapContactEntryInput(src))
		}
		for _, src := range req.RegistrationProfile.AuthorisedCollectors {
			psi.AuthorisedCollectors = append(psi.AuthorisedCollectors, mapContactEntryInput(src))
		}

		if rp := req.RegistrationProfile.Collection; rp != nil {
			psi.Collection = &domain.CollectionInput{
				Over18CollectionAcknowledged: rp.Over18CollectionAcknowledged,
				EmergencyCollectionReviewed:  rp.EmergencyCollectionReviewed,
			}
		}

		if rp := req.RegistrationProfile.FundingSupport; rp != nil {
			psi.FundingSupport = &domain.FundingSupportInput{
				BenefitsContributeToFees: rp.BenefitsContributeToFees,
				WorkingTaxCredit:         rp.WorkingTaxCredit,
				CollegeUniPaidToParent:   rp.CollegeUniPaidToParent,
				CollegeUniPaidToNursery:  rp.CollegeUniPaidToNursery,
				Funding3yoTermTime:       rp.Funding3yoTermTime,
				Funding2yoTermTime:       rp.Funding2yoTermTime,
				FundingSupportNotes:      rp.FundingSupportNotes,
				FundingSupportReviewed:   rp.FundingSupportReviewed,
			}
		}

		if rp := req.RegistrationProfile.RoutineCare; rp != nil {
			psi.RoutineCare = &domain.RoutineCareInput{
				RoutineCareNotes:    rp.RoutineCareNotes,
				RoutineCareReviewed: rp.RoutineCareReviewed,
			}
		}

		if rp := req.RegistrationProfile.GDPRDeclaration; rp != nil {
			psi.GDPRDeclaration = &domain.GDPRDeclarationInput{
				GDPRDeclaredByName:  rp.GDPRDeclaredByName,
				GDPRDeclarationDate: rp.GDPRDeclarationDate,
			}
		}

		input.Profile = psi
	}

	if req.Consents != nil {
		input.Consents = domain.ConsentInput{
			PaperFormOnFile:  req.Consents.PaperFormOnFile,

			UrgentMedicalTreatment:         req.Consents.UrgentMedicalTreatment,
			UrgentMedicalTreatmentExceptions: req.Consents.UrgentMedicalTreatmentExceptions,
			Plasters:                       req.Consents.Plasters,
			SafeguardingReportingAcknowledgement: req.Consents.SafeguardingReportingAcknowledgement,
			InformationSharingConsent:      req.Consents.InformationSharingConsent,
			GDPRDataProcessingConsent:      req.Consents.GDPRDataProcessingConsent,
			AreaSENCOLiaison:               req.Consents.AreaSENCOLiaison,
			HealthVisitorLiaison:           req.Consents.HealthVisitorLiaison,
			TransitionDocuments:            req.Consents.TransitionDocuments,
			LocalOutings:                   req.Consents.LocalOutings,
			FacePainting:                   req.Consents.FacePainting,
			ParentSuppliedSunCream:         req.Consents.ParentSuppliedSunCream,
			ParentSuppliedNappyCream:       req.Consents.ParentSuppliedNappyCream,
			DevelopmentProfilePhotos:       req.Consents.DevelopmentProfilePhotos,
			NurseryDisplayBoards:           req.Consents.NurseryDisplayBoards,
			PromotionalLiterature:          req.Consents.PromotionalLiterature,
			NurseryWebsite:                 req.Consents.NurseryWebsite,
			StaffStudentCoursework:         req.Consents.StaffStudentCoursework,
			SocialMedia:                    req.Consents.SocialMedia,
			SocialMediaChannelNotes:        req.Consents.SocialMediaChannelNotes,
			NotesExceptions:                req.Consents.NotesExceptions,
		}
	}

	if req.CollectionPassword != nil {
		input.CollectionPassword = *req.CollectionPassword
	}

	return input
}

func mapContactEntryInput(src contactEntryPayload) domain.ContactEntryInput {
	return domain.ContactEntryInput{
		FullName:                 src.FullName,
		RelationshipToChild:      src.RelationshipToChild,
		Address:                  toStringMap(src.Address),
		Telephone:                src.Telephone,
		Email:                    src.Email,
		WorkAddress:              toStringMap(src.WorkAddress),
		HasParentalResponsibility: src.HasParentalResponsibility,
	}
}

func toStringMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}
